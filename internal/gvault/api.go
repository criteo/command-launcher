package vault

type Vault interface {
	Write(key string, value string) error

	Read(key string) (string, error)
}
