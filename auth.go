// thank you, Toqueteos!
// https://gist.github.com/toqueteos/5372776
package main

import (
	"bufio"
	"bytes"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

/*
func LoadAuthProfile(c Config) *AuthProfile {
	return &AuthProfile{c.Value("profileId"), c.Value("profileName")}
}
func SaveAuteProfile(c Config, p *AuthProfile) {
	c.SetValue("profileId",   p.Id)
	c.SetValue("profileName", p.Name)
}*/

type AuthProfile struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AuthInfo struct {
	ClientToken       string        `json:"clientToken,omitempty"`
	AccessToken       string        `json:"accessToken,omitempty"`
	SelectedProfile   *AuthProfile  `json:"selectedProfile,omitempty"`
	AvailableProfiles []AuthProfile `json:"availableProfiles,omitempty"`
}

type AuthResponse struct {
	AuthInfo
	Error string `json:"error,omitempty"`
}

type CredentialRequest interface {
	AskLogin(savedUsername string) (username, password string, err error)
}

type YggError string

func (y YggError) Error() string { return "Yggdrasil: " + string(y) }

type yggPayload map[string]interface{}

////////////////////////////////////////////////////////////////////////////////

type YggAuth struct {
	store PersistentStore
	pwreq CredentialRequest
	info  AuthInfo
}

func NewYggAuth(s PersistentStore, crq CredentialRequest) *YggAuth {
	y := &YggAuth{store: s, pwreq: crq}
	err := y.store.Load(&y.info)
	if err != nil {
		log.Println(err)
	}
	return y
}

func (y *YggAuth) ProfileName() string {
	if y.info.SelectedProfile != nil {
		return y.info.SelectedProfile.Name
	}
	return ""
}

func (y *YggAuth) Start() error {
	// try validate first
	if y.Validate() == nil {
		log.Println("Access token still valid.")
		return nil
	}
	// then try refresh
	if y.Refresh() == nil {
		log.Println("Access token refreshed.")
		y.store.Save(&y.info)
		return nil
	}
	// else ask the user for her credentials
	var oldname string
	if y.info.SelectedProfile != nil {
		oldname = y.info.SelectedProfile.Name
	}
	user, passwd, err := y.pwreq.AskLogin(oldname)
	if err != nil {
		return err
	}

	// and try to authenticate
	if err = y.Authenticate(user, passwd); err != nil {
		return err
	}
	log.Println("Authenticated.")
	y.store.Save(&y.info)
	return nil
}

func (y *YggAuth) Authenticate(username, password string) error {
	// Generate an access token using a username and password
	// (Any existing client token is invalidated if not provided)
	// Returns response error on failure
	if y.info.ClientToken == "" {
		var err error
		y.info.ClientToken, err = GenerateV4UUID()
		if err != nil {
			return err
		}
	}
	resp, err := y.request("/authenticate", yggPayload{
		"agent": yggPayload{
			"name":    "Minecraft",
			"version": 1,
		},
		"username":    username,
		"password":    password,
		"clientToken": y.info.ClientToken,
	})
	if err != nil {
		return err
	}
	if err = yggError(resp); err != nil {
		return err
	}
	y.info = resp.AuthInfo
	dumpJson(y.info)
	return nil
}

func (y *YggAuth) Refresh() error {
	// Generate an access token with a client/access token pair
	// (The used access token is invalidated)
	// Returns response error on failure
	if y.info.ClientToken == "" {
		return YggError("Client token necessary for refresh")
	}
	resp, err := y.request("/refresh", yggPayload{
		"accessToken": y.info.AccessToken,
		"clientToken": y.info.ClientToken,
	})
	if err != nil {
		return err
	}
	dumpJson(resp)
	if err = yggError(resp); err != nil {
		return err
	}
	y.info = resp.AuthInfo
	return nil
}

func (y *YggAuth) SignOut(username, password string) (*AuthResponse, error) {
	// Invalidate access tokens with a username and password
	// Returns None on success, otherwise error dict
	resp, err := y.request("/signout", yggPayload{
		"username": username,
		"password": password,
	})
	if err != nil {
		return nil, err
	}
	if err = yggError(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (y *YggAuth) Invalidate(clientToken, accessToken string) error {
	// Invalidate access tokens with y client/access token pair
	// Returns nil on success, otherwise error
	if y.info.AccessToken == "" {
		return YggError("No access token to invalidate")
	}
	resp, err := y.request("/invalidate", yggPayload{
		"accessToken": y.info.AccessToken,
		"clientToken": y.info.ClientToken,
	})
	if err != nil {
		return nil
	}
	return yggError(resp)
}

func (y *YggAuth) Validate() error {
	// Check if an access token is valid
	// Returns nil on success (ie valid access token), otherwise error
	if y.info.AccessToken == "" {
		return YggError("No access token to validate")
	}
	resp, err := y.request("/validate", yggPayload{
		"accessToken": y.info.AccessToken,
	})
	if err != nil {
		return err
	}
	return yggError(resp)
}

func (y *YggAuth) request(endpoint string, payload interface{}) (*AuthResponse, error) {
	url := "https://authserver.mojang.com" + endpoint
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	r := new(AuthResponse)
	if err = json.NewDecoder(resp.Body).Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (y *YggAuth) JoinSession(serverIdString string, publicKey []byte) (*SessionInfo, error) {
	sharedSecret, err := GenerateSharedSecret()
	if err != nil {
		return nil, err
	}

	h := sha1.New()
	io.WriteString(h, serverIdString)
	h.Write(sharedSecret)
	h.Write(publicKey)
	sidSum := McDigest(h.Sum(nil))

	url := "https://sessionserver.mojang.com/session/minecraft/join"
	jd, err := json.Marshal(map[string]interface{}{
		"accessToken":     y.info.AccessToken,
		"selectedProfile": y.info.SelectedProfile,
		"serverId":        sidSum,
	})
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(jd))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	rsacipher, err := NewRSA_PKCS1v15(publicKey)
	if err != nil {
		return nil, err
	}

	return &SessionInfo{sharedSecret, rsacipher}, nil
}

////////////////////////////////////////////////////////////////////////////////

type SessionInfo struct {
	SharedSecret []byte
	Cipher       *RSA_PKCS1v15
}

////////////////////////////////////////////////////////////////////////////////

func McDigest(hash []byte) string {
	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		twosComplement(hash)
	}

	// Trim away zeroes
	res := strings.TrimLeft(hex.EncodeToString(hash), "0")
	if negative {
		res = "-" + res
	}

	return res
}

// little endian
func twosComplement(p []byte) {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = byte(^p[i])
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
}

type RSA_PKCS1v15 struct {
	pubkey *rsa.PublicKey
	err    error
}

func NewRSA_PKCS1v15(pubkey []byte) (*RSA_PKCS1v15, error) {
	k0, err := x509.ParsePKIXPublicKey(pubkey)
	if err != nil {
		return nil, err
	}
	return &RSA_PKCS1v15{pubkey: k0.(*rsa.PublicKey)}, nil
}
func (c *RSA_PKCS1v15) Encrypt(b []byte) []byte {
	o, err := rsa.EncryptPKCS1v15(crand.Reader, c.pubkey, b)
	if err != nil {
		c.err = err
	}
	return o
}
func (c *RSA_PKCS1v15) Error() error {
	return c.err
}

////////////////////////////////////////////////////////////////////////////////

// return a ctypto-inited pseudo random generator
func NewRandom() (*rand.Rand, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(crand.Reader, buf); err != nil {
		return nil, err
	}
	seed := int64(binary.BigEndian.Uint64(buf))
	return rand.New(rand.NewSource(seed)), nil
}
func MustNewRandom() *rand.Rand {
	r, e := NewRandom()
	if e != nil {
		panic(e)
	}
	return r
}

type randReader struct {
	r   *rand.Rand
	buf []byte
	pos int
}

func (r *randReader) Read(b []byte) (nread int, err error) {
	nread = len(b)
	for len(b) != 0 {
		if r.pos == 0 {
			u1, u2 := uint64(r.r.Int63()), uint64(r.r.Int63())
			u := u1 ^ (u2 << 1)
			binary.BigEndian.PutUint64(r.buf, u)
		}
		n := copy(b, r.buf[r.pos:])
		r.pos, b = (r.pos+n)&7, b[n:]
	}
	return
}
func NewRandReader(r *rand.Rand) io.Reader {
	return &randReader{r, make([]byte, 8), 0}
}

var (
	authRand   = MustNewRandom()
	authReader = NewRandReader(authRand)
)

func GenerateSharedSecret() ([]byte, error) {
	sharedSecret := make([]byte, 16)
	if _, err := io.ReadFull(authReader, sharedSecret); err != nil {
		return nil, err
	}
	return sharedSecret, nil
}

func GenerateV4UUID() (string, error) {
	src := bufio.NewReaderSize(authReader, 1)
	b := []byte("xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx")
	vhex := "0123456789abcdef"
	for i := range b {
		switch b[i] {
		case 'x', 'y':
			byt, err := src.ReadByte()
			if err != nil {
				return "", err
			}
			digit := int(byt) % 16
			if b[i] == 'y' {
				digit = 8 + digit%4
			}
			b[i] = vhex[digit]
		}
	}
	return string(b), nil
}

func yggError(resp *AuthResponse) error {
	if resp.Error != "" {
		return YggError(resp.Error)
	}
	return nil
}
