package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/zsichen/oxfordapi/core"
)

var (
	appid  = os.Getenv("app_id")
	appkey = os.Getenv("app_key")
	lang   = os.Getenv("lang")
)

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestStr, _ := json.Marshal(request)
	log.Printf("request:%s\n", string(requestStr))
	param := struct {
		Word string `json:"word"`
		Mode string `json:"mode"`
	}{}

	var ok bool
	if param.Word, ok = request.QueryStringParameters["word"]; !ok {
		return events.APIGatewayProxyResponse{Body: "Missing require parameter", StatusCode: 418}, nil
	}
	param.Mode = request.QueryStringParameters["mode"]
	res, err := core.OxfordAPIRequest(appid, appkey, lang, param.Word)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 418}, err
	}
	if param.Mode == "origin" {
		return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
	}
	tmp := &core.AutoGenerated{}
	err = json.Unmarshal(res, &tmp)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 418}, err
	}
	neatStr, _ := json.Marshal(core.NeatAutoGenerated(tmp))
	return events.APIGatewayProxyResponse{Body: string(neatStr), StatusCode: 200}, nil
}

func main() {
	lambda.Start(handleRequest)
}
