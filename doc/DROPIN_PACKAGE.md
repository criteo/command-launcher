# Dropin Package

## What is dropin package? and why do I need it?

A dropin package is a package that are not managed by the remote repository. It is only available on the developer's machine. It allows developers to integrate their own scripts/tools into command launcher to benefit the feature provided by the command launcher, for example, auto-complete, monitoring etc.

For example, you probably already has lots of shell scripts to maintain your infrastructure. Writing auto-completion for all these scripts is time consuming and it is difficult to remember which script doing what, and what parameters they accept. By writing a simple `manifest.mf` file, you can let command launcher to manage these scripts for you with auto-complete, secret management, and monitoring.

Dropin package is also a good way for you to develop and test your command launcher package, as it follows the same structure as a regular command launcher package.

## How to create a dropin package?

1. Identify the dropins folder, by running the following command: (support your command launcher is named `cl`, see [README.md](../README.md) to build your command launcher with a different name)
```
cl config dropin_folder
```
If the dropin folder returned by previous command doesn't exist, create it.

2. create a package folder in the dropin folder, let's say, a package named `my-first-package`. You can named it whatever you want.

3. add a `manifest.mf` in the newly created package folder, follow [MANIFEST.md](./MANIFEST.md) guide to define your command in the manifest file. Note: you can copy other scripts in the package folder and use `{{.Root}}` to reference the package root folder in your manifest file.

4. run `cl` any time to test your command


## How to share a dropin package with others?

A dropin package is simply a directory with manifest.mf in it, the best way to share a dropin package is to push it to a git repository and ask for others to clone it in their own dropin folder


## How to update dropin package?

For now, the command launcher does not update the dropin folder automatically, it is up to developers themselve to make a git pull to keep these dropin package up-to-date.







