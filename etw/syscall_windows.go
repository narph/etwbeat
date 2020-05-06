package etw

import "C"
import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	ERROR_INSUFFICIENT_BUFFER syscall.Errno = 122
	ERROR_NOT_FOUND           syscall.Errno = 23
)

var (
	advapi                               = syscall.NewLazyDLL("advapi32.dll")
	tdh                                  = syscall.NewLazyDLL("tdh.dll")
	closeTrace                           = advapi.NewProc("CloseTrace")
	controlTraceW                        = advapi.NewProc("ControlTraceW")
	enableTraceEx2                       = advapi.NewProc("EnableTraceEx2")
	openTraceW                           = advapi.NewProc("OpenTraceW")
	processTrace                         = advapi.NewProc("ProcessTrace")
	queryTraceW                          = advapi.NewProc("QueryTraceW")
	startTraceW                          = advapi.NewProc("StartTraceW")
	tdhEnumerateProviderFieldInformation = tdh.NewProc("TdhEnumerateProviderFieldInformation")
)

type EventTraceProperties struct {
	Wnode               WnodeHeader
	BufferSize          uint32
	MinimumBuffers      uint32
	MaximumBuffers      uint32
	MaximumFileSize     uint32
	LogFileMode         uint32
	FlushTimer          uint32
	EnableFlags         uint32
	AgeLimit            int32
	NumberOfBuffers     uint32
	FreeBuffers         uint32
	EventsLost          uint32
	BuffersWritten      uint32
	LogBuffersLost      uint32
	RealTimeBuffersLost uint32
	LoggerThreadId      syscall.Handle
	LogFileNameOffset   uint32
	LoggerNameOffset    uint32
}

type EventTrace struct {
	Header           EventTraceHeader
	InstanceId       uint32
	ParentInstanceId uint32
	ParentGuid       GUID
	MofData          uintptr
	MofLength        uint32
	UnionCtx         uint32
}

type EventTraceHeader struct {
	Size      uint16
	Union1    uint16
	Union2    uint32
	ThreadId  uint32
	ProcessId uint32
	TimeStamp int64
	Union3    [16]byte
	Union4    uint64
}

type EventTraceLogfile struct {
	LogFileName   *uint16
	LoggerName    *uint16
	CurrentTime   int64
	BuffersRead   uint32
	Union1        uint32
	CurrentEvent  EventTrace
	LogfileHeader TraceLogfileHeader
	//BufferCallback *EventTraceBufferCallback
	BufferCallback uintptr
	BufferSize     uint32
	Filled         uint32
	EventsLost     uint32
	Callback       uintptr
	IsKernelTrace  uint32
	Context        uintptr
}
type TraceLogfileHeader struct {
	BufferSize         uint32
	VersionUnion       uint32
	ProviderVersion    uint32
	NumberOfProcessors uint32
	EndTime            int64
	TimerResolution    uint32
	MaximumFileSize    uint32
	LogFileMode        uint32
	BuffersWritten     uint32
	Union1             [16]byte
	LoggerName         *uint16
	LogFileName        *uint16
	TimeZone           TimeZoneInformation
	BootTime           int64
	PerfFreq           int64
	StartTime          int64
	ReservedFlags      uint32
	BuffersLost        uint32
}

type TimeZoneInformation struct {
	Bias         int32
	StandardName [32]uint16
	StandardDate SystemTime
	StandardBias int32
	DaylightName [32]uint16
	DaylightDate SystemTime
	DaylighBias  int32
}

type SystemTime struct {
	Year         uint16
	Month        uint16
	DayOfWeek    uint16
	Day          uint16
	Hour         uint16
	Minute       uint16
	Second       uint16
	Milliseconds uint16
}

type WnodeHeader struct {
	BufferSize    uint32
	ProviderId    uint32
	Union1        uint64
	Union2        int64
	Guid          GUID
	ClientContext uint32
	Flags         uint32
}

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type EnableTraceParameters struct {
	Version          uint32
	EnableProperty   uint32
	ControlFlags     uint32
	SourceId         GUID
	EnableFilterDesc *EventFilterDescriptor
	FilterDescrCount uint32
}

type EventFilterDescriptor struct {
	Ptr  uint64
	Size uint32
	Type uint32
}

type FileTime struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}

type EVENT_FIELD_TYPE uint32
type PROVIDER_FIELD_INFO struct {
	NameOffset        uint32
	DescriptionOffset uint32
	Value             uint64
}
type PROVIDER_FIELD_INFOARRAY struct {
	NumberOfElements uint32
	FieldType        EVENT_FIELD_TYPE
	FieldInfoArray   [ANYSIZE_ARRAY]PROVIDER_FIELD_INFO
}

const ANYSIZE_ARRAY = 100

