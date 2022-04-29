package twktree

import (
	"fmt"
)

var rootDeclaration *VFSDirectoryDeclaration

func Root() *VFSDirectoryDeclaration { return rootDeclaration }

func init() {
	rootDeclaration = NewDirectoryDeclaration()
	rootDeclaration.AddField("TweakTemplates", initTweakTemplates())
}

func initTweakTemplates() *VFSDirectoryDeclaration {
	tweakTemplates := NewDirectoryDeclaration()

	tweakTemplatesDeclaratoin := map[string]*VFSDirectoryDeclaration{
		"WadInfo":       initTweakTemplateWadInfo(),
		"WadInfoGroup":  initTweakTemplateWadInfoGroup(),
		"AI1":           initTweakTemplateAI1(),
		"CSM":           initTweakTemplateCSM(),
		"Orb":           initTweakTemplateOrb(),
		"OrbEmitter":    initTweakTemplateOrbEmitter(),
		"MFX":           initTweakTemplateMFX(),
		"PlayFX":        initTweakTemplatePlayFX(),
		"ForceFeedback": initTweakTemplateForceFeedback(),
		"Bonus":         initTweakTemplateBonus(),
		"RepeatPress":   initTweakTemplateRepeatPress(),
		"CameraShake":   initTweakTemplateCameraShake(),
		"CombatFileSet": initTweakTemplateCombatFileSet(),
		"Brkb":          initTweakTemplateBrkb(),
		"CameraFilter":  initTweakTemplateCameraFilter(),
		"TimedPress":    initTweakTemplateTimedPress(),
		"TAndF":         initTweakTemplateTAndF(),
		"Translate":     initTweakTemplateTranslate(),
	}

	for name, dd := range tweakTemplatesDeclaratoin {
		tweakTemplates.AddField(name, NewIndexedDirectoryDeclaration(dd))
	}

	return tweakTemplates
}

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
	dd.AddFieldN("@hash(a7cf0c6b)", FILE_TYPE_BOOL) // not used by engine? (at least vita version)
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

func initTweakTemplateMFX() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Underwater Alternate", FILE_TYPE_INT16)
	dd.AddFieldA("Hit Particle System", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Deflect Arrows", FILE_TYPE_BOOL)
	dd.AddFieldN("Drop Blood", FILE_TYPE_BOOL)
	dd.AddFieldN("Ignore Invulnerable", FILE_TYPE_BOOL)
	dd.AddFieldN("Drop Blood Idx", FILE_TYPE_INT8)
	dd.AddFieldA("Foot Sound", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Jump Land Effect", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Slice Hit Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Slice Hit Sound", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Run Dust Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Run Dust Attach", FILE_TYPE_BOOL)
	dd.AddFieldN("Run Dust Orient", FILE_TYPE_BOOL)
	dd.AddFieldN("Arrow Sound Index", FILE_TYPE_INT8)
	dd.AddFieldN("Hit Camera Shake", FILE_TYPE_INT16)
	dd.AddFieldN("Movement X", FILE_TYPE_FLOAT)
	dd.AddFieldN("Movement Z", FILE_TYPE_FLOAT)
	return dd
}

func initTweakTemplatePlayFX() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Effect Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Joint Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Particle Effect", FILE_TYPE_BOOL)
	dd.AddFieldN("Looped", FILE_TYPE_BOOL)
	dd.AddFieldN("Driven Animation", FILE_TYPE_BOOL)
	dd.AddFieldN("Attached", FILE_TYPE_BOOL)
	dd.AddFieldN("Leave", FILE_TYPE_BOOL)
	dd.AddFieldN("Left Weapon", FILE_TYPE_BOOL)
	dd.AddFieldN("Right Weapon", FILE_TYPE_BOOL)
	dd.AddFieldN("Remove", FILE_TYPE_BOOL)
	dd.AddFieldN("Blood Drop", FILE_TYPE_BOOL)
	dd.AddFieldN("Scale X", FILE_TYPE_FLOAT)
	dd.AddFieldN("Scale Y", FILE_TYPE_FLOAT)
	dd.AddFieldN("Scale Z", FILE_TYPE_FLOAT)
	dd.AddFieldN("Scale Num Part", FILE_TYPE_FLOAT)
	dd.AddFieldN("Offset X", FILE_TYPE_FLOAT)
	dd.AddFieldN("Offset Y", FILE_TYPE_FLOAT)
	dd.AddFieldN("Offset Z", FILE_TYPE_FLOAT)
	dd.AddFieldN("Rot X", FILE_TYPE_FLOAT)
	dd.AddFieldN("Rot Y", FILE_TYPE_FLOAT)
	dd.AddFieldN("Rot Z", FILE_TYPE_FLOAT)
	return dd
}

