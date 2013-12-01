package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
)

type config struct {
	v map[string]string
}

func (c *config) Set(name, value string) {
	c.v[name] = value
	c.save()
}
func (c *config) Get(name string) string {
	if s, ok := c.v[name]; ok {
		return s
	}
	return ""
}
func (c *config) SetSecret(name, value string) {
	c.Set(name, c.Encrypt(value))
}
func (c *config) GetSecret(name string) string {
	return c.Decrypt(c.Get(name))
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

func (c *config) getCipherBlock() cipher.Block {
	if k, ok := c.v[cipherKey]; ok {
		k, err := hex.DecodeString(k)
		if err == nil && len(k) == cipherLen {
			return makeCipherBlock(k)
		}
	}
	k := make([]byte, cipherLen)
	io.ReadFull(rand.Reader, k)
	c.v[cipherKey] = hex.EncodeToString(k)
	c.save()
	return makeCipherBlock(k)
}

func (c *config) Encrypt(plaintext string) string {
	buf := make([]byte, aes.BlockSize+len(plaintext))
	iv, rest := buf[:aes.BlockSize], buf[aes.BlockSize:]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	copy(rest, plaintext)
	cipher.NewCFBEncrypter(c.getCipherBlock(), iv).XORKeyStream(rest, rest)
	return base64.StdEncoding.EncodeToString(buf)
}

func (c *config) Decrypt(ciphertext string) string {
	buf, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil || len(buf) < aes.BlockSize {
		return ""
	}
	iv, rest := buf[:aes.BlockSize], buf[aes.BlockSize:]
	cipher.NewCFBDecrypter(c.getCipherBlock(), iv).XORKeyStream(rest, rest)
	return string(rest)
}

func (c *config) load() {
	data, err := ioutil.ReadFile(configFileName())
	if err == nil {
		if err = json.Unmarshal(data, &c.v); err != nil {
			log.Println(os.Stderr, "config load:", err)
		}
	}
}

func (c *config) save() {
	data, err := json.MarshalIndent(c.v, "", "  ")
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	ioutil.WriteFile(configFileName(), data, 0700)
}

var Config config

const configName = ".mcbot-config"

func configFileName() string {
	u, err := user.Current()
	if err != nil {
		return configName
	}
	return path.Join(u.HomeDir, configName)
}

func init() {
	Config = config{make(map[string]string)}
	Config.load()
}
