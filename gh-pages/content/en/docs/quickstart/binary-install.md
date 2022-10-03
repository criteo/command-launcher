---
title: "Binary install"
description: "Install command launcher with pre-built binaries"
lead: "Install command launcher with pre-built binaries"
date: 2022-10-02T17:25:20+02:00
lastmod: 2022-10-02T17:25:20+02:00
draft: false
images: []
menu:
  docs:
    parent: "quickstart"
    identifier: "installation-30e697cb4baa85f9bed185936eb70fff"
weight: 110
toc: true
---

## Download pre-built binaries

Pre-built binaries can be downloaded from the [Github release page](https://github.com/criteo/command-launcher/releases/latest). Copy the binary into your PATH.

The pre-built binary is named `cdt` (Criteo Dev Toolkit), if you want to use a different name, you can build your own binaries from source. See [build from source](../build-from-source).

## Setup auto-completion

Command launcher will automatically handle auto completion for all sub commands. You need to setup it once:

### Bash

```
$ source <(cdt completion bash)

# To load completions for each session, execute once:
# Linux:
$ cdt completion bash > /etc/bash_completion.d/cdt
# macOS:
$ cdt completion bash > $(brew --prefix)/etc/bash_completion.d/cdt
```

### Zsh

```
# If shell completion is not already enabled in your environment,
# you will need to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ cdt completion zsh > "${fpath[1]}/_cdt"

# You will need to start a new shell for this setup to take effect.
```

### Powershell

```
PS> cdt completion powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> cdt completion powershell > cdt.ps1
# and source this file from your PowerShell profile.
```

### Fish

```
$ cdt completion fish | source

# To load completions for each session, execute once:
$ cdt completion fish > ~/.config/fish/completions/cdt.fish
```


## Uninstall

Simply delete the binary.

## Build from source

Command launcher is easy to build from source, follow the [instructions](../build-from-source)