func initTweakTemplateForceFeedback() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Duration", FILE_TYPE_FLOAT)
	dd.AddFieldN("SM frequency", FILE_TYPE_FLOAT)
	dd.AddFieldN("SM amplitude", FILE_TYPE_FLOAT)
	dd.AddFieldN("SM phase", FILE_TYPE_FLOAT)
	dd.AddFieldN("SM bias", FILE_TYPE_FLOAT)
	dd.AddFieldN("SM Waveform", FILE_TYPE_INT16)
	dd.AddFieldN("LM frequency", FILE_TYPE_FLOAT)
	dd.AddFieldN("LM amplitude", FILE_TYPE_FLOAT)
	dd.AddFieldN("LM phase", FILE_TYPE_FLOAT)
	dd.AddFieldN("LM bias", FILE_TYPE_FLOAT)
	dd.AddFieldN("LM Waveform", FILE_TYPE_INT16)
	dd.AddFieldN("Attack", FILE_TYPE_FLOAT)
	dd.AddFieldN("Decay", FILE_TYPE_FLOAT)
	dd.AddFieldN("Sustain", FILE_TYPE_FLOAT)
	dd.AddFieldN("Release", FILE_TYPE_FLOAT)
	for i := 0; i < 4; i++ {
		dd.AddFieldN(fmt.Sprintf("Input %d Bias", i), FILE_TYPE_FLOAT)
		dd.AddFieldN(fmt.Sprintf("Input %d Scale", i), FILE_TYPE_FLOAT)
		dd.AddFieldN(fmt.Sprintf("Input %d Operator", i), FILE_TYPE_INT16)
	}
	return dd
}

func initTweakTemplateBonus() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Exp Points", FILE_TYPE_FLOAT)
	dd.AddFieldN("Message Index", FILE_TYPE_INT32)
	return dd
}

func initTweakTemplateRepeatPress() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Total Time", FILE_TYPE_FLOAT)
	dd.AddFieldN("Forward Amount", FILE_TYPE_FLOAT)
	dd.AddFieldN("Push Back Amount", FILE_TYPE_FLOAT)
	dd.AddFieldN("Slip Tolerance", FILE_TYPE_FLOAT)
	dd.AddFieldN("Visual Damping", FILE_TYPE_FLOAT)
	dd.AddFieldN("Pressure Factor", FILE_TYPE_FLOAT)
	dd.AddFieldN("Quick Penalty Time", FILE_TYPE_FLOAT)
	dd.AddFieldN("Fail Move", FILE_TYPE_INT16)
	dd.AddFieldN("Pass Move", FILE_TYPE_INT16)
	dd.AddFieldN("Button", FILE_TYPE_INT16)
	dd.AddFieldA("Master Replace", FILE_TYPE_STRING, 0x18)
	return dd
}

func initTweakTemplateCameraShake() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Duration", FILE_TYPE_FLOAT)
	dd.AddFieldN("Angle Amplitude", FILE_TYPE_FLOAT)
	dd.AddFieldN("Pos Amplitude", FILE_TYPE_FLOAT)
	dd.AddFieldN("Frequency", FILE_TYPE_FLOAT)
	return dd
}

func initTweakTemplateCombatFileSet() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("ID", FILE_TYPE_INT16)                      // Not used by engine?
	dd.AddFieldN("Circle Health Threshold", FILE_TYPE_INT16) // Not used by engine?
	dd.AddFieldN("Circle Scale", FILE_TYPE_FLOAT)            // Not used by engine?
	dd.AddFieldN("Circle X Offset", FILE_TYPE_FLOAT)         // Not used by engine?
	dd.AddFieldN("Circle Y Offset", FILE_TYPE_FLOAT)         // Not used by engine?
	dd.AddFieldN("Circle Z Offset", FILE_TYPE_FLOAT)         // Not used by engine?
	dd.AddFieldN("Reticle Scale", FILE_TYPE_FLOAT)
	dd.AddFieldN("Reticle Y Offset", FILE_TYPE_FLOAT)
	dd.AddFieldN("Target Weight", FILE_TYPE_FLOAT)
	dd.AddFieldN("Impulse Away Scale", FILE_TYPE_FLOAT)
	dd.AddFieldN("Impulse Up Scale", FILE_TYPE_FLOAT)
	dd.AddFieldN("Impulse Right Scale", FILE_TYPE_FLOAT)
	dd.AddFieldN("No Aiming", FILE_TYPE_BOOL)
	dd.AddFieldN("Invulnerable", FILE_TYPE_BOOL)
	dd.AddFieldN("Trail Left", FILE_TYPE_BOOL)
	dd.AddFieldN("Trail Right", FILE_TYPE_BOOL)
	dd.AddFieldN("Friendly Fire", FILE_TYPE_BOOL)
	dd.AddFieldN("@hash(f184266d)", FILE_TYPE_INT16) // Not used by engine?
	dd.AddFieldN("Physics", FILE_TYPE_INT16)
	for i := 0; i < 8; i++ {
		dd.AddFieldN(fmt.Sprintf("File %d", i), FILE_TYPE_INT16)
	}
	return dd
}

