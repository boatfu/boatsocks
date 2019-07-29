package boatsocks

import (
	"encoding/base64"
	"fmt"
)

const byteLength = 256

//Cypher 用于加密解密
type Cypher struct {
	encodeMap [byteLength]byte
	decodeMap [byteLength]byte
}

//NewCypher 注意字符串必须为256位
func NewCypher(pd string) *Cypher {
	password, err := base64.StdEncoding.DecodeString(pd)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var encodeMap [byteLength]byte
	var decodeMap [byteLength]byte
	for i, v := range password {
		encodeMap[i] = v
		decodeMap[v] = byte(i)
	}
	return &Cypher{
		encodeMap: encodeMap,
		decodeMap: decodeMap,
	}
}

// Encode 加密数据
func (c *Cypher) Encode(bs []byte) {
	for i, v := range bs {
		bs[i] = c.encodeMap[v]
	}
}

//Decode 解密数据
func (c *Cypher) Decode(bs []byte) {
	for i, v := range bs {
		bs[i] = c.decodeMap[v]
	}
}
