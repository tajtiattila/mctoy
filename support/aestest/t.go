package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

type test struct {
	sharedSecret string
	response     string
}

/*
serverId:
sharedSecret: ce83fbd62b5675f7315d5ef0dbf2251a
publicKey: 30819f300d06092a864886f70d010101050003818d0030818902818100a1439bc1dd595ca382bcf72c430101e914a5a6b4b8e7d92283e5b61becf320abd75f66d784fecb841a5067fc975b3f3dbf9e31effba5135e01e612409f1dfccbacfc82fb57e0d686e8d20ccb31780702acb09550f4f1c6bdabbc6663899fe5879660ea037975246d1e87f1f8b5857f59ca696dd7d0bb7c022a9c5bc729f64c090203010001
serverIdSum: 2cd0aa20fcab24c2851fe8ef1b25b9b29823327f
sessionserver ok
-> 263/263
    85 02 01 00 80 64 be 2e ad ff 13 97 a8 b9 94 dc   .....d..........
    0f 96 4f 61 1a 13 3c 21 a8 c3 81 c6 ee bc 17 7a   ..Oa..<!.......z
    34 57 f5 65 2a 45 c5 82 4f 8e 7e 0b ac 41 4b e9   4W.e*E..O.~..AK.
    2a 8a 21 7b 2c cd e9 be 3c 51 ab 9c 1c e5 aa 44   *.!{,...<Q.....D
    81 54 64 af b8 1e c7 cd 30 21 3f 9f aa 1c ea c2   .Td.....0!?.....
    67 cd 42 65 b8 02 e5 5d 5b 58 2e 79 28 a8 e6 c9   g.Be...][X.y(...
    b0 c9 15 ad 11 4a c9 46 a5 d8 bc f4 69 4a 6f 71   .....J.F....iJoq
    e0 e1 70 b4 e5 1e a2 8b e4 f4 02 1b 72 3a 8a ce   ..p.........r:..
    1b 66 b6 ad 63 00 80 76 94 60 59 46 db cb e7 c8   .f..c..v.`YF....
    d7 77 a6 9b 41 49 4e 56 06 40 5e c1 23 36 13 d5   .w..AINV.@^.#6..
    c4 9b a4 b3 98 4a d6 53 e5 b1 4d 9d e9 0f 53 4b   .....J.S..M...SK
    90 51 3b 27 c3 ba a5 6d 7e 7a 76 26 a3 eb f6 f2   .Q;'...m~zv&....
    0d b3 5b 2a 1c 9e c5 65 8e 4d 20 d7 3c b0 d2 06   ..[*...e.M .<...
    87 64 79 06 c1 44 ce fb e0 96 08 d4 be 19 37 b3   .dy..D........7.
    08 13 44 41 2f f6 f7 76 4f c1 f3 db d9 1e 59 43   ..DA/..vO.....YC
    9b 9d 38 9b 44 4e be 25 fa c6 f4 2b b5 ea aa b5   ..8.DN.%...+....
...7more bytes
<- 51/4096
    f2 7f 51 27 c6 a6 0b ea 48 53 e6 8c bd 07 04 22   ..Q'....HS....."
    28 a2 b1 d6 22 e6 9e b8 7d 79 29 16 9d d8 1a 32   (..."...}y)....2
    d6 69 73 27 8e 5b 14 0b ff bf 54 18 2a 88 ae 36   .is'.[....T.*..6
    0e a1 34                                          ..4
<- 18/16321
    da d2 06 e2 44 bb 2a 14 04 b4 55 fc 4a 59 19 04   ....D.*...U.JY..
    2b c8                                             +.
<- 333/16303
    75 15 5d 3a 61 c0 b9 c1 18 31 52 b4 c6 23 f5 bc   u.]:a....1R..#..
    ee e9 ab e2 b9 aa 82 2b 2c a7 a9 17 3c 8b 8a a1   .......+,...<...
    eb 01 c8 cd c6 dd ff 67 63 4e fd 5f b4 84 86 bb   .......gcN._....
    d9 f0 72 9e 55 da fd 7d ec c5 20 44 ec 73 a7 bc   ..r.U..}.. D.s..
    dd c6 4e bd 0e 96 b3 0a 67 b1 8b 80 ed f4 3c 11   ..N.....g.....<.
    9f 6d 73 da 70 18 80 34 77 6e 1a 62 2e b9 45 14   .ms.p..4wn.b..E.
    0e d6 4b 19 5e be 23 45 e5 3c 0d 2e 80 1e de 26   ..K.^.#E.<.....&
    23 57 16 2c eb dc 77 88 b4 15 96 0c b1 34 39 b3   #W.,..w......49.
    d0 77 b6 60 8b 51 ca 10 69 22 4c 21 a4 3f 97 06   .w.`.Q..i"L!.?..
    18 19 ce 37 8a 0d 84 6d f2 ec c9 23 35 b2 27 50   ...7...m...#5.'P
    14 cc 82 ad 83 db cf 27 31 67 6a 8c 38 47 88 93   .......'1gj.8G..
    da fe 9e d6 0b 5e 52 90 0b 2d 9e 19 de 20 0b 2c   .....^R..-... .,
    3b 0b 0a 5c 81 53 19 71 48 5c 24 0d a9 77 4e 3e   ;..\.S.qH\$..wN>
    2e 5b 3b 24 c7 77 e2 d6 01 32 7c 75 9c 46 1e d0   .[;$.w...2|u.F..
    3d b6 9e 7a 15 6f d9 de f5 9e 70 1c 9d e5 72 04   =..z.o....p...r.
    c9 8a b9 be a1 2d 24 e9 10 28 a6 e3 68 73 c3 f2   .....-$..(..hs..
*/
var t0 = []test{
	{
		sharedSecret: "ce83fbd62b5675f7315d5ef0dbf2251a",
		response: `
		f2 7f 51 27 c6 a6 0b ea 48 53 e6 8c bd 07 04 22   ..Q'....HS....."
		28 a2 b1 d6 22 e6 9e b8 7d 79 29 16 9d d8 1a 32   (..."...}y)....2
		d6 69 73 27 8e 5b 14 0b ff bf 54 18 2a 88 ae 36   .is'.[....T.*..6
		0e a1 34                                          ..4
		da d2 06 e2 44 bb 2a 14 04 b4 55 fc 4a 59 19 04   ....D.*...U.JY..
		2b c8                                             +.
		75 15 5d 3a 61 c0 b9 c1 18 31 52 b4 c6 23 f5 bc   u.]:a....1R..#..
		ee e9 ab e2 b9 aa 82 2b 2c a7 a9 17 3c 8b 8a a1   .......+,...<...
		eb 01 c8 cd c6 dd ff 67 63 4e fd 5f b4 84 86 bb   .......gcN._....
		d9 f0 72 9e 55 da fd 7d ec c5 20 44 ec 73 a7 bc   ..r.U..}.. D.s..
		dd c6 4e bd 0e 96 b3 0a 67 b1 8b 80 ed f4 3c 11   ..N.....g.....<.
		9f 6d 73 da 70 18 80 34 77 6e 1a 62 2e b9 45 14   .ms.p..4wn.b..E.
		0e d6 4b 19 5e be 23 45 e5 3c 0d 2e 80 1e de 26   ..K.^.#E.<.....&
		23 57 16 2c eb dc 77 88 b4 15 96 0c b1 34 39 b3   #W.,..w......49.
		d0 77 b6 60 8b 51 ca 10 69 22 4c 21 a4 3f 97 06   .w. .Q..i"L!.?..
		18 19 ce 37 8a 0d 84 6d f2 ec c9 23 35 b2 27 50   ...7...m...#5.'P
		14 cc 82 ad 83 db cf 27 31 67 6a 8c 38 47 88 93   .......'1gj.8G..
		da fe 9e d6 0b 5e 52 90 0b 2d 9e 19 de 20 0b 2c   .....^R..-... .,
		3b 0b 0a 5c 81 53 19 71 48 5c 24 0d a9 77 4e 3e   ;..\.S.qH\$..wN>
		2e 5b 3b 24 c7 77 e2 d6 01 32 7c 75 9c 46 1e d0   .[;$.w...2|u.F..
		3d b6 9e 7a 15 6f d9 de f5 9e 70 1c 9d e5 72 04   =..z.o....p...r.
		c9 8a b9 be a1 2d 24 e9 10 28 a6 e3 68 73 c3 f2   .....-$..(..hs..
`,
	},
}

func chk(err error, d ...interface{}) {
	if err != nil {
		fmt.Println(err)
		fmt.Println(d...)
		panic(err)
	}
}

func show(b []byte) {
	d := MakeDumper(os.Stdout)
	d.bytes(b)
}

func runtest(t *test) {
	buf := make([]byte, 0, 65536)
	for _, line := range strings.Split(t.response, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			line = strings.Replace(line[:len(line)-16], " ", "", -1)
			b, err := hex.DecodeString(line)
			chk(err, line)
			buf = append(buf, b...)
		}
	}
	out := make([]byte, len(buf))

	ss, err := hex.DecodeString(t.sharedSecret)
	chk(err, t.sharedSecret)
	aesc, err := aes.NewCipher(ss)
	chk(err, ss)
	cipher.NewCFBDecrypter(aesc, ss).XORKeyStream(out, buf)
	show(out)
	cipher.NewCTR(aesc, ss).XORKeyStream(out, buf)
	show(out)
	cipher.NewOFB(aesc, ss).XORKeyStream(out, buf)
	show(out)
}

func main() {
	for _, t := range t0 {
		runtest(&t)
	}
}
