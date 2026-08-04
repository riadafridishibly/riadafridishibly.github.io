---
published: true
comments: true
layout: post
title: Signal Handling and Graceful Shutdown in Go
description: "Catch SIGINT and SIGTERM in Go, broadcast the shutdown to every goroutine, unblock blocking calls and shut down an HTTP server cleanly."
categories: [programming]
tags: [golang, signals, graceful-shutdown, concurrency]
image: /assets/img/2023-03-11-go-graceful-shutdown/broken_computer.jpg
image_alt: "Graceful shutdown in Go"
---



## Table Of Contents
{:.no_toc}
- 
{:toc}

**All the code samples used in this post are available** [_here._](https://github.com/riadafridishibly/go-graceful-shutdown)

## Graceful Shutdown

When a process is running for a long time, sometimes we want to quit the running program. If it's a CLI process we press `CTRL+C` or send specific kill signals. For GUI applications we quit from the menu or if the process becomes nonresponsive we find the `PID` of the process and then run `kill -9 <PID>`. These actions trigger an event and the event is sent to the specific process which sometimes causes the process to exit. In a stateful program, we want to save the states or perform cleanups before exiting the process. This safe exit process is called graceful shutdown.

In short, to perform a graceful shutdown we need to catch the kill signal then perform the required cleanup and then exit.

## What are signals in OS context?

A _signal_ is a software interrupt delivered to a process. Here we'll be dealing with signals which cause the process to die. To understand better we'll play with lots of example code here in this post.

## An Example Program

To understand better, let's start with a very simple hello world program. We'll print the `PID` of our process first, then we'll wait for the signal to arrive.

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("PID:", os.Getpid())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Wait for signal
	got := <-sigCh

	fmt.Printf("Received Signal: %s, Sig Num: %d\n", got, got)
}
```

Run this program and then press `CTRL+C`. The process will exit and print something like this,

```
PID: 123768
^CReceived Signal: interrupt, Sig Num: 2
```

Now run the program again, open a new terminal window and run this command.

```sh
$ kill -SIGTERM <PID>
```

The `PID` value is printed. Let's grab the `PID` from there.

For example,

```sh
$ kill -SIGTERM 123768
```

Now the output should look like this,

```
PID: 129465
Received Signal: terminated, Sig Num: 15
```

Let's run the program again and kill with `-SIGKILL`. The output should look like this,

```
PID: 131937
signal: killed
```

> We have used the named version of the signal, we can use number as well, for example, `kill -9 <pid>` is similar to running `kill -SIGKILL <pid>`.

Okay, let's recap what's happening here. First of all, we grabbed the `PID` of the running program and print that. Then with these two lines, we've created a buffered channel of size 1 and registered two signals `SIGINT` and `SIGTERM`.

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
```

This way if any of these signals is sent to the process, `signal.Notify` will send that signal to the `sigCh` channel.

Next, we wait for the signals to arrive in `sigCh` channel.

```go
// Wait for signal
got := <-sigCh
```

Normally our app will block elsewhere, for example, a web server will block the `main` function. But in this case, we're simply waiting on `sigCh` for any signal.

## An automated way to send signals

Sending a signal using the `kill` command is okay, but we can automate this for this post. Let's write a simple wrapper that will send a specific signal after certain seconds.

```go
func SimulateSendSignal(after time.Duration, sig os.Signal) {
	go func() {
		pid := os.Getpid()
		p, err := os.FindProcess(pid)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(after)
		fmt.Printf("==== Sending signal %q to PID(%d)\n", sig, pid)
		if err := p.Signal(sig); err != nil {
			log.Fatal(err)
		}
	}()
}
```

We'll see the function in action in future examples.

## What are the available signals?

We've already seen three signals, `SIGINT`, `SIGTERM` and `SIGKILL`. Let's investigate why each one is different.

> **One important thing about signals is, not all signals are catchable. If we try to catch `SIGKILL` or `SIGSTOP` we won't be able to do so. Kernel can catch it, but userspace program can not.**

