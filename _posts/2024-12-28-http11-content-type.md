---
published: true
comments: true
layout: post
title: "Understanding HTTP/1.1: Content-Type in Go (Part 2)"
description: "How the Content-Type header decides what a browser does with your response, plus a hand-rolled http.ResponseWriter implementation in Go."
categories: [programming]
tags: [http, tcp, networking, golang]
---


Previously we've seen how to construct basic HTTP request and response with our simple TCP server. In this blog post we'll go a bit deeper into the `Content-Type` header and see Content-Type in action. 

This is specifically HTTP thing. So when we're sending data to an HTTP client we can specify what type of data we're sending by using a header called `Content-Type`. When the client is browser even if we don't say the content type in header sometimes it'll guess from the bytes sent to it. All these we'll see in this blog post!

So rather than using `curl` as our client we'll be using a browser for the testing purposes. We'll send different types of responses to the client (browser) and let the browser render different content. 

## Before we start

In our previous post we manged to send `Hello, World!` to the client. This time nothing much different. We're going to send some HTML. But before that we're going to introduce a little bit of abstraction. We'll implement `http.ResponseWriter` interface. That way using the API will feel a little bit nicer.

First let's continue with the abstraction. What is the `http.ResponseWriter`? It's an interface which is used in `golang` http library. If you are familiar with handler function like `func handler(w http.ResponseWriter, r *http.Request)` you already know what this is! But you don't need to know about that right away.

The main reason we're implementing this interface is just to keep things consistent. Let's see a simple implementation of `http.ResponseWriter`. I'm not going to go deep with the interface implementation. But here are the methods we need to implement.

```go
type ResponseWriter interface {
    Header() http.Header
    Write([]byte) (int, error)
    WriteHeader(statusCode int)
}
```

Implementing this three method is not that difficult. With my good friend ChatGPT and modifying the code a bit I managed to get this, which will work just fine for our use cases.


```go
func NewResponseWriter(conn net.Conn) http.ResponseWriter {
	return &response{
		conn:    conn,
		headers: make(http.Header),
		status:  http.StatusOK,
	}
}

type response struct {
	conn        net.Conn
	headers     http.Header
	status      int
	wroteHeader bool
}

// interface guard
var _ http.ResponseWriter = (*response)(nil)

func (res *response) Header() http.Header {
	return res.headers
}

func (res *response) Write(data []byte) (int, error) {
	if !res.wroteHeader {
		res.WriteHeader(http.StatusOK)
	}
	return res.conn.Write(data)
}

func (res *response) WriteHeader(statusCode int) {
	if res.wroteHeader {
		return
	}
	res.wroteHeader = true
	res.status = statusCode
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n",
		res.status, http.StatusText(res.status))
	res.conn.Write([]byte(statusLine))
	for key, values := range res.headers {
		for _, value := range values {
			headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
			res.conn.Write([]byte(headerLine))
		}
	}
	res.conn.Write([]byte("\r\n"))
}
```

First we'll mimic the previous hello world example. Now the code for `handleConn` will look like this,

```go
func handleConn(c net.Conn) {
	defer c.Close()
	br := bufio.NewReader(c)
	req, err := http.ReadRequest(br)
	if err != nil {
		log.Fatal("Failed to read request: ", err)
	}

	log.Printf("Received HTTP request, method=%s path=%s\n",
		req.Method, req.URL.Path)

	w := NewResponseWriter(c)
	w.Write([]byte("Hello, World!\n"))
}
```

