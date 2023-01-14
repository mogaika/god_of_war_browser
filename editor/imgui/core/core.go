package core

type Freeable interface {
	Free()
}

type Referencable[T Freeable] interface {
	Freeable
	NewReference() Reference[T]
}

type Reference[T Freeable] struct {
	ref *Referencable[T]
}

type ColorView struct {
	OnChange func()
}
