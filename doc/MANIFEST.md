# Specification of manifest.mf


## What is a manifest.mf file?
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

### Command properties list

| Property      | Required           | Description                                                                                          |
|---------------|--------------------|------------------------------------------------------------------------------------------------------|
| name          | yes                | the name of your command                                                                             |
| type          | yes                | the type of the command, `group` or `executable`                                                     |
| group         | no                 | the group of your command belongs to, default, command launcher root                                 |
| short         | yes                | a short description of your command, it will be display in auto-complete options                     |
| long          | no                 | a long description of your command                                                                   |
| argsUsage     | no                 | custom the one line usage in help                                                                    |
| examples      | no                 | a list of example entries                                                                            |
| executable    | yes for executable | the executable to call when executing your command                                                   |
| args          | no                 | the argument list to pass to the executable, command launcher arguments will be appended to the list |
| validArgs     | no                 | the static list of options for auto-complete the arguments                                           |
| validArgsCmd  | no                 | array of string, command to run to get the dynamic auto-complete options for arguments               |
| requiredFlags | no                 | the static list of options for the command flags                                                     |
| checkFlags    | no                 | whether check the flags defined in manifest before calling the command, default false                |



### Command properties

#### name

The name of the command. A user uses the group and the name of the command to run it:

```
cl {group} {name}
```

You must make sure your command's group and name combination is unique


#### type

There are two types of commands: `group` or `executable`

An executable type of command is meant to be executed. You must fill the `executable` and `args` fields of an executable command.

A group type of command is used to group executable commands.


#### group

The group of your command. A user uses the group and the name of your command to run it:

```
cl {group} {name}
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

The above manifest snippet registered a command: `cl infra reinstall`, when triggered, it will execute the `reinstall` binary located in the package's bin folder

#### short

The short description of the command. It is mostly used as the description in auto-complete options and the list of command in help output. Please keep it in a single line.

#### long

The long description of the command. In case your command doesn't support "-h" or "--help" flags, command launcher will generate one help command for you, and render your long description in the output.

#### argsUsage

Custom the one-line usage message. By default, command launcher will generate a one-line usage in the format of:

```
Usage:
  APP_NAME group command_name [flags]
```

For some commands that accept multiple types of arguments, it would be nice to have a usage that show the different argument names and their orders. For example, for a commend that accepts the 1st argument as country, and 2nd argument as city name, we can custom the usage message with following manifest:

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
```
Usage:
  cl get-city-population country city [flags]
```

#### examples

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

```
...

Usage:
  cl get-city-population country city [flags]

Example:
  # get the city population of Paris, France
  get-city-population France Paris

...
```


#### executable

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

#### args

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
```
cl crawler --url https://example.com
```
It executes following command:
```
java -jar {{package path}}/bin/crawler.jar --url https://example.com
```

Note: you can use variables in `args` fields as well. See [Variables](./VARIABLE.md)

#### validArgs

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

When you type: `[cl] city population [TAB]`, your shell will prompt options: `paris`, `rome`, and `london`


#### validArgsCmd

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

When you type `[cl] city poplution [TAB]`, command launcher will run the command specified in this field, and append all existing flags/arguments to the `validArgsCmd`.

More details see: [Auto-Complete](./AUTO_COMPLETE.md)


#### requiredFlags

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

#### checkFlags

Whether parse and check flags before execute the command. Default: false.

The `requiredFlags`, `validARgs` and `validArgsCmd` are mainly used for auto completion. Command launcher will not parse the arguments by default, instead, she will simply pass the arguments to the callee command. In the other word, in this case, it is the callee command's responsibility to parse the flags and arguments. This works fine when the command is implemented with languages that has better command line supports, like golang.

For some cases, the argument parsing is difficult or has less support, for example, implementing the command in shell script. Enable `checkFlags` will allow command launcher parse the arguments and catch errors. Further more, command launcher will pass the parsed flags and arguments to the callee command through environment varibles:

For flags: `[APP_NAME]_FLAG_[FLAG_NAME]` ('-' is replaced with '_'). Example: flag `--user-name` is passed through environment variable `[APP_NAME]_FLAG_USER_NAME`
For arguments: `[APP_NAME]_ARG_[INDEX]` where the index starts from 1. Example: command `cl get-city-population France Paris` will get environment variable `[APP_NAME]_ARG_1=France` and `[APP_NAME]_ARG_2=Paris`

Another behavior change is that once `checkFlags` is enabled, the `-h` and `--help` flags are handled by command launcher. The original behavior is managed by the callee command itself.

