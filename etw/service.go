package etw

import (
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common/fmtstr"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/libbeat/processors/add_formatted_index"
	"github.com/narph/etwbeat/config"
	"github.com/pkg/errors"
	"syscall"
)

const (
	TRACE_LEVEL_NONE                      = 0
	TRACE_LEVEL_CRITICAL                  = 1
	TRACE_LEVEL_FATAL                     = 1
	TRACE_LEVEL_ERROR                     = 2
	TRACE_LEVEL_WARNING                   = 3
	TRACE_LEVEL_INFORMATION               = 4
	TRACE_LEVEL_VERBOSE                   = 5
	TRACE_LEVEL_RESERVED6                 = 6
	TRACE_LEVEL_RESERVED7                 = 7
	TRACE_LEVEL_RESERVED8                 = 8
	TRACE_LEVEL_RESERVED9                 = 9
	ERROR_ALREADY_EXISTS    syscall.Errno = 183

EventKeywordInformation  EVENT_FIELD_TYPE = 0
EventLevelInformation    EVENT_FIELD_TYPE = 1
EventChannelInformation  EVENT_FIELD_TYPE = 2
EventTaskInformation     EVENT_FIELD_TYPE = 3
EventOpcodeInformation   EVENT_FIELD_TYPE = 4
EventInformationMax      EVENT_FIELD_TYPE = 5

)

func enableTrace(session config.Session) (uintptr, *EventTraceProperties, error) {
	g, _ := GUIDFromString(session.Providers[0])
	//var fieldinfo *PROVIDER_FIELD_INFOARRAY
	//var bufferSize uint32
	err:= TdhEnumerateProviderFieldInformation(g, EventKeywordInformation)
	//err= _TdhEnumerateProviderFieldInformation(g, fieldinfo,EventKeywordInformation, bufferSize)
	if err!= nil {
		_= err
	}
	var sessionHandle uintptr
	sessionProperties := NewSession(session.Name)
	sessionPtr, err := syscall.UTF16PtrFromString(session.Name)
	if err != nil {
		return 0, nil, errors.Wrapf(err, "Failed to convert session %s", session.Name)
	}
	err = _StartTrace(&sessionHandle, sessionPtr, sessionProperties)
	if err != nil {
		if err == ERROR_ALREADY_EXISTS {
			return 0, sessionProperties, nil
		}
		return 0, nil, errors.Wrapf(err, "Failed to start trace %s", session.Name)
	}

	if err := _EnableTraceEx2(
		sessionHandle,
		g,
		EVENT_CONTROL_CODE_ENABLE_PROVIDER,
		getTraceLevel(session.TraceLevel),
		//getTraceLevel(session.TraceLevel),
		0xffffffffffffffff,
		0,
		0,
		nil,
	); err != nil {
		return 0, nil, errors.Wrap(err, "Failed to enable trace")
	}
	return sessionHandle, sessionProperties, nil
}

func getTraceLevel(level string) uint8 {
	switch level {
	case "critical":
		return TRACE_LEVEL_CRITICAL
	case "error":
		return TRACE_LEVEL_ERROR
	case "warning":
		return TRACE_LEVEL_WARNING
	case "information":
		return TRACE_LEVEL_INFORMATION
	case "verbose":
		return TRACE_LEVEL_VERBOSE
	default:
		return TRACE_LEVEL_INFORMATION
	}
}

func stopTrace(sessionHandle uintptr, sessionProperties *EventTraceProperties, providerId *GUID) error {
	_ControlTrace(sessionHandle, nil, sessionProperties, EVENT_TRACE_CONTROL_STOP)
	return _EnableTraceEx2(sessionHandle, providerId, EVENT_CONTROL_CODE_DISABLE_PROVIDER, 0, 0, 0, 0, nil)

}

// processorsForConfig assembles the Processors for an eventLogger.
func processorsForConfig(beatInfo beat.Info, config consumerConfig) (*processors.Processors, error) {
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

func openTrace(sessionName string, bufferCallback interface{}, eventReceivedCallback interface{}) error {
	var etLogFile EventTraceLogfile
	// Consumer Part
	etLogFile.SetProcessTraceMode(PROCESS_TRACE_MODE_EVENT_RECORD | PROCESS_TRACE_MODE_RAW_TIMESTAMP | PROCESS_TRACE_MODE_REAL_TIME)
	etLogFile.BufferCallback = syscall.NewCallback(bufferCallback)
	etLogFile.Callback = syscall.NewCallback(eventReceivedCallback)
	etLogFile.Context = 0
	sessionPtr, err := syscall.UTF16PtrFromString(sessionName)
	if err != nil {
		return errors.Wrapf(err, "Failed to convert session %s", sessionName)
	}
	etLogFile.LoggerName = sessionPtr
	traceHandle, err := _OpenTrace(&etLogFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to open trace for session %s", sessionName)
	}
	if err := _ProcessTrace(&traceHandle, 1, nil, nil); err != nil {
		return errors.Wrapf(err, "Failed to process trace for session %s", sessionName)
	}
	return nil
}
