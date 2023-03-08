package rendercontext

type TempDataHolder interface {
	ClearTempRenderData()
}

var global store = newStore()

func Use(dh TempDataHolder) { global.Use(dh) }
func Swap()                 { global.Swap() }

type store struct {
	used    map[TempDataHolder]struct{}
	notUsed map[TempDataHolder]struct{}
}

func newStore() store {
	return store{
		used:    make(map[TempDataHolder]struct{}),
		notUsed: make(map[TempDataHolder]struct{}),
	}
}

func (s *store) Swap() {
	for dh := range s.notUsed {
		dh.ClearTempRenderData()
	}
	s.notUsed = s.used
	s.used = make(map[TempDataHolder]struct{})
}

func (s *store) Use(dh TempDataHolder) {
	delete(s.notUsed, dh)
	s.used[dh] = struct{}{}
}
