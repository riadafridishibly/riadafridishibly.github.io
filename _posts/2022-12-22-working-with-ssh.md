---
published: true
comments: true
layout: post
title: Working with SSH in Linux and macOS
author: Riad Afridi Shibly
categories: programming, os-setup
tags: [ssh, ssh-agent, keys]
---

# Table of contents
{:.no_toc}
- 
{:toc}

_This post is not yet completed_

## TL;DR

### Generate keys

Default RSA key,

This will generate an RSA key called `id_rsa` in the `~/.ssh/` directory.
```sh
ssh-keygen
```

I mostly prefer using the `ed25519` key. To generate the key run the following,

```sh
ssh-keygen -t ed25519 -f ~/.ssh/id_{name_of_the_key}
```

### Start ssh-agent

```sh
eval $(ssh-agent)
```

### Add keys to ssh-agent

```sh
ssh-add </path/to/your/private/key>
```

Add all private (hopefully) keys to ssh-agent.

```sh
grep -rl PRIVATE ~/.ssh | xargs ssh-add
```

### Add keys to a remote server

```sh
ssh-copy-id -p 22 -i ~/.ssh/id_ed25519 user@example.com
```

### Remove keys from a remote server

What I mostly do, I open the `~/.ssh/authorized_keys` file in the server and delete the key from there. This file lists all the public keys uploaded to the server.

### Remove host from known_hosts file

```sh
ssh-keygen -R hostname
```

### Debug

It's useful to debug using the `ssh -v` (verbose) option. It'll tell you what's happening under the hood. The interesting line to look for is `Offering`.

## What is SSH?

SSH is a widely used and popular secure shell implementation that is mostly used to remotely access machines. To learn more about it search google or run `man ssh` in your terminal.

## Using SSH

To use ssh we'll need a remote machine with the OpenSSH server installed. Then from our computer, we can access that machine through the `ssh` client program. To demonstrate this I'll use two docker containers here one as a remote machine and another as our local machine.

## Our demo machine

Let's create a docker container that will act like our remote machine. First, start with a docker file. We'll build a container from this file.

```docker
FROM debian:bookworm-slim

RUN DEBIAN_FRONTEND=noninteractive \
  apt-get update && \
  apt-get install -y \
  openssh-server iproute2 openssl

RUN useradd -s /bin/bash -m \
  -p $(openssl passwd -1 password) riad

RUN echo "#!/bin/bash" >> /startup.sh && \
    echo "service ssh start" >> /startup.sh && \
    echo "echo 'IP:' \$(hostname -I)" >> /startup.sh && \
    echo "exec bash \"\$@\" " >> /startup.sh && \
    chmod +x /startup.sh

CMD ["/startup.sh"]
```

This is going to be a very simple Debian-based docker image. We'll install some packages as well.

To build the docker image let's create a new `Dockerfile` file and then paste the code above. Then run the following command from the `Dockerfile` directory,

Or you can directly paste this into a terminal and it'll create the file for you,

```wrap
echo 'RlJPTSBkZWJpYW46Ym9va3dvcm0tc2xpbQoKUlVOIERFQklBTl9GUk9OVEVORD1ub25pbnRlcmFjdGl2ZSBcCiAgYXB0LWdldCB1cGRhdGUgJiYgXAogIGFwdC1nZXQgaW5zdGFsbCAteSBcCiAgb3BlbnNzaC1zZXJ2ZXIgaXByb3V0ZTIgb3BlbnNzbAoKUlVOIHVzZXJhZGQgLXMgL2Jpbi9iYXNoIC1tIFwKICAtcCAkKG9wZW5zc2wgcGFzc3dkIC0xIHBhc3N3b3JkKSByaWFkCgpSVU4gZWNobyAiIyEvYmluL2Jhc2giID4+IC9zdGFydHVwLnNoICYmIFwKICAgIGVjaG8gInNlcnZpY2Ugc3NoIHN0YXJ0IiA+PiAvc3RhcnR1cC5zaCAmJiBcCiAgICBlY2hvICJlY2hvICdJUDonIFwkKGhvc3RuYW1lIC1JKSIgPj4gL3N0YXJ0dXAuc2ggJiYgXAogICAgZWNobyAiZXhlYyBiYXNoIFwiXCRAXCIgIiA+PiAvc3RhcnR1cC5zaCAmJiBcCiAgICBjaG1vZCAreCAvc3RhcnR1cC5zaAoKQ01EIFsiL3N0YXJ0dXAuc2giXQo=' | base64 -d > Dockerfile
```

Then let's run this command to build a new image based on the dockerfile.

