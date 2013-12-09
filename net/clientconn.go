package net

import (
	"encoding/json"
	"errors"
	"fmt"
	proto "github.com/tajtiattila/mctoy/protocol"
	"time"
)

type ClientConn struct {
	Conn
}

var (
	ErrNoResponse         = errors.New("No response from server")
	ErrLoginFailed        = errors.New("Login failed")
	ErrUnexpectedResponse = errors.New("Unexpected response")
)

func Connect(addr string) (c *ClientConn, err error) {
	nc := &ClientConn{}
	if err = nc.dial(addr); err == nil {
		c = nc
	}
	return
}

func NewServerStatus(addr string) (*ServerStatus, error) {
	c, err := Connect(addr)
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

	err := c.Handshake(proto.StateStatus)
	if err != nil {
		return nil, err
	}

	err = c.Send(proto.StatusRequest{})
	if err != nil {
		return nil, err
	}

	p, err := c.Recv()
	if err != nil {
		return nil, err
	}
	sr, ok := p.(*proto.StatusResponse)
	if !ok {
		return nil, ErrUnexpectedResponse
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
	if err = c.Send(proto.StatusPing{time.Now().Unix()}); err != nil {
		return
	}
	var pi interface{}
	if pi, err = c.Recv(); err != nil {
		return
	}
	if p, ok := pi.(*proto.StatusPing); ok {
		t = time.Now().Sub(time.Unix(p.Time, 0))
	} else {
		err = ErrUnexpectedResponse
	}
	return
}

func (c *ClientConn) Login(auth Auth) error {
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

	err = c.Handshake(proto.StateLogin)
	if err != nil {
		return err
	}

	if err = auth.Start(); err != nil {
		return err
	}

	c.state = proto.StateLogin

	err = c.Send(proto.LoginStart{auth.ProfileName()})
	if err != nil {
		return err
	}

	p, err := c.Recv()
	if err != nil {
		return err
	}

	if erq, ok := p.(*proto.EncryptionRequest); ok {
		session, err := auth.JoinSession(erq.ServerId, erq.PublicKey)
		if err != nil {
			return err
		}

		c.Send(proto.EncryptionResponse{
			SharedSecret: session.Cipher.Encrypt(session.SharedSecret),
			VerifyToken:  session.Cipher.Encrypt(erq.VerifyToken),
		})

		c.InitIO(session.SharedSecret)

		if p, err = c.Recv(); err != nil {
			return err
		}
	}

	switch pkt := p.(type) {
	case *proto.LoginSuccess:
		fmt.Println("Login successful")
		c.state = proto.StatePlay
	case *proto.LoginDisconnect:
		fmt.Println("LoginDisconnect:", pkt.Reason)
		err = ErrLoginFailed
	default:
		fmt.Println("Unexpected packet received at login")
		err = ErrLoginFailed
	}

	return nil
}

func (c *ClientConn) Handshake(nextstate proto.CxnState) (err error) {
	err = c.Send(proto.Handshake{
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
