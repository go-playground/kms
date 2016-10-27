package kms

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

// NOTES:
// - Run "go test" to run tests
// - Run "gocov test | gocov report" to report on test converage by file
// - Run "gocov test | gocov annotate -" to report on all code and functions, those ,marked with "MISS" were never called
//
// or
//
// -- may be a good idea to change to output path to somewherelike /tmp
// go test -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
//
//
// because 'kms' is a singleton testing required some variables to be reinitialized
//

func reinitialize() {
	notify.Store(make(chan struct{}))
	done.Store(make(chan struct{}))

	AllowSignalHardShutdown(true)

	exitFunc.Store(func(code int) {
		fmt.Println("Exiting")
	})
}

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

func TestListen(t *testing.T) {

	reinitialize()

	m := sync.Mutex{}

	var stopping, stopped bool

	Wait()

	go func() {
		<-ShutdownInitiated()
		m.Lock()
		defer m.Unlock()
		stopping = true
		<-ShutdownComplete()
		stopped = true
	}()

	go func() {
		<-time.After(time.Second * 1)
		Done()
		syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	}()

	Listen(true)

	m.Lock()
	defer m.Unlock()

	if !stopping {
		t.Errorf("Expected '%t' Got '%t'", true, stopping)
	}

	if !stopped {
		t.Errorf("Expected '%t' Got '%t'", true, stopped)
	}
}

func TestListen2(t *testing.T) {

	reinitialize()

	m := sync.Mutex{}

	var stopping, stopped bool

	go func() {
		<-ShutdownInitiated()
		m.Lock()
		defer m.Unlock()
		stopping = true
		<-ShutdownComplete()
		stopped = true
	}()

	go func() {
		defer Wait().Done()
		<-time.After(time.Second * 1)
		syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	}()

	Listen(true)

	m.Lock()
	defer m.Unlock()

	if !stopping {
		t.Errorf("Expected '%t' Got '%t'", true, stopping)
	}

	if !stopped {
		t.Errorf("Expected '%t' Got '%t'", true, stopped)
	}
}

func TestListenTimeout(t *testing.T) {

	reinitialize()

	m := sync.Mutex{}

	var stopping, stopped bool

	go func() {
		<-ShutdownInitiated()
		m.Lock()
		defer m.Unlock()
		stopping = true
		<-ShutdownComplete()
		stopped = true
	}()

	go func() {
		defer Wait().Done()
		<-time.After(time.Second * 1)
		syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	}()

	ListenTimeout(true, time.Second*2)

	m.Lock()
	defer m.Unlock()

	if !stopping {
		t.Errorf("Expected '%t' Got '%t'", true, stopping)
	}

	if !stopped {
		t.Errorf("Expected '%t' Got '%t'", true, stopped)
	}
}

func TestListenTimeoutTimeout(t *testing.T) {

	reinitialize()
	AllowSignalHardShutdown(false)

	exitFunc.Store(func(code int) {
		fmt.Println("Exiting OK")
		close(done.Load().(chan struct{}))
	})

	m := sync.Mutex{}

	var stopping, stopped bool

	go func() {
		<-ShutdownInitiated()
		m.Lock()
		defer m.Unlock()
		stopping = true
		<-ShutdownComplete()
		stopped = true
	}()

	Wait()

	go func() {
		<-time.After(time.Second * 1)
		syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	}()

	ListenTimeout(true, time.Millisecond*200)

	m.Lock()
	defer m.Unlock()

	if !stopping {
		t.Errorf("Expected '%t' Got '%t'", true, stopping)
	}

	if !stopped {
		t.Errorf("Expected '%t' Got '%t'", true, stopped)
	}
}

func TestListenTimeoutDone(t *testing.T) {

	reinitialize()

	exitFunc.Store(func(code int) {
		fmt.Println("Exiting OK")
		close(done.Load().(chan struct{}))
	})

	m := sync.Mutex{}

	var stopping, stopped bool

	go func() {
		<-ShutdownInitiated()
		m.Lock()
		defer m.Unlock()
		stopping = true
		<-ShutdownComplete()
		stopped = true
	}()

	go func() {
		<-time.After(time.Second * 1)
		syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	}()

	ListenTimeout(true, time.Second*10)

	m.Lock()
	defer m.Unlock()

	if !stopping {
		t.Errorf("Expected '%t' Got '%t'", true, stopping)
	}

	if !stopped {
		t.Errorf("Expected '%t' Got '%t'", true, stopped)
	}
}

func TestListenTimeoutDoneNoHardShutdown(t *testing.T) {

	reinitialize()
	AllowSignalHardShutdown(false)

	exitFunc.Store(func(code int) {
		fmt.Println("Exiting OK")
		close(done.Load().(chan struct{}))
	})

	m := sync.Mutex{}

	var stopping, stopped bool

	go func() {
		<-ShutdownInitiated()
		m.Lock()
		defer m.Unlock()
		stopping = true
		<-ShutdownComplete()
		stopped = true
	}()

	go func() {
		<-time.After(time.Second * 1)
		syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	}()

	ListenTimeout(true, time.Second*10)

	m.Lock()
	defer m.Unlock()

	if !stopping {
		t.Errorf("Expected '%t' Got '%t'", true, stopping)
	}

	if !stopped {
		t.Errorf("Expected '%t' Got '%t'", true, stopped)
	}
}

func TestListenTimeoutDoubleKill(t *testing.T) {

	reinitialize()

	exitFunc.Store(func(code int) {
		fmt.Println("Exiting OK")
		close(done.Load().(chan struct{}))
	})

	m := sync.Mutex{}

	var stopping, stopped bool

	go func() {
		<-ShutdownInitiated()
		m.Lock()
		defer m.Unlock()
		stopping = true
		<-ShutdownComplete()
		stopped = true
	}()

	Wait()

	go func() {
		<-time.After(time.Second * 1)
		syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	ListenTimeout(true, time.Second*10)

	m.Lock()
	defer m.Unlock()

	if !stopping {
		t.Errorf("Expected '%t' Got '%t'", true, stopping)
	}

	// no guarantee stopped will or will not get set, it is a hard shutdown
	// if !stopped {
	// 	t.Errorf("Expected '%t' Got '%t'", true, stopped)
	// }
}
