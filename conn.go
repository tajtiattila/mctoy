package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Conn struct {
	cfg     Config
	host    string
	port    int
	c       net.Conn
	r       *PacketScanner
	w       *PacketWriter
	wbuf    []byte
	rbufsiz int
}

const (
	bufLen = 65536
)

var (
	ErrPacketMismatch    = errors.New("Packet id mismatch")
	ErrServerAddrInvalid = errors.New("Server address invalid")
	ErrNoResponse        = errors.New("No response from server")
	ErrLoginMissing      = errors.New("Username or password missing")
	ErrLoginFailed       = errors.New("Login failed")
)

func Connect(cfg Config) (*Conn, error) {

	v := strings.SplitN(cfg.Value("server"), ":", 2)
	host := v[0]
	if len(host) == 0 {
		return nil, ErrServerAddrInvalid
	}
	port := 25565
	var err error
	if len(v) > 0 {
		port, err = strconv.Atoi(v[1])
		if err != nil {
			return nil, ErrServerAddrInvalid
		}
	}

	u, p := cfg.Value("username"), cfg.Secret("password")
	if u == "" || p == "" {
		return nil, ErrLoginMissing
	}

	return dial(cfg, host, port)
}

func dial(cfg Config, host string, port int) (*Conn, error) {
	nc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	var (
		sr io.Reader
		sw io.Writer
	)
	sr, sw = nc, nc
	if PACKETDEBUG {
		sr, sw = NewDebugReader(sr, os.Stdout), NewDebugWriter(sw, os.Stdout)
	}
	c := &Conn{
		cfg,
		host,
		port,
		nc,
		NewPacketScanner(sr),
		NewPacketWriter(sw),
		make([]byte, bufLen),
		0,
	}
	return c, nil
}

type ServerStatus struct {
	Description string `json:"description"`
	Players     struct {
		Online int `json:"online"`
		Max    int `json:"max"`
	} `json:"players"`
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Ping time.Duration
}

func (s *ServerStatus) String() string {
	return fmt.Sprintf("%s %d/%d %s %s", s.Version.Name, s.Players.Online, s.Players.Max, s.Ping, s.Description)
}

func NewServerStatus(host string, port int) (*ServerStatus, error) {
	c, err := dial(nil, host, port)
	if err != nil {
		return nil, err
	}
	return c.ServerStatus()
}

func (c *Conn) ServerStatus() (*ServerStatus, error) {
	/*
		C->S : Handshake State=1
		C->S : Request
		S->C : Response
		C->S : Ping
		S->C : Ping
	*/

	err := c.Send(Handshake{
		ProtocolVersion: 4, // 1.7.2
		ServerAddress:   c.host,
		ServerPort:      uint16(c.port),
		NextState:       StateStatus,
	})

	err = c.Send(StatusRequest{})
	if err != nil {
		return nil, err
	}

	var sr StatusResponse
	err = c.Recv(&sr)
	if err != nil {
		return nil, err
	}

	s := new(ServerStatus)
	err = json.Unmarshal([]byte(string(sr)), s)
	if err != nil {
		return nil, err
	}

	t, err := c.Ping()
	if err != nil {
		return nil, err
	}

	s.Ping = t

	return s, nil
}

func (c *Conn) Ping() (t time.Duration, err error) {
	err = c.Send(Ping{time.Now().Unix()})
	if err != nil {
		return
	}
	var p Ping
	if err = c.Recv(&p); err != nil {
		return
	}
	t = time.Now().Sub(time.Unix(p.Time, 0))
	return
}

