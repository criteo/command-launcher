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


|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| Property      | Required           | Description                                                                                          |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| name          | yes                | the name of your command                                                                             |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| type          | yes                | the type of the command, `group` or `executable`                                                     |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| group         | no                 | the group of your command belongs to, default, command launcher root                                 |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| short         | yes                | a short description of your command, it will be display in auto-complete options                     |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| long          | no                 | a long description of your command                                                                   |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| executable    | yes for executable | the executable to call when executing your command                                                   |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| args          | no                 | the argument list to pass to the executable, command launcher arguments will be appended to the list |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| validArgs     | no                 | the static list of options for auto-complete the arguments                                           |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| validArgsCmd  | no                 | array of string, command to run to get the dynamic auto-complete options for arguments               |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| requiredFlags | no                 | the static list of options for the command flags                                                     |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|




