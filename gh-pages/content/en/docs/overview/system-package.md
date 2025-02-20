---
title: "System package"
description: "Extend command launcher's built-in function with system package"
lead: "Extend command launcher's built-in function with system package"
date: 2022-10-02T21:36:35+02:00
lastmod: 2022-11-13T21:36:35+02:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "system-package-d940a4460a129d45a4ae1158e21b2130"
weight: 800
toc: true
---

## What is a System Package

A System Package is like any other Command Launcher package, with one `manifest.mf` file in it to describe the commands, and containing binaries, scripts, and resources to execute those commands.

The difference is that a System Package contains `system` commands, and it can only be installed from a central repository (not as a dropin package).

You can customize your Command Launcher by providing a system package. In a system package, you can define system commands as functional hooks to extend Command Launcher's built-in functionalities, for example, `login` and `metrics`.

## Define system package

To specify which package is the system package, use the configuration `system_package`.

```shell
cola config system_package your-system-package-name
```

An example system package manifest looks like this:

```yaml
pkgName: system-package-demo
version: 1.0.0
cmds:
    - name: __login__
      type: system
      executable: "{{.PackageDir}}/hooks/login-hook"
    - name: __metrics__
      type: system
      executable: "{{.PackageDir}}/hooks/metrics-hook"
    - name: other-commands
      type: executable
      executable: "{{.PackageDir}}/scripts/other-cmd.sh"

```

> NOTE: The system command will be ignored if the package is not defined as system package.

## System commands

To extend Command Launcher built-in functionality, you need to specify some `system` type commands in a system package.
 The following table lists available system commands:

| system command name | description                                     |
|---------------------|-------------------------------------------------|
| \_\_login\_\_       | calling your IAM system to return `login_token` |
| \_\_metrics\_\_     | collect metrics                                 |

### System command \_\_login\_\_

The built-in `login` command will trigger the `__login__` system command. It takes two arguments:

- username
- password

```shell
__login__ [username] [password]
```

The `__login__` system command outputs the credentials to be stored by command launcher in a JSON format. The credentials could be one or many of following items:

| credential name | description              | environment variable |
|-----------------|--------------------------|----------------------|
| username        | the user name            | COLA_USERNAME        |
| password        | the password             | COLA_PASSWORD        |
| auth_token      | the authentication token | COLA_AUTH_TOKEN      |

For example: the following output tells command launcher to store the `username` and the `auth_token`, but not store the `password`.

```json
{
    "username": "joe",
    "auth_token": "DZ4OfC4vS38!"
}
```

To use these credentials see [Manage resources](../resources)

### System command \_\_metrics\_\_

At the end of each Command Launcher execution, the `__metrics__` system hook will be triggered. The following arguments will be passed to the `__metrics__` system command in order:

1. repository/registry name (see remote command)
2. package name
3. command group name, or "default" if no group
4. command name
5. user partition
6. command exit code
7. command execution duration in nano seconds
8. error message or "nil" if no error
9. command start timestamp in seconds

Here is an example:

```shell
__metrics__ default example cola-example hello 2 0 5000000 nil 1668363339
```

> Note: the `__metrics__` hook will be called at the end of **each Command Launcher call**, please make sure it ends fast to reduce the footprint.
