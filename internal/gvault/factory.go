package vault

func CreateVault(name string) (Vault, error) {
	fileVault := FileVault{
		Name: name,
	}

	if err := fileVault.init(); err != nil {
		return nil, err
	}

	return &fileVault, nil
}
