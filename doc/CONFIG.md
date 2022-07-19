# Configuration & Command environment variables

## Environment variables

For every command launcher command, the following environment variables are available.

The environment variable name is in the form of `[COMMAND_LAUNCHER_BINARY_NAME]_[CONFIG_KEY]`, all in capital.

| config key        | env var example      | description                                                |
|-------------------|----------------------|------------------------------------------------------------|
| HOME              | CL_HOME              | the command launcher home directory                        |
| USERNAME          | CL_USERNAME          | the user name registered in the built-in login command     |
| PASSWORD          | CL_PASSWORD          | the user password registered in the built-in login command |
| CONFIG_FILE       | CL_CONFIG_FILE       | the absolute path of the configuration file                |
| REMOTE_CONFIG_URL | CL_REMOTE_CONFIG_URL | the url to download the remote configuration               |


### XXX_HOME

The XXX_HOME environment variable defines the command laucher's global home directory. This directory will contain the default configuration file, the default local package repository, the default dropins folder.

### XXX_USERNAME

When you login with command launcher's default `login` command, the username and password are stored in the system's default credential vault. The username and password will be passed to your command as an environment variable. `[BINARY_NAME]_USERNAME` holds the username, and `[BINARY_NAME]_PASSWORD` holds the password.

### XXX_PASSWORD

See `XXX_USERNAME`

### XXX_CONFIG_FILE

By default, the configuration file will be loaded from `$XXX_HOME/config.json` file. If you want to override it, you can specify the configuration file in `XXX_CONFIG_FILE` variable.

### XXX_REMOTE_CONFIG_URL

For enterprise user, you can setup a remote configuration file to ensure all users has a default configuration, which will be merged with the user's local configuration.


## Configuration file




## Configuration from configuration



