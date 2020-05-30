package Connection

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Auth struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Id      int         `json:"id"`
}

// 获取zabbix的auth
func Connection(url string, data map[string]interface{}) Auth {
	client := &http.Client{}

	params, error := json.Marshal(data)
	request, error := http.NewRequest("POST", url, bytes.NewBuffer(params))

	if error != nil {
		log.Printf("Connection的request出错:%s\n", error.Error())
	}
	request.Header.Set("Content-Type", "application/json-rpc")
	req, error := client.Do(request)
	defer req.Body.Close()
	if error != nil {
		log.Printf("Connection的client出错:%s\n", error.Error())
	}
	config, error := ioutil.ReadAll(req.Body)

	if error != nil {
		log.Printf("Connection的Get_Auth的读取出错:%s\n", error.Error())
	}
	// fmt.Println(string(config))
	var auth Auth
	error = json.Unmarshal(config, &auth)
	if error != nil {
		log.Printf("Connection的Get_Auth的json解析出错:%s\n", error.Error())
	}
	return auth

}
