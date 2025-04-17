package init

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"log"
)

// awsConfig contains aws config and clients for aws services
type awsConfig struct {
	config        *aws.Config
	cognitoClient *cognitoidentityprovider.Client
}

func (aws *awsConfig) GetCognitoClient() *cognitoidentityprovider.Client {
	return aws.cognitoClient
}

// createAWSConfig creates new aws config
func (a *app) createAWSConfig() *awsConfig {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	log.Println("connected to aws")
	return &awsConfig{
		config:        &cfg,
		cognitoClient: a.createCognitoClient(&cfg),
	}
}

func (a *app) createCognitoClient(cfg *aws.Config) *cognitoidentityprovider.Client {
	return cognitoidentityprovider.NewFromConfig(*cfg)
}