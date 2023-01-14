package uit

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/mogaika/god_of_war_browser/editor/storage"
)

type ReferenceView struct {
	Label    string
	Resource *storage.Resource

	button *gtk.Button
}

func (rv *ReferenceView) updateLabelText() {
	if rv.Resource == nil {
		rv.button.SetLabel("(not selected)")
	} else {
		rv.button.SetLabel(rv.Resource.Name().Get())
	}
}

func (rv *ReferenceView) Create() (label string, w gtk.IWidget) {
	b, err := gtk.ButtonNew()
	if err != nil {
		panic(err)
	}
	rv.button = b
	rv.updateLabelText()
	if rv.Resource != nil {
		rv.Resource.Name().ConnectOnChange(rv, rv.updateLabelText)
	}
	rv.button.Connect("destroy", func() {
		if rv.Resource != nil {
			rv.Resource.Name().DisconnectOnChange(rv)
		}
	})
	return rv.Label, b
}

type FieldView interface {
	Create() (string, gtk.IWidget)
}

type Form struct {
	Inputs []FieldView
}

func (f *Form) Build() gtk.IWidget {
	grid, err := gtk.GridNew()
	if err != nil {
		panic(err)
	}

	grid.SetColumnHomogeneous(true)
	grid.SetRowSpacing(2)
	grid.SetColumnSpacing(4)

	for i, in := range f.Inputs {
		lt, w := in.Create()
		l, _ := gtk.LabelNew(lt)

		grid.Attach(l, 0, i, 1, 1)
		grid.Attach(w, 1, i, 1, 1)
	}
	grid.ShowAll()
	return grid
}
