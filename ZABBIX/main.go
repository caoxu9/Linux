package main

import (
	Connection "./Connection"
	"fmt"
	"github.com/Unknwon/goconfig"
)

// 获取登陆身份验证
func Get_auth(url, user, password string) string {
	data := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "user.login",
		"params": map[string]string{
			"user":     user,
			"password": password,
		},
		"id":   1,
		"auth": nil,
	}
	result := Connection.Connection(url, data)
	return result.Result.(string)
}

// 获取applicationid
func Get_applicationid(url, auth, hostid, name string) string {
	data := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "application.get",
		"params": map[string]string{
			"output":    "extend",
			"hostids":   hostid,
			"sortfield": "name",
		},
		"auth": auth,
		"id":   1,
	}
	result := Connection.Connection(url, data)
	for _, first := range result.Result.([] interface{}) {
		second, s := first.(map[string]interface{})
		if s {
			for key := range second {
				if second[key] == name {
					return second["applicationid"].(string)
				}
			}
		}
	}
	panic("获取applicationid出错!!!!")

}

// 获取hostid和Interfaceid
func Get_hostid_and_interfaceid(url, auth, IP string) (host, hostid, interfaceid string) {
	data := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "host.get",
		"params": map[string]interface{}{
			"output": []string{
				"hostid",
				"host",
			},
			"selectInterfaces": []string{
				"interfaceid",
				"ip",
			},
		},
		"id":   2,
		"auth": auth,
	}
	result := Connection.Connection(url, data)
	for _, first := range result.Result.([] interface{}) {
		second, s := first.(map[string]interface{})
		if s {
			for k := range second {
				if k == "interfaces" {
					three, o := second[k].([]interface{})
					if o {
						for _, t := range three {
							four, f := t.(map[string]interface{})
							if f {
								for key := range four {
									if key == "ip" {
										if four["ip"] == IP {
											return second["host"].(string), second["hostid"].(string), four["interfaceid"].(string)
										}
									}
								}
							}
						}
					}
				}

			}
		}
	}
	panic("无对应主机信息!!!!!!!!")
}

// 创建监听端口
func Create_listen(url, auth, name, hostid, port, interfaceid, applicationid string) {
	data := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "item.create",
		"params": map[string]interface{}{
			"name":         name,
			"key_":         fmt.Sprintf("net.tcp.listen[%s]", port),
			"hostid":       hostid,
			"type":         0,
			"value_type":   3,
			"interfaceid":  interfaceid,
			"applications": []string{applicationid},
			"delay":        "30s",
		},
		"auth": auth,
		"id":   1,
	}
	result := Connection.Connection(url, data)
	fmt.Println(result.Result)
	fmt.Printf("创建%s监控成功\n", name)
}

// 创建一个触发器
func Create_trigger(url, auth, name, IP, port, hostid string) {
	data := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "trigger.create",
		"params": map[string]interface{}{
			"description": fmt.Sprintf(name + "停止运行"),
			"expression":  fmt.Sprintf("{%s:net.tcp.listen[%s].last()}=0", IP, port),
			"hostids":     hostid,
			"priority":    5},
		"auth": auth,
		"id":   1,
	}
	response := Connection.Connection(url, data)
	fmt.Println(response.Result)
	fmt.Printf("添加%s触发器完成\n", name)
}
func main() {
	// 读取app.ini，读取相应的配置
	config, error := goconfig.LoadConfigFile("./app.ini")
	if error != nil {
		panic("配置文件加载错误!!!")
	}

	// 读取zabbix的的地址、用户名、密码
	Zabbix, error := config.GetSection("Zabbix")
	if error != nil {
		panic("读取zabbix配置文件出错！！！")
	}
	var url string = Zabbix["url"]
	var user string = Zabbix["user"]
	var password string = Zabbix["password"]
	var item string = Zabbix["item"]
	auth := Get_auth(url, user, password)

	IPconfig, error := goconfig.LoadConfigFile("./IP.ini")
	if error != nil {
		panic("配置文件加载错误!!!")
	}
	ips := IPconfig.GetSectionList()

	for _, value := range ips {
		//fmt.Println(value)
		session, error := IPconfig.GetSection(value)
		//fmt.Println(session)
		if error != nil {
			panic("读取ip列表出错！！！")
		}
		for port, name := range session {

			host, hostid, interfaceid := Get_hostid_and_interfaceid(url, auth, value)
			applicationid := Get_applicationid(url, auth, hostid, item)
			Create_listen(url, auth, name, hostid, port, interfaceid, applicationid)
			Create_trigger(url, auth, name, host, port, hostid)
		}

	}

}
