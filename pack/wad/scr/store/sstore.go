package store

import "github.com/mogaika/god_of_war_browser/pack/wad"

type ScriptLoader func([]byte, *wad.WadNodeRsrc) (interface{}, error)

var gScriptLoaders = make(map[string]ScriptLoader, 0)

func AddScriptLoader(name string, st ScriptLoader) {
	if _, already := gScriptLoaders[name]; already {
		panic(st)
	}
	gScriptLoaders[name] = st
}

func GetScriptLoader(name string) ScriptLoader {
	return gScriptLoaders[name]
}
