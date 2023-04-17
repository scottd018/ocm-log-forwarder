package stdout

import (
	"fmt"

	v1 "github.com/openshift-online/ocm-sdk-go/servicelogs/v1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/poller"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

type StdOut struct {
	SentMessages []string
}

func (stdout *StdOut) Initialize(proc *processor.Processor) (err error) {
	// nothing to initialize with this backend so we simply return
	return nil
}

func (stdout *StdOut) Send(proc *processor.Processor, response *poller.Response) error {
	logChan := make(chan *v1.LogEntry, len(response.Logs))

	for i := range response.Logs {
		go func(index int) {
			logChan <- response.Logs[index]
		}(i)
	}

	defer close(logChan)

	for i := 0; i < len(response.Logs); i++ {
		logMessage := <-logChan

		if stdout.HasSent(logMessage) {
			continue
		}

		stdout.send(logMessage)
	}

	return nil
}

func (stdout *StdOut) String() string {
	return config.DefaultBackendStdOut
}

func (stdout *StdOut) HasSent(message *v1.LogEntry) bool {
	// send and return if we have no messages on the channel
	if len(stdout.SentMessages) < 1 {
		return false
	}

	for i := range stdout.SentMessages {
		if message.ID() == stdout.SentMessages[i] {
			return true
		}
	}

	return false
}

func (stdout *StdOut) Log(event *zerolog.Event, message string) {
	event.Str("source", fmt.Sprintf("%s-backend", stdout.String())).Msg(message)
}

func (stdout *StdOut) send(logEntry *v1.LogEntry) {
	// log the message to stdout
	stdout.Log(
		log.Info().
			Str("_id", logEntry.ID()).
			Str("@timestamp", logEntry.Timestamp().String()).
			Str("cluster_id", logEntry.ClusterID()).
			Str("external_id", logEntry.ClusterUUID()).
			Str("username", logEntry.Username()).
			Str("severity", string(logEntry.Severity())).
			Str("event_id", logEntry.EventStreamID()).
			Str("service_name", logEntry.ServiceName()),
		logEntry.Summary(),
	)

	// add the message to the list of sent messages
	stdout.SentMessages = append(stdout.SentMessages, logEntry.ID())
}
