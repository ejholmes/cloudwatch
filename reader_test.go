package cloudwatch

import (
	"bytes"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	c := new(mockClient)
	r := &Reader{
		group:  aws.String("group"),
		stream: aws.String("1234"),
		client: c,
	}

	c.On("GetLogEvents", &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String("group"),
		LogStreamName: aws.String("1234"),
	}).Return(&cloudwatchlogs.GetLogEventsOutput{
		Events: []*cloudwatchlogs.OutputLogEvent{
			{Message: aws.String("Hello\n"), Timestamp: aws.Int64(1000)},
			{Message: aws.String("World"), Timestamp: aws.Int64(1000)},
		},
	}, nil)

	b := new(bytes.Buffer)
	n, err := io.Copy(b, r)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), n)
}
