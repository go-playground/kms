##KMS

![Project status](https://img.shields.io/badge/version-1.2.0-green.svg)
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

Doesn't Go 1.8 handle graceful shutdown?
------------
Yes and No, it does not yet handle Hijacked connections, like WebSockets, see:
- https://github.com/golang/go/issues/4674#issuecomment-257549871
- https://github.com/golang/go/issues/17721#issuecomment-257572027

but this package does; furthermore this package isn't just for graceful http shutdown but graceful shutdown of anything and everything.

When Go 1.8 comes out I will transparently make changes to let the std lib handle idle connections as
it has lower level access to them, but continue to handle Hijacked connections.

Installation
-----------

Use go get 

```shell
go get -u github.com/go-playground/kms
```

Built In
--------
There are a few built in graceful shutdown helpers using the kms package for:
- TCP
- Unix Sockets
- HTTP(S) graceful shutdown.

Examples
-------
[see here](https://github.com/go-playground/kms/tree/master/examples) for more

```go
package main

import (
	"net/http"
	"time"

	"github.com/go-playground/kms"
	"github.com/go-playground/kms/kmsnet/kmshttp"
)

func main() {

	// listen for shutdown signal(s) with timeout, non-blocking
	kms.ListenTimeout(false, time.Minute*3)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Home"))
	})

	kmshttp.ListenAndServe(":3007", nil)
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
