---
title: "Manifest.mf specification"
description: "Specification of manifest.mf file"
lead: "Specification of manifest.mf file"
date: 2022-10-02T19:02:32+02:00
lastmod: 2022-10-02T19:02:32+02:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "manifest-c4f5d2abc378574f57d70f0ab85d24fb"
weight: 240
toc: true
---


## What is a manifest.mf file?

A manifest.mf file is a file located at the root of your command launcher package. It describes the commands packaged in the zip file. When cola installs a package, it reads the manifest file and registers the commands in the manifest file.

## Format of manifest.mf

manifest.mf is in JSON or YAML format. It contains 3 fields:

- pkgName: a unique name of your package
- version: the version of your package
- cmds: a list of command definition, see command definition section

Here is an example

```json
{
    "pkgName": "hotfix",
    "version": "1.0.0-44231",
    "cmds": [ ... ]
}
```

## Command Definition

> Command launcher is implemented with [cobra](https://github.com/spf13/cobra). It follows the same command concepts:
>
> Commands represent actions, Args are things and Flags are modifiers for those actions.
>
> The best applications read like sentences when used, and as a result, users intuitively know how to interact with them.
>
> The pattern to follow is APPNAME VERB NOUN --ADJECTIVE or APPNAME COMMAND ARG --FLAG.
>
> Each package contains multiple command definitions. You can specify following definition for your command:

### Command properties list

| Property           | Required           | Description                                                                                          |
|--------------------|--------------------|------------------------------------------------------------------------------------------------------|
| name               | yes                | the name of your command                                                                             |
| type               | yes                | the type of the command, `group` or `executable`                                                     |
| group              | no                 | the group of your command belongs to, default, command launcher root                                 |
| short              | yes                | a short description of your command, it will be display in auto-complete options                     |
| long               | no                 | a long description of your command                                                                   |
| argsUsage          | no                 | custom the one line usage in help                                                                    |
| examples           | no                 | a list of example entries                                                                            |
| executable         | yes for executable | the executable to call when executing your command                                                   |
| args               | no                 | the argument list to pass to the executable, command launcher arguments will be appended to the list |
| validArgs          | no                 | the static list of options for auto-complete the arguments                                           |
| validArgsCmd       | no                 | array of string, command to run to get the dynamic auto-complete options for arguments               |
| requiredFlags      | no                 | the static list of options for the command flags (deprecated in 1.9)                                 |
| flags              | no                 | the flag list (available in 1.9+)                                                                    |
| exclusiveFlags     | no                 | group of exclusive flags (available in 1.9+)                                                         |
| groupFlags         | no                 | group of grouped flags, which must be presented together (available in 1.9+)                         |
| checkFlags         | no                 | whether check the flags defined in manifest before calling the command, default false                |
| requestedResources | no                 | the resources that the command requested, ex, USERNAME, PASSWORD                                     |

## Command properties

### name

The name of the command. A user uses the group and the name of the command to run it:

```shell
cola {group} {name}
```

You must make sure your command's group and name combination is unique

### type

There are three types of commands: `group`, `executable`, or `system`

An `executable` type of command is meant to be executed. You must fill the `executable` and `args` fields of an executable command.

A `group` type of command is used to group executable commands.

A `system` type of command is an executable command which extends the built-in command-launcher functions. More details see [system package](../system-package)

### group

The group of your command. A user uses the group and the name of your command to run it:

```shell
cola {group} {name}
```

You must make sure your command's group and name combination is unique

To registry a command at the root level of command launcher, set `group` to empty string.

> Note: command launcher only supports one level of group, the "group" field of a "group" type command is ignored.

**Example**

```json
{
  ...
  "cmds": [
    {
      "name": "infra",
      "type": "group"
    },
    {
      "name": "reintall",
      "type": "executable",
      "group": "infra",
      "executable": "{{.PackageDir}}/bin/reinstall",
      "args": []
    }
    ...
  ]
}
```

The above manifest snippet registered a command: `cola infra reinstall`, when triggered, it will execute the `reinstall` binary located in the package's bin folder

### short

The short description of the command. It is mostly used as the description in auto-complete options and the list of command in help output. Please keep it in a single line.

### long

The long description of the command. In case your command doesn't support "-h" or "--help" flags, command launcher will generate one help command for you, and render your long description in the output.

### argsUsage

Custom the one-line usage message. By default, command launcher will generate a one-line usage in the format of:

```text
Usage:
  APP_NAME group command_name [flags]
```

For some commands that accept multiple types of arguments, it would be nice to have a usage that show the different argument names and their orders. For example, for a command that accepts the 1st argument as country, and 2nd argument as city name, we can custom the usage message with following manifest:

```json
{
  ...
  "cmds": [
    {
      "name": "get-city-population",
      "type": "executable",
      "executable": "{{.PackageDir}}/bin/get-city-population.sh",
      "args": [],
      "argsUsage": "country city"
    }
    ...
  ]
}
```

The help message looks like:

```text
Usage:
  cola get-city-population country city [flags]
```

### examples

You can add examples to your command's help message. The `examples` property defines a list of examples for your command. Each example contains two fields: `scenario` and `command`:

- **scenario**, describes the use case.
- **cmd**, demonstrates the command to apply for the particular use case.

For example:

```json
{
  ...
  "cmds": [
    {
      "name": "get-city-population",
      "type": "executable",
      "executable": "{{.PackageDir}}/bin/get-city-population.sh",
      "args": [],
      "argsUsage": "country city"
      "examples": [
        {
          "scenario": "get the city population of Paris, France",
          "cmd": "get-city-population France Paris"
        }
      ]
    }
    ...
  ]
}
```

The help message looks like:

```text
...

Usage:
  cola get-city-population country city [flags]

Example:
  # get the city population of Paris, France
  get-city-population France Paris

...
```

### executable

The executable to call when your command is trigger from command launcher. You can inject predefined variables in the executable location string. More detail about the variables see [Manifest Variables](./VARIABLE.md)

**Example**

```json
{
  ...
  "cmds": [
    {
      "executable": "{{.PackageDir}}/bin/my-binary{{.Extension}}"
    }
  ]
}
```

### args

The arguments that to be appended to the executable when the command is triggered. The other arguments passed from command launcher will be appeneded after these arguments that are defined in `args` field.

**Example**

```json
{
  ...
  "cmds": [
    {
      "name": "crawler",
      "type": "executable",
      "group": "",
      "executable": "java",
      "args": [ "-jar", "{{.PackageDir}}/bin/crawler.jar"]
    }
  ]
}
```

When we call this command from command launcher:

```shell
cola crawler --url https://example.com
```

It executes following command:

```shell
java -jar {{package path}}/bin/crawler.jar --url https://example.com
```

Note: you can use variables in `args` fields as well. See [Variables](./VARIABLE.md)

### validArgs

A static list of the arguments for auto-complete.

**Example**

```json
{
  "cmds": [
    {
      "name": "population",
      "type": "executable",
      "group": "city",
      "executable": "get-city-population",
      "args": [],
      "validArgs": [
        "paris",
        "rome",
        "london"
      ]
    }
  ]
}
```

Once you have configured auto-complete for command launcher, the command described above will have auto-complete for its arguments.

When you type: `[cola] city population [TAB]`, your shell will prompt options: `paris`, `rome`, and `london`

### validArgsCmd

A command to execute to get the dynamic list of arguments.

**Example**

```json
{
  "cmds": [
    {
      "name": "population",
      "type": "executable",
      "group": "city",
      "executable": "{{.PackageDir}}/bin/get-city-population",
      "args": [],
      "validArgsCmd": [
        "{{.PackageDir}}/bin/population-cities.sh",
        "-H",
      ]
    }
  ]
}
```

When you type `[cola] city poplution [TAB]`, command launcher will run the command specified in this field, and append all existing flags/arguments to the `validArgsCmd`.

More details see: [Auto-Complete](./AUTO_COMPLETE.md)

### flags

> Available in 1.9+

Define flags (options) of the command. Each flags could have the following properties

| Property | Required | Description                                                                             |
|----------|----------|-----------------------------------------------------------------------------------------|
| name     | yes      | flag name                                                                               |
| short    | no       | flag short name, usually one letter                                                     |
| desc     | no       | flag description                                                                        |
| type     | no       | flag type, default "string", currently support "string" and bool"                       |
| default  | no       | flag default value, only available for string type, bool flag's default is always false |
| required | no       | boolean, is the flag required, default false                                            |

**Example**

```json
{
   "cmds": [
    {
      "name": "population",
      "type": "executable",
      "group": "city",
      "executable": "get-city-population",
      "args": [],
      "validArgs": [
        "paris",
        "rome",
        "london"
      ],
      "flags": [
        {
          "name": "human",
          "short": "H",
          "desc": "return the human readabe format",
          "type": "bool"
        }
      ]
    }
  ]
}
```

### exclusiveFlags

> Available in 1.9+

Declare two or more flags are mutually exclusive. For example, a flag `json` that outputs JSON format will be exclusive to the flag `text`, which outputs in plain text format.

**Example**

```json
{
  "cmds": [
    {
      "name": "population",
      "type": "executable",
      "group": "city",
      "executable": "get-city-population",
      "args": [],
      "validArgs": [
        "paris",
        "rome",
        "london"
      ],
      "flags": [
        {
          "name": "human",
          "short": "H",
          "desc": "return the human readabe format",
          "type": "bool"
        },
        {
          "name": "json",
          "short": "j",
          "desc": "return the JSON format",
          "type": "bool"
        }
      ],
      "exclusiveFlags": [
        [ "human", "json" ]
      ]
    }
  ]
}
```

### groupFlags

> Available in 1.9+

Ensure that two or more flags are presented at the same time.

**Example**

```json
{
  "cmds": [
    {
      "name": "population",
      "type": "executable",
      "group": "city",
      "executable": "get-city-population",
      "args": [],
      "flags": [
        {
          "name": "country",
          "short": "c",
          "desc": "country name",
          "required": true
        },
        {
          "name": "city",
          "short": "t",
          "desc": "city name"
        }
      ],
      "groupFlags": [
        [ "country", "city" ]
      ]
    }
  ]
}
```

### requiredFlags

> Deprecated in 1.9, see [flags](./#flags) property

The static list of flags for your command

**Example**

```json
{
  "cmds": [
    {
      "name": "population",
      "type": "executable",
      "group": "city",
      "executable": "get-city-population",
      "args": [],
      "validArgs": [
        "paris",
        "rome",
        "london"
      ],
      "requiredFlags": [
        "human\t H\t return the human readable format",
      ]
    }
  ]
}
```

It declares a `--human` flags with a short form: `-H`

The flag definition string uses `\t` to separate different fields. A complete fields can be found in the following table

| field position | field name         | field description                                                                                 |
|----------------|--------------------|---------------------------------------------------------------------------------------------------|
| 1              | flag full name     | the full name of the flags, usually in format of x-y-z, note: no need to include `--` in the name |
| 2              | flag short name    | optional, the short name (one letter) for the flag                                                |
| 3              | flag description   | optional, the description of the flag name                                                        |
| 4              | flag type          | optional, the flag type, one of string and bool. Default: string                                  |
| 5              | flag default value | optional, the default value if not specified                                                      |

Beside the complete form, it is also possible to have a short form:

1. only the full name: ex. "user-name"
2. full name + description: ex. "user-name\t the user name"
3. full name + short name + description: ex. "user-name\t u\t the user name"

### checkFlags

Whether parse and check flags before execute the command. Default: false.

The `requiredFlags` (deprecated in 1.9), `flags`, `validArgs` and `validArgsCmd` are mainly used for auto completion. Command launcher will not parse the flag and arguments by default, it will simply pass through them to the callee command. In other words, in this case, it is the callee command's responsibility to parse the flags and arguments. This works fine when the command is implemented with languages that has advanced command line supports, like golang.

For some cases, arguments parsing is difficult or has less support, for example, implementing the command in shell script. Enable `checkFlags` will allow command launcher to parse the arguments and catch errors. Further more, command launcher will pass the parsed flags and arguments to the callee command through environment variables:

- For flags: `COLA_FLAG_[FLAG_NAME]` ('-' is replaced with '_'). Example: flag `--user-name` is passed through environment variable `COLA_FLAG_USER_NAME`

- For arguments: `COLA_ARG_[INDEX]` where the index starts from 1. Example: command `cola get-city-population France Paris` will get environment variable `COLA_ARG_1=France` and `COLA_ARG_2=Paris`. An additional environment variable `COLA_NARGS` (available in 1.9+) is available as well to get the number of arguments.

> Even checkFlags is set to `true`, command launcher will still pass through the original arguments to the callee command as well, in addition, to the original arguments, parsed flags and arguments are passed to the callee as environment variables.

Another behavior change is that once `checkFlags` is enabled, the `-h` and `--help` flags are handled by command launcher. The original behavior is managed by the callee command itself.

### requestedResources

Under the user consent, command launcher can pass several resources to the callee command, for example, the user credential collected and stored securely by the built-in `login` command. The `requestedResources` is used to request such resources. Command launcher will prompt user consent for the first time, and pass requested resources value to the callee command through environment variable. More detail see: [Manage resources](../resources)

The following snippet requests to access the user name and password resources.

```json
{
  "cmds": [
    {
      "name": "population",
      "type": "executable",
      "group": "city",
      "executable": "get-city-population",
      "args": [],
      "requestedResources": [ "USERNAME", "PASSWORD" ]
    }
  ]
}
```

## System commands

Besides the [system command](../system-package/#system-commands) defined in a [system package](../system-package). You can define package level system command as a lifecycle hook.

### \_\_setup\_\_

When a package is installed, sometimes it requires some setups to make it work properly. For example, a command written in javascript could require a `npm install` to install all its dependencies. You can define a system command `__setup__` in your package as a normal command, with type `system`, it will be called when the package is installed. (You can disable this behavior, by turning the configuration `enable_package_setup_hook` to `false`). You can also manually trigger it through the built-in command: `cola package setup`

> Make sure the setup hook is idempotent, when a new version is installed the setup hook will be called again.

**Example**

```yaml
pkgName: package-demo
version: 1.0.0
cmds:
    - name: __setup__
      type: system
      executable: "{{.PackageDir}}/hooks/setup-hook"
      args: [ "predefined-arg1", "predefined-arg2" ]
    - name: other-commands
      type: executable
      executable: "{{.PackageDir}}/scripts/other-cmd.sh"
```
