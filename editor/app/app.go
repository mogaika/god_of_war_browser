package app

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/editor/core"
	"github.com/mogaika/god_of_war_browser/editor/gow"
	"github.com/mogaika/god_of_war_browser/editor/model"
	"github.com/mogaika/god_of_war_browser/editor/view"
	"github.com/mogaika/god_of_war_browser/vfs"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/sqweek/dialog"
)

var app = application{}

func init() {
	if project, err := core.NewProject(""); err != nil {
		panic(err)
	} else {
		// app.project = project
		project.AddResource("game", &model.Texture{})
		project.AddResource("tweak_templates/animation/goHero", &model.Texture{})
		project.AddResource("tweak_templates/mfx/502", &model.Texture{})
		project.AddResource("tweak_templates/mfx/501", &model.Texture{})
		project.AddResource("tweak_templates/cloth/195", &model.Texture{})
		project.AddResource("textures/zopa", &model.Texture{})
		project.AddResource("archive/heh/level", &model.WadArchive{})
		project.AddResource("archive/heh/textures/test", &model.Texture{})
		project.AddResource("archive/heh/textures/test2", &model.Texture{})
	}
}

type application struct {
	project     *core.Project
	projectView view.ProjectEditorView
	size        [2]float32
}

func SetFramebufferSize(size [2]float32) { app.size = size }

func (app *application) renderUIProjectSelect() {
	imgui.SetNextWindowPos(imgui.Vec2{0, 0})
	imgui.SetNextWindowSize(imgui.Vec2{app.size[0], app.size[1]})

	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 0.0)

	imgui.BeginV("Project select", nil, imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoResize)

	imgui.Text("Hello, please create project to store parsed game files")
	if imgui.Button("Create new project") {
		filename, err := dialog.File().Filter(".eprj").Title("Create project").Save()
		log.Printf("Create: %q %v", filename, err)
	}
	if imgui.Button("Open existing project") {
		filename, err := dialog.File().Filter(".eprj").Title("Open project").Load()
		log.Printf("Open: %q %v", filename, err)
	}
	imgui.Separator()
	if imgui.Button("New temporary project") {
		if project, err := core.NewProject(""); err != nil {
			panic(err)
		} else {
			app.project = project
		}
	}

	wadPath := `./editor/cmd/testwads/`
	var debugWADS []string
	if files, err := os.ReadDir(wadPath); err == nil {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if strings.HasSuffix(strings.ToLower(file.Name()), ".wad") {
				debugWADS = append(debugWADS, file.Name())
			}
		}
	} else {
		imgui.Textf("Failed to load wads list: %v", err)
	}

	imgui.Separator()
	for _, wadName := range debugWADS {
		if imgui.Button(fmt.Sprintf("Open wad %q debug", wadName)) {
			dir := vfs.NewDirectoryDriver(wadPath)
			wadFile, _ := vfs.DirectoryGetFile(dir, wadName)
			project, _ := core.NewProject("./editor/cmd/testprojects/importtest")

			config.SetGOWVersion(config.GOW1)
			if err := gow.LoadWadFromReader(project, wadFile); err != nil {
				log.Printf("Failed to import wad %q: %v", wadFile.Name(), err)
			}
			app.project = project
		}
	}
	imgui.Separator()
	if imgui.Button("Open all debug wads") {
		dir := vfs.NewDirectoryDriver(wadPath)
		project, _ := core.NewProject("./editor/cmd/testprojects/importtest")

		config.SetGOWVersion(config.GOW1)
		for _, wadName := range debugWADS {
			wadFile, _ := vfs.DirectoryGetFile(dir, wadName)
			if err := gow.LoadWadFromReader(project, wadFile); err != nil {
				log.Printf("Failed to import wad %q: %v", wadFile.Name(), err)
			}
		}
		app.project = project

		/*
			app.project = project.NewProject("")
			for _, wadName := range debugWADS {
				f, err := os.Open(wadPath + wadName)
				if err != nil {
					log.Printf("Failed to open wad: %v", err)
				}
				defer f.Close()
				w, _ := resources.LoadWadFromReader(app.project, f, wadName)
				app.project.OpenResource(w)
			}
		*/
	}

	imgui.End()
	imgui.PopStyleVar()
}

func renderDebug() {
	imgui.Text(fmt.Sprintf("Application average %.3f ms/frame (%.1f FPS)", 1000.0/imgui.CurrentIO().Framerate(), imgui.CurrentIO().Framerate()))
	var gowVersion string
	switch config.GetGOWVersion() {
	case config.GOW1:
		gowVersion = "gow1"
	case config.GOW2:
		gowVersion = "gow2"
	case config.GOW2018:
		gowVersion = "gow2018"
	default:
		gowVersion = "unknown"
	}
	var psVersion string
	switch config.GetPlayStationVersion() {
	case config.PS2:
		psVersion = "ps2"
	case config.PS3:
		psVersion = "ps3"
	case config.PSVita:
		psVersion = "psvita"
	case config.PC:
		psVersion = "pc"
	}
	imgui.Text(fmt.Sprintf("Version: %v", gowVersion))
	imgui.Text(fmt.Sprintf("Console: %v", psVersion))
}

func RenderUI() {
	if app.project == nil {
		app.renderUIProjectSelect()
	} else {
		app.projectView.RenderUI(app.project, app.size)
	}
}

// func Project() *project.Project { return app.project }
