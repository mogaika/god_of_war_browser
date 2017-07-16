package utils

import "io"

type ResourceSource interface {
	Name() string
	Size() int64
	Save(in *io.SectionReader) error
}
