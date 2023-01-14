package app

import (
	"fmt"
	"log"
	"os"

	"github.com/mogaika/god_of_war_browser/config"
	"github.com/mogaika/god_of_war_browser/editor/imgui/project"
	"github.com/mogaika/god_of_war_browser/editor/imgui/resources"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/sqweek/dialog"
)

var app = application{}

type application struct {
	project *project.Project
	counter int
	size    [2]float32
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
		app.project = project.NewProject("")
	}

	wadPath := `/home/mogaika/gow/`
	debugWADS := []string{
		"R_PERMU.WAD",
		"R_SHELL.WAD",
		"ATHN01B.WAD",
		"ATHN05A.WAD",
		"DEST00.WAD",
		"OLYMP04.WAD",
		"PAND01C.WAD",
		"R_HERO1.WAD",
		"R_BRSRK0.WAD",
		"R_HERO0.WAD",
	}
	imgui.Separator()
	for _, wadName := range debugWADS {
		if imgui.Button(fmt.Sprintf("Open wad %q debug", wadName)) {
			app.project = project.NewProject("")

			config.SetGOWVersion(config.GOW1)
			f, err := os.Open(wadPath + wadName)
			if err != nil {
				log.Printf("Failed to open wad: %v", err)
			}
			defer f.Close()
			w, _ := resources.LoadWadFromReader(app.project, f, wadName)
			app.project.OpenResource(w)
		}
	}
	imgui.Separator()
	if imgui.Button("Open all debug wads") {
		config.SetGOWVersion(config.GOW1)
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
		app.project.RenderUI(app.size)
	}
}

func Project() *project.Project { return app.project }
