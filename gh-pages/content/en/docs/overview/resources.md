---
title: "Manage resources"
description: "Access informations that collected from command launcher"
lead: "Access informations that collected from command launcher"
date: 2022-10-02T19:21:34+02:00
lastmod: 2022-10-02T19:21:34+02:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "resources-9305bb9c59eb0255e5e68e7e938d505f"
weight: 250
toc: true
---

## What are Resources

Resources are the information collected by Command Launcher. One good example is the _username_ and _password_ from the built-in `login` command.

Some of these pieces of information require user consent to access them, a command needs to explicitly request the access to these resources through the `requestedResources` property in the manifest.

Others are automatically passed to the command.

Command Launcher passes resources to managed commands through environment variables. The naming convention is: COLA_[RESOURCE_NAME]. If you compiled command launcher to a different name, command launcher will pass an additional environment variable `[APP_NAME]_[RESOURCE_NAME]` to the managed command as well.

For example, the following snippet of manifest requests the resource `USERNAME` and `AUTH_TOKEN`.

```yaml
pkgName: infra-management
version: 1.0.0
cmds:
  - name: create-pod
    ...
    requestedResources: [ "USERNAME", "AUTH_TOKEN" ]

```

## User consent

Command launcher will pass the resources to the command on runtime through environment variables: `COLA_[RESOURCE_NAME]`, **ONLY IF** user has agreed to do so. This is done through a user consent process, with a prompt message for the first-time run of the command:

```text
Command 'create-pod' requests access to the following resources:
  - USERNAME
  - AUTH_TOKEN

authorize the access? [yN]
```

The user consent will last for a specific period of time define in the `user_consent_life` configuration entry.

## Access resources in your command

Once the user grants access to the requested resources, Command Launcher will pass the resources to the command at runtime through environment variables with the naming convention: `COLA_[RESOURCE_NAME]`. Here is an example of a bash script:

```bash
#!/bin/bash

USERNAME=${COLA_USERNAME}
AUTH_TOKEN=${COLA_AUTH_TOKEN}
```

## Available resources

| Resource Name     | Require User Consent | Description                                               |
|-------------------|----------------------|-----------------------------------------------------------|
| USERNAME          | Yes                  | the username collected from `login` command               |
| PASSWORD          | Yes                  | the password collected from `login` command               |
| AUTH_TOKEN        | Yes                  | the authentication token collected from `login` command   |
| LOG_LEVEL         | No                   | the log level of command launcher                         |
| DEBUG_FLAGS       | No                   | the debug flags defined in command launcher's config      |
| PACKAGE_DIR       | No                   | the absolute path to the package directory                |
| FULL_COMMAND_NAME | No                   | the name of the command executed (includes app and group) |
