---
published: true
comments: true
layout: post
title: "[Part 1] Understanding HTTP/1.1: TCP Server"
author: Riad Afridi Shibly
categories: programming
tags: [http, tcp, network, golang]
---

## Table Of Contents

{:.no_toc}

- {:toc}

The goal of this blog post is to spin up a TCP server and start sending HTTP requests to the server. That way we'll see what type of data an HTTP client like curl sends to the server. Out of all the versions of HTTP, we'll focus on `HTTP 1.x`. As it a text based protocol and we'll be able to see what's going on under the hood.

We'll use an HTTP client application like `curl` or a browser to connect to the TCP server and investigate what type of data they send to the server. Later we'll construct an HTTP response and send it to the client. It's very easy to start with HTTP server directly but we want to go from TCP to HTTP and investigate the requests and response formats along the way. Later in this journey we'll build a simple reverse proxy.

For this experiment we'll use `golang`. It's quite a good fit for network programming and the standard library already includes all the bells and whistles for our convenience.

These are the things we'll do in this blog post,

1. Create a simple TCP echo server
2. Connect with `telent` client and send messages
3. Try to connect with a http client
4. Send http response to the client
5. Understand the http request format
6. Understand the http response format
7. Multipart? How does MIME types work?
8. What if we want to upload binary data?
9. `Expect: 100-continue` header
10. All http requests

## Creating a TCP Server

The first thing we need is a TCP server. It's quite easy to spin up a TCP server in golang. All we need to do is import the `net` package and use some functionality from there. The basic idea is quite simple. We'll listen on a port and wait for the incoming connection. Once someone (any client) connects to us we'll spawn a goroutine to handle the connection. Go's asynchronous programming model allows us to implement concurrent connection handling very easy. A basic TCP server looks like this. This is also the very example given by the go documentation website to demonstrate the `Listen` function. You can find the example right [here](https://pkg.go.dev/net#example-Listener).

```go
package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func handleConn(c net.Conn) {
	defer c.Close()
	io.Copy(c, c)
}

func main() {
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	fmt.Println("Listening on:", l.Addr().String())

	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go handleConn(conn)
	}
}
```

I assume you already understand what does the above code do. As I already provided a little bit of explanation before. If you still have confusion you can run the code and play with it. That's the whole purpose of this blog post. We'll connect to this very server with various client and try to understand what's going on. It's some kind of `EDD`, which I mockingly call "Error Driven Development". So we'll handle error along the way.

So what does the `handleConn` function do? Nothing much. The first line simply saying whenever the function returns close the connection. `defer` in many programming languages simply mean, "execute the function after it returns". After that we've used `io.Copy`. It's a handy function in golang. We're basically reading from a `source` and writing the same thing to the `destination`. In our case we're reading from the connection `c` and writing the same thing to `c`. As you can guess from the behavior it's an echo server. So whatever is sent to the server will be echoed back to the client.

Let's start the server.

> _I assume you already know how to run go programs. If you have go installed in your machine then create a directory. `cd` into the directory. Run `go mod init simple-tcp-server`. Then create a `main.go` file. And copy paste the above code. Now you should be able to run `go run main.go` from the directory._

So how do we connect to the server?

Depending on your OS, you may already have `telnet` installed on your system. If you're already good with `golang` you can write another program which uses `net.Dial` to connect to the server. Let's use `telnet`. You should be able to get `telnet` for your system as well. Try quick google search or ask `ChatGPT` for that.

We're listening on port 2000. So let's connect to that. This command should start the connection `telnet 127.0.0.1 2000`.

Now let's try to connect to the server. Let's write something and press `Enter` after each line. _The `Enter` will flush the message, before that it may keep the thing in a buffer. But when we press enter it'll immediately write the message to the connection._

> To exit `telnet` press `CTRL + ]`

```
$ telnet 127.0.0.1 2000
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
one
one
two
two
helloooo
helloooo
bye
bye
^]
telnet> Connection closed.
```

The previous code snippet is basically copy pasted from a terminal. We connected to the server. Whatever we sent to the server, was echoed back to us.

Not very interesting. But we have a tcp server. Remember our goal is to see what other clients like browser and curl send to the server. To see that fully we'll everything in server side as well. It's also quite easy to do in golang.

```go
func handleConn(c net.Conn) {
	defer c.Close()
	io.Copy(c, io.TeeReader(c, os.Stdout))
}
```

We updated the `handleConn` function with this. We are using a `TeeReader` here. You might ask what is a `TeeReader`? Good question. Simply put, whatever is read from `c` will be written to `os.Stdout` and the same data will be available to read from the `io.TeeReader`. Isn't it cool? Otherwise we need to read in a temporary buffer then write the same data to `os.Stdout` and the connection `c`. You can experiment with this. And there's a unix command `tee` as well. You can check that out too. It's the right tool for these kind of scenarios where data is being streamed.

Now if we restart the server and connect with `telnet` again we'll see something like this. Depending on what type of message you send. Of course you're not gonna send one, two, hellooo, are you?

```
$ go run main.go
Listening on: [::]:2000
one
two
helloooo
```

## What if we try to connect with an http client?

To be honest, our echo server is not that interesting, is it? Now let's see what happens if we try to connect with an HTTP client. Probably the very client you want to use is already installed on your system. Yeah it's none other than `curl`. Let's run a command.

```
$ curl -v http://localhost:2000
```

The `curl` output looks like this.

```text
* Host localhost:2000 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
*   Trying [::1]:2000...
* Connected to localhost (::1) port 2000
> GET / HTTP/1.1
> Host: localhost:2000
> User-Agent: curl/8.7.1
> Accept: */*
>
* Request completely sent off
* Received HTTP/0.9 when not allowed
* Closing connection
curl: (1) Received HTTP/0.9 when not allowed
```

I intentionally enabled verbose mode just to see what type of data curl sends to the server. Ignore the output for now. By default if you don't specify anything, curl sends an http `GET` request to the server. Let's not worry about the last line `curl: (1) Received HTTP/0.9 when not allowed` right now. Our echo server tried to send the same request back to the client. But client expected something different. _We'll construct an HTTP response later, and `curl` will be happy. No worries._

The most interesting thing is logged on the server end. Something like this,

```
GET / HTTP/1.1
Host: localhost:2000
User-Agent: curl/8.7.1
Accept: */*
```

This is a `GET` request. So if we connect to an HTTP server with telnet and write the same thing as above, we'll get the http response. Let's put our theory in test. Here's a simple http server which sends a response with just `Hello, World!`. Go ahead run this code.

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})
	http.ListenAndServe(":8080", mux)
}
```

To check our server is indeed working you can browse the link `http://localhost:8080` on a browser or just run `curl http://localhost:8080`. The output should look like this,

