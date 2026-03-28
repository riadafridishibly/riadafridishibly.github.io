---
published: true
comments: true
layout: post
title: Migrating from GitLab CE to Forgejo
author: Riad Afridi Shibly
categories: programming
tags: [git, forgejo, devops, self-hosted]
---

## Table Of Contents

1. Table Of Contents
{:toc}

Recently we migrated our internal git server from GitLab CE to Forgejo. The experience has been great, IMHO.

We were running GitLab for a very long time. It served us well for so many years. Recently I updated it to the latest version. That's when the problems began. It started failing randomly with lots of network errors. They updated the UI and probably introduced a lot of bloat, but the performance degradation was unmanageable — though to be fair, we were running that on a small server with only 3GB of RAM.

Then I started exploring options. One solid option was Forgejo. It's a fork of Gitea, which is a fork of Gogs. Gogs started as a lightweight git server written in Go. So I thought about it and ended up choosing Forgejo, because why not. Also [codeberg.org](http://codeberg.org/) uses Forgejo with some modifications. Codeberg is an alternative to GitHub hosted in the EU. Recently, the Zig programming language moved off of GitHub and their main development is currently happening on Codeberg.

## Migration

I created a new VM, installed Debian (no thanks to Ubuntu), installed the latest Forgejo, and pointed it to a new domain. Forgejo supports migration from a lot of services. All you need is a link and an API key. I tried that with one of my repos and it worked great.

But there was a problem. Well couple of problems. We have 300+ repos and ff I didn't complete the migration right then, there would be changes on both the old server and the new server. So I pulled an all-nighter and completed it. I could have mirrored them, but then it wouldn't duplicate the issues and PRs. I can't just manually migrate all of those. So I started writing a script — mainly I told Claude to do that. I gave it the OpenAPI spec and it did the rest.

## After Migration

After migration, the performance improvement is visible. The web UI is very snappy (also ugly). SSH performance is also very good.

With GitLab I used `glab` CLI, but for Forgejo the existing options didn't quite work for me. However, there's the [Forgejo SDK](https://codeberg.org/mvdkleijn/forgejo-sdk). Using Claude, I built a GitHub-style CLI for Forgejo based on the SDK — and it's working really well. You can give it a try here: [github.com/riadafridishibly/fj](https://github.com/riadafridishibly/fj).

## Final Thoughts

Now with only 1.5GB of RAM, Forgejo is thriving — super fast performance, snappy UI, and we have all the required features.

Though I haven't explored Forgejo Actions yet. I might, if I can migrate to a beefier server.
