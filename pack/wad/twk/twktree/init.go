package twktree

import (
	"fmt"
)

func initTweakTemplateWadInfo() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	for i := 0; i < 6; i++ {
		dd.AddFieldA(fmt.Sprintf("Extra GO %d Name", i), FILE_TYPE_STRING, 0x18)
		dd.AddFieldN(fmt.Sprintf("Extra GO %d Cap", i), FILE_TYPE_INT8)
	}
	for i := 0; i < 5; i++ {
		dd.AddFieldN(fmt.Sprintf("Group %d Idx", i), FILE_TYPE_INT16)
		dd.AddField(fmt.Sprintf("Group %d", i), NewIndexedDirectoryDeclaration(NewFieldDeclaration(FILE_TYPE_INT8, 0)))
	}
	return dd
}

func initTweakTemplateWadInfoGroup() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	for i := 0; i < 74; i++ {
		dd.AddFieldA(fmt.Sprintf("GO %d Name", i), FILE_TYPE_STRING, 0x18)
	}
	return dd
}

func initTweakTemplateAI1() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("SoldierTemplate", FILE_TYPE_INT32)
	dd.AddFieldN("DecisionTreeTemplate", FILE_TYPE_INT32)
	dd.AddFieldN("Block Modifier Tweaks", FILE_TYPE_INT16)
	dd.AddFieldN("HitPoints", FILE_TYPE_INT32)
	dd.AddFieldN("MaxHitPoints", FILE_TYPE_INT32)
	dd.AddFieldN("Damage Multiplier", FILE_TYPE_FLOAT)
	dd.AddFieldN("AI Mode", FILE_TYPE_INT16)
	dd.AddFieldN("Fight Radius", FILE_TYPE_INT16)
	dd.AddFieldN("Second Fight Radius", FILE_TYPE_INT16)
	dd.AddFieldN("GoodGuy", FILE_TYPE_BOOL)
	dd.AddFieldN("SwitchView", FILE_TYPE_BOOL)
	dd.AddFieldN("Don't Reaquire", FILE_TYPE_BOOL)
	dd.AddFieldN("Ignore AI Block", FILE_TYPE_BOOL)
	dd.AddFieldN("Show Health Meter", FILE_TYPE_BOOL)
	dd.AddFieldN("Health Meter Idx", FILE_TYPE_BOOL)
	dd.AddFieldN("Has Second Radius", FILE_TYPE_BOOL)
	dd.AddFieldN("Get Stuck", FILE_TYPE_BOOL)
	dd.AddFieldN("Ignore Hits", FILE_TYPE_BOOL)
	dd.AddFieldN("Has Grid", FILE_TYPE_BOOL)
	dd.AddFieldN("Initial Underwater", FILE_TYPE_BOOL)
	dd.AddFieldN("Omnipotent", FILE_TYPE_BOOL)
	dd.AddFieldN("No Turn Damp Cancel", FILE_TYPE_BOOL)
	dd.AddFieldN("ScaleCenter", FILE_TYPE_FLOAT)
	dd.AddFieldN("ScaleRange", FILE_TYPE_FLOAT)
	dd.AddFieldN("First Contact Move", FILE_TYPE_INT16)
	dd.AddFieldN("Search Move", FILE_TYPE_INT16)
	dd.AddFieldN("World Idle Move", FILE_TYPE_INT16)
	dd.AddFieldN("Timed Retaliate Move", FILE_TYPE_INT16)
	dd.AddFieldN("Block", FILE_TYPE_FLOAT)
	dd.AddFieldN("Projectile Block", FILE_TYPE_FLOAT)
	dd.AddFieldN("CullRadius", FILE_TYPE_FLOAT)
	dd.AddFieldN("AlertRadius", FILE_TYPE_FLOAT)
	dd.AddFieldN("VisualRadius", FILE_TYPE_FLOAT)
	dd.AddFieldN("VisualDegree", FILE_TYPE_FLOAT)
	dd.AddFieldN("Prox Radius", FILE_TYPE_FLOAT)
	dd.AddFieldN("Safe Zone Radius", FILE_TYPE_FLOAT)
	dd.AddFieldN("WanderSpeed", FILE_TYPE_FLOAT)
	dd.AddFieldN("ChaseSpeed", FILE_TYPE_FLOAT)
	dd.AddFieldN("Toward Mult", FILE_TYPE_INT8)
	dd.AddFieldN("Away Mult", FILE_TYPE_INT8)
	dd.AddFieldN("Flight Height", FILE_TYPE_FLOAT)
	dd.AddFieldN("PauseTime", FILE_TYPE_INT16)
	dd.AddFieldN("Anticipate Jump Time", FILE_TYPE_INT16)
	dd.AddFieldN("Frames Till Lost", FILE_TYPE_INT8)
	dd.AddFieldN("Family Distract Time", FILE_TYPE_FLOAT)
	dd.AddFieldN("Auto Smash Time", FILE_TYPE_INT32)
	dd.AddFieldN("Inside Square Time", FILE_TYPE_INT16)
	dd.AddFieldN("Force Action Time", FILE_TYPE_INT16)
	dd.AddFieldN("Pass Frc Act Time", FILE_TYPE_INT16)
	dd.AddFieldN("Climb", FILE_TYPE_BOOL)
	dd.AddFieldN("Balance", FILE_TYPE_BOOL)
	dd.AddFieldN("Fly", FILE_TYPE_BOOL)
	dd.AddFieldN("Swim", FILE_TYPE_BOOL)
	dd.AddFieldN("Dive", FILE_TYPE_BOOL)
	dd.AddFieldN("DiveDash", FILE_TYPE_BOOL)
	dd.AddFieldN("Jump", FILE_TYPE_BOOL)
	dd.AddFieldN("DblJmp", FILE_TYPE_BOOL)
	dd.AddFieldN("Rope", FILE_TYPE_BOOL)
	dd.AddFieldN("Pole", FILE_TYPE_BOOL)
	dd.AddFieldN("Zipline", FILE_TYPE_BOOL)
	dd.AddFieldN("BiPinnedRope", FILE_TYPE_BOOL)
	dd.AddFieldN("LadderSlide", FILE_TYPE_BOOL)
	dd.AddFieldN("RopeSlide", FILE_TYPE_BOOL)
	dd.AddFieldN("InteractiveObjects", FILE_TYPE_BOOL)
	dd.AddFieldN("WallPullup", FILE_TYPE_BOOL)
	dd.AddFieldN("DiveIntoWater", FILE_TYPE_BOOL)
	dd.AddFieldN("Lean Turns", FILE_TYPE_BOOL)
	dd.AddFieldN("ID", FILE_TYPE_INT16)
	dd.AddFieldN("Level", FILE_TYPE_INT16)
	dd.AddFieldN("Circle Health Threshold", FILE_TYPE_INT16)
	for i := 1; i < 6; i++ {
		dd.AddFieldN(fmt.Sprintf("Circle Health Threshold %d", i), FILE_TYPE_INT16)
	}
	dd.AddFieldN("Circle Scale", FILE_TYPE_FLOAT)
	dd.AddFieldN("Circle X Offset", FILE_TYPE_FLOAT)
	dd.AddFieldN("Circle Y Offset", FILE_TYPE_FLOAT)
	dd.AddFieldN("Circle Z Offset", FILE_TYPE_FLOAT)

	return dd
}

