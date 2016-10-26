package kmsnet

import (
	"log"
	stdnet "net"
	"os"

	"github.com/go-playground/kms"
)

// NewUnixListener returns an instance of a net.Listener that
// is pre-wired with notification and shutdown siganls.
func NewUnixListener(net, laddr string) (stdnet.Listener, error) {

	unixAddr, err := stdnet.ResolveUnixAddr(net, laddr)
	if err != nil {
		return nil, err
	}

	l, err := stdnet.ListenUnix(net, unixAddr)
	if err != nil {
		return nil, err
	}

	go func() {
		<-kms.ShutdownInitiated()
		if err := l.Close(); err != nil {
			log.Println(err)
		}
	}()

	return &unixListener{l}, nil
}

// NewUnixListenerNoShutdown returns an instance of a net.Listener that
// is pre-wired with kms, but no shutdown signals allowing for a custom
// shutdown to be implimented by the caller.
func NewUnixListenerNoShutdown(net, laddr string) (stdnet.Listener, error) {

	unixAddr, err := stdnet.ResolveUnixAddr(net, laddr)
	if err != nil {
		return nil, err
	}

	l, err := stdnet.ListenUnix(net, unixAddr)
	if err != nil {
		return nil, err
	}

	return &unixListener{l}, nil
}

// NewUnixNoShutdown returns an instance of a net.Listener that
// is pre-wired with kms, but no shutdown signals allowing for a custom
// shutdown to be implimented by the caller.
func NewUnixNoShutdown(l *stdnet.UnixListener) stdnet.Listener {
	return &unixListener{l}
}

type unixListener struct {
	*stdnet.UnixListener
}

var _ stdnet.Listener = new(unixListener)

func (l *unixListener) Accept() (stdnet.Conn, error) {

	conn, err := l.UnixListener.AcceptUnix()
	if err != nil {
		return nil, err
	}

	kms.Wait()

	return zeroUinxConn{Conn: conn}, nil
}

// blocking wait for close
func (l *unixListener) Close() (err error) {

	//stop accepting connections - release fd
	err = l.UnixListener.Close()
	return
}

func (l *unixListener) File() *os.File {

	// returns a dup(2) - FD_CLOEXEC flag *not* set
	tl := l.UnixListener
	fl, _ := tl.File()

	return fl
}

//notifying on close net.Conn
type zeroUinxConn struct {
	stdnet.Conn
}

func (conn zeroUinxConn) Close() (err error) {

	err = conn.Conn.Close()
	kms.Done()
	return
}
