package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Dico map[string]string

type FileVault struct {
	Name string
	hash []byte
}

func (fv *FileVault) Write(key string, value string) error {
	vaultDir, err := maybeCreateDir()
	if err != nil {
		return err
	}

	dico, err := fv.readFile()
	if err != nil {
		return err
	}

	dico[key] = value
	data, err := json.Marshal(dico)
	if err != nil {
		return err
	}

	encrypted, err := fv.encrypt(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(vaultDir, fv.Name), encrypted, 0600)
}

func (fv *FileVault) Read(key string) (string, error) {
	dico, err := fv.readFile()
	if err != nil {
		return "", err
	}

	if len(dico) == 0 {
		return "", fmt.Errorf("vault %s is empty", fv.Name)
	}

	return dico[key], nil
}

func (fv *FileVault) readFile() (Dico, error) {
	dico := make(Dico)
	vaultDir, err := maybeCreateDir()
	if err != nil {
		return dico, err
	}

	encrypted, err := ioutil.ReadFile(filepath.Join(vaultDir, fv.Name))
	if err != nil {
		return dico, err
	}

	if len(encrypted) == 0 {
		return dico, err
	}

	data, err := fv.decrypt(encrypted)
	if err != nil {
		return dico, err
	}

	err = json.Unmarshal(data, &dico)
	if err != nil {
		return dico, err
	}

	return dico, nil
}

func (fv *FileVault) init() error {
	hash, err := readSecret()
	if err != nil {
		return err
	}
	fv.hash = hash

	dirVault, err := maybeCreateDir()
	if err != nil {
		return err
	}

	_, err = os.Stat(filepath.Join(dirVault, fv.Name))
	if os.IsNotExist(err) {
		_, err = os.OpenFile(filepath.Join(dirVault, fv.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fv *FileVault) decrypt(encrypted []byte) ([]byte, error) {
	cphr, err := aes.NewCipher(fv.hash)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(cphr)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, msg := encrypted[:nonceSize], encrypted[nonceSize:]
	data, err := gcm.Open(nil, nonce, msg, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (fv *FileVault) encrypt(data []byte) ([]byte, error) {
	cphr, err := aes.NewCipher(fv.hash)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(cphr)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)

	return encrypted, nil
}

func readSecret() ([]byte, error) {
	// first get the secret from environment variable
	secret := os.Getenv("CDT_VAULT_SECRET")
	if secret != "" {
		hash := sha256.Sum256([]byte(secret))
		return hash[:], nil
	}

	empty := []byte{}
	homedir, err := os.UserHomeDir()
	if err != nil {
		return empty, err
	}

	sshDir := filepath.Join(homedir, ".ssh")
	_, err = os.Stat(sshDir)
	if err != nil {
		return empty, err
	}

	// get the secret file from environment variable
	secretFile := os.Getenv("CDT_VAULT_SECRET_FILE")
	if secretFile == "" {
		secretFile = filepath.Join(sshDir, "id_rsa")
	}

	// in case environment variable missing, fallback to default
	data, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return empty, err
	}

	hash := sha256.Sum256(data)
	return hash[:], nil
}

func maybeCreateDir() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dirVault := filepath.Join(homedir, ".file-vault")
	_, err = os.Stat(dirVault)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirVault, 0700)
		if err != nil {
			return "", err
		}
	}

	return dirVault, nil
}
