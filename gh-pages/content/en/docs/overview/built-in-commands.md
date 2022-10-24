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
