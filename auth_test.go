package main

import (
	"crypto/sha1"
	"io"
	"testing"
)

// AuthDigest computes a special SHA-1 digest required for Minecraft web
// authentication on Premium servers (online-mode=true).
// Source: http://wiki.vg/Protocol_Encryption#Server
//
// Also many, many thanks to SirCmpwn and his wonderful gist (C#):
// https://gist.github.com/SirCmpwn/404223052379e82f91e6
func AuthDigest(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return McDigest(h.Sum(nil))
}

var examples = map[string]string{
	"Notch": "4ed1f46bbe04bc756bcb17c0c7ce3e4632f06a48",
	"jeb_":  "-7c9d5b0044c130109a5d7b5fb5c317c02b4e28c1",
	"simon": "88e16a1019277b15d58faf0541e11910eb756f6",
}

func TestMcDigest(t *testing.T) {
	for k, v := range examples {
		if AuthDigest(k) != v {
			t.Errorf("AutDigest mismatch for %s:\n%s\n%s\n\n", k, AuthDigest(k), v)
		} else {
			t.Logf("AutDigest for %s is OK\n", k)
		}
	}
}
