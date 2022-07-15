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
| executable    | yes for executable | the executable to call when executing your command                                                   |
| args          | no                 | the argument list to pass to the executable, command launcher arguments will be appended to the list |
| validArgs     | no                 | the static list of options for auto-complete the arguments                                           |
| validArgsCmd  | no                 | array of string, command to run to get the dynamic auto-complete options for arguments               |
| requiredFlags | no                 | the static list of options for the command flags                                                     |


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
      "executable": "{{.Root}}/bin/reinstall",
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

#### executable

The executable to call when your command is trigger from command launcher. You can inject predefined variables in the executable location string. More detail about the variables see [Manifest Variables](./VARIABLE.md)

**Example**

```json
{
  ...
  "cmds": [
    {
      "executable": "{{.Root}}/bin/my-binary{{.Extension}}"
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
      "args": [ "-jar", "{{.Root}}/bin/crawler.jar"]
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
      "executable": "{{.Root}}/bin/get-city-population",
      "args": [],
      "validArgsCmd": [
        "{{.Root}}/bin/population-cities.sh",
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










