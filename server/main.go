package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/zsichen/oxfordapi/core"
)

var cmd = struct {
	AppId  string `json:"app_id"`
	AppKey string `json:"app_key"`
	Port   string `json:"port"`
	Lang   string `json:"lang"`
	Config string `json:"-"`
}{}

func LoadConfig() error {
	buf, err := ioutil.ReadFile(cmd.Config)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, &cmd)
}

func main() {
	flag.StringVar(&cmd.Config, "config", "config.json", "config file")
	flag.StringVar(&cmd.Port, "port", "8080", "listen port")
	flag.StringVar(&cmd.Lang, "lang", "en-gb", "language code")
	flag.Parse()

	if err := LoadConfig(); err != nil {
		log.Fatalln(err)
	}

	router := gin.Default()
	router.GET("/quest", func(c *gin.Context) {
		core.Handler(c, cmd.AppId, cmd.AppKey, cmd.Lang)
	})
	router.Run(":" + cmd.Port)
}
