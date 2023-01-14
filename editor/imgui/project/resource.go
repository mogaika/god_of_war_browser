package project

import (
	"github.com/google/uuid"
)

type Kind string

type Resource struct {
	id   uuid.UUID
	kind Kind
	name string
	data IResource
}

type IResource interface {
	RenderUI()
	Kind() Kind
}

func (r *Resource) GetID() uuid.UUID   { return r.id }
func (r *Resource) GetName() string    { return r.name }
func (r *Resource) GetKind() Kind      { return r.kind }
func (r *Resource) GetData() IResource { return r.data }
