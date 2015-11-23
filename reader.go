package cloudwatch

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// Reader is an io.Reader implementation that streams log lines from cloudwatch
// logs.
type Reader struct {
	b bytes.Buffer

	group, stream, nextToken *string

	client client
}

func NewReader(group, stream string, client *cloudwatchlogs.CloudWatchLogs) *Reader {
	return &Reader{
		group:  aws.String(group),
		stream: aws.String(stream),
		client: client,
	}
}

func (r *Reader) Read(b []byte) (int, error) {
	if r.b.Len() > 0 {
		return r.b.Read(b)
	}

	resp, err := r.client.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  r.group,
		LogStreamName: r.stream,
		NextToken:     r.nextToken,
	})
	if err != nil {
		return 0, err
	}

	r.nextToken = resp.NextForwardToken

	for _, event := range resp.Events {
		r.b.WriteString(*event.Message)
	}

	if r.nextToken == nil {
		n, err := r.b.Read(b)
		if err != nil {
			return n, err
		}
		return n, io.EOF
	}

	return r.b.Read(b)
}
