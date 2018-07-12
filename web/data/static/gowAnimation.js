'use strict';

var ga_instance;
var globalMul = 360; // 16*90;

function gaAnimationManager() {
    this.lastUpdateTime = window.performance.now() / 1000.0;
    this.matLayerAnimations = [];
    this.matSheetAnimations = [];
	this.objSkeletAnimations = [];
};

gaAnimationManager.prototype.update = function() {
    var currentTime = window.performance.now() / 1000.0;
    var dt = currentTime - this.lastUpdateTime;

    for (var i in this.matLayerAnimations) {
        this.matLayerAnimations[i].update(dt, currentTime);
    }
    for (var i in this.matSheetAnimations) {
        this.matSheetAnimations[i].update(dt, currentTime);
    }
    for (var i in this.objSkeletAnimations) {
        this.objSkeletAnimations[i].update(dt, currentTime);
    }

    this.lastUpdateTime = currentTime;
}

gaAnimationManager.prototype.addAnimation = function(anim) {
    switch (anim.type) {
        case 0:
            this.objSkeletAnimations.push(anim);
            break;
        case 8:
            this.matLayerAnimations.push(anim);
            break;
        case 9:
            this.matSheetAnimations.push(anim);
            break;
        default:
            console.error("Unknown animation type ", anim);
            break;
    }
}

gaAnimationManager.prototype.freeAnimation = function(anim) {
    switch (anim.type) {
        case 0:
            var id = this.objSkeletAnimations.indexOf(anim);
            if (id >= 0) {
                this.objSkeletAnimations.splice(id, 1);
            }
            break;
        case 8:
            var id = this.matLayerAnimations.indexOf(anim);
            if (id >= 0) {
                this.matLayerAnimations.splice(id, 1);
            }
            break;
        case 9:
            var id = this.matSheetAnimations.indexOf(anim);
            if (id >= 0) {
                this.matSheetAnimations.splice(id, 1);
            }
            break;
        default:
            console.error("Unknown animation type ", anim);
            break;
    }
}

function gaMatertialLayerAnimation(anim, act, stateIndex, material) {
    this.type = 8;
    this.enabled = false;
    this.anim = anim;
    this.act = act;
    this.stateIndex = stateIndex;
    this.material = material;
    this.time = 0.0;
    material.anims.push(this);
}

gaMatertialLayerAnimation.prototype.update = function(dt, currentTime) {
    var newTime = this.time + dt;
    if (!this.enabled) {
        return;
    }

    var stateDesc = this.act.StateDescrs[this.stateIndex];
    var dataType = this.anim.DataTypes[this.stateIndex];

    if (dataType.TypeId == 8) {
        var layer = this.material.layers[dataType.Param1 & 0x7f];

        for (var iData in stateDesc.Data) {
            var data = stateDesc.Data[iData];

            var floatStep = newTime / stateDesc.ImportantFloat;

            var step = Math.trunc(floatStep) % (data.Dm.DatasCount1 - 1);
            var nextStep = (step + 1) % data.Dm.DatasCount1;

            var blendFactor2 = floatStep - Math.floor(floatStep);
            var blendFactor1 = (1 - blendFactor2);

            for (var key in data.Stream) {
                var stream = data.Stream[key];
                switch (key) {
                    case "U":
                        layer.uvoffset[0] = stream[step] * blendFactor1 + stream[nextStep] * blendFactor2;
                        break;
                    case "V":
                        layer.uvoffset[1] = stream[step] * blendFactor1 + stream[nextStep] * blendFactor2;
                        break;
                }
            }
        }

        gr_instance.requireRedraw = true;
    } else {
        log.error("incorrect animation typeid");
    }

    this.time = newTime;
}

gaMatertialLayerAnimation.prototype.enable = function(enable) {
    this.enabled = (enable === undefined) || (!!enable);
}

function gaObjSkeletAnimation(anim, act, stateIndex, obj, mdl) {
    this.type = 0;
    this.anim = anim;
    this.stateIndex = stateIndex;
    this.act = act;
	this.object = obj;
    this.step = 0;

	this.jointLocalScale = Array(obj.Joints.length);
	this.jointLocalPos = Array(obj.Joints.length);
	this.jointLocalRots = Array(obj.Joints.length);
	this.jointLocalMatrices = Array(obj.Joints.length);
	this.jointGlobalMatrices = Array(obj.Joints.length);

	if (mdl.animation != undefined) {
		ga_instance.freeAnimation(mdl.animation);
	}

    this.mdl = mdl;
	mdl.animation = this;
	this.reset();
}

