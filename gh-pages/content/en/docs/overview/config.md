---
title: "Configuration"
description: "List of command launcher configurations"
lead: "List of command launcher configurations"
date: 2022-10-02T20:21:40+02:00
lastmod: 2022-10-02T20:21:40+02:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "config-22936c8eb194f0d72a25cab368910542"
weight: 235
toc: true
---

## List of configurations

| Config Name                      | Type     | Description                                                                                                                   |
|----------------------------------|----------|-------------------------------------------------------------------------------------------------------------------------------|
| ci_enabled                       | bool     | whether the CI mode is enabled or not                                                                                         |
| command_repository_base_url      | string   | the base url of the remote repository, it must contain a `/index.json` endpoint to list the available pacakges                |
| command_update_enabled           | bool     | whether auto update managed commands or not                                                                                   |
| dropin_folder                    | string   | the absolute path of the dropin folder                                                                                        |
| enabled_user_consent             | bool     | whether enable the user consent. Be caution, when set to false, all resources are allowed to pass to the managed commands.    |
| experimental_command_enabled     | bool     | whether enable experimental command or not                                                                                    |
| internal_command_enabled         | bool     | whether enable internal command or not                                                                                        |
| local_command_repository_dirname | bool     | the absolute path of the local repository folder.                                                                             |
| log_enabled                      | bool     | whether log is enabled or not                                                                                                 |
| log_level                        | string   | the log level of command launcher. Note, the managed command could also request access to this config                         |
| metric_graphite_host             | string   | graphite url for metrics                                                                                                      |
| package_lock_file                | string   | only available for CI mode (ci_enabled = true). Lock the package version for CI purpose                                       |
| remote_config_check_cycle        | int      | interval in hours to check the remote config                                                                                  |
| remote_config_check_time         | time     | next remote config check time. This configuration is set automatically by command launcher, you shouldn't change it manually. |
| self_update_base_url             | string   | base url to get command launcher binaries                                                                                     |
| self_update_enabled              | bool     | whether auto update command launcher itself                                                                                   |
| self_update_latest_version_url   | string   | url to get the latest command launcher version information                                                                    |
| self_update_timeout              | duration | timeout duration for self update                                                                                              |
| usage_metrics_enabled            | bool     | whether enable metrics                                                                                                        |
| user_consent_life                | duration | the life of user consent                                                                                                      |

## Change configuration

It is recommended to use the built-in `config` command to change the configurations. For duration type configurations, you can use `h`, `m`, and `s` to present hour, minute, and seconds. For example:

```bash
cdt config user_consent_life 24h
```

sets the user consent life to 24 hours.
