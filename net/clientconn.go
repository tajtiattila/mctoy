package net

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type ClientConn struct {
	Conn
	perstor PersistentStore // used by YggAuth
}

var (
	ErrNoResponse  = errors.New("No response from server")
	ErrLoginFailed = errors.New("Login failed")
)

func Connect(addr string, s PersistentStore) (c *ClientConn, err error) {
	nc := &ClientConn{perstor: s}
	if err = nc.dial(addr); err == nil {
		c = nc
	}
	return
}

func NewServerStatus(addr string) (*ServerStatus, error) {
	c, err := Connect(addr, nil)
	if err != nil {
		return nil, err
	}
	return c.ServerStatus()
}

func (c *ClientConn) ServerStatus() (*ServerStatus, error) {
	/*
		C->S : Handshake State=1
		C->S : Request
		S->C : Response
		C->S : Ping
		S->C : Ping
	*/

	err := c.Handshake(StateStatus)
	if err != nil {
		return nil, err
	}

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
	err = json.Unmarshal([]byte(sr.JSON), s)
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

func (c *ClientConn) Ping() (t time.Duration, err error) {
	err = c.Send(StatusPing{time.Now().Unix()})
	if err != nil {
		return
	}
	var p StatusPing
	if err = c.Recv(&p); err != nil {
		return
	}
	t = time.Now().Sub(time.Unix(p.Time, 0))
	return
}

func (c *ClientConn) Login(up UserPassworder) error {
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

	err = c.Handshake(StateLogin)
	if err != nil {
		return err
	}

	ygg := NewYggAuth(c.perstor)
	if err = ygg.Start(up); err != nil {
		return err
	}

	err = c.Send(LoginStart{ygg.ProfileName()})
	if err != nil {
		return err
	}

	var erq EncryptionRequest
	if err = c.Recv(&erq); err != nil {
		return err
	}

	session, err := ygg.JoinSession(erq.ServerId, erq.PublicKey)
	if err != nil {
		return err
	}

	c.Send(EncryptionResponse{
		SharedSecret: session.Cipher.Encrypt(session.SharedSecret),
		VerifyToken:  session.Cipher.Encrypt(erq.VerifyToken),
	})

	c.InitIO(session.SharedSecret)

	if err = c.Scan(); err != nil {
		return err
	}
	pkid, ok := c.PeekId()
	if !ok {
		return ErrNoResponse
	}

	switch pkid {
	case 0x00:
		var d LoginDisconnect
		if err = c.Peek(&d); err != nil {
			return err
		}
		fmt.Println("LoginDisconnect:", d.Reason)
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
		c.state = StatePlay
	}

	return nil
}

func (c *ClientConn) Handshake(nextstate CxnState) (err error) {
	err = c.Send(Handshake{
		ProtocolVersion: 4, // 1.7.2
		ServerAddress:   c.host,
		ServerPort:      uint16(c.port),
		NextState:       uint(nextstate),
	})
	if err == nil {
		c.state = nextstate
	}
	return
}
