---
published: false
comments: true
layout: post
title: Working with `[]byte` slice in golang
categories: programming
tags: [byte, golang]
image: 2022-07-17-working-with-bytes-in-go/banner.jpg
---

# Table Of Contents
{:.no_toc}
- 
{:toc}

# Topics

- String (utf-8) in golang
- `[]byte`
- `io.Reader` and `io.Writer` interface
- Different types of reader
- Representations
    - Hex
    - Base64
- Multi Reader
- Multi Writer
- Tee Reader
- `bufio.Reader` implements all of them
- Working with buffered data, returns with `MultiReader`
