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

Generate the new config file using `bloggy cfg`, you can optionally pass the output filename, otherwise cfg.yaml will be used.

Then edit the newly generated config to meet your needs.

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
