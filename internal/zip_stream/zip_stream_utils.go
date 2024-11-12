package zip_stream

import (
	"encoding/binary"
	"time"
)

func MSDosTimeToTime(dosDate, dosTime uint16) time.Time {
	return time.Date(
		// date bits 0-4: day of month; 5-8: month; 9-15: years since 1980
		int(dosDate>>9+1980),
		time.Month(dosDate>>5&0xf),
		int(dosDate&0x1f),

		// time bits 0-4: second/2; 5-10: minute; 11-15: hour
		int(dosTime>>11),
		int(dosTime>>5&0x3f),
		int(dosTime&0x1f*2),
		0, // nanoseconds

		time.UTC,
	)
}

// timeZone returns a *time.Location based on the provided offset.
// If the offset is non-sensible, then this uses an offset of zero.
func timeZone(offset time.Duration) *time.Location {
	const (
		minOffset   = -12 * time.Hour  // E.g., Baker island at -12:00
		maxOffset   = +14 * time.Hour  // E.g., Line island at +14:00
		offsetAlias = 15 * time.Minute // E.g., Nepal at +5:45
	)
	offset = offset.Round(offsetAlias)
	if offset < minOffset || maxOffset < offset {
		offset = 0
	}
	return time.FixedZone("", int(offset/time.Second))
}

type ReadBuf []byte

func (buff *ReadBuf) Uint8() uint8 {
	v := (*buff)[0]
	*buff = (*buff)[1:]
	return v
}

func (buff *ReadBuf) Uint16() uint16 {
	v := binary.LittleEndian.Uint16(*buff)
	*buff = (*buff)[2:]
	return v
}

func (buff *ReadBuf) Uint32() uint32 {
	v := binary.LittleEndian.Uint32(*buff)
	*buff = (*buff)[4:]
	return v
}

func (buff *ReadBuf) Uint64() uint64 {
	v := binary.LittleEndian.Uint64(*buff)
	*buff = (*buff)[8:]
	return v
}

func (buff *ReadBuf) Sub(n int) ReadBuf {
	b2 := (*buff)[:n]
	*buff = (*buff)[n:]
	return b2
}