gaObjSkeletAnimation.prototype.recalcMatrices = function() {
	var obj = this.object;
	var mdl = this.mdl;
	
	for (var i = 0; i < obj.Joints.length; i++) {
		var joint = obj.Joints[i];

		var localRots = this.jointLocalRots[i];
		var localQ = quat.fromEuler(quat.create(), localRots[0], localRots[1], localRots[2]);
		var localRotMatrix = mat4.fromQuat(mat4.create(), localQ);

		var local = this.jointLocalMatrices[i];

		mat4.identity(local);
		//local = mat4.translate(local, local, this.jointLocalPos[i]);
		//local = mat4.mul(local, local, localRotMatrix);
		//local = mat4.mul(local, local, localRotMatrix);
		local = mat4.fromRotationTranslationScale(local, localQ, this.jointLocalPos[i], this.jointLocalScale[i])
		
		
		
		
		this.jointLocalMatrices[i] = local;
		
		var global = (joint.Parent >= 0) ? mat4.mul(mat4.create(), this.jointGlobalMatrices[joint.Parent], local) : local;
		this.jointGlobalMatrices[i] = global;
		
		var inverseBindPose = joint.IsSkinned ? obj.Matrixes3[joint.InvId] : mat4.create();
		
		var result = mat4.mul(mat4.create(), global, inverseBindPose);
		mdl.setJointMatrix(i, new Float32Array(result));
	}
}

gaObjSkeletAnimation.prototype.reset = function() {
	this.step = 0;
	var obj = this.object;
	
	for (var i = 0; i < obj.Joints.length; i++) {
		this.jointLocalRots[i] = obj.Vectors5[i];
		this.jointLocalPos[i] = obj.Vectors4[i];
		this.jointLocalScale[i] = obj.Vectors6[i];
		this.jointLocalMatrices[i] = mat4.create();
	}
}

gaObjSkeletAnimation.prototype.update = function(dt, currentTime) {
    var newTime = this.time + dt * 0.5;
    var stateDesc = this.act.StateDescrs[this.stateIndex];
    var dataType = this.anim.DataTypes[this.stateIndex];

    if (dataType.TypeId == 0) {
		var mdl = this.mdl;
		var skeleton = this.object.Joints;
		
		for (var iData in stateDesc.Data) {
			var data = stateDesc.Data[iData];

			for (var iStream in data.Stream) {
				var jointId = parseInt(iStream / 4);
				var coord = parseInt(iStream) % 4;
				
				
				var newValue = data.Stream[iStream][parseInt(this.step) % data.Stream[iStream].length] * globalMul;
				//console.log( data.Stream[iStream][parseInt(this.step) % data.Stream[iStream].length], "*", globalMul, "=", newValue);
				this.jointLocalRots[jointId][coord] = newValue;
			}
		}
		
		this.recalcMatrices();
		gr_instance.requireRedraw = true;
    } else {
        console.error("incorrect animation typeid");
    }
	
	this.step += 0.2;
    this.time = newTime;
}


function gaMatertialLayerAnimation(anim, act, stateIndex, material) {
    this.type = 8;
    this.enabled = false;
    this.anim = anim;
    this.act = act;
    this.stateIndex = stateIndex;
    this.material = material;
    this.time = 0.0;
    material.anims.push(this);
}

gaMatertialLayerAnimation.prototype.update = function(dt, currentTime) {
    var newTime = this.time + dt;
    if (!this.enabled) {
        return;
    }

    var stateDesc = this.act.StateDescrs[this.stateIndex];
    var dataType = this.anim.DataTypes[this.stateIndex];

    if (dataType.TypeId == 8) {
        var layer = this.material.layers[dataType.Param1 & 0x7f];

        for (var iData in stateDesc.Data) {
            var data = stateDesc.Data[iData];

            var floatStep = newTime / stateDesc.ImportantFloat;

            var step = Math.trunc(floatStep) % (data.Dm.DatasCount1 - 1);
            var nextStep = (step + 1) % data.Dm.DatasCount1;

            var blendFactor2 = floatStep - Math.floor(floatStep);
            var blendFactor1 = (1 - blendFactor2);

            for (var key in data.Stream) {
                var stream = data.Stream[key];
                switch (key) {
                    case "U":
                        layer.uvoffset[0] = stream[step] * blendFactor1 + stream[nextStep] * blendFactor2;
                        break;
                    case "V":
                        layer.uvoffset[1] = stream[step] * blendFactor1 + stream[nextStep] * blendFactor2;
                        break;
                }
            }
        }

        gr_instance.requireRedraw = true;
    } else {
        log.error("incorrect animation typeid");
    }

    this.time = newTime;
}

function gaInit() {
    if (ga_instance) {
        console.warn("gaAnimationManager already created", ga_instance);
        return;
    }
    ga_instance = new gaAnimationManager();
}