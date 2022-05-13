package archive

import (
	"github.com/mogaika/god_of_war_browser/pack"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/vfs"
	"strings"
)

type Provider interface {
	ListArchives() ([]string, error)
	GetArchive(name string) (*wad.Wad, error)
}

type ProviderPS2 struct {
	Dir vfs.Directory
}

func (ap *ProviderPS2) ListArchives() ([]string, error) {
	list := make([]string, 0)
	dirList, err := ap.Dir.List()
	if err != nil {
		return nil, err
	}
	for _, name := range dirList {
		if strings.HasSuffix(name, ".WAD") {
			list = append(list, name)
		}
	}
	return list, nil
}

func (ap *ProviderPS2) GetArchive(name string) (*wad.Wad, error) {
	data, err := pack.GetInstanceHandler(ap.Dir, strings.ToUpper(name)+".WAD")
	if err != nil {
		return nil, err
	}

	return data.(*wad.Wad), nil
}
