package kms

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// KillingMeSoftly interface contains all functions needed for operation and method chaining
type KillingMeSoftly interface {
	Done()
}

type killingMeSoftly struct {
	wg *sync.WaitGroup
}

// Logger is the default instance of the log package
var (
	once         sync.Once
	killMeSoftly *killingMeSoftly
	// notify       chan struct{}
	// done         chan struct{}
	notify   atomic.Value // chan struct{}
	done     atomic.Value // chan struct{}
	exitFunc atomic.Value // os.Exit
)

func init() {
	once.Do(func() {
		killMeSoftly = &killingMeSoftly{
			wg: &sync.WaitGroup{},
		}

		notify.Store(make(chan struct{}))
		done.Store(make(chan struct{}))
		exitFunc.Store(os.Exit)
		// notify = make(chan struct{})
		// done = make(chan struct{})
	})
}

// ShutdownInitiated returns a notification channel for the package which will be
// closed/notified once a termination signal is recieved.
//
// usefull when other code, such as a custom TCP connection listener needs to be
// notified to stop listening for new connections.
func ShutdownInitiated() <-chan struct{} {
	return notify.Load().(chan struct{})
}

// ShutdownComplete returns a notification channel for the package which will be
// closed/notified once termination is imminent.
func ShutdownComplete() <-chan struct{} {
	return done.Load().(chan struct{})
}

// Wait signifies that your application is busy performing an operation.
//
// best to chain using defer kms.Wait().Done()
func Wait() KillingMeSoftly {
	killMeSoftly.wg.Add(1)
	return killMeSoftly
}

// Done signifies that your application is done performing an operation. it is different from
// the chained version as it does not need to be connected the the wait object.
func Done() {
	killMeSoftly.wg.Done()
}

// Done signifies that your application is done performing an operation.
//
// best to chain using defer kms.Wait().Done()
func (k *killingMeSoftly) Done() {
	k.wg.Done()
}

// Listen sets up signals to listen for interupt or kill signals
// in an attempt to wait for all operations to complete before letting
// the process die.
func Listen(block bool) {

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	done := done.Load().(chan struct{})
	notify := notify.Load().(chan struct{})
	exit := exitFunc.Load().(func(int))

	go func() {

		defer close(s)
		defer signal.Stop(s)

		sig := <-s

		close(notify)

		log.Printf("signal %s recieved, attempting soft shutdown...\n", sig)

		// listen for another signal, if another happens.. force shutdown
		go func() {
			sig := <-s
			fmt.Printf("recieved additional %s, hard shutdown initiated\n", sig)
			exit(1)
		}()

		killMeSoftly.wg.Wait()
		log.Println("soft shutdown complete, ending process")
		close(done)
	}()

	if block {
		<-done
	}
}

// ListenTimeout sets up signals to listen for interupt or kill signals
// in an attempt to wait for all operations to complete before letting
// the process die.
//
// the wait duration is how long to wait before forcefully shutting everything down.
func ListenTimeout(block bool, wait time.Duration) {

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	done := done.Load().(chan struct{})
	notify := notify.Load().(chan struct{})
	exit := exitFunc.Load().(func(int))

	go func() {
		defer close(s)
		defer signal.Stop(s)
		sig := <-s

		close(notify)

		log.Printf("signal %s recieved, attempting soft shutdown for %s...\n", sig, wait)

		go func() {
			select {

			case <-time.After(wait):
				fmt.Println("timeout reached, hard shutdown initiated")
				exit(1)
			case sig := <-s:
				fmt.Printf("recieved additional %s, hard shutdown initiated\n", sig)
				exit(1)
			case <-done:
			}

		}()

		killMeSoftly.wg.Wait()
		log.Println("soft shutdown complete, ending process")
		close(done)
	}()

	if block {
		<-done
	}
}