Why are they different? Well, we can catch different signals and handle them differently.

Let's quickly review three of them and their meaning. Their default behavior is to kill the process.

### SIGINT

The SIGINT (“program interrupt”) signal is sent when the user types the INTR character (normally C-c). 

### SIGTERM

The SIGTERM signal is a generic signal used to cause program termination. Unlike SIGKILL, this signal can be blocked, handled, and ignored. **It is the normal way to politely ask a program to terminate.**

The shell command `kill` generates SIGTERM by default.

### SIGKILL

The SIGKILL signal is used to cause immediate program termination. It cannot be handled or ignored, and is therefore always fatal. It is also **not** possible to block this signal.

This signal is usually generated only by explicit request. Since it cannot be handled, you should generate it only as a last resort, after first trying a less drastic method such as C-c or SIGTERM. If a process does not respond to any other termination signals, sending it a SIGKILL signal will almost always cause it to go away.

If SIGKILL fails to terminate a process, that by itself constitutes an operating system bug.

The system will generate SIGKILL for a process itself under some unusual conditions where the program cannot possibly continue to run (even to run a signal handler).

[Here's a list of different signals and their meaning](https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html).


## Why did we initialize the **sigCh** as buffered channel?

From the [`signal.Notify`](https://pkg.go.dev/os/signal#Notify) docs, 

> Package signal will not block sending to c: the caller must ensure that c has sufficient buffer space to keep up with the expected signal rate. For a channel used for notification of just one signal value, a buffer of size 1 is sufficient.

So if we don't provide a buffer, `signal.Notify` won't wait for sending the signal to the channel. Sending to an unbuffered channel will be successful when there's another goroutine waiting for receiving from that channel. Otherwise, sending operation will block. Let's demonstrate that with another simple code.

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riadafridishibly/go-graceful-shutdown/utils"
)

func main() {
	fmt.Println("PID:", os.Getpid())
	sigCh := make(chan os.Signal, 1) // Change this to unbuffered, make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	utils.SimulateSendSignal(1*time.Second, os.Interrupt)

	fmt.Println("Sleep started. Waiting for 5 sec.")
	time.Sleep(5 * time.Second)
	fmt.Println("Sleep done...")

	got := <-sigCh
	fmt.Printf("Received Signal: %s, Sig Num: %d\n", got, got)
}
```

This program won't exit automatically if we don't provide a buffered channel. With an unbuffered channel, no signal will be registered during the sleep state of the program. But with the buffered channel signal will be successfully sent to the channel, and received from the `sigCh` channel as well.

> If you use `go-staticcheck` it'll warn you like this, 
>
> `the channel used with signal.Notify should be buffered (SA1017)`

This has nothing to do with signals though, it's the specific behavior of go channels.

## Signal Broadcast

We've seen we can capture the signal. But how do we propagate the signal throughout our app? 

Before exploring this area, let's quickly review the channel behaviors.

- Sending to or receiving from nil channel will block.
- Sending to a closed channel will panic.
- Receiving from a closed channel returns immediately, and can be used multiple times.

Let's see a few different cases where we can implement signal broadcasts.

### When we are already dealing with channels

If we have something like this, where we're just sending or receiving data from a channel we can easily implement closing the loop.

```go
func splitString(s string) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, v := range strings.Fields(s) {
			ch <- v
		}
	}()
	return ch
}
```

Let's handle the `done` channel in the next example,

```go
func splitStringDone(s string, done <-chan bool) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, v := range strings.Fields(s) {
			select {
			case ch <- v:
			case <-done:
				return
			}
		}
	}()
	return ch
}
```

Here we're taking a `done` channel. When done is closed, we'll receive from `<-done` immediately and return.

This way we can handle the closing signal. Here's the full example.

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/riadafridishibly/go-graceful-shutdown/utils"
)

func splitStringDone(s string, done <-chan bool) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, v := range strings.Fields(s) {
			select {
			case ch <- v:
				// Just for blocking for 1 sec
				select {
				case <-time.After(1 * time.Second):
				case <-done:
					return
				}
			case <-done:
				return
			}
		}
	}()
	return ch
}

func printer(name string, ch <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for v := range ch {
		fmt.Printf("%s: value = %v\n", name, v)
	}
}

func main() {
	fmt.Println("PID:", os.Getpid())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Comment out this line and run the program again
	utils.SimulateSendSignal(2*time.Second, os.Interrupt)

	done := make(chan bool)
	go func() {
		got := <-sigCh
		fmt.Printf("Received Signal: %s, Sig Num: %d\n", got, got)

		// Close the done channel to signal the `splitStringDone` function that
		// we are no longer interested, we're quiting.
		close(done)
	}()

	ch := splitStringDone("a b c d e f g", done)

	var wg sync.WaitGroup

	wg.Add(2)
	go printer("Printer 1", ch, &wg)
	go printer("Printer 2", ch, &wg)

	wg.Wait()

	fmt.Println("Exited!")

	// Print the goroutine stack trace,
	// to check which goroutines are currently alive
	// debug.SetTraceback("all")
	// panic("show me the stacks")
}
```

