# bloggy
Custom static site generator. 

It uses the combination of yaml and markdown to create the post.

It is currently used by: 
* [https://lateralusd.github.io](https://lateralusd.github.io)
* [https://6en6ar.github.io](https://6en6ar.github.io)

# Installation

```bash
$ git clone https://github.com/lateralusd/bloggy.git
$ cd bloggy
$ go build
```

# Quick setup

Generate the new config file using `bloggy cfg`, you can optionally pass the output filename, otherwise cfg.yaml will be used.

Then edit the newly generated config to meet your needs.

Run `bloggy post NAME OF THE POST` to create new post.

```bash
$ ./bloggy post new post
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

To create the new page which will be shown in the navbar, use `bloggy page NAME OF THE PAGE`.

After you are satisfied with your post content, just run `$ ./bloggy gen` to generate the webpages and then open index.html

# Sample run

```bash
$ bloggy cfg             
New config "cfg.yaml" created
$ cat cfg.yaml
url: https://username.github.io/
title: sample blog
twitter: https://twitter.com/user
github: https://github.com/user
mail: someone@something.com
author: Haxor
about: About page section
outdir: public
$ bloggy new test post
New post posts/test_post.md created
$ cat posts/new_post.md 
---
title: Test post
description: This is short description
date: 2022-02-03T23:09:11.961103+01:00
---

# Introduction

Here comes the content.

## How to

Just write markdown and when you want to reference the image, place the image inside the static/images and reference it in url ../images/nameoftheimage.png.

![Image](../images/sample.png)
$ bloggy gen
Generated 2 posts