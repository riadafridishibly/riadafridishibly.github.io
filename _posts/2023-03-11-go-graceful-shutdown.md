---
published: true
comments: true
layout: post
title: Graceful Shutdown in Golang
author: Riad Afridi Shibly
categories: programming
tags: [signal, graceful, shutdown, channel]
image: 2023-03-11-go-graceful-shutdown/broken_computer.jpg
---

# Table Of Contents
{:.no_toc}
- 
{:toc}

# Graceful Shutdown

When a process is running for a long time, sometimes we may want to quit the running program. To do so we sometimes press `CRTL+C`, send a specific kill signal to the process, quit from the menu and so on. These actions trigger an event and the event is sent to the specific process which sometimes causes the process to exit. In a stateful program maybe we want to save the states or perform cleanups before exiting the process. To fulfill these specific needs we need to perform a graceful shutdown.

In short, we'll catch the signal. Then perform the required cleanup and exit.

## What are signals in OS context?

A _signal_ is a software interrupt delivered to a process. Here we'll be dealing with signals which cause the process to die.


## An Example Program

To start understanding signals let's write a simple program.

<!-- FIXME: Update the link -->

```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Process PID:", os.Getpid())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Wait for signal
	got := <-sigCh

	fmt.Printf("Received Signal: %s, Sig Num: %d\n", got, got)
}
```

Run this program and then press `CTRL+C`. It should output something like this,

```
Process PID: 123768
^CReceived Signal: interrupt, Sig Num: 2
```

Now run the program again, open a new terminal window and run this command.

```sh
$ kill -SIGTERM <PID>
```

The `PID` value is printed. You can grab the `PID` from there.

For example,

```sh
$ kill -SIGTERM 123768
```

Now the output should look like this,

```
Process PID: 129465
Received Signal: terminated, Sig Num: 15
```

Let's run the program again and kill with `-SIGKILL`. The output should look like this,

```
Process PID: 131937
signal: killed
```

Okay, let's recap what's happening here. First of all we grabbed the `PID` of the running program and print that. Then with these two lines we've created a buffered channel of size 1 and register two signals `SIGINT` and `SIGTERM`. 

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
```

This way if any of these signals is sent to out program, `signal.Notify` will send that signal to the `sigCh` channel.

Next, we wait on `sigCh` channel.

```go
// Wait for signal
got := <-sigCh
```

Then we're waiting for the signal to arrive. Normally our app will block elsewhere, for example a web server will block the `main` function. But in this case we're simply waiting on `sigCh` for any signal.

## What are the available signals?

We've already seen three signals, `SIGINT`, `SIGTERM` and `SIGKILL`. If they eventually kill the process then why are they different? 

> **One important thing about signals is, not all signals are catchable. If we try to catch `SIGKILL` or `SIGSTOP` we won't be able to do so. Kernel can catch it, but userspace program can not.**

Why are they different? Well, we can catch different signals and handle them differently, [here's a short description of different signals and their meaning](https://www.gnu.org/software/libc/manual/html_node/Termination-Signals.html).

## Why did we initialize the **sigCh** as buffered channel?

From the [`signal.Notify`](https://pkg.go.dev/os/signal#Notify) docs, 

> Package signal will not block sending to c: the caller must ensure that c has sufficient buffer space to keep up with the expected signal rate. For a channel used for notification of just one signal value, a buffer of size 1 is sufficient.

So if we don't provide a buffer, `signal.Notify` won't wait for sending the signal to the channel. In golang, sending to an unbuffered channel will be successful when there's another goroutine waiting for receiving from that channel. Otherwise, sending operation will block. Let's demonstrate that with another simple code.


<!-- FIXME: Provide repo link -->
```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Process PID:", os.Getpid())
	sigCh := make(chan os.Signal, 1) // Change this to unbuffered
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Sleep started. Now press C-c")
	time.Sleep(10 * time.Second)
	fmt.Println("Sleep done...")

	got := <-sigCh
	fmt.Printf("Received Signal: %s, Sig Num: %d\n", got, got)
}
```

In this case, if we press `CTRL+C` after starting the sleep, then our signal will still be registered. But if we change the `make(chan os.Signal, 1)` line to `make(chan os.Signal)` then, after starting sleep we won't be able to register the signal anymore! Try running this and the program won't exit the first time you press `CTRL+C`.

## Signal Broadcast

We've seen we can capture the signal. But how do we propagate the signal throughout our app? 

Before exploring this area, let's quickly review the channel behaviors.

- Sending to or receiving from nil channel will block.
- Sending to a closed channel will panic.
- Receiving from a closed channel returns immediately, and can be used multiple times.

Let's see a few different cases where we can implement signal broadcast.

### When we already have channel

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

Let's convert this code to this,

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
)

func splitStringDone(s string, done <-chan bool) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, v := range strings.Fields(s) {
			select {
			case ch <- v:
				// This select is for blocking for 1 sec
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
	fmt.Println("Process PID:", os.Getpid())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

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
}
```

