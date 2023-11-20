package blockchain

type Set map[any]struct{}

func (s Set) Add(value any) {
	s[value] = struct{}{}
}

func (s Set) Remove(value any) {
	delete(s, value)
}

func (s Set) Contains(value any) bool {
	_, exists := s[value]
	return exists
}
