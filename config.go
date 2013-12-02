package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	path "path/filepath"
)

type Config interface {
	Value(name string) string
	SetValue(name, value string)

	Secret(name string) string
	SetSecret(name, value string)

	GetAllKeys() []string
	Clear()
}

const (
	ConfigFileMode = 0700
)

var (
	ErrNoFileName = errors.New("Config: No filename")
)

type autoFileConfig struct {
	v            map[string]string
	autoFileName string
}

func NewMemoryConfig() Config {
	return &autoFileConfig{make(map[string]string), ""}
}

func NewFileConfig(configPath string) (Config, error) {
	absPath, err := path.Abs(configPath)
	if err == nil {
		return nil, err
	}
	return newFileConfig(absPath)
}

func NewUserConfig(configPath string) (Config, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	return newFileConfig(path.Join(u.HomeDir, configPath))
}

func newFileConfig(absPath string) (Config, error) {
	err := os.MkdirAll(path.Dir(absPath), ConfigFileMode)
	if err != nil {
		return nil, err
	}
	c := &autoFileConfig{make(map[string]string), absPath}
	err = c.Load()
	if err != nil {
		return nil, err
	}
	return c, nil
}

//////////////////////////////////////////////////////////////////////////////////////////

func (c *autoFileConfig) SetValue(name, value string) {
	if old, ok := c.v[name]; ok && old == value {
		// avoid save if nothing changed
		return
	}
	c.v[name] = value
	c.Save()
}

func (c *autoFileConfig) Value(name string) string {
	if s, ok := c.v[name]; ok {
		return s
	}
	return ""
}

func (c *autoFileConfig) SetSecret(name, value string) {
	c.SetValue(name, c.Encrypt(value))
}

func (c *autoFileConfig) Secret(name string) string {
	return c.Decrypt(c.Value(name))
}

func (c *autoFileConfig) GetAllKeys() []string {
	keys := make([]string, 0, len(c.v))
	for k := range c.v {
		keys = append(keys, k)
	}
	return keys
}

func (c *autoFileConfig) Clear() {
	c.v = make(map[string]string)
}

const (
	cipherKey = "_SecretSecret"
	cipherLen = 32 // must match cipher used in makeCipher
)

func makeCipherBlock(k []byte) cipher.Block {
	c, err := aes.NewCipher(k)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *autoFileConfig) getCipherBlock() cipher.Block {
	if k, ok := c.v[cipherKey]; ok {
		k, err := hex.DecodeString(k)
		if err == nil && len(k) == cipherLen {
			return makeCipherBlock(k)
		}
	}
	k := make([]byte, cipherLen)
	io.ReadFull(rand.Reader, k)
	c.v[cipherKey] = hex.EncodeToString(k)
	c.Save()
	return makeCipherBlock(k)
}

func (c *autoFileConfig) Encrypt(plaintext string) string {
	buf := make([]byte, aes.BlockSize+len(plaintext))
	iv, rest := buf[:aes.BlockSize], buf[aes.BlockSize:]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	copy(rest, plaintext)
	cipher.NewCFBEncrypter(c.getCipherBlock(), iv).XORKeyStream(rest, rest)
	return base64.StdEncoding.EncodeToString(buf)
}

func (c *autoFileConfig) Decrypt(ciphertext string) string {
	buf, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil || len(buf) < aes.BlockSize {
		return ""
	}
	iv, rest := buf[:aes.BlockSize], buf[aes.BlockSize:]
	cipher.NewCFBDecrypter(c.getCipherBlock(), iv).XORKeyStream(rest, rest)
	return string(rest)
}

func (c *autoFileConfig) Load() error {
	if c.autoFileName == "" {
		return ErrNoFileName
	}
	data, err := ioutil.ReadFile(c.autoFileName)
	if os.IsPermission(err) {
		return err
	}
	if os.IsNotExist(err) {
		// make sure we can write the file
		err = ioutil.WriteFile(c.autoFileName, nil, ConfigFileMode)
		return err
	}
	if err == nil {
		if err = json.Unmarshal(data, &c.v); err != nil {
			// don't fail on corrupt config
			log.Println(os.Stderr, "Config", c.autoFileName, "corrupt:", err)
		}
	}
	return nil
}

func (c *autoFileConfig) Save() error {
	if c.autoFileName == "" {
		return nil
	}
	data, err := json.MarshalIndent(c.v, "", "  ")
	if err == nil {
		data = append(data, '\n')
		err = ioutil.WriteFile(c.autoFileName, data, ConfigFileMode)
	}
	if err != nil {
		log.Println(os.Stderr, "Config", c.autoFileName, "can't be saved:", err)
	}
	return err
}
