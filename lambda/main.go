package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/zsichen/oxfordapi/core"
)

var (
	appid  = os.Getenv("app_id")
	appkey = os.Getenv("app_key")
	lang   = os.Getenv("lang")
	table  = os.Getenv("table")
	svc    *dynamodb.DynamoDB
)

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc = dynamodb.New(sess)
}

type LetterItem struct {
	Id    string
	Value string
}

func GetLetterItem(id string) (val []byte, err error) {
	if table == "" {
		err = errors.New("dynamodb table name yet to setting")
		return
	}

	var result *dynamodb.GetItemOutput
	result, err = svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
	if result.Item == nil {
		err = errors.New("entry not found")
		return
	}
	item := &LetterItem{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		log.Println(err)
	}
	val = []byte(item.Value)
	return
}

func PutLetterItem(id, value string) (err error) {
	if table == "" {
		return
	}
	av, err := dynamodbattribute.MarshalMap(&LetterItem{
		Id:    id,
		Value: value,
	})
	if err != nil {
		return err
	}
	_, err = svc.PutItem(&dynamodb.PutItemInput{
		TableName: &table,
		Item:      av,
	})
	return
}

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

	res, err := GetLetterItem(param.Word)
	if err != nil {
		res, err = core.OxfordAPIRequest(appid, appkey, lang, param.Word)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 418}, err
		}
		defer func() {
			if ok {
				if err := PutLetterItem(param.Word, string(res)); err != nil {
					log.Println(err)
				}
			}
		}()
	}
	if param.Mode == "origin" {
		return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
	}
	tmp := &core.AutoGenerated{}
	err = json.Unmarshal(res, &tmp)
	if err != nil {
		ok = false
		return events.APIGatewayProxyResponse{StatusCode: 418}, err
	}
	neatStr, _ := json.Marshal(core.NeatAutoGenerated(tmp))
	return events.APIGatewayProxyResponse{Body: string(neatStr), StatusCode: 200}, nil
}

func main() {
	lambda.Start(handleRequest)
}
