---
published: true
comments: true
layout: post
title: Trace Logging in Golang
description: "Add a TRACE log level to a Go program to observe code execution paths, using runtime caller info to report the file and line that logged."
categories: [programming]
tags: [golang, logging, trace, debugging]
---

_Disclaimer: This article is not about `runtime/trace` package or trace profiling functionalities._

Tracing is used to observe code execution paths. If you log something in a function and you see that output in the log file, then you know that the function has been executed. There is a log level defined in different logger implementations called `TRACE`, which is the lowest log level in most loggers. The log won't be present in the production environment, but when debugging, you'll receive the traces. This is useful in situations where a debugger is not accessible.

## Code

[Playground](https://go.dev/play/p/BhZEMnJw2m8)

```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var TraceLogger = log.New(os.Stderr, "TRACE ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)

func Trace(skip int, format string, args ...any) {
	TraceLogger.Output(skip+1+1, fmt.Sprintf(format, args...)) // Skip: [Output, Trace]
}

func trace(msg string) (m string, start time.Time) {
	Trace(1, "Entering: %s", msg)
	return msg, time.Now()
}

func un(m string, start time.Time) {
	Trace(1, "Exiting: %s, TOOK: %v", m, time.Since(start))
}

func main() {
	defer un(trace("main"))
	time.Sleep(400 * time.Millisecond)
}
```

The output will look like this, 

```
// 2023/10/08 21:20:30 main.go:26: TRACE Entering: main
// 2023/10/08 21:20:31 main.go:28: TRACE Exiting: main, TOOK: 401.070667ms
```

So any function we want to trace we can add the following line on top, 

```go
func myFunction() {
	defer un(trace("<any message or function name>"))
    // rest of the body...
}
```

It will also report at which point the function exited! If your function has multiple return paths, you will also see the log there along with the line number at which point the function exited.

[Here is an example](https://go.dev/play/p/o9QYOUrMxV0)


## Explanation

At the entry point `trace` is immediately executed and returned. However the function `un` is deferred, so we see the `Entering` log first which is logged by the trace function. Finally, when we exit out of the function, `un` is called.

Another adjustment we've made is to the `calldepth` of the `Trace` and `log.Output` functions. I mapped the skip variable like this, 

```go
Trace(0, ...) // skip = 0; means Log the caller of Trace
Trace(1, ...) // skip = 1; means log the caller of caller of Trace
```

for more information about skip, please check the docs for [runtime.Caller](https://pkg.go.dev/runtime#Caller)

That's it! Happy Coding!