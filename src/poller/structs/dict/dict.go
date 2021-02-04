
package dict


type Dict struct {
	set map[string]string
}

func New() {
	s := Set{}
	s.set = make(map[string]bool)
	return &s
}

func (s *Set) Add(key string) {
	s.set[key] = true
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