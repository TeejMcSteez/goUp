package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
)

const (
	FrameSnapshot byte = 0
	FrameAppend   byte = 1
	headerSize         = 12
)

// rounds integer to next multiple of 4
// adds 3 so it rounds up instead of down
// pushing any number that isn't aligned passed the
// next multiple of 4 and already aligned number doesnt get far enough
// then clears bottom two bits to round down to multiple of 4
func align4(n int) int { return (n + 3) &^ 3 }

// Frame Structure:
//
//	 0  u8   frame type (0 = snapshot, 1 = append)
//	 1  u8   padding
//	 2  u16  column count
//	 4  u32  seq
//	 8  u32  row count
//	12  name table: per column [u8 len][utf8 bytes],
//	    padded to a 4 byte boundary
//	    ts:   rows x u32 unix seconds (ascending)
//	    cols: columns x rows x u32 response time in microseconds, 0 = no sample
func encodeFrame(typ byte, seq uint32, names []string, ts []uint32, cols [][]uint32) []byte {
	rows := len(ts)

	nameBytes := 0

	names = slices.Clone(names)
	for i, n := range names {
		// Name length is a single byte
		if len(n) > math.MaxUint8 {
			names[i] = strings.ToValidUTF8(n[:math.MaxUint8], "")
		}
		nameBytes += 1 + len(names[i])
	}
	tableEnd := align4(headerSize + nameBytes)
	buf := make([]byte, tableEnd+rows*4*(1+len(cols)))

	le := binary.LittleEndian

	buf[0] = typ
	le.PutUint16(buf[2:], uint16(len(cols)))
	le.PutUint32(buf[4:], seq)
	le.PutUint32(buf[8:], uint32(rows))

	o := headerSize

	for _, n := range names {
		buf[o] = byte(len(n))
		o++
		o += copy(buf[o:], n)
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

// Frame is a decoded response time frame. Cols holds microseconds, 0 = no sample.
type Frame struct {
	Typ   byte
	Seq   uint32
	Names []string
	Ts    []uint32
	Cols  [][]uint32
}

// DecodeFrame parses a frame produced by encodeFrame.
func DecodeFrame(buf []byte) (Frame, error) {
	if len(buf) < headerSize {
		return Frame{}, errors.New("frame too short")
	}
	le := binary.LittleEndian

	typ := buf[0]
	if typ != FrameSnapshot && typ != FrameAppend {
		return Frame{}, fmt.Errorf("unknown frame type %d", typ)
	}
	colCount := int(le.Uint16(buf[2:]))
	rows := int(le.Uint32(buf[8:]))
	dataSize := rows * 4 * (1 + colCount)

	// Every name takes at least its length byte
	if len(buf) < align4(headerSize+colCount)+dataSize {
		return Frame{}, errors.New("frame too short for its name table and rows (truncated, or server predates named append frames)")
	}

	o := headerSize
	names := make([]string, colCount)
	for c := range names {
		if o >= len(buf) || o+1+int(buf[o]) > len(buf) {
			return Frame{}, errors.New("name table truncated")
		}
		n := int(buf[o])
		names[c] = string(buf[o+1 : o+1+n])
		o += 1 + n
	}
	o = align4(o)

	if o+dataSize != len(buf) {
		return Frame{}, fmt.Errorf("frame is %d bytes, want %d", len(buf), o+dataSize)
	}

	read := func() []uint32 {
		out := make([]uint32, rows)
		for i := range out {
			out[i] = le.Uint32(buf[o:])
			o += 4
		}
		return out
	}
	ts := read()
	cols := make([][]uint32, colCount)
	for c := range cols {
		cols[c] = read()
	}

	return Frame{Typ: typ, Seq: le.Uint32(buf[4:]), Names: names, Ts: ts, Cols: cols}, nil
}
