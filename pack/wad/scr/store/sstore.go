package store

type ScriptLoader func([]byte) interface{}

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
