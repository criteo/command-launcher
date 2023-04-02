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

setup auto completion. See help to get instructions:

```shell
cola completion --help
```

## login

Store your credentials securely and pass them to managed commands when requested and under your agreements. More details see: [Managed resources](../resources)

## update

Check updates for command launcher and managed commands.

## version

Return command launcher version information.

## package

A collection of commands to manage installed packages and commands

### package list

List installed packages and commands

```shell
# list local installed packages
cola package list --local

# list local installed packages and commands
cola package list --local --include-cmd

# list dropin packages
cola package list --dropin

# list local dropin packages and commands
cola package list --dropin --include-cmd

# list remote packages
cola package list --remote
```

### package install

Install a dropin package from a git repo or from a zip file

```shell
# install a dropin package from git repository
cola package install --git https://github.com/criteo/command-launcher-package-example

# install a dropin package from zip file
cola package install --file https://github.com/criteo/command-launcher/raw/main/examples/remote-repo/command-launcher-demo-1.0.0.pkg
```

### package delete

Remove a dropin package from the package name defined in manifest

```shell
cola package delete command-launcher-example-package
```

### package setup

Manually trigger the package [setup hook](../manifest/#__setup__)

```shell
cola package setup command-launcher-example-package
```

## remote

A collection of commands to manage extra remote registries. A registry is a URI that hosts multiple packages. The list of available packages of the registry is defined in its `/index.json` endpoint.

### remote list

List remote registries.

```shell
cola remote list
```

### remote add

Add a new remote registry. Command launcher will synchronize from this remote registry once added.

```shell
cola remote add myregistry https://raw.githubusercontent.com/criteo/command-launcher/main/examples/remote-repo
```

### remote delete

Delete a remote registry by its name.

```shell
cola delete myregistry
```

## rename

### rename

Rename a command into a different name

To avoid command conflicts, each command has a unique full name in form of `[name]@[group]@[package]@[repository]`, for group command and root level command, it's group is empty. For example: `hello@@my-package@dropin` is the full name of the command `hello` in `my-package` package, which can be found in the `dropin` repository.

Usually, such command is launched through: `cola [group] [name]`. You can rename the group and the name of the command to a different name, so that you can call it through: `cola [new group] [new name]`

To rename a command to a different name, use following commands:

```shell
# To change the group name:
cola rename [group]@@[package]@[repository] [new group]

# To change the command name:
cola rename [name]@[group]@[package]@[repository] [new name]
```

For example, you can rename the `hello` command to `bonjour` using following rename command:

```shell
cola rename hello@@my-package@dropin bonjour

# now call it from cola will trigger the original hello command
cola bonjour
```

### rename --delete

To delete a renamed command name

```shell
cola rename --delete [command full name]
```

Now you have to use its original name to call the command

### rename --list

> available in 1.10.0+

To list all renamed command

```shell
cola rename --list
```
