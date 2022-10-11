---
title: "Quick Start"
description: "Integrate your first command to command launcher"
lead: "In this tutorial, we will integrate our first command to command launcher"
date: 2022-10-04T17:01:29+02:00
lastmod: 2022-10-04T17:01:29+02:00
draft: false
images: []
menu:
  docs:
    parent: "quickstart"
    identifier: "first-command-884b250e24cf7581b3266bc090a7cfd1"
weight: 130
toc: true
---

## Preparation

If you want to use the default command launcher binary name (`cdt`): install command launcher following the guide [Binary install](../binary-install). Or if you prefer a different binary name, build command launcher from source: [Build from source](../build-from-source)

> Understand why binary name matters: [why command launcher binary name matters?](../build-from-source/#why-does-the-binary-name-matter)

In this tutorial, we will use the binary name `cdt` (which stands for `Criteo Dev Toolkits`).

To check your command launcher installation, run following command:

```bash
cdt config
```

It should list all your configurations. The important ones for this tutorial is the following:

```json
dropin_folder                    : [command launcher home]/dropins
local_command_repository_dirname : [command launcher home]/current
```

Default command launcher home:
{{< details "MacOS" >}}
_/Users/[user_name]/[.binary_name]_
{{< /details >}}
{{< details "Linux" >}}
_/home/[user_name]/[.binary_name]_
{{< /details >}}
{{< details "Windows" >}}
_C:\Users\\[user_name]\\[.binary_name]_
{{< /details >}}

For example, on my Macbook (with user `criteo`), my command launcher home is `/Users/criteo/.cdt`

Command launcher also provide auto-completion feature to all managed commands, you only need to setup auto-completion once: [setup auto-completion](../binary-install/#setup-auto-completion)

## Let's build a command

If you already have a command you can skip this step. Command launcher is technology agnostic to your command. You can build your command in any tech stack that suits your use case. In this tutorial, let's build a simple "Hello World" command line tool with bash scripts.

In command launcher, commands are packaged into `package`s. Let's first create a package for our newly created command:

```bash
cd $HOME/.cdt/dropins
mkdir my-first-package
cd my-first-package
```

Create a script file `first-command-launcher-cmd.sh`, which prints greeting message according to the language

```bash
#!/bin/bash

LANG=${LANG:-en}

if [ $LANG == "fr" ]; then
  echo "Bonjour! $1"
  exit 0
else
  echo "Hello! $1"
  exit 0
fi
```

You can run it directly as a normal bash script:

```bash
$ ./first-command-launcher-cmd.sh "command launcher"
Hello! command launcher

$ LANG=fr ./first-command-launcher-cmd.sh "command launcher"
Bonjour! command launcher
```

## Prepare a minimal manifest.mf

Now we have a working command/script. To make command launcher aware of it, we need to create a manifest file (in JSON or YAML format) at the root folder of the package:

_manifest.mf_

```yaml
pkgName: my-first-package
version: 0.0.1
cmds:
  - name: greeting
    type: executable
    executable: "{{.PackageDir}}/first-command-launcher-cmd.sh"
```

That's it! Your command has been integrated to command launcher with a subcommand named `greeting`, to test it:

```bash
$ cdt greeting "command launcher"
Hello! command launcher

$ LANG=fr cdt greeting "command launcher"
Bonjour! command launcher
```

Command launcher will pass all the environment variables, arguments to itself to your command.

## Tell more about your command to command launcher

We can go even further to turn our bash scripts into a native-like program. Let's add extra information in the manifest, and make some improvements of its user interface:

- the short and long description
- some examples
- use a flag to take language input instead of environment variable `LANG`

```yaml
pkgName: my-first-package
version: 0.0.2
cmds:
  name: greeting
  type: executable
  short: Simple greeting command
  executable: "{{.PackageDir}}/first-command-launcher-cmd.sh"
  requiredFlags:
    - "language\t greeting language"
  examples:
    - scenario: Greeting with default language
      cmd: greeting [name]
    - scenario: Specify the greeting language
      cmd: greeting --language fr [name]
  checkFlags: true
```

The above manifest tells command launcher that the `greeting` command requires a flags called `language`, and let command launcher to check the flags before calling the script

Now when you run the greeting command with `-h` or `--help`, you will get a nice help message like a native command:

```shell
$ cdt greeting -h
Usage:
  cdt greeting [flags]

Examples:
  # Greeting with default language
  greeting [name]

  # Specify the greeting language
  greeting --language fr [name]


Flags:
  -h, --help              help for greeting
      --language string   greeting language
```

But how can our script get the language passed from the `--language` flag? Command launcher passes an environment variable to your script with the name `CDT_FLAG_LANGUAGE` (More details see [checkFlags](../../overview/manifest/#checkflags)). You can modify your script to get the language from it like so:

```bash
#!/bin/bash

LANG=${CDT_FLAG_LANGUAGE:-en}

if [ $LANG == "fr" ]; then
  echo "Bonjour! $1"
  exit 0
else
  echo "Hello! $1"
  exit 0
fi
```

Now we have turned our bash script into a native-like command.

## Auto-completion

Command launcher will automatically enable auto-completion for your command. Subcommand, flags, arguments will will be auto-completed when you type `[TAB][TAB]`, for example:

```shell
$ cdt g[TAB][TAB]
greeting - Simple greeting command
grepx    - Enhanced grep command

$ cdt greeting --[TAB][TAB]
$ cdt greeting --language
```
