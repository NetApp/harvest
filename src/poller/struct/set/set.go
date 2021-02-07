
package set


type Set struct {
	set map[string]bool
}

func New() *Set {
	s := Set{}
	s.set = make(map[string]bool)
	return &s
}

func NewFrom(values []string) *Set {
	s := New()
	s.AddValues(values)
	return s
}

func (s *Set) Add(value string) {
	s.set[value] = true
}

func (s *Set) AddValues(values []string) {
	for _, v := range values {
		s.set[v] = true
	}
}

func (s *Set) Delete(value string) {
	delete(s.set, value)
}

func (s *Set) Has(value string) bool {
	_, has := s.set[value]
	return has
}

func (s *Set) Values() []string {
	values := make([]string, len(s.set))
	for v, _ := range s.set {
		values = append(values, v)
	}
	return values
}

func (s *Set) IsEmpty() bool {
	return len(s.set) == 0
}

func (s *Set) Size() int {
	return len(s.set)
}