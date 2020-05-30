package main

import (
	Heketi "./Heketi"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func main() {
	heketi := &Heketi.Heketi{Url: "http://127.0.0.1:8080", User: "admin", Key: "My Secret"}
	heketi.NewClient()
	router := gin.Default()

	v1 := router.Group("/caoxu")
	{

		// 帮助文档
		v1.POST("/help", func(context *gin.Context) {
			var data map[string]string = make(map[string]string)
			data["/help"] = "帮助文档"
			data["/getClusterList"] = "查询集群ID"
			data["/getClusterInfo"] = "集群信息获取,根据id获取，参数id"
			data["/CreateCluster"] = "创建一个集群"
			data["/AddNode"] = "添加一个node，clusterId string, zone int, hostnameManage []string, hostnamesStorage []string, Tags map[string]string"
			data["/VolumeCreate"] = "创建一个卷,参数size int, name string, replica int, clustersIdList []string, block bool"
			context.JSON(200, data)
		})

		// 获取集群列表
		v1.POST("/getClusterList", func(context *gin.Context) {
			// 集群ID列表返回
			clusterList, error := heketi.ClusterList()
			var datas map[int]string = make(map[int]string)
			if error != nil {
				fmt.Println("集群列表查询失败:", error.Error())
				datas[0] = "当前无集群"
				context.JSON(200, datas)
			} else {
				for i := 0; i < len(clusterList); i++ {
					fmt.Printf("第%d个集群ID为%s\n", i+1, clusterList[i])
					datas[i+1] = clusterList[i]
				}
				context.JSON(200, datas)
			}
		})

		// 获取指定集群信息
		v1.POST("/getClusterInfo", func(context *gin.Context) {

			clusterId := context.PostForm("id")
			if clusterId == "" {
				context.JSON(400, map[string]string{"id": "请输入集群id"})
			} else {
				// 集群信息获取
				NodesList, VolumesList, BlockVolumesList, error := heketi.ClusterInfo(clusterId)
				if error != nil {
					context.JSON(400, map[string]interface{}{"error": error.Error()})
				} else {

					type ClusterInfo struct {
						Node    string
						Volumes struct {
							Volume []string
							Block  []string
						}
					}

					var datas ClusterInfo
					for i := 0; i < len(NodesList); i++ {
						for j := 0; j < len(VolumesList); j++ {
							for k := 0; k < len(BlockVolumesList); k++ {
								datas.Node = NodesList[i]
								datas.Volumes.Volume[j] = VolumesList[j]
								datas.Volumes.Block[k] = BlockVolumesList[k]
							}
						}
					}
					context.JSON(200, map[string]interface{}{
						"datas": datas,
					})
				}
			}
		})

		// 创建一个集群
		v1.POST("/CreateCluster", func(context *gin.Context) {
			// 创建一个集群,返回集群id
			clusterId, error := heketi.CreateCluster(true, true)
			if error != nil {
				fmt.Println("创建集群失败:", error.Error())
				context.JSON(400, map[string]interface{}{"error": error.Error()})
			} else {
				fmt.Println("新建集群ID为:", clusterId)
				context.JSON(200, map[string]interface{}{"新建集群ID": clusterId})
			}

		})

		// 创建一个卷
		v1.POST("/VolumeCreate", func(context *gin.Context) {
			size, error := strconv.Atoi(context.PostForm("size"))
			fmt.Println(size)
			if error != nil {
				fmt.Println("数字转换异常：", error.Error())
				context.JSON(400, map[string]string{
					"size": "size异常",
				})
			}
			replica, error := strconv.Atoi(context.PostForm("replica"))
			fmt.Println(replica)
			if error != nil {
				fmt.Println("数字转换异常：", error.Error())
				context.JSON(400, map[string]string{
					"replica": "replica异常",
				})
			}
			block, error := strconv.ParseBool(context.PostForm("block"))
			fmt.Println(block)
			if error != nil {
				fmt.Println("block转换异常：", error.Error())
				context.JSON(400, map[string]string{
					"block": "block异常",
				})
			}
			name := context.PostForm("name")
			fmt.Println(name)
			clustersIdList := context.PostForm("clustersIdList")
			fmt.Println(clustersIdList)
			if name == "" && clustersIdList == "" {
				context.JSON(400, map[string]string{
					"name or clustersIdList": "name or clustersIdList异常",
				})
			}
			clustersIdLists := strings.Split(clustersIdList, ",")
			fmt.Println(clustersIdLists)
			clusterId, error := heketi.VolumeCreate(size, name, replica, clustersIdLists, block)
			fmt.Println(error.Error())
			if error != nil {

				context.JSON(400, map[string]string{
					"error": "创建存储卷失败",
				})
			} else {
				context.JSON(200, map[string]string{
					"success": "创建存储卷成功" + clusterId,
				})
			}

		})

		// 添加一个node
		v1.POST("/AddNode", func(context *gin.Context) {

			clusterId := context.PostForm("clusterId")

			zone, error := strconv.Atoi(context.PostForm("zone"))
			fmt.Println(zone)
			if error != nil {
				fmt.Println("数字转换异常：", error.Error())
				context.JSON(400, map[string]string{
					"zone": "zone异常",
				})
			}

			hostnameManage := context.PostForm("hostnameManage")
			fmt.Println(hostnameManage)
			hostnameManages := strings.Split(hostnameManage, ",")
			fmt.Println(hostnameManages)

			hostnamesStorage := context.PostForm("hostnamesStorage")
			fmt.Println(hostnamesStorage)
			hostnamesStorages := strings.Split(hostnamesStorage, ",")
			fmt.Println(hostnamesStorages)

			datas, error := heketi.AddNode(clusterId, zone, hostnameManages, hostnamesStorages, map[string]string{})

			if error != nil {

				context.JSON(400, map[string]string{
					"error": "添加node失败",
				})
			} else {
				context.JSON(200, map[string]interface{}{
					"success": datas.DevicesInfo,
				})
			}

		})

	}

	router.Run("0.0.0.0:8000")

}
