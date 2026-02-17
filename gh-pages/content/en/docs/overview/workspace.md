---
title: "Workspace package"
description: "Project-scoped commands that are automatically discovered from your workspace"
lead: "Project-scoped commands that are automatically discovered from your workspace"
date: 2024-01-01T00:00:00+00:00
lastmod: 2024-01-01T00:00:00+00:00
draft: false
images: []
menu:
  docs:
    parent: "overview"
    identifier: "workspace-d1a2b3c4e5f6a7b8c9d0e1f2a3b4c5d6"
weight: 247
toc: true
---

## What are workspace packages?

Workspace packages are command packages that live inside your project directory and are automatically discovered when you run Command Launcher from within that project. Unlike [dropin packages](../dropin) (which are global to your machine) or managed packages (which come from a remote repository), workspace packages are **scoped to a specific project** and are only visible when your working directory is inside that project.

This is useful for:

- **Project-specific tooling** — build scripts, deployment helpers, or dev workflows that only make sense for a particular project
- **Team-shared commands** — check the packages and the `.cdt-packages` file into version control so every team member gets the same CLI tools
- **Quick prototyping** — develop and test new commands directly in your project without touching the global dropin folder

## Enabling workspace packages

Workspace packages are disabled by default. Enable the feature with:

```shell
cdt config ENABLE_WORKSPACE_PACKAGES true
```

## Setting up a workspace

### 1. Create a packages file

Create a file named **`.cdt-packages`** in your project root (the file name matches your app name — if your binary is called `cola`, the file is `.cola-packages`).

This file lists the relative paths to your package directories, one per line:

```text
# Lines starting with # are comments
tools/my-build-tool
tools/my-deploy-tool
scripts/dev-helpers
```

### 2. Create package directories

Each path listed in the packages file must be a directory containing a `manifest.mf` file, following the standard [manifest format](../manifest). The directory structure looks like this:

```text
my-project/
├── .cdt-packages          # lists package paths
├── src/
│   └── ...
└── tools/
    ├── my-build-tool/
    │   ├── manifest.mf    # command definitions
    │   └── build.sh       # scripts
    └── my-deploy-tool/
        ├── manifest.mf
        └── deploy.sh
```

### 3. Write a manifest

A workspace package manifest is identical to a dropin or managed package manifest. Here is a minimal example:

```json
{
  "pkgName": "my-build-tool",
  "version": "1.0.0",
  "cmds": [
    {
      "name": "build",
      "type": "executable",
      "short": "Build the project",
      "executable": "{{.PackageDir}}/build.sh"
    }
  ]
}
```

See the [manifest reference](../manifest) for the full list of supported fields (flags, auto-completion, groups, etc.).

## How discovery works

When you run Command Launcher, it walks **up** from your current working directory toward the filesystem root, looking for `.cdt-packages` files at each level. Every matching file is loaded, with **deepest-first** priority — packages found closer to your working directory take precedence over those found higher up.

For example, given this directory tree:

```text
/home/user/
├── .cdt-packages          # project-wide tools
└── repo/
    ├── .cdt-packages      # repo-specific tools (higher priority)
    └── src/
        └── (you are here)
```

If you run `cdt` from `/home/user/repo/src/`, both packages files are discovered. Commands from `/home/user/repo/.cdt-packages` take priority over those from `/home/user/.cdt-packages`.

### Priority order

Workspace packages have the **highest** priority among all package sources:

1. **Workspace packages** (deepest-first) — highest priority
2. **Dropin packages**
3. **Default managed packages**
4. **Extra remote packages** — lowest priority

If a workspace command has the same name as a command from another source, the workspace command wins.

## Security

### Consent prompt

Because workspace packages contain arbitrary scripts that execute on your machine, Command Launcher requires your **explicit consent** before running any workspace command for the first time.

When you invoke a workspace command, you will see a prompt like:

```text
This command is provided by workspace: /home/user/my-project
Do you trust and want to run commands from this workspace? [yN]
```

