---
title: "Enterprise Setup Guide"
description: "Setup command launcher for enterprise scenarios"
lead: "Setup command launcher for enterprise scenarios"
date: 2022-10-30T11:34:47+02:00
lastmod: 2022-10-30T11:34:47+02:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "overview-enterprise-setup"
weight: 245
toc: true
---

## Setup remote configuration

For enterprise use case, it is common to enforce some configurations for all users. For example, the remote registry to synchronise local commands. You can specify a remote configuration file using `COLA_REMOTE_CONFIG_URL` environment variable for this purpose.

```shell
export COLA_REMOTE_CONFIG_URL=https://remote-server/remote-config.json
```

The remote configuration is a JSON file contains the configuration items that you would like to enforce. Command launcher will override the local configuration item when the item is defined in the remote configuration.

For example, you can setup `self_update_enabled` to `true` in the remote configuration file. This will ensure that all users get the latest version of command launcher automatically.

### Remote configuration synchronise cycle

It is always nice to be able to change the configuration temporarily. Command launcher will check the remote configuration periodically. You can setup this check period from the configuration: `remote_config_check_cycle`.

For example, the following configuration set up a 24-hour check period.

```json
{
    ...
    "remote_config_check_cycle": 24,
    ...
}
```

> For the configuration items, which are missing from the remote configuration. The local value is always respected.

## Command auto-update: setup remote registry

Another common use case for enterprise scenario is to ensure the same set and same version of commands available on a group of users. (For example, all engineers have the same version of build and test command).

You can setup a remote package registry to list the available packages. Command launcher will synchronise with it at the end of each command call and make sure the local copy is synchronised with the remote package registry (always use the latest version available on the remote registry).

Remote repository registry is a json file, which contains all available packages:

The following example demonstrates a registry, which has three packages. Note that the package "hotfix" has two different versions, and the version `1.0.0-45149` targets to 30% of the user (partition 6, 7, and 8). More details about the partition see [Progressive Rollout](../provider-guide#progressive-rollout)

```json
[
  {
    "name": "hotfix",
    "version": "1.0.0-44733",
    "checksum": "5f5f47e4966b984a4c7d33003dd2bbe8fff5d31bf2bee0c6db3add099e4542b3",
    "url": "https://the-url-of-the-env-package/any-name.zip",
    "startPartition": 0,
    "endPartition": 9
  },
  {
    "name": "hotfix",
    "version": "1.0.0-45149",
    "checksum": "773a919429e50346a7a002eb3ecbf2b48d058bae014df112119a67fc7d9a3598",
    "url": "https://the-url-of-the-env-package/hotfix-1.0.0-45149.zip",
    "startPartition": 6,
    "endPartition": 8
  },
  {
    "name": "env",
    "version": "0.0.1",
    "checksum": "c87a417cce3d26777bcc6b8b0dea2ec43a0d78486438b1bf3f3fbd2cafc2c7cc",
    "url": "https://the-url-of-the-env-package/package.zip",
    "startPartition": 0,
    "endPartition": 9
  }
]
```

You can host this `index.json` file on an http server, for example: `https://my-company.com/cola-remote-registry/index.json`.

To make command launcher be aware of the remote package registry, setup the configuration:

```shell
cola config command_repository_base_url https://my-company.com/cola-remote-registry
```

## Use command launcher on CI

Another use case of using command launcher is for Continuous Integration (CI). In this case, we would like to pin the version of command to have a deterministic behavior.

Two configurations will help us achieve it: `ci_enabled` and `package_lock_file`.

When enabled, the `ci_enabled` bool configuration tells command launcher to read a "lock" file to get the package version instead of using the latest one from remote registry. The lock file is specified in the `package_lock_file` configuration.

When `ci_enabled` config set to `false`. The lock file is ignored by command launcher.

### Package lock JSON file

Package lock file pins the package version in command launcher:

```json
{
    "hotfix": "1.2.0",
    "infra-ops": "3.1.2",
    ...
}
```

The example above demonstrates a lock file, which pins the `hotfix` package version to `1.2.0`, and `infra-ops` package version to `3.1.2`

> NOTE: please make sure the version pinned in lock file are available on the remote package registry.
>
> Partition will be ignored when the version is pinned in a lock file.

## Self Auto-update

Command launcher looks for a version metadata endpoint to recognize its latest version, and download the binary follows a URL convention.

The latest version endpoint is defined by `self_update_latest_version_url` configuration. It must return the latest command launcher version in JSON or YAML format:

In JSON

```JSON
{
    "version": "45861",
    "releaseNotes": "- feature 1\n-feature 2\nfeature 3",
    "startPartition": 0,
    "endPartition": 9
}
```

Or in YAML

```YAML
version: "45861"
releaseNotes: |
  * feature 1
  * feature 2
  * feature 3
startPartition: 0
endPartition: 9
```

The binary download URL must follow the convention:

```text
[SELF_UPDATE_BASE_URL]/{version}/{binaryName}_{OS}_{ARCH}_{version}{extension}
```

The `[SELF_UPDATE_BASE_URL]` should be defined in `self_update_base_url` configuration. The latest `{version}` can be found in the version metadata from `self_update_latest_version_url` endpoint. `{binaryName}` is the short name when building command launcher (default `cola`). Pre-built `{OS}` and `{ARCH}` see following table:

| OS      | Architecture | Pre-built  |
|---------|--------------|------------|
| windows | amd64        | yes        |
| windows | arm64        | no         |
| linux   | amd64        | yes        |
| linux   | arm64        | yes         |
| darwin  | amd64        | yes        |
| darwin  | arm64        | yes        |

## Custom command launcher with system package

See [System Package](../system-package)
