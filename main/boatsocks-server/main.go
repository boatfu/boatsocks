package main

import (
	"boatsocks"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//Config 从本地导入配置文件
type Config struct {
	LocalPort string
	Password  string
}

func main() {
	//读取config文件
	bs, err := ioutil.ReadFile("/etc/gopath/src/boatsocks/config/serverConfig.json")
	//在windows中
	// bs, err := ioutil.ReadFile("C:\\Users\\boate\\go\\src\\boatsocks\\config\\serverConfig.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var config Config
	err = json.Unmarshal(bs, &config)
	if err != nil {
		fmt.Println(err)
		return
	}
	server := boatsocks.NewServer(config.LocalPort, config.Password)
	err = server.Link()
	if err != nil {
		fmt.Println(err)
		return
	}
}
