---
title: "Dropin package"
description: "The easiest way to integrate your own scripts/tools to command launcher"
lead: "The easiest way to integrate your own scripts/tools to command launcher"
date: 2022-10-02T19:29:47+02:00
lastmod: 2022-10-02T19:29:47+02:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "dropin-affc160f6b7478c1050512a4037f5ca2"
weight: 245
toc: true
---

## What is dropin package? and why do I need it?

A _dropin_ package is a package that is only available on the local machine, rather than managed by the remote repository. It allows developers to integrate their own scripts/tools into Command Launcher to benefit from the features provided by the command launcher, such as auto-complete, monitoring, etc.

For example, you probably already have lots of shell scripts to maintain your infrastructure. Writing auto-completion for all these scripts is time consuming and it is difficult to remember which script does what, and what parameters they accept. Writing a `manifest.mf` file is enough to let Command Launcher manage these scripts for you with auto-complete, secret management, and monitoring.

A dropin package is also a good way for you to develop and test your Command Launcher package, as it follows the same structure as a regular package.

## How to create a dropin package?

1. identify the _dropins folder_: run the following command:

    ```shell
    cola config dropin_folder
    ```

    If the dropin folder returned by the command doesn't exist, create it.
1. create a package folder in the dropin folder, let's say, a package named `my-first-package`. You can name it whatever you want.
1. add a `manifest.mf` file in the newly created package folder, follow the [MANIFEST](../manifest) guide to define your commands in this file. Note: you can copy your scripts in the package folder and use `{{.PackageDir}}` to reference the package location in your manifest file.
1. run `cola` any time to test your command

## How to share a dropin package with others?

A dropin package is simply a directory with `manifest.mf` file in it; the best way to share a dropin package is to push it to a git repository and ask others to clone it in their own dropin folder.

Starting from 1.7.0, you can use the built-in `install` command to install a dropin package hosted in a git repository or a zip file:

```shell
cola package install --git https://github.com/criteo/command-launcher-package-example
```

If you uploaded your package to an HTTP server as a zip file, you can install it with `cola install --file`

```shell
cola package install --file https://github.com/criteo/command-launcher/raw/main/examples/remote-repo/command-launcher-demo-1.0.0.pkg
```

## How to update a dropin package?

Command Launcher does not currently update the dropin folder automatically, it is up to developers themselves to keep these dropin packages up-to-date.