func (e *EventTraceLogfile) SetProcessTraceMode(ptm uint32) {
	e.Union1 = ptm
}

func (g *GUID) String() string {
	return fmt.Sprintf("{%08X-%04X-%04X-%02X%02X-%02X%02X%02X%02X%02X%02X}",
		g.Data1,
		g.Data2,
		g.Data3,
		g.Data4[0], g.Data4[1],
		g.Data4[2], g.Data4[3], g.Data4[4], g.Data4[5], g.Data4[6], g.Data4[7])
}

func _StartTrace(traceHandle *uintptr, instanceName *uint16, properties *EventTraceProperties) error {
	r1, _, _ := startTraceW.Call(
		uintptr(unsafe.Pointer(traceHandle)),
		uintptr(unsafe.Pointer(instanceName)),
		uintptr(unsafe.Pointer(properties)))
	if r1 == 0 {
		return nil
	}
	return syscall.Errno(r1)
}

func _EnableTraceEx2(traceHandle uintptr, providerId *GUID, controlCode uint32, level uint8, matchAnyKeyword uint64, matchAllKeyword uint64, timeout uint32, enableParameters *EnableTraceParameters) error {
	r1, _, _ := enableTraceEx2.Call(
		uintptr(traceHandle),
		uintptr(unsafe.Pointer(providerId)),
		uintptr(controlCode),
		uintptr(level),
		uintptr(matchAnyKeyword),
		uintptr(matchAllKeyword),
		uintptr(timeout),
		uintptr(unsafe.Pointer(enableParameters)))
	if r1 == 0 {
		return nil
	}
	return syscall.Errno(r1)
}

func _ProcessTrace(handleArray *uint64, handleCount uint32, startTime *FileTime, endTime *FileTime) error {
	r1, _, _ := processTrace.Call(
		uintptr(unsafe.Pointer(handleArray)),
		uintptr(handleCount),
		uintptr(unsafe.Pointer(startTime)),
		uintptr(unsafe.Pointer(endTime)))
	if r1 == 0 {
		return nil
	}
	return syscall.Errno(r1)
}

func _OpenTrace(logfile *EventTraceLogfile) (uint64, error) {
	r1, _, err := openTraceW.Call(
		uintptr(unsafe.Pointer(logfile)))
	if err.(syscall.Errno) == 0 {
		return uint64(r1), nil
	}
	return uint64(r1), err
}

func _ControlTrace(traceHandle uintptr, instanceName *uint16, properties *EventTraceProperties, controlCode uint32) (uint32, error) {
	r1, _, err := controlTraceW.Call(
		uintptr(traceHandle),
		uintptr(unsafe.Pointer(instanceName)),
		uintptr(unsafe.Pointer(properties)),
		uintptr(controlCode))
	if err.(syscall.Errno) == 0 {
		return uint32(r1), nil
	}
	return uint32(r1), err
}

func _CloseTrace(traceHandle uint64) (uint32, error) {
	r1, _, err := closeTrace.Call(
		uintptr(traceHandle))
	if err.(syscall.Errno) == 0 {
		return uint32(r1), nil
	}
	return uint32(r1), err
}

func _TdhEnumerateProviderFieldInformation(providerId *GUID, field_type EVENT_FIELD_TYPE, lpBuffer *byte, size *uint32) error {
	r0, _, _ := tdhEnumerateProviderFieldInformation.Call(uintptr(unsafe.Pointer(providerId)),
		uintptr(field_type),
		uintptr(unsafe.Pointer(lpBuffer)),
		uintptr(unsafe.Pointer(size)))
	if r0 == 0 {
		return nil
	}
	return syscall.Errno(r0)
}

func TdhEnumerateProviderFieldInformation(providerId *GUID, field_type EVENT_FIELD_TYPE) error {
	var bufSize uint32
	var buff []byte
	err := _TdhEnumerateProviderFieldInformation(providerId, field_type, nil, &bufSize)
	if err == ERROR_INSUFFICIENT_BUFFER {
		buff = make([]byte, bufSize)
		bufSize = uint32(len(buff))
		err = _TdhEnumerateProviderFieldInformation(providerId, field_type, &buff[0], &bufSize)
	}
	if err == ERROR_NOT_FOUND {
		return err
	}
	if err == nil {
		providerFields := (*PROVIDER_FIELD_INFOARRAY)(unsafe.Pointer(&buff[0]))
		for i := 0; i < int(providerFields.NumberOfElements); i++ {
			pProvider:= uintptr(unsafe.Pointer(providerFields))
			pNameOffset:= uintptr(providerFields.FieldInfoArray[i].NameOffset)
			wchar := FromPtrToUTF16(unsafe.Pointer(pProvider + pNameOffset))
			var keyword string
			for _, e:= range  wchar {
				 keyword += string(e)
			 }
			_= keyword
		}
	}

	return nil
}
