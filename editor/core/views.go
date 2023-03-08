package core

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/inkyblackness/imgui-go/v4"
)

type RefView[T Resource] struct {
	ref Ref[T]
}

func NewRefView[T Resource](ref Ref[T]) RefView[T] {
	return RefView[T]{
		ref: ref,
	}
}

func (rv RefView[T]) RenderUI(p *Project) {
	imgui.Selectable(fmt.Sprintf("ref: %q", rv.ref.Uid().String()))
}

var coreRefPkgPath = reflect.TypeOf(NewRef[Resource](uuid.Nil)).PkgPath()

func reflectViewSingleElement(p *Project, key string, rv reflect.Value) {
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}
	rt := rv.Type()

	typeName := rt.Name()

	if rt.Kind() == reflect.Struct && rt.PkgPath() == coreRefPkgPath && strings.HasPrefix(typeName, "Ref[") && p != nil {
		if !rv.IsZero() {
			reflectViewSingleElement(p, key, reflect.ValueOf(rv.Interface().(RefI).ResolveAny(p)))
		}
		return
	}

	imgui.TableNextRow()

	isFolder := false

	switch rt.Kind() {
	case reflect.Map, reflect.Slice, reflect.Struct, reflect.Interface, reflect.Array:
		isFolder = true
	}

	if rt.PkgPath() != "" {
		if rt.PkgPath() == coreRefPkgPath && strings.HasPrefix(typeName, "Ref[") {
			typeName = strings.Replace(typeName, "github.com/mogaika/god_of_war_browser/", "", 1)
		} else {
			typeName = path.Base(rt.PkgPath()) + "." + typeName
		}
	}

	if isFolder {
		imgui.TableNextColumn()
		open := imgui.TreeNodeV(key, imgui.TreeNodeFlagsSpanFullWidth)

		imgui.TableNextColumn()
		imgui.Text("")

		imgui.TableNextColumn()
		imgui.Text(typeName)

		if open {
			switch rt.Kind() {
			case reflect.Map:
				for iter := rv.MapRange(); iter.Next(); {
					reflectViewSingleElement(p, iter.Key().String(), iter.Value())
				}
			case reflect.Struct:
				for i := 0; i < rt.NumField(); i++ {
					if rv.Field(i).CanInterface() {
						reflectViewSingleElement(p, rt.Field(i).Name, rv.Field(i))
					}
				}
			case reflect.Array, reflect.Slice:
				for i := 0; i < rv.Len(); i++ {
					reflectViewSingleElement(p, fmt.Sprint(i), rv.Index(i))
				}
			}
			imgui.TreePop()
		}
	} else {
		imgui.TableNextColumn()
		imgui.TreeNodeV(key, imgui.TreeNodeFlagsLeaf|imgui.TreeNodeFlagsNoTreePushOnOpen|imgui.TreeNodeFlagsSpanFullWidth)

		imgui.TableNextColumn()
		switch rt.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			v := rv.Interface()
			imgui.Textf("0x%x | %d", v, v)
		case reflect.Bool, reflect.Complex64, reflect.Complex128, reflect.Float32, reflect.Float64:
			imgui.Textf("%v", rv.Interface())
		default:
			imgui.Textf("%q", rv.String())
		}

		imgui.TableNextColumn()
		imgui.Text(typeName)
	}
}

func ReflectViewAny(o any) {
	ReflectView(nil, o)
}

func ReflectView(p *Project, o any) {
	flags := imgui.TableFlagsBordersV | imgui.TableFlagsBordersOuterH | imgui.TableFlagsResizable | imgui.TableFlagsRowBg | imgui.TableFlagsNoBordersInBody

	rv := reflect.ValueOf(o)

	if imgui.BeginTableV("reflect", 3, flags, imgui.Vec2{}, 0) {
		reflectViewSingleElement(p, "", rv)
		imgui.EndTable()
	}
}
