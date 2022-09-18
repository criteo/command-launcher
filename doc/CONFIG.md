# Configuration & Command environment variables

## Environment variables

For every command in launcher command, the following environment variables are available.

The environment variable name is in the form of `[APP_NAME]_[CONFIG_KEY]`, all in capital. The binary name is what you specified in the build command `-X main.appName=`

| config key        | env var example (app name: cl) | description                                                |
|-------------------|--------------------------------|------------------------------------------------------------|
| HOME              | CL_HOME                        | the command launcher home directory                        |
| USERNAME          | CL_USERNAME                    | the user name registered in the built-in login command     |
| PASSWORD          | CL_PASSWORD                    | the user password registered in the built-in login command |
| CONFIG_FILE       | CL_CONFIG_FILE                 | the absolute path of the configuration file                |
| REMOTE_CONFIG_URL | CL_REMOTE_CONFIG_URL           | the url to download the remote configuration               |
| LOG_LEVEL         | CL_LOG_LEVEL                   | the log level defined in command launcher's config         |


## Configuration Keys

These configuration keys will be passed to your command with environment variable named [APP_NAME]_[CONFIG_KEY]

### HOME

The HOME configuration key defines the command laucher's global home directory. This directory will contain the default configuration file, the default local package repository, the default dropins folder.

### USERNAME

When you login with command launcher's default `login` command, the username and password are stored in the system's default credential vault. The username and password will be passed to your command as an environment variable. `[APP_NAME]_USERNAME` holds the username, and `[APP_NAME]_PASSWORD` holds the password.

### PASSWORD

See `USERNAME`

### CONFIG_FILE

By default, the configuration file will be loaded from `$[APP_NAME]_HOME/config.json` file. If you want to override it, you can specify the configuration file in `[APP_NAME]_CONFIG_FILE` variable.

### REMOTE_CONFIG_URL

For enterprise user, you can setup a remote configuration file to ensure all users has a default configuration, which will be merged with the user's local configuration.

### LOG_LEVEL

Command launcher's built in command `config` allows you to setup a log level for it. If your command also logs according to a log level, you can take this environment variable as input.

### Configuration keys related to parsed flags and arguments

When `checkFlags: true` is declared in the command's manifest, command launcher will first parse the arguments throwing errors if found any (ex. unknown flag name), and pass the following variables to the command runtime:

**FLAG_[FLAG_NAME]**

For example, a command declared a flag 'user-name', an environment named `[APP_NAME]_FLAG_USER_NAME` can be accessed during the command's runtime.

**ARG_[ARG_INDEX]**

The parsed arguments (exclude the flags) starting with index 1. For example, command `test -O opt1 my-args1` can access `my-arg1` from the environment variable `[APP_NAME]_ARG_1`

## Configuration load sequence

When command launcher executes, it will search for a configuration file in several places in order:

1. The file defined in [APP_NAME]_CONFIG_FILE environment variable
2. The `[APP_NAME].json` file in current working directory and all its parents until the root folder
3. The default `config.json` file in the its home directory `$HOME/.[APP_NAME]/config.json`


## Get/Set configuration

Use the built-in `config` command


