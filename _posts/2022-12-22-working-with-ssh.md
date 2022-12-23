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

## TL;DR

### Generate keys

### Start ssh-agent

```sh
eval $(ssh-agent)
```

### Add keys to ssh-agent

```sh
grep -rl PRIVATE ~/.ssh | xargs ssh-add
```

### Add keys to a remote server

### Remove keys from a remote server

### Keyscan

### Debug

## What is SSH?

SSH is a widely used and popular secure shell implementation that is mostly used to remotely access machines. To learn more about it search google or run `man ssh` in your terminal.

## Using SSH

To use ssh we'll need a remote machine (we'll run one in a minute) with `openssh` server installed. Then from our computer, we can access that machine through `ssh` client program. 

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

To build the docker image create a new `Dockerfile` file and then paste the code above. Then run the following command from the `Dockerfile` directory,

Or you can directly paste this in a terminal and it'll create the file for you,

```wrap
echo 'RlJPTSBkZWJpYW46Ym9va3dvcm0tc2xpbQoKUlVOIERFQklBTl9GUk9OVEVORD1ub25pbnRlcmFjdGl2ZSBcCiAgYXB0LWdldCB1cGRhdGUgJiYgXAogIGFwdC1nZXQgaW5zdGFsbCAteSBcCiAgb3BlbnNzaC1zZXJ2ZXIgaXByb3V0ZTIgb3BlbnNzbAoKUlVOIHVzZXJhZGQgLXMgL2Jpbi9iYXNoIC1tIFwKICAtcCAkKG9wZW5zc2wgcGFzc3dkIC0xIHBhc3N3b3JkKSByaWFkCgpSVU4gZWNobyAiIyEvYmluL2Jhc2giID4+IC9zdGFydHVwLnNoICYmIFwKICAgIGVjaG8gInNlcnZpY2Ugc3NoIHN0YXJ0IiA+PiAvc3RhcnR1cC5zaCAmJiBcCiAgICBlY2hvICJlY2hvICdJUDonIFwkKGhvc3RuYW1lIC1JKSIgPj4gL3N0YXJ0dXAuc2ggJiYgXAogICAgZWNobyAiZXhlYyBiYXNoIFwiXCRAXCIgIiA+PiAvc3RhcnR1cC5zaCAmJiBcCiAgICBjaG1vZCAreCAvc3RhcnR1cC5zaAoKQ01EIFsiL3N0YXJ0dXAuc2giXQo=' | base64 -d > Dockerfile
```

```sh
docker build -t debian:ssh-test .
```

To run the image run with,

```sh
docker run --rm -it -h machine debian:ssh-test
```

It'll run a docker container and open a shell for you. But we want to access it through ssh.

## Logging into the machine

To login into the machine, you need to start another terminal (maybe your terminal tab or a new window) and then write the following,

```sh
docker run --rm -it -h local debian:ssh-test
```

Then enter this,

```sh
ssh riad@<the-printed-ip>
```

For example,

```sh
ssh riad@172.17.0.2
```

Then it'll ask for the password, you can simply write `password` (that's what is set in the container).

You'll be presented with a prompt like this,

```
riad@remote$ 
```

Congratulations! You've Successfully logged in!

## Generating keys

Now that we've logged in with the password, let's try something cool. Let's try logging in with keys! The idea here is to generate a key pair. We'll keep the private key and send the public key to the server. Then we should be able to log in without any password. Let's try this.

From our local container, let's run this command

```
ssh-keygen
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

Before going crazy with `ssh-agent` let's try to setup the environment.

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

It'll ask for password!! _If not then the key probably is already added to ssh-agent!_ To remove that run `ssh-add -D`, Now try again!

Okay! Let's try this command differently,

```
root@local:/# ssh -i ~/.ssh/mykey riad@remote
```

This time you should be able to log in. We've specified our key!

We can do this another way! This is where `ssh-agent` comes into the picture. 

## Adding keys to ssh-agent

Let's see if our `ssh-agent` is running! To check, let's try to print out the socket path of `ssh-agent`,

```
root@local:/# echo $SSH_AUTH_SOCK
```

If it prints nothing then the `ssh-agent` is not running. To start `ssh-agent` let's run the following command,

```
root@local:/# eval $(ssh-agent)
```

## ssh verbose mode

## Inspecting the offerings

## Using ssh config file

## Closing