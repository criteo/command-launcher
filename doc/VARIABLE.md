# Variables in package manifest

The two common use cases of integrating commands in command launcher are:
1. Reference resources (files, binraries) that are located in the package itself
2. Provide system/architecture-aware commands, for example, .sh script for linux, and .bat script for windows

To cover these use cases, in certain fields of the manifest file, predefined variables can used in the field values.

## Available Variables

| Variable Name   | Variable Description                                                   |
|-----------------|------------------------------------------------------------------------|
| Root            | The absolute path of the package                                       |
| Cache           | Same as "Root" variable                                                |
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
    "executable": "{{.Root}}/bin/script{{.ScripteExtension}}",
  }
]
```
The executable on linux will be a script called `script.sh` located in the bin folder of the package. On windows, the executable will be a script called `script.bat`.

## Advanced usage of variables

See golang [text/template](https://pkg.go.dev/text/template)
