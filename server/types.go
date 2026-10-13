package server

type ResponseTimePayload struct {
	names   []string
	ts      []uint32
	cols    [][]float32
	seq     uint32
	maxRows int
}
