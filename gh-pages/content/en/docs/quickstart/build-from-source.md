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

Command launcher is designed for both enterprise and individual usage. According to your context, you might want to call it differently. For example, at Criteo, we call it "Criteo Dev Toolkit". The binary name is used for several default configurations, for example, command launcher home `$HOME/.[APP_NAME]`, additional resources environment prefix `[APP_NAME]_`, etc.

The default pre-built binary is call `cola` (**Co**mmand **La**uncher), which means that the default home folder is `$HOME/.cola` and the resources environment variables are all starts with `COLA_`.

Another pre-built binary is called `cdt` (Criteo Dev Toolkit), its home folder will be `$HOME/.cdt`, and its commands can access the resource environment variables with both prefix `COLA_` and `CDT_`.

> For compatibility concern, we highly recommend to reference resources in your command with prefix `COLA_`

The easiest way to use a different name is to simply copy or rename the pre-built binary. The app name is derived from the binary's file name at startup. You can then set the long display name via config:

```shell
cp cola myapp
myapp config app_long_name "My App"
```

Symlinks are treated as aliases (they resolve to the original binary name), while copies create a separate instance with its own config directory (`~/.myapp/`).

You can also set the name at build time if you prefer.

## Build from source

Requirements: golang >= 1.17

You can build the command launcher with your prefered name (in the example: `Command Launcher`, a.k.a `cola`).

```shell
go build -o cola -ldflags='-X main.version=dev -X main.appName=cola -X "main.appLongName=Command Launcher"' main.go
```

Or simply call the `build.sh` scripts

```shell
./build.sh [version] [app name] [app long name]
```

## Run tests

Run unit tests

```shell
go test -v ./...
```

Run all integration tests

```shell
./test/integration.sh
```

You can run one integration test by specify the name of the integration test file (without the .sh extension). The integration tests can be found in [`test/integration`](https://github.com/criteo/command-launcher/tree/main/test/integration) folder, for example:

```shell
./test/integration.sh test-remote
```
