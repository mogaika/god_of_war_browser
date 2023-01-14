package wadtree

/*
	Level nodes order:
	888 wad heap size
	30 servers init
	999 pop heap
	666 wad start
	24 EntityCount
	30 SBK_ sound banks
	for ctx:
		30 mat_fog
	RIB_Sheet
	for ctx:
		CXT_
	113 TWK_
	500 RSRCS
	30 ANM_ external anims
*/

type ServerType int16

const (
	SERVER_GO                      ServerType = 1
	SERVER_RENDER_PRIMITIVE_MASTER            = 2
	SERVER_ANIMATION                          = 3
	SERVER_SCRIPT                             = 4
	SERVER_MASTER                             = 5
	SERVER_LIGHT                              = 6
	SERVER_TEXTURE                            = 7
	SERVER_MATERIAL                           = 8
	SERVER_CAMERA                             = 9
	SERVER_PROLOGUE                           = 10
	SERVER_EPILOGUE                           = 11
	SERVER_GFX                                = 12
	SERVER_RENDER_MASTER                      = 13
	SERVER_RENDER_MODEL                       = 15
	SERVER_COLLISION                          = 17
	SERVER_RENDER_PARTICLE                    = 19
	SERVER_WAYPOINT                           = 20
	SERVER_EVENT                              = 21
	SERVER_BEHAVIOR                           = 23
	SERVER_SOUND                              = 24
	SERVER_EMMITTER                           = 26
	SERVER_WAD                                = 27
	SERVER_RENDER_EE_PRIMITIVE                = 28
	SERVER_EFFECTS                            = 30
	SERVER_RENDER_FLASH                       = 33
	SERVER_RENDER_LINE                        = 35
	SERVER_RENDER_SHADOW                      = 39
)

type Reference string

type AnimationAct struct {
}

type AnimationGroup struct {
}

type Animation struct {
}

type MaterialLayer struct {
}

type Material struct {
	Animation *Animation
}

type Script struct {
}

type RenderMesh struct {
}

type RenderModel struct {
	Meshes    []*RenderMesh
	Materials []Reference
	Scripts   []*Script
}

type Collision struct {
}

type Object struct {
	Entities []*Entity
}

type Entity struct {
}

type GameObject struct {
	Object  *Object
	Scripts []*Script
}

type Light struct {
}

type Camera struct {
}

type SoundEmitter struct {
}

type Sound struct {
}

type Wad struct {
	Name    string
	IsLevel bool

	Map map[Reference]interface{}

	Objects map[string]*Object
	Sounds  map[string]*Sound
	Cameras []*Camera

	Resources []*Wad
	Perm      *Wad
}

type Instance interface {
	ServerType() uint32
}

type wadLoader struct {
	namespace map[string]*Instance
}
