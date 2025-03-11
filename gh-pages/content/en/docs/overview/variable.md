---
title: "Variables in manifest"
description: "Use of variables in the manifest.mf file"
lead: "Use of variables in the manifest.mf file"
date: 2022-10-02T21:36:35+02:00
lastmod: 2022-10-02T21:36:35+02:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "variable-d940a4460a129d45a4ae1158e21b2130"
weight: 999
toc: true
---

The values of certain fields in the manifest file can be set using predefined variables. These variables are replaced with their actual values at runtime.

The two most common use cases of variables in a Command Launcher's manifest.mf file are:

1. Referencing files that are located in the package directory itself
2. Providing system/architecture-aware commands, for example, .sh script for linux, and .bat script for windows

## Available Variables

| Variable Name   | Variable Description                                                   |
|-----------------|------------------------------------------------------------------------|
| PackageDir      | The absolute path of the package                                       |
| Root            | Same as the "PackageDir" variable                                      |
| Cache           | Same as the "PackageDir" variable                                      |
| Os              | The OS: "windows", "linux", or "darwin"                                |
| Arch            | The system architecture: "arm64", "amd64"                              |
| Binary          | The binary file name of the Command Launcher                           |
| Extension       | The system-aware binary extension, "" for linux, ".exe" for windows    |
| ScriptExtension | The system-aware script extension, ".sh" for linux, ".bat" for windows |

## Fields that accept variables

Variables can only be used in the command properties: `executable`, `args`, and `validArgsCmd`

## How to use these variables

You can reference them in this format: `{{.Variable}}`. For example:

```json
"cmds": [
  {
    "name": "variable-demo",
    "type": "executable",
    "executable": "{{.PackageDir}}/bin/script{{.ScripteExtension}}",
  }
]
```

The executable on Linux will be a script called `script.sh` located in the `bin` folder of the package. On windows, the executable will be a script called `script.bat`.

## If Else

One common scenario is to have a different path or file name, depending on what OS Command Launcher is running on. You can use a conditional structure (if-else) in the fields that accept variables. For example:

```json
"cmds": [
  {
    "name": "variable-demo",
    "type": "executable",
    "executable": "{{.PackageDir}}/bin/script{{if eq .Os \"windows\"}}.ps1{{else}}.sh{{end}}",
  }
]
```

The executable on Linux will be a script called `script.sh` located in the `bin` folder of the package. On Windows, the executable will be a script called `script.ps1`.

## Advanced usage of variables

See golang [text/template](https://pkg.go.dev/text/template) for advanced usage.
