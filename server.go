package boatsocks

import (
	"fmt"
	"net"
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
	buf := make([]byte, 256)
	n, err := local.DecodeRead(buf)
	if err != nil {
		return
	}
	//打印真实的地址
	str := string(buf[:n])
	if len(str) > 30 {
		return
	}
	
	remoteServer, err := net.Dial("tcp", str)
	if err != nil {
		fmt.Println(err)
		return
	}
	server := NewSocks(remoteServer, s.Cypher)

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
