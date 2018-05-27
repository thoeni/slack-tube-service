package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/thoeni/slack-tube-service/lines"
	"github.com/thoeni/slack-tube-service/tfl"
	"github.com/thoeni/slack-tube-service/users"
)

var tokenStore TokenRepository
var svc *dynamodb.DynamoDB

type tubeServuceLambda struct {
	tfl       tfl.Service
	userRepo  users.Repo
	linesRepo lines.Repo
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	l := NewLambda()

	query := request.Body
	v, err := url.ParseQuery(query)

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	switch request.Path {
	case "/api/slack/tubestatus/":
		return slackRequestHandler(l, request.HTTPMethod, v)
	case "/api/slack/token/":
		token := strings.Replace(request.Path, "/api/slack/token/", "", -1)
		return slackTokenRequestHandler(request.HTTPMethod, token, v)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
		}, nil
	}
}

func main() {
	lambda.Start(handler)
}

func NewLambda() *tubeServuceLambda {
	// DynamoDB
	sess := session.Must(session.NewSession())
	svc = dynamodb.New(sess, aws.NewConfig().WithRegion("eu-west-1"))

	return &tubeServuceLambda{
		tfl:       tfl.NewService(),
		userRepo:  users.NewRepoWithClient(svc),
		linesRepo: lines.NewRepoWithClient(svc),
	}
}
