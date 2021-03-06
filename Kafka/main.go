package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/Unknwon/goconfig"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
	"time"
)

type AutoGenerated struct {
	Timestamp time.Time `json:"@timestamp"`
	Metadata  struct {
		Beat    string `json:"beat"`
		Type    string `json:"type"`
		Version string `json:"version"`
		Topic   string `json:"topic"`
	} `json:"@metadata"`
	Input struct {
		Type string `json:"type"`
	} `json:"input"`
	Fields struct {
		LogTopic string `json:"log_topic"`
	} `json:"fields"`
	Ecs struct {
		Version string `json:"version"`
	} `json:"ecs"`
	Host struct {
		Architecture string `json:"architecture"`
		Os           struct {
			Codename string `json:"codename"`
			Platform string `json:"platform"`
			Version  string `json:"version"`
			Family   string `json:"family"`
			Name     string `json:"name"`
			Kernel   string `json:"kernel"`
		} `json:"os"`
		Name          string `json:"name"`
		ID            string `json:"id"`
		Containerized bool   `json:"containerized"`
		Hostname      string `json:"hostname"`
	} `json:"host"`
	Agent struct {
		Type        string `json:"type"`
		EphemeralID string `json:"ephemeral_id"`
		Hostname    string `json:"hostname"`
		ID          string `json:"id"`
		Version     string `json:"version"`
	} `json:"agent"`
	Log struct {
		Offset int `json:"offset"`
		File   struct {
			Path string `json:"path"`
		} `json:"file"`
	} `json:"log"`
	Message string `json:"message"`
}

type Messages struct {
	Timestamp            string `json:"@timestamp"`
	Version              string `json:"@version"`
	RemoteAddr           string `json:"remote_addr"`
	RemotePort           string `json:"remote_port"`
	ServerAddr           string `json:"server_addr"`
	ServerPort           string `json:"server_port"`
	HTTPXForwardedFor    string `json:"http_x_forwarded_for"`
	Host                 string `json:"host"`
	RequestLength        string `json:"request_length"`
	ContentLength        string `json:"content_length"`
	BodyBytesSent        string `json:"body_bytes_sent"`
	RequestURI           string `json:"request_uri"`
	HTTPUserAgent        string `json:"http_user_agent"`
	RequestMethod        string `json:"request_method"`
	HTTPCookie           string `json:"http_cookie"`
	HTTPReferer          string `json:"http_referer"`
	UpstreamAddr         string `json:"upstream_addr"`
	UpstreamStatus       string `json:"upstream_status"`
	UpstreamResponseTime string `json:"upstream_response_time"`
	RequestTime          string `json:"request_time"`
	Status               string `json:"status"`
}

var (
	wg sync.WaitGroup
)

func main() {
	// 加载配置
	ini, error := goconfig.LoadConfigFile("./app.ini")
	if error != nil {
		log.Error(error.Error())
	}
	con, error := ini.GetSection("config")
	if error != nil {
		log.Error(error.Error())
	}
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// 打印读取到的配置
	log.WithFields(log.Fields{"kafka": con["kafka"]}).Info("读取的kafka配置")
	log.WithFields(log.Fields{"topic": con["topic"]}).Info("读取的topic配置")
	log.WithFields(log.Fields{"USERNAME": con["USERNAME"]}).Info("读取的USERNAME配置")
	log.WithFields(log.Fields{"PASSWORD": con["PASSWORD"]}).Info("读取的PASSWORD配置")
	log.WithFields(log.Fields{"NETWORK": con["NETWORK"]}).Info("读取的NETWORK配置")
	log.WithFields(log.Fields{"SERVER": con["SERVER"]}).Info("读取的SERVER配置")
	log.WithFields(log.Fields{"DATABASE": con["DATABASE"]}).Info("读取的DATABASE配置")
	log.WithFields(log.Fields{"log": con["log"]}).Info("读取的log配置")

	// 设置日志格式
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	if con["log"] == "1" {
		log.WithFields(log.Fields{"日志等级": "info"}).Info("读取的log配置")
		log.SetLevel(log.InfoLevel)
	} else {
		log.WithFields(log.Fields{"日志等级": "error"}).Info("读取的log配置")
		log.SetLevel(log.ErrorLevel)
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:3306)/%s", con["USERNAME"], con["PASSWORD"], con["NETWORK"], con["SERVER"], con["DATABASE"])

	DB, err := sql.Open("mysql", dsn)

	if err != nil {

		log.Error("Open mysql failed,err:%v\n", err)

		return

	}
	// 给db设置一个超时时间，时间小于数据库的超时时间即可
	DB.SetConnMaxLifetime(100 * time.Second)

	// 用于设置最大打开的连接数，默认值为0表示不限制。
	DB.SetMaxOpenConns(100)

	// 用于设置闲置的连接数
	DB.SetMaxIdleConns(16)

	consumer, error := sarama.NewConsumer([]string{con["kafka"]}, config)
	if error != nil {
		log.WithFields(log.Fields{"连接kafka报错": "kafka"}).Error(err.Error())
	}

	partitionList, err := consumer.Partitions(con["topic"])
	if err != nil {
		log.WithFields(log.Fields{"获取topic异常": "topic"}).Error(err.Error())
	}
	for partition := range partitionList {
		//ConsumePartition方法根据主题，分区和给定的偏移量创建创建了相应的分区消费者
		//如果该分区消费者已经消费了该信息将会返回error
		//sarama.OffsetNewest:表明了为最新消息
		pc, err := consumer.ConsumePartition(con["topic"], int32(partition), sarama.OffsetNewest)
		if err != nil {
			log.WithFields(log.Fields{"获取topic异常2": "topic2"}).Error(err.Error())
		}
		defer pc.AsyncClose()
		wg.Add(1)
		go func(sarama.PartitionConsumer) {
			defer wg.Done()
			//Messages()该方法返回一个消费消息类型的只读通道，由代理产生
			for msg := range pc.Messages() {
				var data AutoGenerated
				var m Messages
				if err := json.Unmarshal([]byte(string(msg.Value)), &data); err != nil {
					log.WithFields(log.Fields{"第一层解析报错": "代码166行"}).Error(err.Error())
				} else {
					d := strings.ReplaceAll(data.Message, "\\x", "")
					if err := json.Unmarshal([]byte(d), &m); err != nil {
						log.WithFields(log.Fields{"第二层解析报错": "代码data.Message报错"}).Error(err.Error())
					} else {
						log.Info(data.Message)
						_, err := DB.Exec("INSERT INTO nginx(timestamps, version, remote_addr, remote_port, server_addr, server_port, http_x_forwarded_for, host_s, request_length, content_length, body_bytes_sent, request_uri, http_user_agent, request_method, http_cookie, http_referer, upstream_addr, upstream_status, upstream_response_time, request_time, statu_s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.Timestamp,
							m.Version, m.RemoteAddr, m.RemotePort, m.ServerAddr, m.ServerPort, m.HTTPXForwardedFor, m.Host, m.RequestLength, m.ContentLength, m.BodyBytesSent, m.RequestURI, m.HTTPUserAgent, m.RequestMethod, m.HTTPCookie, m.HTTPReferer, m.UpstreamAddr, m.UpstreamStatus, m.UpstreamResponseTime, m.RequestTime, m.Status)
						if err != nil {
							log.WithFields(log.Fields{"插入数据库异常": "mysql"}).Error(err.Error())
						}
					}
				}

			}
		}(pc)
	}
	wg.Wait()
	consumer.Close()
}