- **Accept (y):** The command runs, and your consent is saved so you won't be prompted again (until it expires).
- **Deny (N or Enter):** The command is not executed. Additionally, the denied workspace's commands are **hidden** from help and autocompletion on subsequent runs, preventing accidental re-prompting.

Consent is stored securely in your system keychain and expires after the duration configured by `USER_CONSENT_LIFE` (default: 30 days). After expiration, you will be prompted again.

> **Note:** Workspace commands appear in `--help` and autocompletion even **before** you consent. This lets you discover what commands are available. Consent is only checked when you actually execute a command.

### Path traversal protection

For security, paths in the `.cdt-packages` file **must not** contain `..` (parent directory traversal). Any line containing `..` is rejected with a warning. This ensures that workspace packages can only reference directories within or below the workspace root.

### Validation

Each path listed in the packages file must point to a directory that contains a valid `manifest.mf` file. Paths that don't meet this requirement are skipped with a warning (visible when logging is enabled).

## Configuration reference

| Config key | Type | Default | Description |
|---|---|---|---|
| `ENABLE_WORKSPACE_PACKAGES` | bool | `false` | Enable or disable workspace package discovery |
| `USER_CONSENT_LIFE` | duration | `720h` (30 days) | How long workspace consent is remembered before re-prompting |

Set these with:

```shell
cdt config ENABLE_WORKSPACE_PACKAGES true
cdt config USER_CONSENT_LIFE 720h
```

## Complete example

Here is a step-by-step example of adding a workspace command to a project.

**1.** Enable the feature (one-time setup):

```shell
cdt config ENABLE_WORKSPACE_PACKAGES true
```

**2.** In your project root, create the packages file:

```shell
echo "tools/hello" > .cdt-packages
```

**3.** Create the package directory and script:

```shell
mkdir -p tools/hello
```

**4.** Create `tools/hello/manifest.mf`:

```json
{
  "pkgName": "hello",
  "version": "1.0.0",
  "cmds": [
    {
      "name": "hello",
      "type": "executable",
      "short": "Say hello from workspace",
      "executable": "{{.PackageDir}}/hello.sh"
    }
  ]
}
```

**5.** Create `tools/hello/hello.sh`:

```bash
#!/bin/sh
echo "Hello from workspace!"
```

```shell
chmod +x tools/hello/hello.sh
```

**6.** Run it (from anywhere inside the project):

```shell
$ cdt hello
This command is provided by workspace: /home/user/my-project
Do you trust and want to run commands from this workspace? [yN] y
Hello from workspace!
```

On subsequent runs, the consent prompt is skipped:

```shell
$ cdt hello
Hello from workspace!
```

## Workspace packages vs. dropin packages

| | Workspace packages | Dropin packages |
|---|---|---|
| **Scope** | Per-project (based on working directory) | Global (available everywhere) |
| **Discovery** | Automatic via `.cdt-packages` file | Manual placement in dropin folder |
| **Version control** | Designed to be checked into git | Typically per-machine |
| **Consent** | Required before first execution | Not required |
| **Updates** | Manual (edit files directly) | Manual (or `package install --git`) |
| **Priority** | Highest (overrides all other sources) | Second highest |

## Troubleshooting

**Commands not showing up:**

- Verify the feature is enabled: `cdt config ENABLE_WORKSPACE_PACKAGES`
- Check that you are running `cdt` from inside the workspace (or a subdirectory)
- Ensure the `.cdt-packages` file name matches your app name (e.g., `.cola-packages` for `cola`)
- Verify each listed path contains a valid `manifest.mf`
- Enable logging for detailed diagnostics: `cdt config LOG_ENABLED true && cdt config LOG_LEVEL debug`

**Consent prompt keeps appearing:**

- The consent has expired. Increase the duration with `cdt config USER_CONSENT_LIFE 2160h` (90 days).

**Command was denied and now it's hidden:**

- Denial is remembered for the configured `USER_CONSENT_LIFE` duration. Wait for it to expire, or the consent will reset automatically after the configured period.

**Path rejected with "parent directory traversal" warning:**

- Paths in `.cdt-packages` must not contain `..`. Use paths relative to the packages file location that point downward into the project tree.
