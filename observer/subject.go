package observer

import "strings"

// Subject ...
type Subject interface {
}

type subject struct {
	sources map[string][]Source
}

// Ping ...
func (s *subject) Ping() {

}

// RegisterSource ...
func (s *subject) RegisterSource(name string, source Source) {
	if sources, b := s.sources[name]; b {
		for _, s2 := range sources {
			if strings.Compare(source.Network(), s2.Network()) == 0 &&
				strings.Compare(source.String(), s2.String()) == 0 {
				return
			}
		}
		s.sources[name] = append(s.sources[name], source)
		return
	}
	s.sources[name] = []Source{
		source,
	}
	return
}
