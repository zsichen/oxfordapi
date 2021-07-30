package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var URL = func(lang, word string) string {
	return fmt.Sprintf("https://od-api.oxforddictionaries.com/api/v2/entries/%s/%s", lang, strings.ToLower(word))
}

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
		param := struct {
			Word string `form:"word"`
			Mode string `form:"mode"`
		}{}
		err := c.BindQuery(&param)
		if err != nil {
			c.JSON(418, gin.H{
				"error": err.Error(),
			})
			return
		}
		req, err := http.NewRequest(http.MethodGet, URL(cmd.Lang, param.Word), nil)
		if err != nil {
			c.JSON(418, gin.H{
				"error": err.Error(),
			})
			return
		}
		req.Header.Add("app_id", cmd.AppId)
		req.Header.Add("app_key", cmd.AppKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(418, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer resp.Body.Close()
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		if resp.StatusCode != 200 {
			c.JSON(418, gin.H{
				"error":  fmt.Sprintf("status %d", resp.StatusCode),
				"reason": buf.String(),
			})
			return
		}
		m := AutoGenerated{}
		dec := json.NewDecoder(&buf)
		err = dec.Decode(&m)
		if err != nil {
			c.JSON(418, gin.H{
				"error": err.Error(),
			})
			return
		}
		if param.Mode == "origin" {
			c.JSON(200, json.RawMessage(buf.Bytes()))
		} else {
			c.JSON(200, NeatAutoGenerated(&m))
		}
	})
	router.Run(":" + cmd.Port)
}