Here, we're handling the signal in a goroutine. So either our loop ends or we initiate cancellation with a signal. When we catch any signal we simply close the `done` channel. And in the select block `<-done` is selected and we return.

### Dealing with blocking functions

Sometimes we may have a blocking function. With a blocking function, we can't simply use select, if we do we'll just block the case (that's why we didn't put `time.Sleep(1 * time.Second)` in the previous example. we've used another select.).

When we are in a blocking state, the select switch won't help us. In the next example, we'll see the problem in action. First, let's simulate the blocking state with this function,

```go
func BlockingFunc() (string, error) {
	n := 5 * time.Second
	fmt.Printf("Blocking func started, will sleep for %v\n", n)
	defer fmt.Println("Blocking func finished")

	time.Sleep(n)
	return "foo bar baz", nil
}
```

This function prints its status at the start, then it sleeps for 10 seconds and returns a string and an error. Finally, it prints its status again that the function has exited.

If we call this function directly we'll block our program for 5 seconds. In the meantime, the signal catcher won't help us. To demonstrate the problem let's run the following program. Our signal won't exit the program, rather it'll hang for 5 seconds and then the program will exit. The problem is in the select block. Because as soon as we start executing `BlockingFunc` we blocked the main thread. We are already in the `default` case of the select block. so `case <-done:` won't do anything.

Here's the full code. 

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riadafridishibly/go-graceful-shutdown/utils"
)

func nonresponsive(done <-chan bool) (string, error) {
	select {
	case <-done:
		return "", errors.New("operation cancelled")
	default:
		return utils.BlockingFunc() // select won't do anything
	}
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	utils.SimulateSendSignal(1*time.Second, os.Interrupt)

	done := make(chan bool)
	go func() {
		<-sig
		close(done)
	}()

	v, err := nonresponsive(done)
	if err == nil {
		fmt.Println(">>> CANCEL DID NOT WORK")
	}
	fmt.Printf("Value: %q, err: %v\n", v, err)
}
```

We don't want this behavior, we want our program more responsive. To make it responsive we can execute the blocking function in another goroutine and send the results to another channel. Let's rewrite the `nonresponsive` function.

```go
func responsive(done <-chan bool) (string, error) {
	type result struct {
		value string
		err   error
	}
	ch := make(chan result)
	go func() {
		v, err := utils.BlockingFunc()
		ch <- result{v, err}
	}()
	select {
	case <-done:
		return "", errors.New("process cancelled")
	case v := <-ch:
		return v.value, v.err
	}
}
```

Here we've defined a new type called `result`. This struct represents the return values of the `BlockingFunc`. We create a new channel `ch` of type `result`, spawn a new goroutine and send the result back to the channel. Now the select is blocking. It's waiting for either of the two, value from the `done` channel or value from the `ch` channel.

So if we receive value from `done` before `ch` then we'll return immediately. So our blocking state is now gone. 

Let's try the next code snippet.

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riadafridishibly/go-graceful-shutdown/utils"
)

func responsive(done <-chan bool) (string, error) {
	type result struct {
		value string
		err   error
	}
	ch := make(chan result)
	go func() {
		v, err := utils.BlockingFunc()
		ch <- result{v, err}
	}()
	select {
	case <-done:
		return "", errors.New("process cancelled")
	case v := <-ch:
		return v.value, v.err
	}
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	utils.SimulateSendSignal(1*time.Second, os.Interrupt)

	done := make(chan bool)
	go func() {
		<-sig
		close(done)
	}()
	v, err := responsive(done)
	fmt.Printf("Value: %q, err: %v\n", v, err)
}
```

