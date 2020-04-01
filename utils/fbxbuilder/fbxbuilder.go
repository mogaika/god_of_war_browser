package fbxbuilder

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mogaika/fbx/builders/bfbx73"

	"github.com/mogaika/fbx"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/pkg/errors"
)

const FBX_CREATOR = "FBX SDK/FBX Plugins version 2013.3 build=20121223"
const FBX_APPLICATION_VENDOR = "GoW Fan Commnuity"
const FBX_APPLICATION_NAME = "god_of_war_browser"
const FBX_APPLICATION_VERSION = "1.0"
const FBX_DATE_TIME_GMT = "01/01/1970 00:00:00.000"
const FBX_CREATION_TIME = "1970-01-01 10:00:00:000"

var FBX_FILE_ID []byte = []byte{
	0x28, 0xb3, 0x2a, 0xeb, 0xb6, 0x24, 0xcc, 0xc2,
	0xbf, 0xc8, 0xb0, 0x2a, 0xa9, 0x2b, 0xfc, 0xf1}

type FBXBuilder struct {
	f      *fbx.FBX
	c      map[wad.TagId]interface{}
	lastId int64
	files  map[string][]byte

	objects     *fbx.Node
	connections *fbx.Node
}

func NewFBXBuilder(filename string) *FBXBuilder {
	f := &FBXBuilder{
		c:           make(map[wad.TagId]interface{}),
		files:       make(map[string][]byte),
		lastId:      1000000,
		f:           fbx.NewFBX(7400),
		objects:     bfbx73.Objects(),
		connections: bfbx73.Connections(),
	}
	f.createHeaders(filename)
	return f
}