```
$ curl http://localhost:8080
Hello, World!
```

Now the interesting part, let's use telnet again. Run `telnet localhost 8080` and paste the following code and press `Enter` twice,

```
GET / HTTP/1.1
Host: localhost:2000
User-Agent: curl/8.7.1
Accept: */*
```

Then the output should look like this!

```
$ telnet localhost 8080
Trying ::1...
Connected to localhost.
Escape character is '^]'.
GET / HTTP/1.1
Host: localhost:2000
User-Agent: curl/8.7.1
Accept: */*

HTTP/1.1 200 OK
Date: Tue, 24 Dec 2024 20:08:22 GMT
Content-Length: 14
Content-Type: text/plain; charset=utf-8

Hello, World!
```

Well, we've constructed an HTTP request and send it to the server and server responded with a response. As you can already tell just like the http request, http response also has special format. We haven't looked at that carefully yet. But soon we'll do.

Let's get back to our echo server. If we try to visit `http://localhost:2000` with a browser, the browser will tell us `localhost sent an invalid response`.

But if we check the server log, we'll see the request is being logged as this,

```http
GET / HTTP/1.1
Host: localhost:2000
Connection: keep-alive
sec-ch-ua: "Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"
sec-ch-ua-mobile: ?0
sec-ch-ua-platform: "macOS"
Upgrade-Insecure-Requests: 1
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) [..truncated..]
Accept: text/html,application/xhtml+xml,application/xml;q=0.9, [..truncated..]
Sec-Fetch-Site: none
Sec-Fetch-Mode: navigate
Sec-Fetch-User: ?1
Sec-Fetch-Dest: document
Accept-Encoding: gzip, deflate, br, zstd
Accept-Language: en-GB,en-US;q=0.9,en;q=0.8
```

Ah ha! Something is happening! What's going on?

This is a very specific message. Notice the first line for each case. It says `GET / HTTP/1.1` for both cases. As we can already guess, this format of message is called a HTTP GET request. And this version of http is a text based protocol. The next thing to do is parse the http request. We're not going to parse it manually. We'll use [http.ReadRequest](https://pkg.go.dev/net/http#ReadRequest) function from go's http library.

But before doing any of that, let's see what's ana http get request?

## Parsing the request and sending a response

Now that we can see that request, let's try to parse it. Currently the request is pretty bare bone. Later we'll see how does it look like when request is made with some data and even with a large binary file.

The new `handleConn` method will look like this.

```go
func handleConn(c net.Conn) {
	defer c.Close()

	br := bufio.NewReader(c)
	req, err := http.ReadRequest(br)
	if err != nil {
		log.Fatalln("Failed to read request", err)
	}

	log.Println("Parsed http request:")
	log.Println("\tMethod:", req.Method)
	log.Println("\tURI:", req.URL.String())
	log.Println("\tProto:", req.Proto)
	log.Println("\tUser-Agent:", req.Header.Get("user-agent"))

	resp := http.Response{
		StatusCode: 200,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Body:       io.NopCloser(strings.NewReader("Hello, World!")),
		Request:    req,
	}

	err = resp.Write(c)
	if err != nil {
		log.Fatalln("Failed to write response", err)
	}
}
```

Now if we run `curl -v http://localhost:2000` the output will look like this,

```
* Host localhost:2000 was resolved.
* IPv6: ::1
* IPv4: 127.0.0.1
*   Trying [::1]:2000...
* Connected to localhost (::1) port 2000
> GET / HTTP/1.1
> Host: localhost:2000
> User-Agent: curl/8.7.1
> Accept: */*
>
* Request completely sent off
< HTTP/1.1 200 OK
< Connection: close
<
* Closing connection
Hello, World!⏎
```

And the server logs will show the parsed messages.

```
Parsed http request:
    Method: GET
    URI: /
    Proto: HTTP/1.1
    User-Agent: curl/8.7.1
```

If we request `curl -v http://localhost:2000/foobar` from a browser the output will look like this,

```
Parsed http request:
    Method: GET
    URI: /foobar
    Proto: HTTP/1.1
    User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36
```

I've requested `/foobar` and is correctly reflected on the server side.

## What if we try to upload file?

- NOTE: `curl` command to upload file
- `curl -v --form firstfile=@main.go --form secondfile=@tcp-http-server "http://127.0.0.1:6969"`

## Where are we going with this example?
