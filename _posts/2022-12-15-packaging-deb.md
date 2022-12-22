---
published: true
comments: true
layout: post
title: Packaging for Debian
author: Riad Afridi Shibly
categories: programming
tags: [linux, debian, packaging]
# image: /banner.jpg
---

# Table Of Contents
{:.no_toc}
- 
{:toc}


# Why?

When we want to install something in Debian based distro like Ubuntu, Debian or PopOS we mostly use a `.deb` file. In most cases, it's not ideal to just give anyone the binary (executable). `deb-file` is an `ar` archive (something like `zip` or `tar`). Packaging for `deb` is not that hard, but if you want to submit to the Debian group, then that's a different story. 

## What's inside a `.deb` file?

If the deb file is an archive then we should be able to see what's inside it, right? Sure we can. Let's unpack some deb files. We're mostly interested in the file structure. 

- Download discord?
- Download flameshot?

Let's run this command,

```sh
ar x </path/to/your/file>.deb
```


## What's a `.desktop` file?

## Packaging tools.