func (f *FBXBuilder) createHeaders(filename string) {
	f.Root().AddNodes(
		bfbx73.FBXHeaderExtension().AddNodes(
			bfbx73.FBXHeaderVersion(1003),
			bfbx73.FBXVersion(7400),
			bfbx73.EncryptionType(0),
			bfbx73.CreationTimeStamp().AddNodes(
				bfbx73.Version(1000),
				/*
					bfbx73.Year(int32(currentTime.Year())),
					bfbx73.Month(int32(currentTime.Month())),
					bfbx73.Day(int32(currentTime.Day())),
					bfbx73.Hour(int32(currentTime.Hour())),
					bfbx73.Minute(int32(currentTime.Minute())),
					bfbx73.Second(int32(currentTime.Second())),
					bfbx73.Millisecond(0),
				*/
				bfbx73.Year(1970),
				bfbx73.Month(1),
				bfbx73.Day(1),
				bfbx73.Hour(10),
				bfbx73.Minute(0),
				bfbx73.Second(0),
				bfbx73.Millisecond(0),
			),
			bfbx73.Creator(FBX_CREATOR),
			bfbx73.SceneInfo("GlobalInfo\x00\x01SceneInfo", "UserData").AddNodes(
				bfbx73.Type("UserData"),
				bfbx73.Version(100),
				bfbx73.MetaData().AddNodes(
					bfbx73.Version(100),
					bfbx73.Title(""),
					bfbx73.Subject(""),
					bfbx73.Author(""),
					bfbx73.Keywords(""),
					bfbx73.Revision(""),
					bfbx73.Comment(""),
				),
				bfbx73.Properties70().AddNodes(
					bfbx73.P("DocumentUrl", "KString", "Url", "", filename),
					bfbx73.P("SrcDocumentUrl", "KString", "Url", "", filename),
					bfbx73.P("Original", "Compound", "", ""),
					bfbx73.P("Original|ApplicationVendor", "KString", "", "", FBX_APPLICATION_VENDOR),
					bfbx73.P("Original|ApplicationName", "KString", "", "", FBX_APPLICATION_NAME),
					bfbx73.P("Original|ApplicationVersion", "KString", "", "", FBX_APPLICATION_VERSION),
					bfbx73.P("Original|DateTime_GMT", "DateTime", "", "", FBX_DATE_TIME_GMT),
					bfbx73.P("Original|FileName", "KString", "", "", filepath.Base(filename)),
					bfbx73.P("LastSaved", "Compound", "", ""),
					bfbx73.P("LastSaved|ApplicationVendor", "KString", "", "", FBX_APPLICATION_VENDOR),
					bfbx73.P("LastSaved|ApplicationName", "KString", "", "", FBX_APPLICATION_NAME),
					bfbx73.P("LastSaved|ApplicationVersion", "KString", "", "", FBX_APPLICATION_VERSION),
					bfbx73.P("LastSaved|DateTime_GMT", "DateTime", "", "", FBX_DATE_TIME_GMT),
				),
			),
		),
		bfbx73.FileId(FBX_FILE_ID),
		bfbx73.CreationTime(FBX_CREATION_TIME),
		// bfbx73.CreationTime(currentTime.Format("2006-01-02 15:04:05:000")),
		bfbx73.Creator(FBX_CREATOR),
		bfbx73.GlobalSettings().AddNodes(
			bfbx73.Version(1000),
			bfbx73.Properties70().AddNodes(
				bfbx73.P("UpAxis", "int", "Integer", "", int32(1)),
				bfbx73.P("UpAxisSign", "int", "Integer", "", int32(1)),
				bfbx73.P("FrontAxis", "int", "Integer", "", int32(2)),
				bfbx73.P("FrontAxisSign", "int", "Integer", "", int32(1)),
				bfbx73.P("CoordAxis", "int", "Integer", "", int32(0)),
				bfbx73.P("CoordAxisSign", "int", "Integer", "", int32(1)),
				bfbx73.P("OriginalUpAxis", "int", "Integer", "", int32(1)),
				bfbx73.P("OriginalUpAxisSign", "int", "Integer", "", int32(1)),
				bfbx73.P("UnitScaleFactor", "double", "Number", "", float64(1)),
				bfbx73.P("OriginalUnitScaleFactor", "double", "Number", "", float64(1)),
				bfbx73.P("AmbientColor", "ColorRGB", "Color", "", float64(0), float64(0), float64(0)),
			),
		),
		bfbx73.Documents().AddNodes(
			bfbx73.Count(1),
			bfbx73.Document(f.GenerateId(), "Scene", "Scene").AddNodes(
				bfbx73.Properties70().AddNodes(
					bfbx73.P("SourceObject", "object", "", ""),
					bfbx73.P("ActiveAnimStackName", "KString", "", "", ""),
				),
				bfbx73.RootNode(0),
			),
		),
		bfbx73.References(),
		bfbx73.Definitions().AddNodes(
			bfbx73.Version(100),
			bfbx73.Count(1),
			bfbx73.ObjectType("GlobalSettings").AddNodes(
				bfbx73.Count(1),
			),
			bfbx73.ObjectType("Model").AddNodes(
				bfbx73.Count(0),
				bfbx73.PropertyTemplate("FbxNode").AddNodes(
					bfbx73.Properties70().AddNodes(
						bfbx73.P("QuaternionInterpolate", "enum", "", "", int32(0)),
						bfbx73.P("Show", "bool", "", "", int32(1)),
						bfbx73.P("Lcl Translation", "Lcl Translation", "", "A", float64(0), float64(0), float64(0)),
						bfbx73.P("Lcl Rotation", "Lcl Rotation", "", "A", float64(0), float64(0), float64(0)),
						bfbx73.P("Lcl Scaling", "Lcl Scaling", "", "A", float64(1), float64(1), float64(1)),
						bfbx73.P("Visibility", "Visibility", "", "A", float64(1)),
						bfbx73.P("Visibility Inheritance", "Visibility Inheritance", "", "", int32(1)),
					),
				),
			),
			bfbx73.ObjectType("Material").AddNodes(
				bfbx73.Count(0),
				bfbx73.PropertyTemplate("FbxSurfacePhong").AddNodes(
					bfbx73.Properties70().AddNodes(
						bfbx73.P("ShadingModel", "KString", "", "", "Phong"),
						bfbx73.P("MultiLayer", "bool", "", "", int32(0)),
						bfbx73.P("EmissiveColor", "Color", "", "A", float64(0), float64(0), float64(0)),
						bfbx73.P("EmissiveFactor", "Number", "", "A", float64(1)),
						bfbx73.P("AmbientColor", "Color", "", "A", float64(0.2), float64(0.2), float64(0.2)),
						bfbx73.P("AmbientFactor", "Number", "", "A", float64(1)),
						bfbx73.P("DiffuseColor", "Color", "", "A", float64(1), float64(1), float64(1)),
						bfbx73.P("DiffuseFactor", "Number", "", "A", float64(1)),
						bfbx73.P("SpecularColor", "Color", "", "A", float64(0.2), float64(0.2), float64(0.2)),
						bfbx73.P("SpecularFactor", "Number", "", "A", float64(1)),
					),
				),
			),
			bfbx73.ObjectType("Texture").AddNodes(
				bfbx73.Count(0),
				bfbx73.PropertyTemplate("FbxFileTexture").AddNodes(
					bfbx73.Properties70().AddNodes(
						bfbx73.P("TextureTypeUse", "enum", "", "", int32(0)),
						bfbx73.P("Texture alpha", "Number", "", "A", float64(1)),
						bfbx73.P("CurrentMappingType", "enum", "", "", int32(0)),
						bfbx73.P("WrapModeU", "enum", "", "", int32(0)),
						bfbx73.P("WrapModeV", "enum", "", "", int32(0)),
						bfbx73.P("UVSwap", "bool", "", "", int32(0)),
						bfbx73.P("PremultiplyAlpha", "bool", "", "", int32(1)),
						bfbx73.P("UseMaterial", "bool", "", "", int32(0)),
						bfbx73.P("UseMipMap", "bool", "", "", int32(0)),
					),
				),
			),
			bfbx73.ObjectType("Video").AddNodes(
				bfbx73.Count(0),
				bfbx73.PropertyTemplate("FbxVideo").AddNodes(
					bfbx73.Properties70().AddNodes(
						bfbx73.P("ImageSequence", "bool", "", "", int32(0)),
						bfbx73.P("Width", "int", "Integer", "", int32(0)),
						bfbx73.P("Height", "int", "Integer", "", int32(0)),
						bfbx73.P("Path", "KString", "XRefUrl", "", ""),
					),
				),
			),
			bfbx73.ObjectType("Geometry").AddNodes(
				bfbx73.Count(0),
				bfbx73.PropertyTemplate("FbxMesh").AddNodes(
					bfbx73.Properties70().AddNodes(
						bfbx73.P("Color", "ColorRGB", "Color", "", float64(1), float64(1), float64(1)),
						bfbx73.P("Primary Visibility", "bool", "", "", int32(1)),
						bfbx73.P("Casts Shadows", "bool", "", "", int32(1)),
						bfbx73.P("Receive Shadows", "bool", "", "", int32(1)),
					),
				),
			),
			bfbx73.ObjectType("NodeAttribute").AddNodes(
				bfbx73.Count(0),
				bfbx73.PropertyTemplate("FbxNull").AddNodes(
					bfbx73.Properties70().AddNodes(
						bfbx73.P("Size", "double", "Number", "", float64(100)),
						bfbx73.P("Look", "enum", "", "", int32(1)),
					),
				),
			),
		),
		f.objects,
		f.connections,
		bfbx73.Takes().AddNodes(
			bfbx73.Current(""),
		),
	)
}

