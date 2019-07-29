package boatsocks

import (
	"io"
)

//Socks 用于网络间的通信
type Socks struct {
	Conn   io.ReadWriteCloser
	Cypher *Cypher
}

//NewSocks 新建Socks对象
func NewSocks(conn io.ReadWriteCloser, cypher *Cypher) *Socks {
	return &Socks{
		Conn:   conn,
		Cypher: cypher,
	}
}

//EncodeWrite 编码发送
func (s *Socks) EncodeWrite(bs []byte) (int, error) {
	s.Cypher.Encode(bs)
	return s.Conn.Write(bs)

}

//DecodeRead 解码读取
func (s *Socks) DecodeRead(bs []byte) (n int, err error) {
	n, err = s.Conn.Read(bs)
	if err != nil {
		return
	}
	s.Cypher.Decode(bs[:n])
	return
}

//EncodeCopy 数据加密后复制到另一个Socks
func (s *Socks) EncodeCopy(dst io.ReadWriteCloser) error {
	buf := make([]byte, 256)
	n, err := s.Conn.Read(buf)
	if err != nil {
		return err
	}
	if n > 0 {
		newSocks := NewSocks(dst, s.Cypher)
		_, err = newSocks.EncodeWrite(buf[:n])
		if err != nil {
			return err
		}
	}
	return nil
}

//DecodeCopy 数据解密后复制到另一个Socks
func (s *Socks) DecodeCopy(dst io.ReadWriteCloser) error {
	buf := make([]byte, 256)
	n, err := s.DecodeRead(buf)
	if err != nil {
		return err
	}
	if n > 0 {
		newSocks := NewSocks(dst, s.Cypher)

		_, err = newSocks.Conn.Write(buf[:n])
		if err != nil {
			return err
		}

	}
	return nil
}