```sh
docker build -t debian:ssh-test .
```

Now we want to do something different. We want to address the containers by host lookup. So we'll connect both containers to the same network. To create a new network let's run the following command.

```sh
docker network create --driver bridge ssh-net
```

This command will create a new network called `ssh-net`. 

The next thing we want to do is start two terminal sessions and open two docker container shells in two of them with the following commands. 

In one of the terminals let's run this command.

```sh
docker run --rm -it -h local --name local --network ssh-net debian:ssh-test
```

This terminal will act like our local machine.

On the other terminal let's run the following,

```sh
docker run --rm -it -h remote --name remote --network ssh-net debian:ssh-test
```

This one will act like our remote machine.

## Logging into the machine

To log in to the machine, all we need to do is using `ssh` command, we can run ssh command like this,

```sh
ssh username@remotemachine
```

In the demo case, we can run the following command from the local machine (notice the hostname in the terminal prompt `root@local`:/#`), 

```
ssh riad@remote
```

Then it'll ask for the password, write `password` (yee! very secure password!)

Congratulation! Now we're in the remote machine! The prompt will look like this!


```
riad@remote$ 
```

## Generating keys

Now that we've logged in with the password, let's try something cool. Let's try logging in with keys! The idea here is to generate a key pair. We'll keep the private key and send the public key to the server. Then we should be able to log in without any password. Let's try this.

Simply write `exit` and it'll get you out of the remote machine.

```sh
riad@remote$ exit
```

Now to generate keys, from our local container, let's run this command

```
root@local:/# ssh-keygen
```

It'll ask for a bunch of questions, press `ENTER` for all of them. Now run the following command,


## Upload our keys

Now that we've generated keys in our machine it's time to upload the public part to the server. There's already a helper utility available for that. Let's run the following.

```
ssh-copy-id riad@remote
```

This command will try to upload any key that is recognized by the `ssh-agent` to the server. It'll ask for the password. Let's type our password and the keys will be uploaded to the remote machine.

## Using password-less authentication

Now our public key is in the remote machine so if we again run `ssh riad@remote` then it'll log into the remote machine without any password.

## Disabling the password auth

We can now disable password authentication altogether. Sometimes it's recommended for the servers to make the server a bit more secure.

## SSH agent

We've mentioned `ssh-agent` before. But what is it? If we look into the man page, by simply running `man ssh-agent` then we will get this.

> ssh-agent is a program to hold private keys used for public key authentication.

Before going crazy with `ssh-agent` let's try to set up the environment.

## Removing keys from ssh-agent!

To prove a point let's do the following, we're deleting the authorized_keys from the server.

```
root@remote:/# rm /home/riad/.ssh/authorized_keys
```

Let's generate a new key,

```
root@local:/# ssh-keygen -f ~/.ssh/mykey -t ed25519
```

Upload the key to the remote server,

```
root@local:/# ssh-copy-id -i ~/.ssh/mykey riad@remote
```

And try logging into the server!

```
root@local:/# ssh riad@remote
```

It'll ask for the password!! _If not then the key probably is already added to the ssh-agent!_ To remove that run `ssh-add -D`, Now try again!

Okay! Let's try this command differently,

```
root@local:/# ssh -i ~/.ssh/mykey riad@remote
```

This time you should be able to log in. We've specified our key!

We can do this another way! This is where `ssh-agent` comes into the picture. 

## Adding keys to ssh-agent

Let's see if our `ssh-agent` is running! To check, let's try to print out the socket path of the `ssh-agent`,

```
root@local:/# echo $SSH_AUTH_SOCK
```

If it prints nothing then the `ssh-agent` is not running. To start `ssh-agent` let's run the following command,

```
root@local:/# eval $(ssh-agent)
```

## ssh verbose mode

Verbose mode helps to debug common ssh issues! Like everything seems set, still can't log in! Maybe `ssh-agent` doesn't know about the key! Maybe the ssh config is wrong. And so on.

Try running `ssh` command with `-v` flag.

```sh
ssh -v user@example.com
```

This command will spit out lots of logs and the most interesting one is the offering lines. 

## Using ssh config file

Sometimes it's handy to use the ssh config file. The ssh config file is located `~/.ssh/config`. The syntax is like this.

```sh
Host prod
    HostName prod01.example.com
    User riad
    Port 2398
```

Now with that config in place, previously we had to run,

```sh
ssh -p 2398 riad@prod01.example.com
```

Now we can simply write,

```sh
ssh prod
```

And the effect will be the same.

## Closing

This article is mainly written to document some ssh things. It's not yet fully complete! I'll gradually update this article.