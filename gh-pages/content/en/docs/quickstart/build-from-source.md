---
title: "Build from source"
description: "Build command launcher from source"
lead: "Build command launcher from source"
date: 2022-10-02T17:34:04+02:00
lastmod: 2022-10-02T17:34:04+02:00
draft: false
images: []
menu:
  docs:
    parent: "quickstart"
    identifier: "build-from-source-818500deb2b1cbde4714cb2bda54ecaa"
weight: 120
toc: true
---

## Why does the binary name matter?

Command launcher is designed for both enterprise and individual usage. According to your context, you might want to call it differently. For example, at Criteo, we call it "Criteo Dev Toolkit". The binary name is used for several default configurations, for example, command launcher home `$HOME/.[APP_NAME]`, resources environment prefix `[APP_NAME]_`, etc.

The pre-built binary is call `cdt (Criteo Dev Toolkit)`, which means that the default home folder is `$HOME/.cdt` and the resources environment variables are all starts with `CDT_`.

To use a different name, you need to build command launcher from source and pass the desired short and long name to the build scripts.

## Build from source

Requirements: golang >= 1.17

You can build the command launcher with your prefered name (in the example: `Criteo Developer Toolkit`, a.k.a `cdt`).
```
go build -o cdt -ldflags='-X main.version=dev -X main.appName=cdt -X "main.appLongName=Criteo Dev Toolkit"' main.go
```

Or simply call the `build.sh` scripts
```
./build.sh [version] [app name] [app long name]
```

## Run tests

```
go test -v ./...
```
