package boatsocks

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

//Server 对象,用于服务端
type Server struct {
	LocalPort string
	Cypher    *Cypher
}

//NewServer 新建Server对象
func NewServer(localPort string, password string) *Server {
	return &Server{
		LocalPort: localPort,
		Cypher:    NewCypher(password),
	}
}

//Link 建立和本地连接,以及和Server端连接,本地触发
func (s *Server) Link() error {
	listen, err := net.Listen("tcp", "0.0.0.0:"+s.LocalPort)
	if err != nil {
		return err
	}
	defer listen.Close()
	for {
		localConn, err := listen.Accept()
		if err != nil {
			continue
		}
		//将连接打包
		local := NewSocks(localConn, s.Cypher)
		go s.Handle(local)

	}
}

//Handle 处理服务端的local和server
func (s *Server) Handle(local *Socks) {
	// defer local.Conn.Close()
	//解socks5协议
	buf := make([]byte, 256)

	/**
	     The localConn connects to the dstServer, and sends a ver
	     identifier/method selection message:
	  			  +----+----------+----------+
	  			  |VER | NMETHODS | METHODS  |
	  			  +----+----------+----------+
	  			  | 1  |    1     | 1 to 255 |
	  			  +----+----------+----------+
	     The VER field is set to X'05' for this ver of the protocol.  The
	     NMETHODS field contains the number of method identifier octets that
	     appear in the METHODS field.
	*/
	// 第一个字段VER代表Socks的版本，Socks5默认为0x05，其固定长度为1个字节
	_, err := local.DecodeRead(buf)
	// 只支持版本5
	if err != nil || buf[0] != 0x05 {
		return
	}

	/**
	     The dstServer selects from one of the methods given in METHODS, and
	     sends a METHOD selection message:

	  			  +----+--------+
	  			  |VER | METHOD |
	  			  +----+--------+
	  			  | 1  |   1    |
	  			  +----+--------+
	*/
	// 不需要验证，直接验证通过
	local.EncodeWrite([]byte{0x05, 0x00})

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/

	// 获取真正的远程服务的地址
	n, err := local.DecodeRead(buf)

	// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
	if err != nil || n < 7 {
		return
	}
	fmt.Println("第二次验证")

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
	fmt.Println(remoteAddr)
	fmt.Println(remotePort)
	serverConn, err := net.Dial("tcp", remoteAddr+":"+remotePort)
	if err != nil {
		return
	}
	local.EncodeWrite([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	server := NewSocks(serverConn, s.Cypher)

	//新版
	//进行转发
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			err = local.DecodeCopy(server.Conn)
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
		err = server.EncodeCopy(local.Conn)
		if err != nil {
			local.Conn.Close()
			server.Conn.Close()
			fmt.Println(err)
			break
		}
	}

	wg.Wait()

}
