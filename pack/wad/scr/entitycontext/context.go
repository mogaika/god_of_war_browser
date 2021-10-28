package entitycontext

type EntityLevelContext struct {
	LevelData        map[uint16]Variable
	GlobalData       map[uint16]Variable
	EntityIdNameMap  map[uint16]string
	EntityReferences map[uint16][]uint16
}

type Variable struct {
	Type uint8
	Name string
}

func NewContext() EntityLevelContext {
	return EntityLevelContext{
		LevelData:        make(map[uint16]Variable),
		GlobalData:       make(map[uint16]Variable),
		EntityIdNameMap:  make(map[uint16]string),
		EntityReferences: make(map[uint16][]uint16),
	}
}