func initTweakTemplateBrkb() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Type", FILE_TYPE_INT16)
	dd.AddFieldN("Pickup", FILE_TYPE_INT16)
	dd.AddFieldN("Hit Points", FILE_TYPE_FLOAT)
	dd.AddFieldN("Opaque Time", FILE_TYPE_FLOAT)
	dd.AddFieldN("Fade Time", FILE_TYPE_FLOAT)
	dd.AddFieldN("Orb Emitter Template", FILE_TYPE_INT16)
	dd.AddFieldN("Air Orb Emitter Template", FILE_TYPE_INT16)
	dd.AddFieldN("Break Bonus Template", FILE_TYPE_INT16)
	dd.AddFieldN("Air Break Bonus Template", FILE_TYPE_INT16)
	dd.AddFieldN("Shard Physics", FILE_TYPE_BOOL)
	dd.AddFieldN("Force Feedback", FILE_TYPE_BOOL)
	return dd
}

func initTweakTemplateCameraFilter() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Top:R", FILE_TYPE_FLOAT)
	dd.AddFieldN("Top:G", FILE_TYPE_FLOAT)
	dd.AddFieldN("Top:B", FILE_TYPE_FLOAT)
	dd.AddFieldN("Top:A", FILE_TYPE_FLOAT)
	dd.AddFieldN("Middle:R", FILE_TYPE_FLOAT)
	dd.AddFieldN("Middle:G", FILE_TYPE_FLOAT)
	dd.AddFieldN("Middle:B", FILE_TYPE_FLOAT)
	dd.AddFieldN("Middle:A", FILE_TYPE_FLOAT)
	dd.AddFieldN("Bottom:R", FILE_TYPE_FLOAT)
	dd.AddFieldN("Bottom:G", FILE_TYPE_FLOAT)
	dd.AddFieldN("Bottom:B", FILE_TYPE_FLOAT)
	dd.AddFieldN("Bottom:A", FILE_TYPE_FLOAT)
	dd.AddFieldN("Toppoint", FILE_TYPE_FLOAT)
	dd.AddFieldN("Midpoint", FILE_TYPE_FLOAT)
	dd.AddFieldN("BottomPoint", FILE_TYPE_FLOAT)
	dd.AddFieldN("Blend Mode", FILE_TYPE_INT16)
	return dd
}

func initTweakTemplateTimedPress() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Button", FILE_TYPE_INT16)
	dd.AddFieldA("On Sound Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Press Sound Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Max Wrong Zones", FILE_TYPE_INT8)
	return dd
}

func initTweakTemplateTAndF() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Acc", FILE_TYPE_FLOAT)
	dd.AddFieldN("Dec", FILE_TYPE_FLOAT)
	dd.AddFieldN("TF Time", FILE_TYPE_FLOAT)
	dd.AddFieldN("Max Windup Speed", FILE_TYPE_FLOAT)
	dd.AddFieldN("Max Unwind Speed", FILE_TYPE_FLOAT)
	dd.AddFieldA("Pl Grab Anim", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Pl TandF Anim", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Pl Flourish Anim", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("IO TandF Anim", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("IO Flourish Anim", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("IO Idle Anim", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Grab Sound", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Reset Sound", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("Flourish Sound", FILE_TYPE_STRING, 0x18)
	dd.AddFieldA("TandF Loop", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Full Volume Y Vel", FILE_TYPE_FLOAT)
	dd.AddFieldN("Max Volume Change", FILE_TYPE_INT16)
	dd.AddFieldN("Use Sync Joint", FILE_TYPE_BOOL)
	dd.AddFieldN("Final Sync Joint Y", FILE_TYPE_FLOAT)
	dd.AddFieldN("Handle Y Range", FILE_TYPE_FLOAT)
	dd.AddFieldN("Handle XZ Range", FILE_TYPE_FLOAT)
	dd.AddFieldN("Handle Angle", FILE_TYPE_FLOAT)
	dd.AddFieldN("Handle Behind Range", FILE_TYPE_FLOAT)
	dd.AddFieldN("Show R2", FILE_TYPE_BOOL)
	dd.AddFieldN("Ratch Click FCount", FILE_TYPE_INT32)
	dd.AddFieldA("Ratch Forward Sound", FILE_TYPE_STRING, 0x18)
	dd.AddFieldN("Ratch Click BCount", FILE_TYPE_INT32)
	dd.AddFieldA("Ratch Backward Sound", FILE_TYPE_STRING, 0x18)
	return dd
}

func initTweakTemplateTranslate() *VFSDirectoryDeclaration {
	dd := NewDirectoryDeclaration()
	dd.AddFieldA("Instance Name", FILE_TYPE_STRING, 0x18)
	for i := 0; i < 10; i++ {
		dd.AddFieldN(fmt.Sprintf("Template %d", i), FILE_TYPE_INT16)
	}
	return dd
}
