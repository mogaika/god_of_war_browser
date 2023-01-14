package archive

type ServerWad struct {
	PlaceholderInstancesHolder
	PlaceholderReferencesHolder
}

func (sw *ServerWad) GetName() string { return "ServerWAD" }
