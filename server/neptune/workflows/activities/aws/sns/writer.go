package sns

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	awsSns "github.com/aws/aws-sdk-go/service/sns"
)

type snsPublisher interface {
	Publish(*sns.PublishInput) (*sns.PublishOutput, error)
}

type Writer struct {
	Client   snsPublisher
	TopicArn *string
}

func (w *Writer) Write(payload []byte) (int, error) {
	if _, err := w.Client.Publish(&awsSns.PublishInput{
		Message:  aws.String(string(payload)),
		TopicArn: w.TopicArn,
	}); err != nil {
		return 0, err
	}

	return len(payload), nil
}
