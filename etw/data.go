package etw

import (
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"time"
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

func (e *EventRecord) ToEvent() beat.Event {
	// Windows Log Specific data
	win := common.MapStr{
		"provider_id": e.EventHeader.ProviderId.String(),
		"process_id":  e.EventHeader.ProcessId,
		"thread_id": e.EventHeader.ThreadId,
		"event_id": e.EventHeader.EventDescriptor.Id,
		"channel": e.EventHeader.EventDescriptor.Channel,
		"op_code": e.EventHeader.EventDescriptor.Opcode,
		"keyword": e.EventHeader.EventDescriptor.Keyword,
		"activity_id": e.EventHeader.ActivityId.String(),
		"level": e.EventHeader.EventDescriptor.Level,
	}



	m := common.MapStr{
		"winlog": win,
	}

	// ECS data
	m.Put("event.kind", "event")
	m.Put("event.created", time.Now())
	//parse, err:= time.Parse(string(e.EventHeader.Time))
	return beat.Event{
		Timestamp: time.Now(),
		Fields:    m,
		//Private:   e.Offset,
	}
}
