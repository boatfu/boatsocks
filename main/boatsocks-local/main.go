package main

import (
	"boatsocks"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//Config 从本地导入配置文件
type Config struct {
	LocalPort  string
	ServerPort string
	ServerAddr string
	Password   string
}

func main() {
	//读取config文件
	bs, err := ioutil.ReadFile("C:\\Users\\boate\\go\\src\\boatsocks\\config\\localConfig.json")

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
	//执行
	local := boatsocks.NewLocal(config.LocalPort, config.ServerPort, config.ServerAddr, config.Password)
	err = local.Link()
	if err != nil {
		fmt.Println(err)
		return
	}
}
