package utils

func NewStore() *SharedData {
	return &SharedData{
		data: make([]ServiceData, 0),
	}
}
// Writer: Sets new data
func (s *SharedData) Set(data []ServiceData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = data
}

// Reader: get a copy of the current slice
func (s *SharedData) Get() []ServiceData {
    s.mu.RLock()
    defer s.mu.RUnlock()

    out := make([]ServiceData, len(s.data))
    copy(out, s.data)
    return out
}