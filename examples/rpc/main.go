package main

import (
	"fmt"
	"log"
	"net/rpc"
	"syscall"
	"time"

	"github.com/go-playground/kms"
	"github.com/go-playground/kms/kmsnet"
)

// RPCListener ...
type RPCListener struct {
}

// Hello ...
func (l *RPCListener) Hello(in *string, out *string) (err error) {
	res := fmt.Sprintf("Hello %s", *in)
	*out = *(&res)
	return nil
}

func main() {

	go testAndKill()

	inbound, err := kmsnet.NewTCPListener("tcp", ":4444")
	if err != nil {
		log.Fatal(err)
	}
	defer inbound.Close()

	// for TLS just wrap 'inbound'
	// newInbound := tls.NewListener(inbound, tlsConfig)
	// defer newInbound.Close()
	//
	// ...
	// s.Accept(newInbound)
	// ...

	s := rpc.NewServer()

	err = s.Register(new(RPCListener))
	if err != nil {
		log.Fatal(err)
	}

	// listen for shutdown signal(s), non-blocking with timeout
	kms.ListenTimeout(false, time.Minute*3)

FOR:
	for {
		select {
		case <-kms.ShutdownInitiated():
			break FOR
		default:
			s.Accept(inbound)
		}
	}

	<-kms.ShutdownComplete()
}

func testAndKill() {

	time.Sleep(time.Second * 1)

	var in, out string

	in = "Joeybloggs"

	client, err := rpc.Dial("tcp", "localhost:4444")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	err = client.Call("RPCListener.Hello", &in, &out)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server returned '%s'\n", out)

	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
}
