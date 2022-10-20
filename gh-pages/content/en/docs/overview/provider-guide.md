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

A remote command repository is a simple http server, with following endpoints:

- `/index.json`: package registry, which returns the list of packages available.
- `/version`: returns the metadata of the latest version of command launcher.
- `/{version}/{binaryName}_{OS}_{ARCH}_{version}{extension}`: endpoints that download command launcher binary.

It is up-to-you to implement such an http server. You can configure command launcher to point to your remote repository with following command:

```shell
cola config command_repository_base_url https://my-remote-repository/root/url
```

You need to config an endpoint to auto update command launcher as well:

```shell
cola config self_update_base_url https://my-remote-repository/cola/root/url
cola config self_update_latest_version_url https://my-remote-repository/cola/root/url/version
```

### Remote repository registry /index.json

Remote repository registry is a json file, which contains all available packages:

The following example demonstrates a registry, which has three packages. Note that the package "hotfix" has two different versions, and the version `1.0.0-45149` targets to 30% of the user (partition 6, 7, and 8). More details about the partition see [Progressive Rollout](#progressive-rollout)

```json
[
  {
    "name": "hotfix",
    "version": "1.0.0-44733",
    "checksum": "5f5f47e4966b984a4c7d33003dd2bbe8fff5d31bf2bee0c6db3add099e4542b3",
    "startPartition": 0,
    "endPartition": 9
  },
  {
    "name": "hotfix",
    "version": "1.0.0-45149",
    "checksum": "773a919429e50346a7a002eb3ecbf2b48d058bae014df112119a67fc7d9a3598",
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

> Please note that for "env" package, the presence of a `url` field indicates a different location to download the package than the url defined in configuration: `command_repository_base_url`. It enables the distributed package storage.

### Command launcher version metadata /version

Command launcher update itself by checking an endpoint defined in config `self_update_latest_version_url`. This endpoint returns the command version metadata:

```json
{
    "version": "45861",
    "releaseNotes": "- feature 1\n-feature 2\nfeature 3",
    "startPartition": 0,
    "endPartition": 9
}
```

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

Once you have your command package ready, you can upload it to the remote command repository. Depends on how you implement the http server of your remote command repository, the upload process could be different, The only requirement here is to ensure your package can be downloaded from: `https://command_repository_base_url/[package-name]-[package-version].pkg`

> Upload your package to the remote repository server is optional, you can upload your package to any http server that can be accessed by command launcher, and specify the url in the remote repository index.json

You also need to update the [index.json](#remote-repository-registry-indexjson) endpoint to include your package in it.

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

## Credential Management
