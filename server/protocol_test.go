package server

import (
	"encoding/binary"
	"reflect"
	"testing"
)

func TestDecodeFrame_RoundTrip(t *testing.T) {
	want := Frame{
		Seq:   7,
		Names: []string{"a", "beta", "gamma-service"},
		Ts:    []uint32{100, 200},
		Cols:  [][]uint32{{1, 0}, {0, 2}, {3, 4}},
	}
	for _, typ := range []byte{FrameSnapshot, FrameAppend} {
		want.Typ = typ
		got, err := DecodeFrame(encodeFrame(typ, want.Seq, want.Names, want.Ts, want.Cols))
		if err != nil {
			t.Fatalf("type %d: %v", typ, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("type %d: got %+v, want %+v", typ, got, want)
		}
	}
}

func TestDecodeFrame_Empty(t *testing.T) {
	got, err := DecodeFrame(encodeFrame(FrameSnapshot, 0, nil, nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Names) != 0 || len(got.Ts) != 0 || len(got.Cols) != 0 {
		t.Errorf("expected empty frame, got %+v", got)
	}
}

func TestDecodeFrame_Invalid(t *testing.T) {
	valid := encodeFrame(FrameAppend, 1, []string{"a"}, []uint32{100}, [][]uint32{{5}})

	// The layout before append frames carried names
	nameless := make([]byte, headerSize+8)
	copy(nameless, valid[:headerSize])
	binary.LittleEndian.PutUint32(nameless[headerSize:], 100)

	unknownType := append([]byte{}, valid...)
	unknownType[0] = 9

	// Name length byte pointing past the end of the frame
	badName := append([]byte{}, valid...)
	badName[headerSize] = 255

	for name, buf := range map[string][]byte{
		"too short":    valid[:headerSize-1],
		"truncated":    valid[:len(valid)-1],
		"trailing":     append(append([]byte{}, valid...), 0, 0, 0, 0),
		"nameless":     nameless,
		"unknown type": unknownType,
		"bad name":     badName,
	} {
		if _, err := DecodeFrame(buf); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
