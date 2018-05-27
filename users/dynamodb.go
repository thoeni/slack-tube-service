package users

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
	"github.com/aws/aws-sdk-go/aws/session"
)

type dynamodbUserRepo struct {
	svc *dynamodb.DynamoDB
}

func NewRepo() *dynamodbUserRepo {
	s := session.Must(session.NewSession())
	return &dynamodbUserRepo{dynamodb.New(s, aws.NewConfig().WithRegion("eu-west-1"))}
}

func NewRepoWithClient(svc *dynamodb.DynamoDB) *dynamodbUserRepo {
	return &dynamodbUserRepo{svc}
}

func (r *dynamodbUserRepo) PutNewSlackUser(id string, username string, subscribedLines []string) error {

	item, err := dynamodbattribute.MarshalMap(User{
		ID:              id,
		Username:        username,
		SubscribedLines: subscribedLines,
	})
	if err != nil {
		return errors.Wrap(err, "Something went wrong while marshalling the user")
	}

	ce := "attribute_not_exists(id)"
	rv := "ALL_OLD"

	_, err = r.svc.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String("slack-users"),
		ReturnValues:        &rv,
		Item:                item,
		ConditionExpression: &ce,
	})
	if err != nil {
		if ae, ok := err.(awserr.RequestFailure); ok && ae.Code() == "ConditionalCheckFailedException" {
			return errors.Wrap(err, "UserAlreadyExists")
		} else {
			return errors.Wrap(err, "Something went wrong while inserting the user")
		}
	}
	return nil
}

func (r *dynamodbUserRepo) UpdateExistingSlackUser(id string, username string, subscribedLines []string) error {
	idAv, _ := dynamodbattribute.Marshal(id)
	usernameAv, _ := dynamodbattribute.Marshal(username)
	subscribedLinesAv, _ := dynamodbattribute.Marshal(subscribedLines)
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{":valSubLines": subscribedLinesAv, ":username": usernameAv}
	ue := "set username = :username, subscribedLines = list_append(subscribedLines, :valSubLines)"

	_, err := r.svc.UpdateItem(&dynamodb.UpdateItemInput{
		TableName:                 aws.String("slack-users"),
		Key:                       map[string]*dynamodb.AttributeValue{"id": idAv},
		UpdateExpression:          &ue,
		ExpressionAttributeValues: expressionAttributeValues,
	})

	if err != nil {
		return errors.Wrap(err, "UpdateFailed")
	}
	return nil
}
