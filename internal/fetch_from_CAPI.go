package internal

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/comprehend"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
)

func HelloWorld() string {
	fmt.Print("hello")
	return "Hello world"
}

type ArticleFields struct {
	Headline string `json:"headline"`
	Byline   string `json:"byline"`
	BodyText string `json:"bodyText"`
}

type EntityResult struct {
	Entity string
}

func GetArticleFieldsFromPath(path string, apiKey string) *ArticleFields {
	urlPrefix := "https://content.guardianapis.com"
	urlSuffix := "?api-key=" + apiKey + "&show-fields=byline,bodyText,headline"
	resp, err := http.Get(urlPrefix + path + urlSuffix)
	if err != nil {
		panic("no response from CAPI")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("no response from CAPI")
	}

	fields := gjson.Get(string(body), "response.content.fields").Raw
	fieldsBytes := []byte(fields)
	var p = new(ArticleFields)
	articleFieldsError := json.Unmarshal(fieldsBytes, &p)
	if articleFieldsError != nil {
		panic(articleFieldsError)
	}
	return p
}

func GetEntities(client *comprehend.Comprehend, bodyText string) ([]*comprehend.Entity, error) {
	input := &comprehend.DetectEntitiesInput{}
	input.SetText(bodyText)
	input.SetLanguageCode("en")
	result, err := client.DetectEntities(input)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get entities")
	}

	return result.Entities, nil
}

func CreateComprehendClient(profile string) (*comprehend.Comprehend, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})

	if err != nil {
		return nil, errors.Wrap(err, "unable to create new sessions")
	}

	sess.Config.Credentials = credentials.NewCredentials(
		&credentials.SharedCredentialsProvider{
			Profile: profile,
		},
	)

	if _, err := sess.Config.Credentials.Get(); err != nil {
		return nil, errors.Wrap(err, "unable to get credentials")
	}

	return comprehend.New(sess), nil
}

func GetEntitiesForPath(path string) []*comprehend.Entity {
	articleFields := GetArticleFieldsFromPath(path, "test")
	client, err := CreateComprehendClient("developerPlayground")
	if err != nil {
		panic("help!")
	}
	res, err := GetEntities(client, articleFields.BodyText)

	if err != nil {
		panic("help!")
	}

	// value only
	for _, entity := range res {
		fmt.Println(entity.GoString())
	}
	return res
}