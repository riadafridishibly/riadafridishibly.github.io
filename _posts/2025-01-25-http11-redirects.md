---
published: true
comments: true
layout: post
title: "Understanding HTTP/1.1: Redirects in Go (Part 3)"
description: "301 vs 302 vs 307 vs 308 — which HTTP redirect to use, which ones browsers cache, which preserve the request method, tested against a small Go server."
categories: [programming]
tags: [http, redirects, networking, golang]
---


When I first learned about HTTP redirects I was confused about them. Which one do I need? One good way to always start from is the [MDN Docs](https://developer.mozilla.org/en-US/docs/Web/HTTP/Redirections) for any web standard related questions.

Read the docs and you don't even need to read this at all. Bye. Or you can stay...

**TL;DR**

`301` is permanent redirect. So if you hit it once the browser will cache it and won't bother to reach out to the server for subsequent requests. On the other hand `302` won't be cached and browser will hit our server every time user requests one of the urls. Same goes for `308` and `307`. But they'll preserve the HTTP method. There's a nice table for that in MDN Docs. 

Now let's test things out!

Here we'll spawn a golang server and experiment with it, that's what I prefer.

We'll only need two things. An `index.html` file and a simple golang server. You can use whatever language you like.

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Redirect test</title>
	<style>a { display: block }</style>
  </head>
  <body>
    <form action="/301" method="post">
      <input type="text" name="data" id="r301" />
      <button type="submit">301 Submit</button>
    </form>
    <form action="/302" method="post">
      <input type="text" name="data" id="r302" />
      <button type="submit">302 Submit</button>
    </form>
    <form action="/307" method="post">
      <input type="text" name="data" id="r307" />
      <button type="submit">307 Submit</button>
    </form>
    <form action="/308" method="post">
      <input type="text" name="data" id="r308" />
      <button type="submit">308 Submit</button>
    </form>


    <a href="/301">301 Redirect</a>
    <a href="/302">302 Redirect</a>
    <a href="/307">307 Redirect</a>
    <a href="/308">308 Redirect</a>
  </body>
</html>
```

And our golang server, 

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	_ "embed"
)

//go:embed index.html
var index string

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		data := r.FormValue("data")
		log.Printf("Received request, Method=%q URI=%q Data=%q", r.Method, r.RequestURI, data)
		h.ServeHTTP(rw, r)
	})
}

type redirect struct {
	code int
	to   string
}

func (redir redirect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, redir.to, redir.code)
}

func main() {
	log.SetFlags(0)
	mux := http.NewServeMux()
	// We want exact match, thus `/{$}`, golang's new syntax from 1.22
	// otherwise all unmached routes will fallback to this one
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, index) })
	mux.HandleFunc("/welcome", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "Welcome") })

	to := "/welcome"
	mux.Handle("/301", redirect{301, to})
	mux.Handle("/302", redirect{302, to})
	mux.Handle("/307", redirect{307, to})
	mux.Handle("/308", redirect{308, to})

	log.Println("Starting http server at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", WithLogging(mux)))
}
```

Run `go run main.go` and the http server will start on port `8080`.

Now Let's start with the `302` redirect. It's a temporary redirect. Open a new browser tab and the devtool as well. Now visit `http://localhost:8080/`.

It'll show some forms and buttons. We'll follow the server logs. That's where the interesting things are happening. Go to network tab and **Uncheck Disable Cache**.

Home screen should look like this,
![Redirect demo home page](/assets/img/2025-01-25-http11-redirects/home.png)

Now if we press `302 Redirect` button, we'll see these lines on server logs, 

```
Received request, Method="GET" URI="/302" Data=""
Received request, Method="GET" URI="/welcome" Data=""
```


What if we go back and press the button again? Let's try that. In this case we can see these two logs again!

```
Received request, Method="GET" URI="/302" Data=""
Received request, Method="GET" URI="/welcome" Data=""
```

Let's try this with the `301` redirect. 

The first time server browser will hit the server. So the server logs will look like this,

```
Received request, Method="GET" URI="/301" Data=""
Received request, Method="GET" URI="/welcome" Data=""
```

But if you request it next time you'll see only one log on the server,

> Make sure the disable cache is unchecked

```
Received request, Method="GET" URI="/welcome" Data=""
```

This is because browser knows it's permanent redirect. So browser cached it, and directly visits the URL.

Now you can check the later methods, 

If you hit 301 submit with some data, 

```
Received request, Method="POST" URI="/301" Data="hello"
Received request, Method="GET" URI="/welcome" Data=""
```

Here the first method requested is POST, but later it changed to GET. and our data `hello` is lost. 

But if you try with `308`, the log will look like this, 


```
Received request, Method="POST" URI="/308" Data="hello"
Received request, Method="POST" URI="/welcome" Data="hello"
```

Both method and data is preserved here. [Next JS by default uses `307` and `308` status code for redirection. You can read here about that.](https://nextjs.org/docs/app/api-reference/functions/redirect#why-does-redirect-use-307-and-308)


Okay, what's inside the request then? how does the response look like to the client? It's basically a status code with a location header.

```
~❯❯❯ curl -I localhost:8080/301
HTTP/1.1 301 Moved Permanently
Content-Type: text/html; charset=utf-8
Location: /welcome
Date: Sun, 26 Jan 2025 12:31:03 GMT

~❯❯❯ curl -I localhost:8080/302
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Location: /welcome
Date: Sun, 26 Jan 2025 12:31:06 GMT

~❯❯❯ curl -I localhost:8080/307
HTTP/1.1 307 Temporary Redirect
Content-Type: text/html; charset=utf-8
Location: /welcome
Date: Sun, 26 Jan 2025 12:31:09 GMT

~❯❯❯ curl -I localhost:8080/308
HTTP/1.1 308 Permanent Redirect
Content-Type: text/html; charset=utf-8
Location: /welcome
Date: Sun, 26 Jan 2025 12:31:11 GMT
```

So these are the responses you'll get. Simple, right?

That's it! Thanks.