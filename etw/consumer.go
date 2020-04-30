package etw

import (
	"github.com/narph/etwbeat/config"

	"github.com/pkg/errors"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/fmtstr"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/processors"
)

const (
	EVENT_TRACE_CONTROL_STOP            = 1
	EVENT_CONTROL_CODE_DISABLE_PROVIDER = 0
	EVENT_CONTROL_CODE_ENABLE_PROVIDER  = 1
	TRACE_LEVEL_VERBOSE                 = 5
	PROCESS_TRACE_MODE_REAL_TIME        = 0x00000100
	PROCESS_TRACE_MODE_RAW_TIMESTAMP    = 0x00001000
	PROCESS_TRACE_MODE_EVENT_RECORD     = 0x10000000
	EVENT_HEADER_FLAG_STRING_ONLY       = 0x0004
	EVENT_HEADER_PROPERTY_XML           = 0x0001
)

type Consumer struct {
	log        *logp.Logger
	Config     consumerConfig
	eventMeta  common.EventMetadata
	processors beat.ProcessorList
	client     beat.Client
}

type consumerConfig struct {
	common.EventMetadata `config:",inline"`       // Fields and tags to add to events.
	Processors           processors.PluginConfig  `config:"processors"`
	Index                fmtstr.EventFormatString `config:"index"`
}

func NewConsumer(options *common.Config, beatInfo beat.Info) (*Consumer, error) {
	c := consumerConfig{}
	if err := options.Unpack(&c); err != nil {
		return nil, err
	}
	processorsForConfig, err := processorsForConfig(beatInfo, c)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		log:        logp.NewLogger("etw"),
		Config:     c,
		processors: processorsForConfig,
	}, nil
}

func (c *Consumer) Run(done <-chan struct{}, pipeline beat.Pipeline, provider config.Provider) {
	client, err := c.connect(pipeline)
	if err != nil {
		logp.Warn("EventLog[%s] Pipeline error. Failed to connect to publisher pipeline",
			"session")
		return
	}
	c.client = client
	// close client on function return or when `done` is triggered (unblock client)
	defer c.client.Close()
	go func() {
		<-done
		c.client.Close()
	}()
	sessionHandle, sessionProperties, err := enableTrace(provider.Id, provider.SessionName)
	if err != nil {
		c.log.Errorf("session %s could not be enabled: %v", provider.SessionName, err)
		return
	}
	defer func() {
		guid, _ := GUIDFromString(provider.Id)

		c.log.Info("Session %s, stop processing.", provider.SessionName)
		if err := stopTrace(sessionHandle, sessionProperties, guid); err != nil {
			c.log.Errorf("session %s could not be closed: %v", provider.SessionName, err)
			return
		}
	}()
	err = c.readEvents(provider.SessionName)
	if err != nil {
		c.log.Errorf("session %s could not read events: %v", provider.SessionName, err)
		return
	}
}

func (c *Consumer) connect(pipeline beat.Pipeline) (beat.Client, error) {
	return pipeline.ConnectWith(beat.ClientConfig{
		PublishMode: beat.GuaranteedSend,
		Processing: beat.ProcessingConfig{
			EventMetadata: c.eventMeta,
			Processor:     c.processors,
		},
		ACKCount: func(n int) {
			c.log.Info("successfully published %d events", n)
		},
	})
}

func (c *Consumer) readEvents(sessionName string) error {
	err := openTrace(sessionName, c.bufferCallback, c.eventReceivedCallback)
	if err != nil {
		return errors.Wrap(err, "Failed to open trace")
	}
	return nil
}

func (c *Consumer) bufferCallback(etl *EventTraceLogfile) uintptr {
	_ = etl
	return 1
}

func (c Consumer) eventReceivedCallback(er *EventRecord) uintptr {
	c.client.Publish(er.ToEvent())
	return 0
}
