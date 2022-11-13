package helper

func GetUsername() (string, error) {
	return GetSecret("username")
}

func SetUsername(value string) error {
	return SetSecret("username", value)
}

func GetPassword() (string, error) {
	return GetSecret("password")
}

func SetPassword(value string) error {
	return SetSecret("password", value)
}

func GetAuthToken() (string, error) {
	return GetSecret("auth_token")
}
