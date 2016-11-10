##KMS

![Project status](https://img.shields.io/badge/version-1.1.0-green.svg)
[![Build Status](https://semaphoreci.com/api/v1/joeybloggs/kms/branches/master/badge.svg)](https://semaphoreci.com/joeybloggs/kms)
[![Coverage Status](https://coveralls.io/repos/github/go-playground/kms/badge.svg)](https://coveralls.io/github/go-playground/kms)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-playground/kms)](https://goreportcard.com/report/github.com/go-playground/kms)
[![GoDoc](https://godoc.org/github.com/go-playground/kms?status.svg)](https://godoc.org/github.com/go-playground/kms)
![License](https://img.shields.io/dub/l/vibe-d.svg)

KMS(Killing Me Softly) is a library that aids in graceful shutdown of a process/application.

Why does this even exist? Aren't there other graceful shutdown librarys?
-------------
Sure there are other libraries, but they are focused on handling graceful shutdown a specific portion of
code such as graceful shutdown of an http server, but this library allows you to glue together multiple
unrelated(or related) services running within the same process/application.

eg.

```go

// start CRON service running tasks..
go startCRON()

// listen for shutdown signal(s), non-blocking
kms.Listen(false)

// serve http site ( using built in http listener)
// this site also has WebSocket functionality
kmshttp.ListenAndServe(":3007", nil)

```

This library will allow you to gracefully shutdown the http server, any WebSockets and any CRON jobs that may be running using the following:

```go

kms.Wait()
kms.Done()

// chaining is also supported eg. defer kms.Wait().Done()

<-kms.ShutdownInitiated()

and is some cases

<-kms.ShutdownComplete()

```

Installation
-----------

Use go get 

```shell
go get -u github.com/go-playground/kms
```

Built In
--------
There are a few built in graceful shutdown helpers using the kms package for TCP, Unix Sockets and HTTP graceful shutdown.

Example
-------
```go
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
```

Package Versioning
----------
I'm jumping on the vendoring bandwagon, you should vendor this package as I will not
be creating different version with gopkg.in like allot of my other libraries.

Why? because my time is spread pretty thin maintaining all of the libraries I have + LIFE,
it is so freeing not to worry about it and will help me keep pouring out bigger and better
things for you the community.

I am open versioning with gopkg.in should anyone request it, but this should be stable going forward.

Licenses
--------
- [MIT License](https://raw.githubusercontent.com/go-playground/kms/master/LICENSE) (MIT), Copyright (c) 2016 Dean Karn
