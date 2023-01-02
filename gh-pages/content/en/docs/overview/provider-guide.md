---
title: "CLI provider guide"
description: "Complete guide to integrate your CLI to command launcher"
lead: "Complete guide to integrate your CLI to command launcher"
date: 2022-10-02T18:45:11+02:00
lastmod: 2022-10-02T18:45:11+02:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "provider-guide-e4586ef1fd15acd1ef9a2ca69711c418"
weight: 230
toc: true
---


> NOTE: in this page, we use `cola` as the command launcher's binary name, you can build your own command launcher with a different name. See: [build from source](../../quickstart/build-from-source)

Command launcher synchronizes commands from the remote command repository. Commands are packaged into a `package`, then uploaded to remote command repository. The following diagram shows this architecture.

```text
               ┌─────────────────────────────┐
               │  Remote Command Repository  │
               └──────────────┬──────────────┘
                  ┌───────────┴────────────┐
                  ▼                        ▼
            ┌────────────┐           ┌───────────┐
            │  pacakge 1 │           │ pakcage 2 │
            └────────────┘           └───────────┘
                  │                        │
          ┌───────┤─────────┐          ┌───┴────┐
          ▼       ▼         ▼          ▼        ▼
      ┌───────┐┌───────┐┌───────┐  ┌───────┐┌───────┐
      │ cmd A ││ cmd B ││ cmd C │  │ cmd D ││ cmd E │
      └───────┘└───────┘└───────┘  └───────┘└───────┘
```

## Remote command repository

A remote command repository is a simple http server with following endpoints:

- `/index.json`, which returns the list of packages available.

It is up-to-you to implement such an http server. You can configure command launcher to point to your remote repository with following command:

```shell
cola config command_repository_base_url https://my-remote-repository/root/url
```

> NOTE: the command launcher will search for [command_repository_base_url]/index.json for the remote registry. You can also use a local folder as the "base url", for example, `/tmp/cola-remote-repository/`. In this case, you need to create the `index.json` registry file in the folder. This is useful for test purpose.

You need to config an endpoint to auto update command launcher itself as well:

```shell
cola config self_update_latest_version_url https://my-remote-repository/cola/root/url/version
cola config self_update_base_url https://my-remote-repository/cola/root/url
```

The `self_update_latest_version_url` configuration defines the url to download the metadata of command launchers's latest version in YAML or JSON format, see: [Command Launcher version metadata](#command-launcher-version-metadata)

The `self_update_base_url` configuratin defines the base url to download command launcher binary. It follows the following pattern: `[SELF_UPDATE_BASE_URL]/{version}/{binaryName}_{OS}_{ARCH}_{version}{extension}`. If you build your own binary, you should make it available following the same convention.

### Remote repository registry /index.json

Remote repository registry is a json file, which contains all available packages:

The following example demonstrates a registry, which has three packages. Note that the package "hotfix" has two different versions, and the version `1.0.0-45149` targets to 30% of the user (partition 6, 7, and 8). More details about the partition see [Progressive Rollout](#progressive-rollout)

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

### Command launcher version metadata

Command launcher update itself by checking an endpoint defined in config `self_update_latest_version_url`. This endpoint returns the command version metadata:

In JSON

```json
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

> The version metadata endpoint supports both YAML and JSON format. It is recommandded to use YAML for this endpoint because of the multiple line string support in YAML

You can also target a small portion of your command user by specifying the partition. More details see: [Progressive Rollout](#progressive-rollout)

## Integrate your command into command launcher

### Package your command into a command package

> A command package = zip(your commands, manifest file)

A command package is simply a zip of your command and a manifest file that tell command launcher how to run your command. It is up-to-you to organize the structure of the package, the only requirement here is to keep the `manifest.mf` file in the root of the package.

For example, the following structure keeps the binary in different folder according to the os.

```text
my-package.pkg
├─linux/
├─windows/
├─macosx/
└─manifest.mf
```

### Package manifest file, manifest.mf

See [manifest.mf specification](../manifest)

## Upload your package, and update package registry

Once you have your command package ready, you can upload it to a remote server that command launcher have access. Depends on how you implement the http server of your remote command repository, the upload process could be different, The only requirement here is to ensure your package is accessible from command launcher.

You also need to update the [index.json](#remote-repository-registry-indexjson) endpoint to include your package in it and specify the package url in the `url` property.

## Progressive Rollout

Command launcher will assign each machine a unique partition ID from 0 to 9. When you roll out your package, you can specify the partition that you want to target to. For example, you just developed a new version, packaged into package `my-pkg 1.1.0`, uploaded it to remote repository. You can edit the `/index.json` registry, add following entry to target 40% of your audience (partition 4, 5, 6, and 7):

```json
{
  "name": "my-pkg",
  "version": "1.1.0"
  "checksum": "",
  "startPartition": 4,
  "endPartition": 7
}
```

You will have different monitoring vectors for each partition, which will help you making A/B tests.

## Monitoring

Command launcher current implements a built-in graphite exporter. It reports the following metrics to graphite:

1. success command execution count: `devtools.cdt.[package name].[group].[name].ok.count`
2. success command duration: `devtools.cdt.[package name].[group].[name].ok.duration`
3. fail command execution count: `devtools.cdt.[package name].[group].[name].ko.count`
4. fail command duration: `devtools.cdt.[package name].[group].[name].ko.duration`

You can add your custom metrics exporter by a `__metrics__` command hook in a system package, see [system package](../system-package)

## Credential Management

Command launcher has a built-in login command, which will prompt the user to enter his user name and password. The default implementation will store the user name and password securely in the system credential manager. Each command can request the access to such credential in its manifest. Command launcher will ensure user consent on accessing these credentials and pass them to the underlying command through environment variables. More detail see [Manage resources](../resources)
