package boatsocks

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

//Local 对象,用于本地端
type Local struct {
	LocalPort  string
	ServerPort string
	ServerAddr string
	Cypher     *Cypher
}

//NewLocal 新建Local对象
func NewLocal(localPort string, serverPort string, serverAddr string, password string) *Local {
	return &Local{
		LocalPort:  localPort,
		ServerPort: serverPort,
		ServerAddr: serverAddr,
		Cypher:     NewCypher(password),
	}
}

//Link 建立和本地连接,以及和server端连接,本地触发
func (l *Local) Link() error {
	listen, err := net.Listen("tcp", "localhost:"+l.LocalPort)

	if err != nil {
		return err
	}
	defer listen.Close()
	for {
		localConn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		//将连接打包
		local := NewSocks(localConn, l.Cypher)
		//在本地处理local和server
		go l.Handle(local)

	}

}

//Handle local和server进行信息传输
func (l *Local) Handle(local *Socks) {

	//解socks5协议
	buf := make([]byte, 256)

	/**

	  +----+----------+----------+
	  |VER | NMETHODS | METHODS  |
	  +----+----------+----------+
	  | 1  |    1     | 1 to 255 |
	  +----+----------+----------+

	*/
	// 第一个字段VER代表Socks的版本，Socks5默认为0x05，其固定长度为1个字节
	_, err := local.Conn.Read(buf)
	// 只支持版本5
	if err != nil || buf[0] != 0x05 {
		return
	}

	/**


	  +----+--------+
	  |VER | METHOD |
	  +----+--------+
	  | 1  |   1    |
	  +----+--------+
	*/
	// 不需要验证，直接验证通过
	local.Conn.Write([]byte{0x05, 0x00})

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/

	// 获取真正的远程服务的地址
	n, err := local.Conn.Read(buf)

	// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
	if err != nil || n < 7 {
		return
	}

	// CMD代表客户端请求的类型，值长度也是1个字节，有三种类型
	// CONNECT X'01'
	if buf[1] != 0x01 {
		// 目前只支持 CONNECT
		return
	}

	var dIP []byte
	// aType 代表请求的远程服务器地址类型，值长度1个字节，有三种类型
	switch buf[3] {
	case 0x01:
		//	IP V4 address: X'01'
		dIP = buf[4 : 4+net.IPv4len]
	case 0x03:

		dIP = buf[5 : n-2]
	case 0x04:
		//	IP V6 address: X'04'
		dIP = buf[4 : 4+net.IPv6len]
	default:
		return
	}

	remotePort := strconv.Itoa(int(buf[n-2])<<8 | int(buf[n-1]))
	remoteAddr := string(dIP)

	if err != nil {
		return
	}
	local.Conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	serverConn, err := net.Dial("tcp", l.ServerAddr+":"+l.ServerPort)
	if err != nil {
		fmt.Println(err)
		return
	}
	server := NewSocks(serverConn, l.Cypher)
	defer server.Conn.Close()
	str := remoteAddr + ":" + remotePort
	fmt.Println(str)
	server.EncodeWrite([]byte(str))

	//进行转发
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			err = server.DecodeCopy(local.Conn)
			if err != nil {
				local.Conn.Close()
				server.Conn.Close()
				fmt.Println(err)
				break
			}
		}

	}()

	defer wg.Done()
	for {
		err = local.EncodeCopy(server.Conn)
		if err != nil {
			local.Conn.Close()
			server.Conn.Close()
			fmt.Println(err)
			break
		}
	}

	wg.Wait()

}
