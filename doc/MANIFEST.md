# Specification of manifest.mf


## What is CDT manifest.mf file?
A manifest.mf file is a file located at the root of your cdt commands package. It describes the command packaged in the zip file. When cdt install a package, it read the manifest file and register commands to it.

## Format of manifest.mf
Manifest.mf is in JSON format. It contains 3 fields:

- pkgName: a unique name of your package
- version: the version of your package
- cmds: a list of command definition, see command definition section

Here is an example
```
{
    "pkgName": "hotfix",
    "version": "1.0.0-44231",
    "cmds": [ ... ]
}
```

## Command Definition
Each package contains multiple command definitions. You can specify following definition for your command:


TODO: add the command definition here
