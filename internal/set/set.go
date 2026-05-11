package set

type Set[T comparable] map[T]struct{}

func New[T comparable]() Set[T] {
	return map[T]struct{}{}
}

func (s Set[T]) Add(t T) {
	s[t] = struct{}{}
}

func (s Set[T]) AddAll(ts ...T) {
	for _, t := range ts {
		s[t] = struct{}{}
	}
}

func (s Set[T]) Remove(t T) {
	delete(s, t)
}

func (s Set[T]) RemoveAll(ts ...T) {
	for _, t := range ts {
		delete(s, t)
	}
}

func (s Set[T]) Contains(t T) bool {
	_, ok := s[t]
	return ok
}
