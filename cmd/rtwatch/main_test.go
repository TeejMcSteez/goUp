package main

import (
	"goUp/server"
	"reflect"
	"testing"
)

func TestMergeFrame(t *testing.T) {
	prev := &series{
		names: []string{"a", "b", "c"},
		ts:    []uint32{100, 200},
		cols:  [][]uint32{{1, 2}, {3, 0}, {4, 5}},
	}
	// Only b and c reported. The row at 200 fills b's gap in prev's last row
	// and the row at 300 is new; a stays empty there.
	frame := server.Frame{
		Seq:   1,
		Names: []string{"b", "c"},
		Ts:    []uint32{200, 300},
		Cols:  [][]uint32{{9, 6}, {0, 7}},
	}

	got := mergeFrame(prev, frame, mapColumns(prev, frame))

	want := &series{
		names: prev.names,
		seq:   1,
		ts:    []uint32{100, 200, 300},
		cols:  [][]uint32{{1, 2, 0}, {3, 9, 6}, {4, 5, 7}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if prev.cols[1][1] != 0 || len(prev.ts) != 2 {
		t.Errorf("prev was mutated: %+v", prev)
	}
}

func TestMapColumns(t *testing.T) {
	prev := &series{names: []string{"a", "b"}}

	// Same column count as prev but a different service set
	if got := mapColumns(prev, server.Frame{Names: []string{"b", "c"}}); got != nil {
		t.Errorf("unknown service: expected nil, got %v", got)
	}
	got := mapColumns(prev, server.Frame{Names: []string{"b"}})
	if !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("subset: expected [1], got %v", got)
	}
}