We can use the previous example, but I think the context way is cleaner. Go 1.20 introduced [WithCancelCause](https://pkg.go.dev/context#WithCancelCause), we can use that here. 

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riadafridishibly/go-graceful-shutdown/utils"
)

func responsive(ctx context.Context) (string, error) {
	type ret struct {
		value string
		err   error
	}
	ch := make(chan ret)
	go func() {
		v, err := utils.BlockingFunc()
		ch <- ret{v, err}
	}()
	select {
	case <-ctx.Done():
		return "", context.Cause(ctx)
	case v := <-ch:
		return v.value, v.err
	}
}

func main() {
	fmt.Println("PID:", os.Getpid())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancelCause(context.Background())

	utils.SimulateSendSignal(1*time.Second, os.Interrupt)

	go func() {
		got := <-sig
		cancel(fmt.Errorf("signal %s", got))
	}()

	v, err := responsive(ctx)
	fmt.Printf("Value: %q, err: %v\n", v, err)
}
```

It's also common practice in golang to pass `context.Context` as the first parameter in blocking functions. 

### Shutting down the HTTP server 

The graceful shutdown makes more sense while exiting any kind of server. Here's an example of exiting the default HTTP server of go `net/http`.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func reqLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("Method = %s, Path = %s, Took = %v",
			r.Method, r.URL.Path, time.Since(now))
	})
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)

	srv := &http.Server{
		Handler: reqLogMiddleware(mux),
		Addr:    ":8083",
	}

	go func() {
		<-sig
		log.Println("Shutdown sequence initiated")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Println("Error shutting down server. err:", err)
		}
	}()

	log.Println("Server started at: http://localhost:8083/")
	if err := srv.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			log.Println("Http server stopped")
		} else {
			log.Fatal(err)
		}
	}
}
```

## Signal reset

Sometimes we want to handle the first signal and, the subsequent signals sent to the process we may not want to handle (we want to fall back to the default behavior; remember the default behavior of `SIGINT`, `SIGTERM` is to kill the process). Let's say if the graceful shutdown takes more time user may want to exit the process right away. To enable this we need to reset the signal handler after capturing the first signal. Let's see the example in action.


```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/riadafridishibly/go-graceful-shutdown/utils"
)

func main() {
	fmt.Println("PID:", os.Getpid())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	utils.SimulateSendSignal(1*time.Second, os.Interrupt)
	utils.SimulateSendSignal(2*time.Second, syscall.SIGTERM)
	utils.SimulateSendSignal(3*time.Second, os.Interrupt)

	got := <-sigCh
	fmt.Printf("Received Signal: %s, Sig Num: %d\n", got, got)

	// Comment out the next line and run the program again
	signal.Reset(os.Interrupt, syscall.SIGTERM)

	go func() {
		// To show that we're still receiving signals
		for got := range sigCh {
			fmt.Printf("Received Signal: %s, Sig Num: %d\n", got, got)
		}
	}()

	for i := 0; i < 5; i++ {
		fmt.Printf("Exiting in %d sec\n", 5-i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Exited")
}
```

## Conclusion

We may not need to handle signals for all applications, but for stateful applications like web servers, it's a good idea to handle graceful shutdown so that all connections are closed properly and all data is flushed to the disk or database. 

Thank you. :)