Here we're handling the signal in a goroutine. So either our loop ends or we initiate cancellation with a signal. When we catch any signal we simply close the `done` channel. And in the select block `<-done` is selected and we return.

### Dealing with blocking functions

Sometimes we may have a blocking function. With a blocking function, we can't simply use select, if we do we'll just block the case (that's why we didn't put `time.Sleep(1 * time.Second)` in the previous example. we've used another select.).

When we are in blocking state, the select switch won't help us. Let's simulate blocking state with this function,

```go
func blockingFunc() (string, error) {
	fmt.Println("Blocking func started, will sleep for 10 sec")
	defer fmt.Println("Blocking func finished")

	time.Sleep(10 * time.Second)
	return "some value", nil
}
```

This function prints something at the start, then it sleeps for 10 seconds and returns a string and an error. Finally, it prints its status that the function has exited.

If we call this function directly we'll block our program for 10 seconds. In the meantime, the signal catcher won't work. To demonstrate the problem let's run the following program and press `CTRL+C` when the program prints `Blocking func started`. Our signal won't exit the program, rather it'll hang for 10 seconds and the program will exit. The problem is in the select block. Because as soon as we start executing `blockingFunc` we block the main thread. We are already in `default` case of the select block. so `case <-done:` won't be executed anymore.

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
)

func blockingFunc() (string, error) {
	fmt.Println("Blocking func started, will sleep for 10 sec")
	defer fmt.Println("Blocking func finished")

	time.Sleep(10 * time.Second)
	return "some value", nil
}

func nonresponsive(done <-chan bool) (string, error) {
	select {
	case <-done:
		return "", errors.New("cancelled operation")
	default:
		return blockingFunc() // select won't do anything
	}
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	done := make(chan bool)
	go func() {
		<-sig
		close(done)
	}()
	v, err := nonresponsive(done)
	fmt.Printf("Value: %q, err: %v\n", v, err)
}
```

We don't want this behavior, we want our program more responsive. To make it responsive we can execute the blocking function in another goroutine and send it results to another channel. Let's rewrite the `nonresponsive` function in a responsive manner.

```go
func responsive(done <-chan bool) (string, error) {
	type result struct {
		value string
		err   error
	}
	ch := make(chan result)
	go func() {
		v, err := blockingFunc()
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

Here we've defined new type `result`. This struct is simply represents the return values of the `blockingFunc`. We create a new channel `ch`, and spawn a new goroutine and send the result back to the channel. Now the select is blocking. It's waiting for either of the two, value from `done` channel or value from `ch` channel.

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
)

func blockingFunc() (string, error) {
	fmt.Println("Blocking func started, will sleep for 10 sec")
	defer fmt.Println("Blocking func finished")

	time.Sleep(10 * time.Second)
	return "some value", nil
}

func responsive(done <-chan bool) (string, error) {
	type result struct {
		value string
		err   error
	}
	ch := make(chan result)
	go func() {
		v, err := blockingFunc()
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
)

func blockingFunc() (string, error) {
	fmt.Println("Blocking func started, will sleep for 30 sec")
	defer fmt.Println("Blocking func finished")

	time.Sleep(30 * time.Second)
	return "some value", nil
}

func responsive(ctx context.Context) (string, error) {
	type ret struct {
		value string
		err   error
	}
	ch := make(chan ret)
	go func() {
		v, err := blockingFunc()
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

	go func() {
		got := <-sig
		cancel(fmt.Errorf("signal %s", got))
	}()

	v, err := responsive(ctx)
	fmt.Printf("Value: %q, err: %v\n", v, err)
}
```

## Signal Reset