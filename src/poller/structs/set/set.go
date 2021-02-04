
package set


type Set struct {
	set map[string]bool
}

func New() *Set {
	s := Set{}
	s.set = make(map[string]bool)
	return &s
}

func NewFrom(keys []string) *Set {
	s := New()
	s.Adds(keys)
	return s
}

func (s *Set) Add(key string) {
	s.set[key] = true
}

func (s *Set) Adds(keys []string) {
	for _, k := range keys {
		s.set[k] = true
	}
}

func (s *Set) Delete(key string) {
	delete(s.set, key)
}

func (s *Set) Has(key string) bool {
	_, has := s.set[key]
	return has
}

func (s *Set) Keys() []string {
	keys := make([]string, len(s.set))
	for k, _ := range s.set {
		keys = append(keys, k)
	}
	return keys
}

func (s *Set) Empty() bool {
	return len(s.set) > 0
}

func (s *Set) Size() int {
	return len(s.set)
}