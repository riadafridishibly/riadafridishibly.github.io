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

So if we don't provide buffer, `signal.Notify` won't wait for sending the signal to the channel. 