package etw

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

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

