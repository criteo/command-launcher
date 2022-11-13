---
title: "Built-in commands"
description: "Command launcher built-in commands"
lead: "Command launcher built-in commands"
date: 2022-10-02T18:41:12+02:00
lastmod: 2022-10-02T18:41:12+02:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "built-in-commands-5a823aed27c15b41cdaff2e32e58944e"
weight: 220
toc: true
---

## config

Get or set command launcher configuration.

Use `cola config` to list all configurations.

Use `cola config [key]` to get one configuration.

Use `cola config [key] [value]` to set one configuration.

## completion

Setup auto completion. See help to get instructions:

```shell
cola completion --help
```

## login

Store your credentials securely and pass them to managed commands when requested and under your agreements. More details see: [Managed resources](../resources)

## update

Check updates for command launcher and managed commands.

## version

Return command launcher version information.

## list

List installed packages and commands

```shell
# list local installed packages
cola list --local

# list local installed packages and commands
cola list --local --include-cmd

# list dropin packages
cola list --dropin

# list local dropin packages and commands
cola list --dropin --include-cmd

# list remote packages
cola list --remote
```

## install

Install a dropin package from a git repo or from a zip file

```shell
# install a dropin package from git repository
cola install --git https://github.com/criteo/command-launcher-package-example

# install a dropin package from zip file
cola install --file https://github.com/criteo/command-launcher/raw/main/examples/remote-repo/command-launcher-demo-1.0.0.pkg
```

## delete

Remove a dropin package from the package name defined in manifest

```shell
cola delete command-launcher-example-package
```

