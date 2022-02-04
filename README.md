# bloggy
Custom static site generator. 

It uses the combination of yaml and markdown to create the post.

It is currently used by: 
* [https://lateralusd.github.io](https://lateralusd.github.io)

# Installation

```bash
$ git clone https://github.com/lateralusd/bloggy.git
$ cd bloggy
$ go build
```

# Quick setup

First edit `cfg.yaml` to meet your needs.

```yaml
title: lateralusd
twitter: https://www.twitter.com/yourUsername
github: https://www.github.com/yourUsername
author: lateralusd
outdir: public
about: |
      Here comes the about section.
```

Run `$ ./bloggy new nameOfThePost` to create new post.

```bash
$ ./bloggy new new post
New post posts/new_post.md created
$ cat posts/new_post.md
---
title: Test post
description: This is short description
date: 2022-02-03T21:11:45.776829+01:00
---

# Introduction

Here comes the content.
```

After you are satisfied with your post content, just run `$ ./bloggy gen` to generate the webpages and then open index.html
