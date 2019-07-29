package boatsocks

import (
	"fmt"
	"net"
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
	// defer local.Conn.Close()

	remoteConn, err := net.Dial("tcp", l.ServerAddr+":"+l.ServerPort)

	if err != nil {
		fmt.Println(err)
		return
	}
	server := NewSocks(remoteConn, l.Cypher)
	// defer server.Conn.Close()

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