func (c *Conn) Login() error {
	/*
		C->S : Handshake State=2
		C->S : Login Start
		S->C : Encryption Key Request
		(Client Auth)
		C->S : Encryption Key Response
		(Server Auth, Both enable encryption)
		S->C : Login Success
	*/

	var err error

	clientToken, accessToken := c.cfg.Value("clientToken"), c.cfg.Value("accessToken")

	var ygg YggAuth
	profile := LoadYggProfile(c.cfg)
	if accessToken == "" || ygg.Validate(accessToken) != nil {
		tryRefresh := true
		if clientToken == "" {
			clientToken, err = GenerateV4UUID()
			if err != nil {
				return err
			}
			c.cfg.SetValue("clientToken", clientToken)
			tryRefresh = false
		}
		var yr *YggResponse
		if tryRefresh {
			yr, err = ygg.Refresh(clientToken, accessToken)
			if err == nil {
				fmt.Println("Access token refreshed.")
			}
		}
		if yr == nil {
			user, passwd := c.cfg.Value("username"), c.cfg.Secret("password")
			yr, err = ygg.Authenticate(user, passwd, clientToken)
			if err != nil {
				return err
			}
			fmt.Println("Authenticated.")
		}
		clientToken, accessToken = yr.ClientToken, yr.AccessToken
		c.cfg.SetValue("clientToken", clientToken)
		c.cfg.SetValue("accessToken", accessToken)
		profile = yr.SelectedProfile
		SaveYggProfile(c.cfg, profile)
	} else {
		fmt.Println("Access token still valid.")
	}

	err = c.Send(Handshake{
		ProtocolVersion: 4, // 1.7.2
		ServerAddress:   c.host,
		ServerPort:      uint16(c.port),
		NextState:       StateLogin,
	})

	err = c.Send(LoginStart{profile.Name})
	if err != nil {
		return err
	}

	var erq EncryptionRequest
	if err = c.Recv(&erq); err != nil {
		return err
	}
	dumpJson(erq)

	sharedSecret, err := GenerateSharedSecret()
	if err != nil {
		return err
	}

	h := sha1.New()
	io.WriteString(h, erq.ServerId)
	h.Write(sharedSecret)
	h.Write(erq.PublicKey)
	sidSum := McDigest(h.Sum(nil))

	rsacipher, err := NewRSA_PKCS1v15(erq.PublicKey)
	if err != nil {
		return err
	}

	url := "https://sessionserver.mojang.com/session/minecraft/join"
	jd, err := json.Marshal(map[string]interface{}{
		"accessToken":     accessToken,
		"selectedProfile": profile,
		"serverId":        sidSum,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(jd))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		return err
	}

	c.Send(EncryptionResponse{
		SharedSecret: rsacipher.Encrypt(sharedSecret),
		VerifyToken:  rsacipher.Encrypt(erq.VerifyToken),
	})
	if rsacipher.Error() != nil {
		return rsacipher.Error()
	}

	c.EnableCrypto(sharedSecret)

	c.Scan()
	pkid, ok := c.PeekId()
	if !ok {
		return ErrNoResponse
	}

	switch pkid {
	case 0x00:
		var d Disconnect
		if err = c.Peek(&d); err != nil {
			return err
		}
		fmt.Println("Disconnect:", string(d))
		return ErrLoginFailed
	default:
		fmt.Println("Unexpected packet received, id:", pkid)
		return ErrLoginFailed
	case 0x02:
		var ls LoginSuccess
		if err = c.Peek(&ls); err != nil {
			return err
		}
		fmt.Println("Login successful")
	}

	return nil
}

func (c *Conn) EnableCrypto(secret []byte) {
	aesc, err := aes.NewCipher(secret)
	if err != nil {
		panic(err)
	}
	var (
		sr io.Reader
		sw io.Writer
	)
	sr = cipher.StreamReader{
		R: c.c,
		S: NewCFB8Decrypter(aesc, secret),
	}
	sw = cipher.StreamWriter{
		W: c.c,
		S: NewCFB8Encrypter(aesc, secret),
	}
	if PACKETDEBUG {
		sr, sw = NewDebugReader(sr, os.Stdout), NewDebugWriter(sw, os.Stdout)
	}
	c.r, c.w = NewPacketScanner(sr), NewPacketWriter(sw)
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
	o, err := rsa.EncryptPKCS1v15(rand.Reader, c.pubkey, b)
	if err != nil {
		c.err = err
	}
	return o
}
func (c *RSA_PKCS1v15) Error() error {
	return c.err
}

func (c *Conn) Send(p Packet) error {
	pkc := MakePacketEncoder(c.wbuf)
	pkc.PutUvarint(p.Id())
	pkc.Encode(p)
	if pkc.Error() != nil {
		return pkc.Error()
	}
	_, err := c.w.Write(pkc.Bytes())
	return err
}

func (c *Conn) Scan() bool {
	return c.r.Scan()
}

func (c *Conn) PeekId() (uint, bool) {
	return c.r.PacketId()
}

func (c *Conn) Peek(p Packet) (err error) {
	pkc := MakePacketDecoder(c.r.Bytes())
	id := pkc.Uvarint()
	if p.Id() != id {
		return ErrPacketMismatch
	}
	pkc.Decode(p)
	if err == nil {
		err = pkc.Error()
	}
	return
}

func (c *Conn) Recv(p Packet) (err error) {
	c.r.Scan()
	err = c.Peek(p)
	return
}
