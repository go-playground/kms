package kms

import (
	"fmt"
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

// SignalFn is the function type used to signal kms of a shutdown siganl.
// by default this library listens for syscall.SIGINT, syscall.SIGTERM and syscall.SIGHUP
// os.Signal's but you can override with whatever signals or logic you wish.
type SignalFn func() <-chan os.Signal

// Logger is the default instance of the log package
var (
	once         sync.Once
	killMeSoftly *killingMeSoftly

	// note only atomic.Value for tests especially "go test -race"
	notify       atomic.Value // chan struct{}
	done         atomic.Value // chan struct{}
	exitFunc     atomic.Value // os.Exit aka func(int)
	hardShutdown atomic.Value // bool
	sigFn        atomic.Value // SignalFn
)

func init() {
	once.Do(func() {
		killMeSoftly = &killingMeSoftly{
			wg: &sync.WaitGroup{},
		}

		notify.Store(make(chan struct{}))
		done.Store(make(chan struct{}))
		exitFunc.Store(os.Exit)

		AllowSignalHardShutdown(true)

		signalFn := func() <-chan os.Signal {
			s := make(chan os.Signal, 1)
			signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

			go func() {
				<-ShutdownComplete()
				signal.Stop(s)
				close(s)
			}()

			return s
		}

		SetSignalFn(signalFn)
	})
}

// AllowSignalHardShutdown allows you to set whether the application
// should allow hard shutdown of the application if two signals for shutdown
// should cause a hard shutdown. eg. user running application from the command
// line types CTRL + C, the application begins to shut down gracefully, if the
// user types another CTRL + C should the application shut down hard?
//
// Default: true
func AllowSignalHardShutdown(allow bool) {
	hardShutdown.Store(allow)
}

// SetSignalFn allows registering of a custom signal function
// if you wish to listen for different signals, or even your own custom
// logic not related to signals. By default this function listens for
// syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP
func SetSignalFn(fn SignalFn) {
	sigFn.Store(fn)
}

// ShutdownInitiated returns a notification channel for the package which will be
// closed/notified once a termination signal is received.
//
// useful when other code, such as a custom TCP connection listener needs to be
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

// Listen sets up signals to listen for interrupt or kill signals
// in an attempt to wait for all operations to complete before letting
// the process die.
func Listen(block bool) {

	s := sigFn.Load().(SignalFn)()
	done := done.Load().(chan struct{})
	notify := notify.Load().(chan struct{})
	exit := exitFunc.Load().(func(int))

	go func() {

		sig := <-s

		close(notify)

		fmt.Printf("Gracefully stopping (signal: %s)... ", sig)

		if hardShutdown.Load().(bool) {
			// listen for another signal, if another happens.. force shutdown
			go func() {
				<-s
				fmt.Println("done")
				exit(1)
			}()
		}

		killMeSoftly.wg.Wait()
		fmt.Println("done")
		close(done)
	}()

	if block {
		<-done
	}
}

// ListenTimeout sets up signals to listen for interrupt or kill signals
// in an attempt to wait for all operations to complete before letting
// the process die.
//
// the wait duration is how long to wait before forcefully shutting everything down.
func ListenTimeout(block bool, wait time.Duration) {

	s := sigFn.Load().(SignalFn)()
	done := done.Load().(chan struct{})
	notify := notify.Load().(chan struct{})
	exit := exitFunc.Load().(func(int))

	go func() {
		sig := <-s

		close(notify)

		fmt.Printf("Gracefully stopping (signal: %s, timeout: %s)... ", sig, wait)

		if hardShutdown.Load().(bool) {
			go func() {
				select {

				case <-time.After(wait):
					fmt.Println("timed out")
					exit(1)
				case <-s:
					fmt.Println("done")
					exit(1)
				case <-done:
				}

			}()
		} else {
			go func() {
				select {

				case <-time.After(wait):
					fmt.Println("timed out")
					exit(1)
				case <-done:
				}

			}()
		}

		killMeSoftly.wg.Wait()
		fmt.Println("done")
		close(done)
	}()

	if block {
		<-done
	}
}
