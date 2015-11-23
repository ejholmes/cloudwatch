package cloudwatch

import (
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
)

func init() {
	now = func() time.Time {
		return time.Unix(1, 0)
	}
}

func TestWriter(t *testing.T) {
	c := new(mockClient)
	w := &Writer{
		group:  aws.String("group"),
		stream: aws.String("1234"),
		client: c,
	}

	c.On("PutLogEvents", &cloudwatchlogs.PutLogEventsInput{
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			{Message: aws.String("Hello\n"), Timestamp: aws.Int64(1000)},
			{Message: aws.String("World"), Timestamp: aws.Int64(1000)},
		},
		LogGroupName:  aws.String("group"),
		LogStreamName: aws.String("1234"),
	}).Return(&cloudwatchlogs.PutLogEventsOutput{}, nil)

	n, err := io.WriteString(w, "Hello\nWorld")
	assert.NoError(t, err)
	assert.Equal(t, 11, n)
}

func TestWriter_Rejected(t *testing.T) {
	c := new(mockClient)
	w := &Writer{
		group:  aws.String("group"),
		stream: aws.String("1234"),
		client: c,
	}

	c.On("PutLogEvents", &cloudwatchlogs.PutLogEventsInput{
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			{Message: aws.String("Hello\n"), Timestamp: aws.Int64(1000)},
			{Message: aws.String("World"), Timestamp: aws.Int64(1000)},
		},
		LogGroupName:  aws.String("group"),
		LogStreamName: aws.String("1234"),
	}).Return(&cloudwatchlogs.PutLogEventsOutput{
		RejectedLogEventsInfo: &cloudwatchlogs.RejectedLogEventsInfo{
			TooOldLogEventEndIndex: aws.Int64(2),
		},
	}, nil)

	_, err := io.WriteString(w, "Hello\nWorld")
	assert.Error(t, err)
	assert.IsType(t, &RejectedLogEventsInfoError{}, err)
}
