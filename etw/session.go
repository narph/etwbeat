package etw

import "unsafe"

const (
	WNODE_FLAG_ALL_DATA        = 0x00000001
	EVENT_TRACE_REAL_TIME_MODE = 0x00000100
)

func NewSession(sessionName string) *EventTraceProperties {
	size := ((len(sessionName) + 1) * 2) + int(unsafe.Sizeof(EventTraceProperties{}))
	s := make([]byte, size)
	sessionProperties := (*EventTraceProperties)(unsafe.Pointer(&s[0]))

	// Necessary fields for SessionProperties struct
	sessionProperties.Wnode.BufferSize = uint32(size)
	sessionProperties.Wnode.Guid = GUID{} // To set
	sessionProperties.Wnode.ClientContext = 0
	sessionProperties.Wnode.Flags = WNODE_FLAG_ALL_DATA
	sessionProperties.LogFileMode = EVENT_TRACE_REAL_TIME_MODE
	sessionProperties.LogFileNameOffset = 0
	sessionProperties.LoggerNameOffset = uint32(unsafe.Sizeof(EventTraceProperties{}))

	return sessionProperties
}