Now if we have the previous setup ready, visiting [https://localhost:2000](https://localhost:2000) will show `Hello, World!` on the screen.

Now we are ready to experiment with the returned data. We need to open browser dev tool and look into network calls. _Quick google search will show you how to use browser dev tool. In this post I'll write the headers as plain text._ 


## Sending HTML response

Sending text data to the browser is nice and simple. But we can go a little bit further. Browsers are very good at rendering HTML. So we'll send some html to the browser. And after that, we'll send some JS and CSS along the way.

Browser is very forgiving. So If we send any html tag it'll try to render that as well. We don't need to send the whole HTML document with html tag. Even browsers don't require us to send closing tags. But we'll not go that far.

Let's update our `handleConn` function to send an html response. 

```go
func handleConn(c net.Conn) {
	defer c.Close()
	br := bufio.NewReader(c)
	req, err := http.ReadRequest(br)
	if err != nil {
		log.Fatal("Failed to read request: ", err)
	}

	log.Printf("Received HTTP request, method=%s path=%s\n",
		req.Method, req.URL.Path)

	w := NewResponseWriter(c)
    // formatting has no real value to the browser
	w.Write([]byte(`
		<h1>Example Header</h1>
		<hr>
		<p>Sample paragraph</p>
	`))
}
```

Now if we visit [https://localhost:2000](https://localhost:2000), the page will look like this. 

![Browser rendering a simple HTML response](/assets/img/2024-12-28-http11-content-type/simple-html-response.png)

The question is how does browser know we're sending HTML? Why didn't browser render the text just like this?

```
<h1>Example Header</h1>
<hr>
<p>Sample paragraph</p>
```

Well, we need to get ourselves introduced to [MIME sniffing](https://developer.mozilla.org/en-US/docs/Web/HTTP/MIME_types#mime_sniffing). 
> In the absence of a MIME type, or in certain cases where browsers believe they are incorrect, browsers may perform MIME sniffing — guessing the correct MIME type by looking at the bytes of the resource.

But wait! How do we say to the browser treat this data as something else? Or, how do we say to the browser not to guess the type at all? Well we're in luck! We have two headers specifically for this use case. With `Content-Type` header we can tell the browser the data type and with `X-Content-Type-Options` header we can tell the browser not to guess the content type.

Let's set the content type header to something else. What about `plain text`?

```diff
	w := NewResponseWriter(c)
+	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(`
		<h1>Example Header</h1>
		<hr>
		<p>Sample paragraph</p>
	`))
```

Now if we reload the page in browser, it'll show as plain text, rather than the rendered html!

And if we inspect the response headers in browser dev tool under network tab, we'll see the content type is set to text/plain.

```
Content-Type: text/plain
```

Let's use the 2nd header that we previously talked about. Let's tell the browser not to guess the type.

```diff
-	w.Header().Set("Content-Type", "text/plain")
+	w.Header().Set("X-Content-Type-Options", "nosniff")
```

And the browser won't try to sniff the content to get the mime type! It'll simply be rendered as plain text!

Let's play a little bit more! What if we set `Content-Type` to `application/octet-stream`? Setting this content type basically means we're telling the browser we're sending arbitrary data that browser shouldn't try to interpret. Rather it'll pop open `Save As` dialog to let us download the file and do whatever we prefer to do with the data. If we want something to be downloadable we can set the content type from the server.

[Read about all the mime types here](https://developer.mozilla.org/en-US/docs/Web/HTTP/MIME_types)

Let's finalize with a simple web page! We'll try to render this html with our server. 

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>HTML Page</title>
    <link rel="stylesheet" href="style.css" />
    <script src="index.js"></script>
  </head>
  <body>
    <h1>Example header</h1>
    <hr />
    <p>Example Content</p>
  </body>
</html>
```

We've referenced two other files from this html, after browser fetch this html then it'll request for the `style.css` and `index.js` file. We need to modify our `handleConn` function to be able to send three different files. Let's handle them.

```go
func handleConn(c net.Conn) {
	defer c.Close()

	br := bufio.NewReader(c)
	req, err := http.ReadRequest(br)
	if err != nil {
		log.Fatalln("Failed to read request", err)
	}

	log.Printf("Received HTTP request, method=%s path=%s\n",
		req.Method, req.URL.Path)

	w := NewResponseWriter(c)

	switch req.URL.Path {
	case "/index.js":
		sendJS(w)
	case "/style.css":
		sendCSS(w)
	case "/": // send html
		sendHTML(w)
	default:
		w.WriteHeader(404)
	}
}
```

Here we've handled requests for three paths, 

- http://localhost:2000/
- http://localhost:2000/index.js
- http://localhost:2000/style.css

If anything else is requested we're sending `404` which basically means Not Found!

The three functions used here look like this, 

```go
func sendHTML(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	html := `
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>HTML Page</title>
    <link rel="stylesheet" href="style.css" />
    <script src="index.js"></script>
  </head>
  <body>
    <h1>Example header</h1>
    <hr />
    <p>Example Content</p>
  </body>
</html>`
	w.Write([]byte(html))
}

func sendCSS(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/css")
	css := `
h1 {
	color: red;
}
p {
	text-decoration: underline;
}`
	w.Write([]byte(css))
}

func sendJS(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/javascript")
	fmt.Fprint(w, `window.alert("Javascript Executed! Hi!")`)
}
```

Now if we rerun the server it should show an alert with a message saying, 'Javascript Executed! Hi!' and if we press OK, it should show us an HTML page with Red header and underlined paragraph. Now try changing the content type of different files and check what happens. What if instead of sending `text/css` for `CSS`, we send `text/plain`. Try it! Browser won't bother to apply the stylesheet!

There are other MIME types and different use cases for those. For example `application/json` is mostly used for APIs to tell the other clients that the data type is JSON. And we haven't touched the image mime types yet. Now it's a good time to visit MDN and check the other mime types.

We haven't talked about request headers! What if we submit a form? What is the multipart thing mentioned in MIME type docs? We'll see those in action in near future! For now, bye!
