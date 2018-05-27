package lines

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type dynamodbLinesRepo struct {
	svc *dynamodb.DynamoDB
}

func NewRepo() *dynamodbLinesRepo {
	s := session.Must(session.NewSession())
	return &dynamodbLinesRepo{dynamodb.New(s, aws.NewConfig().WithRegion("eu-west-1"))}
}

func NewRepoWithClient(svc *dynamodb.DynamoDB) *dynamodbLinesRepo {
	return &dynamodbLinesRepo{svc}
}

func (r dynamodbLinesRepo) GetLinesFor(id string) ([]string, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String("slack-users"),
	}

	fmt.Println("Calling for", input)
	result, err := r.svc.GetItem(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return nil, aerr
		} else {
			return nil, fmt.Errorf(err.Error())
		}
	}

	if len(result.Item) == 0 {
		return nil, fmt.Errorf("UserNotFound")
	}

	var subscribedLines []string
	dynamodbattribute.Unmarshal(result.Item["subscribedLines"], &subscribedLines)
	return subscribedLines, nil
}
