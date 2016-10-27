package kmsnet

import (
	"log"
	stdnet "net"
	"os"
	"time"

	"github.com/go-playground/kms"
)

// NewTCPListener returns an instance of a net.Listener that
// is pre-wired with notification and shutdown siganls.
func NewTCPListener(net, laddr string) (stdnet.Listener, error) {

	tcpAddr, err := stdnet.ResolveTCPAddr(net, laddr)
	if err != nil {
		return nil, err
	}

	l, err := stdnet.ListenTCP(net, tcpAddr)
	if err != nil {
		return nil, err
	}

	go func() {
		<-kms.ShutdownInitiated()
		if err := l.Close(); err != nil {
			log.Println(err)
		}
	}()

	return &tcpListener{l}, nil
}

// NewTCPListenerNoShutdown returns an instance of a net.Listener that
// is pre-wired with kms, but no shutdown signals allowing for a custom
// shutdown to be implimented by the caller.
func NewTCPListenerNoShutdown(net, laddr string) (stdnet.Listener, error) {

	tcpAddr, err := stdnet.ResolveTCPAddr(net, laddr)
	if err != nil {
		return nil, err
	}

	l, err := stdnet.ListenTCP(net, tcpAddr)
	if err != nil {
		return nil, err
	}

	return &tcpListener{l}, nil
}

// NewTCPNoShutdown returns an instance of a net.Listener that
// is pre-wired with kms, but no shutdown signals allowing for a custom
// shutdown to be implimented by the caller.
func NewTCPNoShutdown(l *stdnet.TCPListener) stdnet.Listener {
	return &tcpListener{l}
}

type tcpListener struct {
	*stdnet.TCPListener
}

var _ stdnet.Listener = new(tcpListener)

func (l *tcpListener) Accept() (stdnet.Conn, error) {

	conn, err := l.TCPListener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	conn.SetKeepAlive(true)                  // see http.tcpKeepAliveListener
	conn.SetKeepAlivePeriod(time.Minute * 3) // see http.tcpKeepAliveListener
	// conn.SetLinger(0) // is the default already accoring to the docs https://golang.org/pkg/net/#TCPConn.SetLinger

	kms.Wait()

	return &zeroTCPConn{TCPConn: conn}, nil
}

// blocking wait for close
func (l *tcpListener) Close() (err error) {

	//stop accepting connections - release fd
	err = l.TCPListener.Close()
	return
}

func (l *tcpListener) File() *os.File {

	// returns a dup(2) - FD_CLOEXEC flag *not* set
	tl := l.TCPListener
	fl, _ := tl.File()

	return fl
}

//notifying on close net.Conn
type zeroTCPConn struct {
	*stdnet.TCPConn
}

func (conn zeroTCPConn) Close() (err error) {
	if err = conn.TCPConn.Close(); err == nil {
		kms.Done()
	}
	return
}
