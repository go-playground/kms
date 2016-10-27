package kmshttp

import (
	"crypto/tls"
	"net"
	"net/http"

	"github.com/go-playground/kms"
	"github.com/go-playground/kms/kmsnet"
)

const (
	http2NextProtoTLS = "h2"
	http2Rev14        = "h2-14"
	http11            = "http/1.1"
)

// ListenAndServe listens on the TCP network address addr and then calls Serve with handler to handle requests
// on incoming connections. Accepted connections are configured to enable TCP keep-alives. Handler is typically
// nil, in which case the DefaultServeMux is used.
func ListenAndServe(addr string, handler http.Handler) (err error) {

	if handler == nil {
		handler = http.DefaultServeMux
	}

	l, err := kmsnet.NewTCPListenerNoShutdown("tcp", addr)
	if err != nil {
		return err
	}

	s := &http.Server{Addr: l.Addr().String(), Handler: handler}

	server := &serverConnState{
		Server:    s,
		l:         l,
		idleConns: make(map[net.Conn]struct{}),
		active:    make(chan net.Conn),
		idle:      make(chan net.Conn),
		closed:    make(chan net.Conn),
		shutdown:  make(chan struct{}),
	}

	server.handleConnState()

	err = server.Serve(server.l)

	// wait for process shutdown to complete or timeout.
	<-kms.ShutdownComplete()

	return err
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it expects HTTPS connections. Additionally,
// files containing a certificate and matching private key for the server must be provided. If the certificate is signed
// by a certificate authority, the certFile should be the concatenation of the server's certificate, any intermediates,
// and the CA's certificate.
func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) (err error) {

	tlsConfig := &tls.Config{
		NextProtos:   []string{http2NextProtoTLS, http2Rev14, http11},
		Certificates: make([]tls.Certificate, 1),
	}

	tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return
	}

	if handler == nil {
		handler = http.DefaultServeMux
	}

	l, err := kmsnet.NewTCPListenerNoShutdown("tcp", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(l, tlsConfig)

	s := &http.Server{Addr: tlsListener.Addr().String(), Handler: handler, TLSConfig: tlsConfig}

	server := &serverConnState{
		Server:    s,
		l:         tlsListener,
		idleConns: make(map[net.Conn]struct{}),
		active:    make(chan net.Conn),
		idle:      make(chan net.Conn),
		closed:    make(chan net.Conn),
		shutdown:  make(chan struct{}),
	}

	server.handleConnState()

	err = server.Serve(server.l)

	// wait for process shutdown to complete or timeout.
	<-kms.ShutdownComplete()

	return err
}

// Serve accepts incoming HTTP connections on the listener l,
// creating a new service goroutine for each. The service goroutines
// read requests and then call handler to reply to them.
// Handler is typically nil, in which case the DefaultServeMux is used.
//
// currently only net.TCPListener and net.UnixListener is supported
func Serve(l net.Listener, handler http.Handler) (err error) {

	if handler == nil {
		handler = http.DefaultServeMux
	}

	s := &http.Server{Handler: handler}

	var lis net.Listener

	switch l.(type) {
	case *net.TCPListener:
		lis = kmsnet.NewTCPNoShutdown(l.(*net.TCPListener))
	case *net.UnixListener:
		lis = kmsnet.NewUnixNoShutdown(l.(*net.UnixListener))
	default:
		panic("unsupported listener type")
	}

	server := &serverConnState{
		Server:    s,
		l:         lis,
		idleConns: make(map[net.Conn]struct{}),
		active:    make(chan net.Conn),
		idle:      make(chan net.Conn),
		closed:    make(chan net.Conn),
		shutdown:  make(chan struct{}),
	}

	server.handleConnState()

	err = server.Serve(server.l)

	// wait for process shutdown to complete or timeout.
	<-kms.ShutdownComplete()

	return err
}

// RunServer wraps an runs the given http.Server instance
func RunServer(s *http.Server) (err error) {

	l, err := kmsnet.NewTCPListenerNoShutdown("tcp", s.Addr)
	if err != nil {
		return err
	}

	if s.TLSConfig != nil {

		// if not configured
		if len(s.TLSConfig.NextProtos) == 0 {
			s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, http2NextProtoTLS, http2Rev14, http11)
		}

		l = tls.NewListener(l, s.TLSConfig)
	}

	server := &serverConnState{
		Server:    s,
		l:         l,
		idleConns: make(map[net.Conn]struct{}),
		active:    make(chan net.Conn),
		idle:      make(chan net.Conn),
		closed:    make(chan net.Conn),
		shutdown:  make(chan struct{}),
	}

	server.handleConnState()

	err = server.Serve(server.l)

	// wait for process shutdown to complete or timeout.
	<-kms.ShutdownComplete()

	return err
}

type serverConnState struct {
	*http.Server
	l         net.Listener
	idleConns map[net.Conn]struct{}
	active    chan net.Conn
	idle      chan net.Conn
	closed    chan net.Conn
	shutdown  chan struct{}
}

func (s *serverConnState) handleConnState() {

	// we do not listen for hijacked, they are a lost cause at this level
	// as we don't know how they are being used, however, you the user can use
	// the kms package to Wait() and Notify() and Done() within your implementation;
	// this is the power of the kms package, being able to tie multiple disparate
	//  things together
	s.ConnState = func(conn net.Conn, state http.ConnState) {

		switch state {
		case http.StateActive:
			s.active <- conn
		case http.StateIdle:
			s.idle <- conn

		// it's up to the implementation to cleanup/close Hijacked connections
		// as we have no details about nor any control over the implementation
		// however this should be pretty trivial using kms.ShutdownInitiated()
		// and a select statement to close the hijacked connections eg. WebSockets
		//
		// case http.StateHijacked:
		// 	fmt.Println("Hijacked")

		case http.StateClosed:
			s.closed <- conn
		}
	}

	go func() {

		var conn net.Conn
		var shuttingDown bool

		defer close(s.shutdown)

		for {
			select {
			case conn = <-s.active:
				if shuttingDown {
					conn.Close()
				}

				delete(s.idleConns, conn)

			case conn = <-s.idle:

				if shuttingDown {
					conn.Close()
					return
				}

				s.idleConns[conn] = struct{}{}

			case conn = <-s.closed:

				delete(s.idleConns, conn)

			case <-s.shutdown:
				s.l.Close()
				shuttingDown = true

				// NOTE: possible race condition if an idle connection
				// transitions from StateIdle to StateActive but there is
				// no way to make it 100% race proof and this is the least
				// of all evils + StateIdle should handle closing the connection
				// after shuttingDown variable is set to true.
				for c := range s.idleConns {
					c.Close()
				}
			}
		}
	}()

	go func() {
		<-kms.ShutdownInitiated()
		s.shutdown <- struct{}{}
	}()
}
