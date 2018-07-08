'use strict';

var ga_instance;

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

function gaObjSkeletAnimation(anim, act, stateIndex, mdl) {
    this.type = 0;
    this.anim = anim;
    this.stateIndex = stateIndex;
    this.act = act;
    this.mdl = mdl;
    this.step = 0;

    var skeleton = mdl.getSkeleton();
	for (var i in skeleton) {
		mdl.setNodeMatrix(i, skeleton[i].RenderMat);
	}
}

gaObjSkeletAnimation.prototype.update = function(dt, currentTime) {
    var newTime = this.time + dt * 0.5;
    var stateDesc = this.act.StateDescrs[this.stateIndex];
    var dataType = this.anim.DataTypes[this.stateIndex];

    if (dataType.TypeId == 0) {
		var mdl = this.mdl;
		var skeleton = mdl.getSkeleton();
		
		for (var iData in stateDesc.Data) {
			var data = stateDesc.Data[iData];

			for (var iStream in data.Stream) {
				var bone = parseInt(iStream / 4);
				var coord = parseInt(iStream) % 4;
				
				var rots = [0, 0, 0];
				rots[coord] = data.Stream[iStream][parseInt(this.step) % data.Stream[iStream].length] * 16 * 90.0;
				var localRot = quat.fromEuler(quat.create(), rots[0], rots[1], rots[2]);
				var joint = skeleton[bone];

				//mdl.setNodeMatrix(bone, mat4.mul(mat4.create(), mat4.fromQuat(mat4.create(), localRot), mdl.skeleton[bone].BindToJointMat));
				//mdl.setNodeMatrix(bone, mat4.fromQuat(mat4.create(), localRot));
				//mdl.setNodeMatrix(bone, mat4.mul(mat4.create(), joint.OurJointToIdleMat, mat4.fromQuat(mat4.create(), localRot)));
				var rotmatrix = mat4.fromQuat(mat4.create(), localRot);
				

				//var result = mat4.mul(mat4.create(), joint.BindToJointMat, rotmatrix);
				// result = mat4.mul(mat4.create(), rotmatrix, result);
				//mdl.setNodeMatrix(bone, rotmatrix);
				joint.TransformMat = rotmatrix;
			}
		}
		
		for (var iJoint in skeleton) {
			var j = skeleton[iJoint];
	
			if (iJoint != 0) {
				var result = mat4.mul(mat4.create(), j.TransformMat, j.ParentToJoint);
				result = mat4.mul(mat4.create(), skeleton[j.Parent].OurJointToIdleMat, result);
				//var result = mat4.mul(mat4.create(), j.OurJointToIdleMat, mdl.matrices[j.Parent]);
				//result = mat4.mul(mat4.create(), result, mdl.matrices[j.Parent]);
				
				mdl.setNodeMatrix(iJoint, result);
			}
		}
		
		gr_instance.requireRedraw = true;
    } else {
        console.error("incorrect animation typeid");
    }
	
	this.step+=0.2;
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