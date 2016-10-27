package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/kms"
	"github.com/go-playground/kms/kmsnet/kmshttp"
)

var (
	// simple variable to prove the CRON job finishes gracefully.
	complete bool
)

func main() {

	go fakeWebsocketHandler()
	go cronJob()

	kms.Listen(false)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Home"))
	})

	kmshttp.ListenAndServe(":3007", nil)

	fmt.Println("CRON completed gracefully? ", complete)
}

// crude CRON job examples
func cronJob() {

	// a long running job that fires every x minutes
	for {
		time.Sleep(time.Second * 1)
		kms.Wait()
		complete = false

		// long running DB calls
		time.Sleep(time.Second * 10)

		complete = true
		kms.Done()
	}
}

// faking a single WebSocket, just to show how
func fakeWebsocketHandler() {

	message := make(chan []byte)

FOR:
	for {
		select {
		case <-kms.ShutdownInitiated():
			close(message)
			// close WebSocket connection here
			break FOR
		case b := <-message:
			fmt.Println(string(b))
		}
	}

}
