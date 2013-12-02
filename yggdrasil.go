package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type YggProfile struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func LoadYggProfile(c Config) *YggProfile {
	return &YggProfile{c.Value("profileId"), c.Value("profileName")}
}
func SaveYggProfile(c Config, p *YggProfile) {
	c.SetValue("profileId", p.Id)
	c.SetValue("profileName", p.Name)
}

type YggResponse struct {
	AccessToken       string       `json:"accessToken,omitempty"`
	ClientToken       string       `json:"clientToken,omitempty"`
	AvailableProfiles []YggProfile `json:"availableProfiles,omitempty"`
	SelectedProfile   *YggProfile  `json:"selectedProfile,omitempty"`
	Error             string       `json:"error,omitempty"`
}

type YggError string

func (y YggError) Error() string { return "Yggdrasil: " + string(y) }

type yggPayload map[string]interface{}

type YggAuth struct {
	username, password string
	clientToken        string
	accessToken        string
}

func selStr(v1, v2 string) string {
	if v1 == "" {
		return v2
	}
	return v1
}

func yggError(resp *YggResponse) error {
	if resp.Error != "" {
		return YggError(resp.Error)
	}
	return nil
}

func (y *YggAuth) Authenticate(username, password, clientToken string) (*YggResponse, error) {
	// Generate an access token using a username and password
	// (Any existing client token is invalidated if not provided)
	// Returns response dict on success, otherwise error dict
	if username != "" {
		y.username = username
	}
	if password != "" {
		y.password = password
	}
	resp, err := y.request("/authenticate", yggPayload{
		"agent": yggPayload{
			"name":    "Minecraft",
			"version": 1,
		},
		"username":    y.username,
		"password":    y.password,
		"clientToken": selStr(clientToken, y.clientToken),
	})
	if err != nil {
		return nil, err
	}
	if err = yggError(resp); err != nil {
		return nil, err
	}
	return y.update(resp)
}

func (y *YggAuth) Refresh(clientToken, accessToken string) (*YggResponse, error) {
	// Generate an access token with a client/access token pair
	// (The used access token is invalidated)
	// Returns response dict on success, otherwise error dict
	resp, err := y.request("/refresh", yggPayload{
		"accessToken": selStr(accessToken, y.accessToken),
		"clientToken": selStr(clientToken, y.clientToken),
	})
	if err != nil {
		return nil, err
	}
	dumpJson(resp)
	if err = yggError(resp); err != nil {
		return nil, err
	}
	return y.update(resp)
}

func (y *YggAuth) SignOut(username, password string) (*YggResponse, error) {
	// Invalidate access tokens with a username and password
	// Returns None on success, otherwise error dict
	resp, err := y.request("/signout", yggPayload{
		"username": selStr(username, y.username),
		"password": selStr(password, y.password),
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
	// Invalidate access tokens with a client/access token pair
	// Returns nil on success, otherwise error
	resp, err := y.request("/invalidate", yggPayload{
		"accessToken": selStr(accessToken, y.accessToken),
		"clientToken": selStr(clientToken, y.clientToken),
	})
	if err != nil {
		return nil
	}
	return yggError(resp)
}

func (y *YggAuth) Validate(accessToken string) error {
	// Check if an access token is valid
	// Returns nil on success (ie valid access token), otherwise error
	resp, err := y.request("/validate", yggPayload{
		"accessToken": selStr(accessToken, y.accessToken),
	})
	if err != nil {
		return err
	}
	return yggError(resp)
}

func (y *YggAuth) request(endpoint string, payload interface{}) (*YggResponse, error) {
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
	r := new(YggResponse)
	if err = json.NewDecoder(resp.Body).Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (y *YggAuth) update(resp *YggResponse) (*YggResponse, error) {
	y.accessToken = resp.AccessToken
	y.clientToken = resp.ClientToken
	return resp, nil
}

type Auth struct {
}
