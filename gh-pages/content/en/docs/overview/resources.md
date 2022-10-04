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

## What is resources

Resources are the information collected by command launcher. One good example is the user name and password from the built-in `login` command.

To access these information, a command needs to explicitly request the access to these resources through `requestedResources` property in the manifest.

For example, the following snippet of manifest requests the resource `USERNAME` and `LOGIN_TOKEN`.

```yaml
pkgName: infra-management
version: 1.0.0
cmds:
  - name: create-pod
    ...
    requestedResources: [ "USERNAME", "LOGIN_TOKEN" ]

```

## User consent

Command launcher will pass the resources to the command on runtime through environment variables: `[APP_NAME]_[RESOURCE_NAME]`, **ONLY IF** user has agreed to do so. This is done through a user consent process, with a prompt message for the first-time run of the command:

```text
Command 'create-pod' requests access to the following resources:
  - USERNAME
  - LOGIN_TOKEN

authorize the access? [yN]
```

The user consent will last for a specific period of time define in `user_consent_life` configuration.

## Access resources in your command

Once user grant the access to the requested resources, command launcher will pass the resources to the command in runtime through environment variable with naming convention: `[APP_NAME]_[RESOURCE_NAMe]`. Here is an example of bash script:

```bash
#!/bin/bash

USERNAME=${CDT_USERNAME}
LOGIN_TOKEN=${CDT_LOGIN_TOKEN}
```

## Available resources

| Resource Name | Description                                          |
|---------------|------------------------------------------------------|
| USERNAME      | the username collected from `login` command          |
| PASSWORD      | the password collected from `login` command          |
| LOG_LEVEL     | the log level of command launcher                    |
| DEBUG_FLAGS   | the debug flags defined in command launcher's config |
