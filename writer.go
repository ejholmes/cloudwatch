package cloudwatch

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type RejectedLogEventsInfoError struct {
	Info *cloudwatchlogs.RejectedLogEventsInfo
}

func (e *RejectedLogEventsInfoError) Error() string {
	return fmt.Sprintf("log messages were rejected")
}

// Writer is an io.Writer implementation that writes lines to a cloudwatch logs
// stream.
type Writer struct {
	group, stream, sequenceToken *string

	client client
}

func NewWriter(group, stream string, client *cloudwatchlogs.CloudWatchLogs) *Writer {
	return &Writer{
		group:  aws.String(group),
		stream: aws.String(stream),
		client: client,
	}
}

func (w *Writer) Write(b []byte) (int, error) {
	r := bufio.NewReader(bytes.NewReader(b))

	var (
		n      int
		events []*cloudwatchlogs.InputLogEvent
		eof    bool
	)

	for !eof {
		b, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				eof = true
			} else {
				break
			}
		}

		if len(b) == 0 {
			continue
		}

		events = append(events, &cloudwatchlogs.InputLogEvent{
			Message:   aws.String(string(b)),
			Timestamp: aws.Int64(now().UnixNano() / 1000000),
		})

		n += len(b)
	}

	resp, err := w.client.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  w.group,
		LogStreamName: w.stream,
		SequenceToken: w.sequenceToken,
	})
	if err != nil {
		return n, err
	}

	if resp.RejectedLogEventsInfo != nil {
		return n, &RejectedLogEventsInfoError{Info: resp.RejectedLogEventsInfo}
	}

	w.sequenceToken = resp.NextSequenceToken

	return n, nil
}
