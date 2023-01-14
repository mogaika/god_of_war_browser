package uihelper

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type MenuAction struct {
	Label  string
	Action func()
}

func ShowMenuFromActions(actions []MenuAction, triggerEvent *gdk.Event) {
	if len(actions) == 0 {
		return
	}

	menu, err := gtk.MenuNew()
	if err != nil {
		panic(err)
	}

	for _, a := range actions {
		menuitem, err := gtk.MenuItemNewWithLabel(a.Label)
		if err != nil {
			panic(err)
		}
		menuitem.Connect("activate", a.Action)
		menu.Append(menuitem)
	}
	menu.ShowAll()
	menu.PopupAtPointer(triggerEvent)
}
