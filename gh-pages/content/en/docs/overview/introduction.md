---
title: "Introduction"
description: "Overall introduction of command launcher"
lead: "Overall introduction of command launcher"
date: 2022-10-02T17:15:37+02:00
lastmod: 2022-10-02T17:15:37+02:00
draft: false
images: []
menu:
  docs:
    parent: "Overview"
    identifier: "intro-7bb54696d0a61c0b18319f6b3e32f884"
weight: 210
toc: true
mermaid: true
---

## What is command launcher?

Command launcher is a small-footprint, feature-rich CLI management tool for both enterprise and individual CLI developers. It eases the command line tool development by providing built-in common functionalities like: monitoring, progressive rollout, auto-completion, credential management, and more to your commands.

## Why a command launcher?

At Criteo, we have many teams who provide command line applications for developers. These CLI providers repeatly handle the same features and functionalities for their CLI apps, such as auto-completion, credential management, release, delivery, monitoring, etc.

On the developer side, they have to manually download these tools and keep them up-to-date. It is also difficult for them to discover available new tools. On the other hand, different developers have developed lots of similar handy scripts/tools by themselves without an easy way to share with others to avoid "re-inventing" the wheel.

To improve both developer and CLI provider's experience, we developed a command launcher to solve the above issues. It has built-in features like auto-completion, credential management, progressive roll-out, and monitoring, so that the CLI app provider can focus on the functionality of their CLI app. Developers only need to download the command launcher to access all these CLI apps. The command launcher will keep their CLI application up-to-date. The dropin feature allows developers to integrate their own scripts/tools into command launcher and share with others. These scripts and tools can also benefit from built-in features like auto-completion and monitoring.

## How it works?

Command Launcher is a small binary downloaded by a developer in their development environment. *CLI providers* package new commands or new versions of commands into a package, upload them to a remote repository, and update the package index of the repository. This process can be automated. For more details about the remote repository, see [CLI Provider Guide](../provider-guide)

Developers can integrate their own commands into Command Launcher as a "dropin" package. These dropin packages will be only accessible on the developer's machine. To share such commands see [Dropin Package](../dropin)

Developers run Command Launcher to access these commands. For example, if you have a command called `toto`, instead of running it directly from the command line, you use `cl toto`, where `cl` is the binary name of the Command Launcher (you can name it anything that suits you).

Every time you run the Command Launcher, it will synchronize with the remote command, and propose updates if they are available.

```text

                           ┌──────────────────┐    Synch    ┌───────────────────────────┐
            ┌──────────────│ Command Launcher │◄────────────│ Remote Command Repository │
            │              └──────────────────┘             └───────────────────────────┘
            │                       │                                      │
            │            ┌──────────┼──────────┐              ┌────────────┼────────────┐
            ▼            ▼          ▼          ▼              ▼            ▼            ▼
       ┌─────────┐   ┌───────┐  ┌───────┐  ┌───────┐     ┌─────────┐  ┌─────────┐  ┌─────────┐
       │ dropins │   │ cmd A │  │ cmd B │  │ cmd C │     │  cmd A  │  │  cmd B  │  │  cmd C  │
       └────┬────┘   └───────┘  └───────┘  └───────┘     └─────────┘  └─────────┘  └─────────┘
     ┌──────┴──────┐
     ▼             ▼
 ┌────────┐   ┌────────┐
 │  cmd D │   │ cmd E  │
 └────────┘   └────────┘
```

## Features

- **Small footprint**. Command Launcher is around 10M, with no dependency to your OS.
- **Technology agnostic**. It can launch commands implemented in any technology, and integrate them with a simple manifest file.
- **Auto-completion**. It will auto-complete the name of your commands out of the box, and their arguments if you provide the right information in the manifest file.
- **Auto-update**. Command Launcher can keep itself and its command packages up to date.
- **Credential management**. With the built-in `login` command, it securely passes user credentials to your commands.
- **Progressive rollout**. Target a new version of command to a group of beta test users, and rollout progressively to all your users.
- **Monitoring**. Built-in monitoring feature to monitor the usage of your commands.
- **Dropins**. Easy to integrate your own command line scripts/tools by dropping your manifest in the "dropins" folder.

## Installation

A pre-built binary can be downloaded from the release page. Unzip it, and place the binary in a directory in your `PATH`.

The two pre-built binaries are named `cola` (**Co**mmand **La**uncher) and `cdt` (**C**riteo **D**ev **T**oolkit), if you want to use a different name, you can pass your preferred name in the build. See the *build* section below.

## Building

Requirements: golang >= 1.17

You can build the command launcher with your preferred name (in the example: `Criteo Developer Toolkit`, a.k.a `cdt`).

```shell
go build -o cdt -ldflags='-X main.version=dev -X main.appName=cdt -X "main.appLongName=Criteo Dev Toolkit"' main.go
```

Or simply call the `build.sh` script

```shell
./build.sh [version] [app name] [app long name]
```

## Running tests

Run unit tests:

```shell
go test -v ./...
```

Run all integration tests:

```shell
./test/integration.sh
```

You can run one integration test by specify the name of the integration test file (without the .sh extension). The integration tests can be found in [`test/integration`](https://github.com/criteo/command-launcher/tree/main/test/integration) folder, for example:

```shell
./test/integration.sh test-remote
```

## Releasing

Simply tag a commit with format 'x.y.z', and push it.

```shell
git tag x.y.z
git push origin x.y.z
```

The supported release tag format:

- \*.\*.\*
- \*.\*.\*-\*

Examples: `1.0.0`, `1.0.1-preview`
