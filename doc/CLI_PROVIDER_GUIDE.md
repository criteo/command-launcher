# CLI Provider Guide

> NOTE: in this page, we use `cdt` as the command launcher's binary name, you can build your own command launcher with a different name. See: [README](../README.md)

Command launcher synchronizes commands from the remote command repository. Commands are packaged into a `package`, then uploaded to remote command repository. The following diagram shows this relationship.

```
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
- `/{package-name}-{version}.pkg`: download endpoint of a particular package.
- `/version`: returns the metadata of the latest version of command launcher.
- `/{version}/{os}/{arch}/{binary-name}`: endpoints that download command launcher binary.

It is up-to-you to implement such an http server. You can configure command launcher to point to your remote repository with following command:

```
cdt config command_repository_base_url https://my-remote-repository/root/url
```

You need to config an endpoint to auto update command launcher as well:
```
cdt config self_update_base_url https://my-remote-repository/cdt/root/url
cdt config self_update_latest_version_url https://my-remote-repository/cdt/root/url/version
```

### Remote repository registry /index.json

Remote repository registry is a json file, which contains all available packages:

The following example demonstrates a registry, which has three packages. Note that the package "hotfix" has two different versions, and the version `1.0.0-45149` targets to 30% of the user (partition 6, 7, and 8). More details about the partition see [Progressive Rollout](#progressive-rollout)
```
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
    "startPartition": 0,
    "endPartition": 9
  }
]
```

### Command Launcher version metadata /version

Command launcher update itself by checking an endpoint defined in config `self_update_latest_version_url`. This endpoint returns the command version metadata:

```
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

> A command package = zip(your command, manifest file)

A command package is simply a zip of your command and a manifest file that tell command launcher how to run your command. It is up-to-you to organize the structure of the package, the only requirement here is to keep the `manifest.mf` file in the root of the package.

For example, the following structure keeps the binary in different folder according to the os.
```
my-package.pkg
├─linux/
├─windows/
├─macosx/
└─manifest.mf
```

### Package manifest file, manifest.mf

See [manifest.mf specification](MANIFEST.md)

## Upload your package, and update package registry

Once you have your command package ready, you need to upload it to the remote command repository. Depends on how you implement the remote command repository http server, you need to make it availeble from the download endpoint.

To make all command launcher aware about your package, you need to update the index.json endpoint to include your package in it.

## Progressive Rollout

Command launcher will assign each machine a unique partition ID from 0 to 9. When you roll out your package, you can specify the partition that you want to target to. For example, you just developed a new version, packaged into package `my-pkg 1.1.0`, uploaded it to remote repository. You can edit the `/index.json` registry, add following entry to target 40% of your audience (partition 4, 5, 6, and 7):

```
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
