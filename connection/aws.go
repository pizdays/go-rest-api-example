package connection

import (
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	awsOnce sync.Once
	awsSess *session.Session
)

// GetAWSSession creates AWS session once. Any further calls
// returns cached session.
func GetAWSSession() *session.Session {
	awsOnce.Do(func() {
		// The session the S3 Uploader will use
		awsSess = session.Must(session.NewSession(
			&aws.Config{
				Region: aws.String(os.Getenv("AWS_REGION")),
				Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"),
					os.Getenv("AWS_SECRET_ACCESS_KEY"),
					os.Getenv("AWS_SESSION_TOKEN")),
			}))
	})

	return awsSess
}
