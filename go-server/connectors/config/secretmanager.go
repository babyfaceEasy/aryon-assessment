package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func NewSecretClient(env Env) *secretsmanager.SecretsManager {
	// Setup AWS session for LocalStack
	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(env.AWSCredentialsID, env.AWSCredentialsSecret, env.AWSCredentialsToken),
		Region:           aws.String(env.AWSRegion),
		Endpoint:         aws.String(env.AWSEndpoint), // Points to the LocalStack container
		S3ForcePathStyle: aws.Bool(env.AWSForcePathStyle),
	}
	sess := session.Must(session.NewSession(awsConfig))
	smClient := secretsmanager.New(sess)

	return smClient

	/*
		_, err = smClient.CreateSecretWithContext(context.Background(), &secretsmanager.CreateSecretInput{
			Name:         aws.String("slack-connector"),
			SecretString: aws.String("my-super-secret"),
		})
		if err != nil {
			slogger.Error("failed to create secret", "err", err)
		}
	*/
}
