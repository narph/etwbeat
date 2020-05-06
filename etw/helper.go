package etw

import (
	"encoding/binary"
	"fmt"
	"github.com/gofrs/uuid"
	"regexp"
	"strconv"
	"strings"
	"unsafe"
)

var guidRE = regexp.MustCompile(`\{[A-F0-9]{8}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{4}-[A-F0-9]{12}\}`)

func GUIDFromString(guid string) (*GUID, error) {
	g := GUID{}
	guid = strings.ToUpper(guid)
	if !guidRE.MatchString(guid) {
		return nil, fmt.Errorf("Bad GUID format")
	}
	guid = strings.Trim(guid, "{}")
	sp := strings.Split(guid, "-")
	c, _ := strconv.ParseUint(sp[0], 16, 32)
	g.Data1 = uint32(c)
	c, _ = strconv.ParseUint(sp[1], 16, 16)
	g.Data2 = uint16(c)
	c, _ = strconv.ParseUint(sp[2], 16, 16)
	g.Data3 = uint16(c)
	i64, _ := strconv.ParseUint(fmt.Sprintf("%s%s", sp[3], sp[4]), 16, 64)
	buf, err := Marshal(&i64, binary.BigEndian)
	if err != nil {
		return nil, err
	}
	copy(g.Data4[:], buf)

	return &g, nil
}

func randomGUID() (GUID, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return GUID{}, err
	}
	validGuid := fmt.Sprintf("{%s}", id.String())
	guid, err := GUIDFromString(validGuid)
	if err != nil {
		return GUID{}, err
	}
	return *guid, nil
}

func FromPtrToUTF16(offset unsafe.Pointer) []uint16 {
	if uintptr(offset) == 0x0 {
		return nil
	}
	offsetPtr := uintptr(offset)
	utf := make([]uint16, 0)
	var w uint16
	for {
		w = *((*uint16)(unsafe.Pointer(offsetPtr)))
		if w == 0 {
			break
		}
		utf = append(utf, w)
		offsetPtr += 2
	}
	return utf
}
