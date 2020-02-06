package fbx

import (
	"time"
)

func NewFbx() *FBX {
	timenow := time.Now()
	return &FBX{
		FBXHeaderExtension: FBXHeaderExtension{
			FBXHeaderVersion: 1003,
			FBXVersion:       7400,
			Creator:          "God Of War Browser (GowBrowser) 1.0.0",
			CreationTimeStamp: &CreationTimeStamp{
				Version:     1000,
				Year:        timenow.Year(),
				Month:       int(timenow.Month()),
				Day:         timenow.Day(),
				Hour:        timenow.Hour(),
				Minute:      timenow.Minute(),
				Second:      timenow.Second(),
				Millisecond: timenow.Nanosecond() / 1000000,
			},
		},
		Documents: Documents{
			Count: 1,
			Document: []*Document{
				&Document{
					Id:       1000000,
					Element:  "Scene",
					RootNode: 0,
				},
			},
		},
		files: make(map[string][]byte),
	}
}

func (f *FBX) CountDefinitions() {
	allcount := 0

	doCount := func(count int, name string) {
		if count > 0 {
			allcount += len(f.Objects.Model)
			f.Definitions.ObjectType = append(f.Definitions.ObjectType, &ObjectTypeDefinition{
				Name: name, Count: count,
			})
		}
	}

	if f.GlobalSettings != nil {
		doCount(1, "GlobalSettings")
	}

	doCount(len(f.Objects.Model), "Model")
	doCount(len(f.Objects.Geometry), "Geometry")
	doCount(len(f.Objects.Material), "Material")
	doCount(len(f.Objects.Texture), "Texture")
	doCount(len(f.Objects.AnimationStack), "AnimationStack")
	doCount(len(f.Objects.AnimationLayer), "AnimationLayer")
	doCount(len(f.Objects.NodeAttribute), "NodeAttribute")
	doCount(len(f.Objects.Video), "Video")
	doCount(len(f.Objects.AnimationCurveNode), "AnimationCurveNode")
	doCount(len(f.Objects.AnimationCurve), "AnimationCurve")
	doCount(len(f.Objects.Pose), "Pose")
	doCount(len(f.Objects.Deformer), "Deformer")

	f.Definitions.Version = 100
	f.Definitions.Count = allcount
}
