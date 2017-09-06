package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

type slackUserItem struct {
	ID              string   `dynamodbav:"id""`
	Username        string   `dynamodbav:"username""`
	SubscribedLines []string `dynamodbav:"subscribedLines""`
}

func getLinesFor(id string) ([]string, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String("slack-users"),
	}

	result, err := svc.GetItem(input)

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

func putNewSlackUser(id string, username string, subscribedLines []string) error {

	item, err := dynamodbattribute.MarshalMap(slackUserItem{
		ID:              id,
		Username:        username,
		SubscribedLines: subscribedLines,
	})
	if err != nil {
		return errors.Wrap(err, "Something went wrong while marshalling the user")
	}

	ce := "attribute_not_exists(id)"
	rv := "ALL_OLD"

	_, err = svc.PutItem(&dynamodb.PutItemInput{
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

func updateExistingSlackUser(id string, username string, subscribedLines []string) error {
	idAv, _ := dynamodbattribute.Marshal(id)
	usernameAv, _ := dynamodbattribute.Marshal(username)
	subscribedLinesAv, _ := dynamodbattribute.Marshal(subscribedLines)
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{":valSubLines": subscribedLinesAv, ":username": usernameAv}
	ue := "set username = :username, subscribedLines = list_append(subscribedLines, :valSubLines)"

	_, err := svc.UpdateItem(&dynamodb.UpdateItemInput{
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
