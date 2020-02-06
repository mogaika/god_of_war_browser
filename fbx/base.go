package fbx

type FBX struct {
	FBXHeaderExtension FBXHeaderExtension
	GlobalSettings     *GlobalSettings
	Documents          Documents
	References         References
	Definitions        Definitions
	Objects            Objects
	Connections        Connections

	lastAllocatedId uint64

	files map[string][]byte
}

func (f *FBX) GenerateId() uint64 {
	if f.lastAllocatedId == 0 {
		f.lastAllocatedId = 1000000 // used for document id
	}
	f.lastAllocatedId++
	return f.lastAllocatedId
}

type References struct {
}

type Document struct {
	Id      uint64 `fbx:"p"`
	Name    string `fbx:"p"`
	Element string `fbx:"p"`

	Properties70 Properties70

	RootNode int
}

type Documents struct {
	Count    int
	Document []*Document
}

type CreationTimeStamp struct {
	Version     int
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Millisecond int
}

type FBXHeaderExtension struct {
	FBXHeaderVersion  int
	FBXVersion        int
	CreationTimeStamp *CreationTimeStamp
	Creator           string
}

type Propertie70 struct {
	Name    string      `fbx:"p"`
	Type    string      `fbx:"p"`
	Purpose string      `fbx:"p"`
	Idk     string      `fbx:"p"`
	Value   interface{} `fbx:"p"`
}

type Properties70 struct {
	P []*Propertie70
}

type GlobalSettings struct {
	Properties70 Properties70
	Version      int
}

type PropertyTemplate struct {
	TemplateName string `fbx:"p"` // ?
	Properties70 Properties70
}

type ObjectTypeDefinition struct {
	Name             string `fbx:"p"`
	Count            int
	PropertyTemplate *PropertyTemplate
}

type Definitions struct {
	Version int
	Count   int

	ObjectType []*ObjectTypeDefinition
}

type LayerElementShared struct {
	Id uint64 `fbx:"p"`

	Version int
	Name    string

	MappingInformationType   string
	ReferenceInformationType string
	Normals                  interface{} `fbx:"a"`
	NormalsW                 interface{} `fbx:"a"`
	Binormals                interface{} `fbx:"a"`
	BinormalsW               interface{} `fbx:"a"`
	Tangents                 interface{} `fbx:"a"`
	TangentsW                interface{} `fbx:"a"`
	UV                       interface{} `fbx:"a"`
	UVIndex                  interface{} `fbx:"a"`
	Smoothing                interface{} `fbx:"a"`
	Materials                interface{} `fbx:"a"`
	Colors                   interface{} `fbx:"a"`
}

type LayerElement struct {
	Type       string
	TypedIndex int
}

type Layer struct {
	Id      uint64 `fbx:"p"`
	Version int

	LayerElement []LayerElement
}

type Geometry struct {
	Id      uint64 `fbx:"p"`
	Name    string `fbx:"p"`
	Element string `fbx:"p"`

	Properties70 Properties70

	Vertices           interface{} `fbx:"a"`
	PolygonVertexIndex interface{} `fbx:"a"`
	Edges              interface{} `fbx:"a"`
	GeometryVersion    int

	LayerElementNormal    *LayerElementShared
	LayerElementBinormal  *LayerElementShared
	LayerElementTangent   *LayerElementShared
	LayerElementUV        *LayerElementShared
	LayerElementSmoothing *LayerElementShared
	LayerElementMaterial  *LayerElementShared
	LayerElementColor     *LayerElementShared

	Layer *Layer
}

type Model struct {
	Id           uint64 `fbx:"p"`
	Name         string `fbx:"p"`
	Element      string `fbx:"p"`
	Version      int
	Properties70 Properties70
	Shading      bool
	Culling      string
}

type Material struct {
	Id           uint64 `fbx:"p"`
	Name         string `fbx:"p"`
	Element      string `fbx:"p"`
	Version      int
	Properties70 Properties70
	ShadingModel string
	MultiLayer   int
}

type Texture struct {
	Id                   uint64 `fbx:"p"`
	Name                 string `fbx:"p"`
	Element              string `fbx:"p"`
	Type                 string
	Version              int
	TextureName          string
	Properties70         Properties70
	FileName             string
	RelativeFilename     string
	ModelUVTranslation   []int `fbx:"i"`
	ModelUVScaling       []int `fbx:"i"`
	Texture_Alpha_Source string
	Cropping             []int `fbx:"i"`
}

type AnimationStack struct{}
type AnimationLayer struct{}
type NodeAttribute struct{}
type Video struct {
	Id      uint64 `fbx:"p"`
	Name    string `fbx:"p"`
	Element string `fbx:"p"`

	Type             string
	Content          string
	Properties70     Properties70
	UseMipMap        int
	Filename         string
	RelativeFilename string
}
type AnimationCurveNode struct{}
type AnimationCurve struct{}
type Pose struct{}

type Deformer struct {
	Id      uint64 `fbx:"p"`
	Name    string `fbx:"p"`
	Element string `fbx:"p"`

	Version       int
	UserData      string
	Indexes       interface{} `fbx:"a"`
	Weights       interface{} `fbx:"a"`
	Transform     interface{} `fbx:"a"`
	TransformLink interface{} `fbx:"a"`
}

type Objects struct {
	Geometry           []*Geometry
	Material           []*Material
	Video              []*Video
	Texture            []*Texture
	Model              []*Model
	AnimationStack     []*AnimationStack
	AnimationLayer     []*AnimationLayer
	NodeAttribute      []*NodeAttribute
	AnimationCurveNode []*AnimationCurveNode
	AnimationCurve     []*AnimationCurve
	Pose               []*Pose
	Deformer           []*Deformer
}

type Connection struct {
	Type   string   `fbx:"p"`
	Child  uint64   `fbx:"p"`
	Parent uint64   `fbx:"p"`
	Extra  []string `fbx:"p"`
}

type Connections struct {
	C []Connection
}