func initTweakTemplateCSM() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Pl Anim1", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Pl Anim2", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("IO Anim1", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("IO Anim2", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("IO Idle Anim", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Orb Emitter Template", FILE_TYPE_INT32)
	dd.AddFieldN("Orb Emitter Template 2", FILE_TYPE_INT32)
	dd.AddFieldN("Orb Emitter Template 3", FILE_TYPE_INT32)
	dd.AddFieldN("Pickup", FILE_TYPE_INT16)
	dd.AddFieldN("Pickup Delay", FILE_TYPE_FLOAT)
	dd.AddFieldN("Chest Type", FILE_TYPE_INT16)
	dd.AddFieldN("Raiden Cycle Time", FILE_TYPE_FLOAT)
	dd.AddFieldN("Handle Y Range", FILE_TYPE_FLOAT)
	dd.AddFieldN("Handle XZ Range", FILE_TYPE_FLOAT)
	dd.AddFieldN("Handle Angle", FILE_TYPE_FLOAT)
	dd.AddFieldN("Handle Behind Range", FILE_TYPE_FLOAT)
	dd.AddFieldN("Show R2", FILE_TYPE_BOOL)
	dd.AddFieldN("Hold Use World", FILE_TYPE_BOOL)
	dd.AddFieldN("Breakable", FILE_TYPE_BOOL)
	return dd
}

func initTweakTemplateOrb() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Oscillate Freq", FILE_TYPE_FLOAT)
	dd.AddFieldN("Oscillate Ampl", FILE_TYPE_FLOAT)
	dd.AddFieldN("Oscillate Y Offset", FILE_TYPE_FLOAT)
	dd.AddFieldN("Life Span", FILE_TYPE_FLOAT)
	dd.AddFieldN("Spawn Time", FILE_TYPE_FLOAT)
	dd.AddFieldN("Attract Time", FILE_TYPE_FLOAT)
	dd.AddFieldN("Attract Radius", FILE_TYPE_FLOAT)
	dd.AddFieldN("Gravity", FILE_TYPE_FLOAT)
	dd.AddFieldN("Type", FILE_TYPE_INT16)
	dd.AddFieldN("Points", FILE_TYPE_FLOAT)
	dd.AddFieldA("Pickup Sound", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("GO Name", FILE_TYPE_STRING, 0x18)
	return dd
}

func initTweakTemplateOrbEmitter() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Health Template", FILE_TYPE_INT32)
	dd.AddFieldN("Magic Template", FILE_TYPE_INT32)
	dd.AddFieldN("Weapon Template", FILE_TYPE_INT32)
	dd.AddFieldA("Emitter Joint", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Emitter Radius", FILE_TYPE_FLOAT)
	dd.AddFieldA("Target Joint", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Target Radius", FILE_TYPE_FLOAT)
	dd.AddFieldN("Normal Health Prob", FILE_TYPE_FLOAT)
	dd.AddFieldN("Normal Health Min", FILE_TYPE_INT16)
	dd.AddFieldN("Normal Health Max", FILE_TYPE_INT16)
	dd.AddFieldN("Normal Magic Prob", FILE_TYPE_FLOAT)
	dd.AddFieldN("Normal Magic Min", FILE_TYPE_INT16)
	dd.AddFieldN("Normal Magic Max", FILE_TYPE_INT16)
	dd.AddFieldN("Normal Weapon Prob", FILE_TYPE_FLOAT)
	dd.AddFieldN("Normal Weapon Min", FILE_TYPE_INT16)
	dd.AddFieldN("Normal Weapon Max", FILE_TYPE_INT16)
	dd.AddFieldN("Health Threshold", FILE_TYPE_INT32)
	dd.AddFieldN("Low Health Prob", FILE_TYPE_FLOAT)
	dd.AddFieldN("Low Health Min", FILE_TYPE_INT16)
	dd.AddFieldN("Low Health Max", FILE_TYPE_INT16)
	dd.AddFieldN("Magic Threshold", FILE_TYPE_INT32)
	dd.AddFieldN("Low Magic Prob", FILE_TYPE_FLOAT)
	dd.AddFieldN("Low Magic Min", FILE_TYPE_INT16)
	dd.AddFieldN("Low Magic Max", FILE_TYPE_INT16)
	return dd
}

func initTweakTemplates() *VFSDirectoryDeclaration {
	tweakTemplates := NewDirectoryDeclaration()

	tweakTemplatesDeclaratoin := map[string]*VFSDirectoryDeclaration{
		"WadInfo":      initTweakTemplateWadInfo(),
		"WadInfoGroup": initTweakTemplateWadInfoGroup(),
		"AI1":          initTweakTemplateAI1(),
		"CSM":          initTweakTemplateCSM(),
		"Orb":          initTweakTemplateOrb(),
		"OrbEmitter":   initTweakTemplateOrbEmitter(),
	}

	for name, dd := range tweakTemplatesDeclaratoin {
		tweakTemplates.AddField(name, NewIndexedDirectoryDeclaration(dd))
	}

	return tweakTemplates
}

var rootDeclaration *VFSDirectoryDeclaration

func Root() *VFSDirectoryDeclaration { return rootDeclaration }

func init() {
	rootDeclaration = NewDirectoryDeclaration()
	rootDeclaration.AddField("TweakTemplates", initTweakTemplates())
}
