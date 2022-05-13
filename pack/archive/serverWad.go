package archive

type ServerWad struct {
	PlaceholderName
	PlaceholderInstancesHolder
	PlaceholderReferencesHolder
}

func (sw *ServerWad) GetName() string { return "ServerWAD" }
