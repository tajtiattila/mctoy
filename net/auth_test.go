package net

import (
	"crypto/sha1"
	"io"
	"math/rand"
	"testing"
	"time"
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

func TestRandReader(t *testing.T) {
	seed := int64(time.Now().Unix())
	rn := rand.New(rand.NewSource(seed))
	r1, r2 := rand.New(rand.NewSource(seed)), rand.New(rand.NewSource(seed))
	rr1, rr2 := NewRandReader(r1), NewRandReader(r2)
	nreads := 8
	v1, v2 := make([]int, nreads), make([]int, nreads)
	t1, t2 := make([]byte, 4096), make([]byte, 4096)
	for n := 0; n < 10; n++ {
		for i := range v1 {
			v1[i] = int(rn.Int31n(64)) + 1
		}
		for i, pi := range rand.Perm(len(v1)) {
			v2[pi] = v1[i]
		}
		for i := range v1 {
			l1, l2 := v1[i], v2[i]
			rr1.Read(t1[:l1])
			rr2.Read(t2[:l2])
		}
		cmp := 16
		rr1.Read(t1[:cmp])
		rr2.Read(t2[:cmp])
		eq := true
		for i := 0; i < cmp; i++ {
			if t1[i] != t2[i] {
				eq = false
			}
		}
		if !eq {
			t.Error("Random reader mismatch based on read buffer lengths")
		} else {
			t.Log("reading lengths OK:", v1, v2)
		}
	}
}
