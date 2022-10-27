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

The default pre-built binary is named `cola` (**Co**mmand **La**uncher), if you want to use a different name, you can build your own binaries from source. See [build from source](../build-from-source).

For example, in each release, we also build a binary named `cdt` (Criteo Dev Toolkit). If you prefer to download `cdt`, please replace `cola` to `cdt` in the examples from the documents.

## Setup auto-completion

Command launcher will automatically handle auto completion for all sub commands. You need to setup it once:

### Bash

```bash
$ source <(cola completion bash)

# To load completions for each session, execute once:
# Linux:
$ cola completion bash > /etc/bash_completion.d/cola
# macOS:
$ cola completion bash > $(brew --prefix)/etc/bash_completion.d/cola
```

### Zsh

```bash
# If shell completion is not already enabled in your environment,
# you will need to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ cola completion zsh > "${fpath[1]}/_cola"

# You will need to start a new shell for this setup to take effect.
```

### Powershell

```powershell
PS> cola completion powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> cola completion powershell > cola.ps1
# and source this file from your PowerShell profile.
```

### Fish

```bash
$ cola completion fish | source

# To load completions for each session, execute once:
$ cola completion fish > ~/.config/fish/completions/cola.fish
```

## Uninstall

Simply delete the binary.

## Build from source

Command launcher is easy to build from source, follow the [instructions](../build-from-source)
