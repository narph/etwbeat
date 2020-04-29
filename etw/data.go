package etw

import (
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
)

func (e *EventRecord) ToEvent() beat.Event {
	// Windows Log Specific data
	win := common.MapStr{
		"provider_id": e.EventHeader.ProviderId.String(),
		"process_id":  e.EventHeader.ProcessId,
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