func (f *FBXBuilder) countDefinitions() {
	counts := make(map[string]int32)
	for _, object := range f.objects.Nodes {
		if count, ex := counts[object.Name]; ex {
			counts[object.Name] = count + 1
		} else {
			counts[object.Name] = 1
		}
	}

	definitions := f.Root().GetNode("Definitions")
	totalCount := int32(1) // 1 for GlobalSettings

	for name, count := range counts {
		totalCount += count

		var objectType *fbx.Node
		for _, ot := range definitions.GetNodes("ObjectType") {
			if ot.Properties[0].(string) == name {
				objectType = ot
			}
		}
		if objectType == nil {
			objectType = bfbx73.ObjectType(name)
			definitions.AddNode(objectType)
		}

		objectType.GetOrAddNode(bfbx73.Count(0)).Properties[0] = count
		log.Printf("counting %v %v %v", name, count)
	}

	definitions.GetOrAddNode(bfbx73.Count(0)).Properties[0] = totalCount
}

func (f *FBXBuilder) Root() *fbx.Node {
	return &f.f.Root
}

func (f *FBXBuilder) AddCache(id wad.TagId, d interface{}) {
	f.c[id] = d
}

func (f *FBXBuilder) GetCached(id wad.TagId) interface{} {
	if v, e := f.c[id]; e {
		return v
	} else {
		return nil
	}
}

func (f *FBXBuilder) GenerateId() int64 {
	f.lastId++
	return f.lastId
}

// TODO: remove this shitty workaround with tempfile
func (f *FBXBuilder) Write(w io.Writer) error {
	f.countDefinitions()

	print(f.f.SPrint())
	f.f.PrintConnections(0, 0)

	tempFile, err := ioutil.TempFile("", "fbxexport.*.fbx")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	if err := fbx.Write(tempFile, f.f); err != nil {
		return err
	}

	if _, err := tempFile.Seek(0, os.SEEK_SET); err != nil {
		return errors.Wrapf(err, "Unable to seek")
	}
	_, err = io.Copy(w, tempFile)
	return err
}

func (f *FBXBuilder) AddExportFile(name string, data []byte) {
	f.files[name] = data
}

func (f *FBXBuilder) WriteZip(w io.Writer, name string) error {
	zw := zip.NewWriter(w)

	fbxW, err := zw.Create(name)
	if err != nil {
		return errors.Wrapf(err, "Can't create zip fbx for %q", name)
	}
	if err := f.Write(fbxW); err != nil {
		return errors.Wrapf(err, "Fbx exporting failed")
	}

	for name, file := range f.files {
		fw, err := zw.Create(name)
		if err != nil {
			return errors.Wrapf(err, "Can't create zip for %q", name)
		}
		if _, err := fw.Write(file); err != nil {
			return errors.Wrapf(err, "Can't write zip for %q", name)
		}
	}

	return zw.Close()
}

func (f *FBXBuilder) AddObjects(nodes ...*fbx.Node)     { f.objects.AddNodes(nodes...) }
func (f *FBXBuilder) AddConnections(nodes ...*fbx.Node) { f.connections.AddNodes(nodes...) }
