package etw

import (
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/fmtstr"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/libbeat/processors/add_formatted_index"
	"github.com/narph/etwbeat/config"
	"github.com/pkg/errors"
	"regexp"
	"syscall"
)

var (
	guidRE = regexp.MustCompile(`\{[A-F0-9]{8}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{12}\}`)
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

type EventRecord struct {
	EventHeader       EventHeader
	BufferContext     EtwBufferContext
	ExtendedDataCount uint16
	UserDataLength    uint16
	ExtendedData      EventHeaderExtendedDataItem
	UserData          uintptr
	UserContext       uintptr
}

type EtwBufferContext struct {
	Union    uint16
	LoggerId uint16
}

type EventHeaderExtendedDataItem struct {
	Reserved1      uint16
	ExtType        uint16
	InternalStruct uint16
	DataSize       uint16
	DataPtr        uint64
}

type EventHeader struct {
	Size            uint16
	HeaderType      uint16
	Flags           uint16
	EventProperty   uint16
	ThreadId        uint32
	ProcessId       uint32
	TimeStamp       int64
	ProviderId      GUID
	EventDescriptor EventDescriptor
	Time            int64
	ActivityId      GUID
}

type EventDescriptor struct {
	Id      uint16
	Version uint8
	Channel uint8
	Level   uint8
	Opcode  uint8
	Task    uint16
	Keyword uint64
}

type Consumer struct {
	Sessions   []Session
	log        *logp.Logger
	Config     consumerConfig
	eventMeta  common.EventMetadata
	processors beat.ProcessorList
	client      beat.Client
}

type Session struct {
	Name          string
	ProviderId    *GUID
	Handle        uintptr
	SessionHandle uint64
	Properties *EventTraceProperties
}

type consumerConfig struct {
	common.EventMetadata `config:",inline"`       // Fields and tags to add to events.
	Processors           processors.PluginConfig  `config:"processors"`
	Index                fmtstr.EventFormatString `config:"index"`
	Providers []config.Provider
}

func NewConsumer(options *common.Config, beatInfo beat.Info) (*Consumer, error) {
	c := consumerConfig{}
	if err := options.Unpack(&c); err != nil {
		return nil, err
	}
	processors, err := processorsForConfig(beatInfo, c)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		Sessions:   []Session{},
		log:        logp.NewLogger("etw"),
		Config:     c,
		processors: processors,
	}, nil
}

func (c *Consumer) Run(
	done <-chan struct{},
	pipeline beat.Pipeline,
) {
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
	guid := "{EDD08927-9CC4-4E65-B970-C2560FB5C289}"
	session := "TestGoSessionSession"
	err = c.EnableTrace(guid, session)
	if err != nil {
		logp.Warn("EventLog[%s] Open() error. No events will be read from "+
			"this source. %v", "session", err)
		return
	}
	defer func() {
		c.log.Info("EventLog[%s] Stop processing.", "session")
		if err := c.CloseTrace(session); err != nil {
			logp.Warn("EventLog[%s] Close() error. %v", session, err)
			return
		}
	}()


	err= c.ReadEvents()
	if err != nil {
		logp.Warn("EventLog[%s] Close() error. %v", session, err)
		return
	}
	for stop := false; !stop; {
		select {
		case <-done:
			return
		default:
		}
	}
}

func (c *Consumer) connect(pipeline beat.Pipeline) (beat.Client, error) {
	//api := e.source.Name()
	return pipeline.ConnectWith(beat.ClientConfig{
		PublishMode: beat.GuaranteedSend,
		Processing: beat.ProcessingConfig{
			EventMetadata: c.eventMeta,
			Processor:     c.processors,
			//KeepNull:      e.keepNull,
		},
		ACKCount: func(n int) {
			//addPublished(api, n)
			logp.Info("EventLog[] successfully published %d events", n)
		},
	})
}

// processorsForConfig assembles the Processors for an eventLogger.
func processorsForConfig(
	beatInfo beat.Info, config consumerConfig,
) (*processors.Processors, error) {
	procs := processors.NewList(nil)

	// Processor order is important! The index processor, if present, must be
	// added before the user processors.
	if !config.Index.IsEmpty() {
		staticFields := fmtstr.FieldsForBeat(beatInfo.Beat, beatInfo.Version)
		timestampFormat, err :=
			fmtstr.NewTimestampFormatString(&config.Index, staticFields)
		if err != nil {
			return nil, err
		}
		indexProcessor := add_formatted_index.New(timestampFormat)
		procs.AddProcessor(indexProcessor)
	}

	userProcs, err := processors.New(config.Processors)
	if err != nil {
		return nil, err
	}
	procs.AddProcessors(*userProcs)

	return procs, nil
}

func (c *Consumer) EnableTrace(guid string, session string) error {
	var sessionHandle uintptr
	sessionProperties := NewSession(session)
	err := StartTrace(&sessionHandle, syscall.StringToUTF16Ptr(session), sessionProperties)
	if err != nil {
		return errors.Wrap(err, "Failed to create trace")
	}
	g, _ := GUIDFromString(guid)
	if err := EnableTraceEx2(
		sessionHandle,
		g,
		EVENT_CONTROL_CODE_ENABLE_PROVIDER,
		TRACE_LEVEL_VERBOSE,
		0xffffffffffffffff,
		0,
		0,
		nil,
	); err != nil {
		return errors.Wrap(err, "Failed to enable trace")
	}
	c.Sessions = append(c.Sessions, Session{Name: session, ProviderId: g, Handle: sessionHandle, Properties:sessionProperties})
	return nil
}

func (c *Consumer) CloseTrace(sessionName string) error {
	for _, session := range c.Sessions {
		if session.Name == sessionName {
			 ControlTrace(session.Handle, nil, session.Properties, EVENT_TRACE_CONTROL_STOP)
			return EnableTraceEx2(session.Handle, session.ProviderId, EVENT_CONTROL_CODE_DISABLE_PROVIDER, 0, 0, 0, 0, nil)
		}
	}
	return errors.Errorf("session %s not found", sessionName)
}

func (c *Consumer) ReadEvents() error {
	var etLogFile EventTraceLogfile
	// Consumer Part
	etLogFile.SetProcessTraceMode(PROCESS_TRACE_MODE_EVENT_RECORD | PROCESS_TRACE_MODE_RAW_TIMESTAMP | PROCESS_TRACE_MODE_REAL_TIME)
	etLogFile.BufferCallback = syscall.NewCallback(BufferCallback)
	etLogFile.Callback = syscall.NewCallback(c.EventRecCallback)
	etLogFile.Context = 0
	etLogFile.LoggerName = syscall.StringToUTF16Ptr(c.Sessions[0].Name)

	traceHandle, err := OpenTrace(&etLogFile)
	if err != nil {
		return errors.Wrap(err, "Failed to open trace")
	}
	c.Sessions[0].SessionHandle = traceHandle
	go c.Process(traceHandle)

	return nil
}

func (c *Consumer) Process(handle uint64) {
	if err := ProcessTrace(&handle, 1, nil, nil); err != nil {
		c.log.Error(err)
	}
}

func (c *Consumer) CloseSession(handle uint64) {
	CloseTrace(handle)
}


func BufferCallback(etl *EventTraceLogfile) uintptr {
	_ = etl
	return 1
}


func (c Consumer) EventRecCallback(er *EventRecord) uintptr {
	c.client.Publish(er.ToEvent())
	return 0
}
