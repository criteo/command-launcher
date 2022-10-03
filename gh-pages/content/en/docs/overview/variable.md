---
title: "Variable"
description: "Use variables in manifest.mf file"
lead: "Use variables in manifest.mf file"
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

The two common use cases of integrating commands in command launcher are:
1. Reference files that are located in the package itself
2. Provide system/architecture-aware commands, for example, .sh script for linux, and .bat script for windows

To cover these use cases, in certain fields of the manifest file, predefined variables can used in the field values.

## Available Variables

| Variable Name   | Variable Description                                                   |
|-----------------|------------------------------------------------------------------------|
| PackageDir      | The absolute path of the package                                       |
| Root            | Same as "PackageDir" variable                                          |
| Cache           | Same as "PackageDir" variable                                          |
| Os              | The OS, "windows", "linux", and "darwin"                               |
| Arch            | The system architecture: "arm64", "amd64"                              |
| Binary          | The binary file name of the command launcher                           |
| Extension       | The system-aware binary extension, "" for linux, ".exe" for windows    |
| ScriptExtension | The system-aware scritp extension, ".sh" for linux, ".bat" for windows |


## Fields that accepts variables

The command fields: `executable`, `args`, and `validArgsCmd`

## How to use these variables

You can reference them in form of `{{.Variable}}`. For example:

```json
"cmds": [
  {
    "name": "variable-demo",
    "type": "executable",
    "executable": "{{.PackageDir}}/bin/script{{.ScripteExtension}}",
  }
]
```
The executable on linux will be a script called `script.sh` located in the bin folder of the package. On windows, the executable will be a script called `script.bat`.

## Advanced usage of variables

See golang [text/template](https://pkg.go.dev/text/template) for advanced usage (ex, if else)


