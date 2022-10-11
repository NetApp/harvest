/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package set

type Set struct {
	set map[string]struct{}
}

func New() *Set {
	s := Set{}
	s.set = make(map[string]struct{})
	return &s
}

func NewFrom(values []string) *Set {
	s := New()
	s.AddValues(values)
	return s
}

func (s *Set) Add(value string) {
	s.set[value] = struct{}{}
}

func (s *Set) AddValues(values []string) {
	for _, v := range values {
		s.set[v] = struct{}{}
	}
}

// Remove removes an element from the set
// See also Delete
func (s *Set) Remove(value string) {
	delete(s.set, value)
}

// Delete removes an element from the set
// return true if the element was in the set, false otherwise
func (s *Set) Delete(value string) bool {
	if s.Has(value) {
		delete(s.set, value)
		return true
	}
	return false
}

func (s *Set) Has(value string) bool {
	_, has := s.set[value]
	return has
}

func (s *Set) Values() []string {
	values := make([]string, 0, len(s.set))
	for v := range s.set {
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

func (s *Set) Slice() []string {
	keys := make([]string, 0, len(s.set))
	for k := range s.set {
		keys = append(keys, k)
	}
	return keys
}

func (s *Set) Iter() map[string]struct{} {
	return s.set
}
