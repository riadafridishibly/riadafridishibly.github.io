---
published: true
comments: true
layout: post
title: Structured Logging in Golang
author: Riad Afridi Shibly
categories: programming
tags: [golang]
---

## Table Of Content

1. Table Of Content
{:toc}

_So you want to use the new shiny structured logging library in golang_

## log/slog

`slog` (a _structured_ logging library in Go) has been included in the Go standard library starting from version 1.21. While Go already had a logging library, it was considered too simple for most use cases. For starters, it lacked the leveled logging feature. To achieve leveled logging, you had to create multiple loggers with the level as a prefix and direct them all to the same `io.Writer`. This package, [spf13/jwalterweatherman](https://github.com/spf13/jwalterweatherman), accomplishes exactly that. There are also several third-party, excellent logging libraries available for structured logging, such as [rs/zerolog](https://github.com/rs/zerolog), [uber-go/zap](https://github.com/uber-go/zap), and [sirupsen/logrus](https://github.com/sirupsen/logrus), to name a few. Now that structured logging is included in the standard library, let's give it a try.

We can start using `slog` by simply importing `log/slog`. The documentation already provides plenty of examples; you can read it [here](https://pkg.go.dev/log/slog). Here's an overview of `slog` and some configuration options we may want to use in a new project!

## A very basic example

```go
package main

import (
	"context"
	"log/slog"
	"os"
)

var LogLevel = new(slog.LevelVar)

func main() {
	// setting the log level
	LogLevel.Set(slog.LevelDebug - 1)

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: LogLevel,
	})

	slog.SetDefault(slog.New(handler))

	slog.Error("error message")
	slog.Warn("warn message")
	slog.Info("info message")
	slog.Debug("debug message")

	slog.Info("info with some key value data", "foo", "bar")
	slog.Info("same as previous but using attribute",
		slog.String("foo", "bar"))

	slog.Info("info with as group",
		slog.Group("num",
			slog.Int("x", 32),
			slog.String("name", "jon")),
		slog.String("foo", "bar"))

	slog.LogAttrs(context.Background(), slog.LevelInfo,
		"well enforcing attribute",
		slog.String("foo", "bar"),
	)

	slog.Log(context.Background(), slog.LevelDebug-1, "trace log")
}
```

Output will look like this,

```
{"time":"2023-09-15T02:17:20.328534+06:00","level":"ERROR","msg":"error message"}
{"time":"2023-09-15T02:17:20.328854+06:00","level":"WARN","msg":"warn message"}
{"time":"2023-09-15T02:17:20.328859+06:00","level":"INFO","msg":"info message"}
{"time":"2023-09-15T02:17:20.328864+06:00","level":"DEBUG","msg":"debug message"}
{"time":"2023-09-15T02:17:20.328868+06:00","level":"INFO","msg":"info with some key value data","foo":"bar"}
{"time":"2023-09-15T02:17:20.328873+06:00","level":"INFO","msg":"same as previous but using attribute","foo":"bar"}
{"time":"2023-09-15T02:17:20.32888+06:00","level":"INFO","msg":"info with as group","num":{"x":32,"name":"jon"},"foo":"bar"}
{"time":"2023-09-15T02:17:20.328907+06:00","level":"INFO","msg":"well enforcing attribute","foo":"bar"}
{"time":"2023-09-15T02:17:20.328912+06:00","level":"DEBUG-1","msg":"trace log"}
```

## Configuration

We want to change the time format, replace `DEBUG-1` level with the keyword `TRACE` and use only short filename for source. To add the source (filename and line number) information we need to add `AddSource: true` in the `slog.HandlerOptions`. To do this, we are going to define a function. See [this example](https://pkg.go.dev/log/slog#example-HandlerOptions-CustomLevels).

Let's modify the example for our needs.

```go
func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		a.Value = slog.StringValue(a.Value.Time().Format(time.DateTime))
	}

	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		switch {
		case level < slog.LevelDebug:
			a.Value = slog.StringValue("TRACE")
		case level < slog.LevelInfo:
			a.Value = slog.StringValue("DEBUG")
		case level < slog.LevelWarn:
			a.Value = slog.StringValue("INFO")
		case level < slog.LevelError:
			a.Value = slog.StringValue("WARN")
		default:
			a.Value = slog.StringValue("ERROR")
		}
	}

	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.File = filepath.Base(source.File)
	}

	return a
}
```

Now the handler creation will look like this,

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level:       LogLevel,
	AddSource:   true,
	ReplaceAttr: replaceAttr,
})
```

Now our output will look as expected. We can do the same thing for the `function` key as well. See `slog.Source` struct for that.

## Wrapping

We want to wrap our `log.Log` call from another function and still keep the relevant file and line number. The example is also given in documentation. [Check it here](https://pkg.go.dev/log/slog#example-package-Wrapping). Don't forget to read the friendly manual. ;) 


```go
func log(ctx context.Context, skip int, 
        logger *slog.Logger, level slog.Level, msg string, attrs ...slog.Attr) {
	if !logger.Enabled(context.Background(), level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2+skip, pcs[:]) // skip [Callers, log]
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(attrs...)
	_ = logger.Handler().Handle(context.Background(), r)
}

func InfoWithSkip(skip int, logger *slog.Logger, msg string, attrs ...slog.Attr) {
	log(context.Background(), skip+1, logger, slog.LevelInfo, msg, attrs...)
}

func DoNotLogMe() {
	InfoWithSkip(1, slog.Default(), "Logging from DoNotLogMe", 
	    slog.String("foo", "baz"))
}
```

Try calling `DoNotLogMe` function from any function. You won't see `DoNotLogMe` in the `source` information.

Upon calling `DoNotLogMe` from `main` the output will look like this,

```
{"time":"2023-09-15 02:55:20","level":"INFO","source":{"function":"main.main","file":"main.go","line":96},"msg":"Logging from DoNotLogMe","foo":"baz"}
```

_I try to design my function such a way where skip value 0 means caller of the function, 1 means caller of the caller of the function and so on. This is similar to runtime.Caller_

## Writing to file

We have initialized the handler with `os.Stdout`, writing to file is easy. We need to open a file and throw it in the handler. Or we can use [lumberjack](https://github.com/natefinch/lumberjack). Which will manage the file for us, rotate the log and also compress it. Very handy.

If we want to write in multiple location, for example while developing software I want to write in `stdout` as well, then we can simply use [`io.MultiWriter`](https://pkg.go.dev/io#MultiWriter).


## Tips

Use log valuer when you wanna throw your full object in logger. Take a look [Working With Records](https://pkg.go.dev/log/slog#hdr-Working_with_Records). Here is [an example](https://pkg.go.dev/log/slog#example-LogValuer-Group).

If you use `slog.Info/Debug/Error` you can miss some key value pair, because the key values are just comma separated. Use `go vet ./...` (you should run `go vet` anyway) to detect the cases where you've missed the key value pairs. Otherwise it'll show bad key. For example, 

```go
slog.Info("info with some key value data",
	"key1", "value1", "key2")
```

Will output something like this, 

```
2023/09/15 03:11:49 INFO info with some key value data key1=value1 !BADKEY=key2
```

If you run `go vet` it'll catch your error, 

```
# command-line-arguments
vet/main.go:8:2: call to slog.Info missing a final value
```

Here is the analyzer for [vet/slog](https://pkg.go.dev/golang.org/x/tools@v0.13.0/go/analysis/passes/slog). When I've defined my `InfoWithSkip` function I didn't use `args ...any` as the final param, rather I've used `slog.Attr`. Because if you wrap your log function in another function `vet` can't detect if you've passed the right number of arguments. This will probably fixed in a future vet version.

