---
title: "✨ 1.6.0 Release"
description: "Command launcher 1.6.0 Release"
excerpt: "Command launcher 1.6.0 Release"
date: 2022-09-18T09:19:42+01:00
lastmod: 2022-09-18T09:19:42+01:00
draft: false
weight: 50
images: []
categories: ["News"]
tags: ["release"]
contributors: ["Bo HOU"]
pinned: false
homepage: false
---

# Release notes

- Add the `PackageDir` context variable to reference the package's directory, same as the `Cache` and `Root` context variables. `Cache` and `Root` variables are deprecated.
- Extend the format of the flag definition. Now it is possible to add bool type flags: `[name] \t [shorthand] \t [description] \t [type] \t [default]`. Currently, two types are supported: `string` and `bool`. both `type` and `default value` are optional, by default, the type is "string", and the default value is the empty string.
- New command definition field in manifest, `checkFlags`. The default value is false, when it is true, before executing the command, the arguments will be parsed, and the parsed flags and args will be passed to the command in the form of an environment variable: `[APP_NAME]_FLAG_[FLAG_LONG_NAME]` and `[APP_NAME]_ARG_[ARG_INDEX_STARTS_FROM_1]`.
- New command definition field in manifest: `argsUsage` to customize the one-line help message. This field will take effect when `checkFlags=true`
- New command definition field in manifest: `examples` to customize the examples in help message. This field will take effect when `checkFlags=true`

