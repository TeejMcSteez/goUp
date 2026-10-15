package server

import (
	"encoding/binary"
	"math"
	"slices"
	"strings"
)

const (
	FrameSnapshot byte = 0
	FrameAppend   byte = 1
	headerSize         = 12
)

func align4(n int) int { return (n + 3) &^ 3 }

func encodeFrame(typ byte, seq uint32, names []string, ts []uint32, cols [][]uint32) []byte {
	rows := len(ts)

	nameBytes := 0

	if typ == FrameSnapshot {
		names = slices.Clone(names)
		for i, n := range names {
			// Name length is a single byte
			if len(n) > math.MaxUint8 {
				names[i] = strings.ToValidUTF8(n[:math.MaxUint8], "")
			}
			nameBytes += 1 + len(names[i])
		}
	}
	tableEnd := align4(headerSize + nameBytes)
	buf := make([]byte, tableEnd+rows*4*(1+len(cols)))

	le := binary.LittleEndian

	buf[0] = typ
	le.PutUint16(buf[2:], uint16(len(cols)))
	le.PutUint32(buf[4:], seq)
	le.PutUint32(buf[8:], uint32(rows))

	o := headerSize

	if typ == FrameSnapshot {
		for _, n := range names {
			buf[o] = byte(len(n))
			o++
			o += copy(buf[o:], n)
		}
	}

	o = tableEnd
	for _, t := range ts {
		le.PutUint32(buf[o:], t)
		o += 4
	}
	for _, col := range cols {
		for _, v := range col {
			le.PutUint32(buf[o:], v)
			o += 4
		}
	}
	return buf
}
