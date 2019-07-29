package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// resp, err := http.Get("http://boatfu.top")
	content := "hello"
	resp, err := http.Post("http://127.0.0.1:7777", "application/json;charset=utf-8", bytes.NewBuffer([]byte(content)))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
