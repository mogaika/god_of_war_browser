'use strict';

let ga_instance;
const fp14_conv = (1<<14);

function gaAnimationManager() {
    this.lastUpdateTime = window.performance.now() / 1000.0;
    this.matLayerAnimations = [];
    this.matSheetAnimations = [];
	this.objSkeletAnimations = [];
};

gaAnimationManager.prototype.update = function() {
    let currentTime = window.performance.now() / 1000.0;
    let dt = currentTime - this.lastUpdateTime;

    for (let i in this.matLayerAnimations) {
        this.matLayerAnimations[i].update(dt, currentTime);
    }
    for (let i in this.matSheetAnimations) {
        this.matSheetAnimations[i].update(dt, currentTime);
    }
    for (let i in this.objSkeletAnimations) {
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
        case 0: {
	            let id = this.objSkeletAnimations.indexOf(anim);
	            if (id >= 0) {
	                let a = this.objSkeletAnimations.splice(id, 1)[0];
					a.reset();
					a.recalcMatrices();
	            }
			} break;
        case 8: {
	            let id = this.matLayerAnimations.indexOf(anim);
	            if (id >= 0) {
	                this.matLayerAnimations.splice(id, 1);
	            }
			}
            break;
        case 9: {
	            let id = this.matSheetAnimations.indexOf(anim);
	            if (id >= 0) {
	                this.matSheetAnimations.splice(id, 1);
	            }
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
    let newTime = this.time + dt;
    if (!this.enabled) {
        return;
    }

    let stateDesc = this.act.StateDescrs[this.stateIndex];
    let dataType = this.anim.DataTypes[this.stateIndex];

    if (dataType.TypeId == 8) {
        let layer = this.material.layers[dataType.Param1 & 0x7f];

        for (let iData in stateDesc.Data) {
            let data = stateDesc.Data[iData];

            let floatStep = newTime / stateDesc.ImportantFloat;

            let step = Math.trunc(floatStep) % (data.Dm.DatasCount1 - 1);
            let nextStep = (step + 1) % data.Dm.DatasCount1;

            let blendFactor2 = floatStep - Math.floor(floatStep);
            let blendFactor1 = (1 - blendFactor2);

            for (let key in data.Stream) {
                let stream = data.Stream[key];
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
	this.recalcMatrices();
}

gaObjSkeletAnimation.prototype.recalcMatrices = function() {
	let obj = this.object;
	let mdl = this.mdl;
	
	for (let i = 0; i < obj.Joints.length; i++) {
		let joint = obj.Joints[i];

		let localRots = this.jointLocalRots[i];
		let localQ = quat.fromEuler(quat.create(), localRots[0], localRots[1], localRots[2]);
		//localQ = quat.fromEuler(quat.create(), 0,0,0);

		let local = this.jointLocalMatrices[i];

		//mat4.identity(local);
		//let localRotMatrix = mat4.fromQuat(mat4.create(), localQ);
		//local = mat4.translate(local, local, this.jointLocalPos[i]);
		//local = mat4.mul(local, local, localRotMatrix);
		local = mat4.fromRotationTranslationScale(local, localQ, this.jointLocalPos[i], this.jointLocalScale[i])
		this.jointLocalMatrices[i] = local;

		let global = (joint.Parent >= 0) ? mat4.mul(this.jointGlobalMatrices[i], this.jointGlobalMatrices[joint.Parent], local) : local;
		this.jointGlobalMatrices[i] = global;

		let result = global;
		if (joint.IsSkinned) {
			let inverseBindMat = obj.Matrixes3[joint.InvId];
			result = mat4.mul(mat4.create(), global, inverseBindMat);
		}

		mdl.setJointMatrix(i, result);
	}
}

gaObjSkeletAnimation.prototype.reset = function() {
	this.step = 0;
	let obj = this.object;
	
	for (let i = 0; i < obj.Joints.length; i++) {
		this.jointLocalRots[i] =  vec4.scale(vec4.create(), obj.Vectors5[i], 360.0/fp14_conv);
		//console.log(i, obj.Vectors5[i], this.jointLocalRots[i]);
		this.jointLocalPos[i] = obj.Vectors4[i];
		this.jointLocalScale[i] = obj.Vectors6[i];
		this.jointLocalMatrices[i] = mat4.create();
		this.jointGlobalMatrices[i] = mat4.create();
	}
}

gaObjSkeletAnimation.prototype.update = function(dt, currentTime) {
    let newTime = this.time + dt * 0.5;
    let stateDesc = this.act.StateDescrs[this.stateIndex];
    let dataType = this.anim.DataTypes[this.stateIndex];

    if (dataType.TypeId == 0) {
		let mdl = this.mdl;
		let skeleton = this.object.Joints;
		let changed = false;
		
		for (let iData in stateDesc.Data) {
			let data = stateDesc.Data[iData];

			for (let iStream in data.Stream) {
				changed = true;
				let jointId = parseInt(iStream / 4);
				let coord = parseInt(iStream) % 4;
				
				let frame = parseInt(Math.floor(this.step));
				let nextFrame = (frame + 1) % data.Stream[iStream].length;
				let curFrame = nextFrame - 1;
				let framePos = this.step - Math.floor(this.step);
				//let framePos = this.step % 1;
				//console.log("step", this.step, "nf", nextFrame, "fp", framePos, "curFram", curFrame, "le", data.Stream[iStream].length);
				
				let newValue;
				let curVal = data.Stream[iStream][curFrame];
				let nextVal = data.Stream[iStream][nextFrame];
				if (framePos < 0.0001) {
					newValue = curVal;
				} else if (framePos > 0.9999) {
					newValue = nextVal;
				} else {
					newValue = (nextVal - curVal) * framePos + curVal;
				}
				
				//let newValue = data.Stream[iStream][parseInt(this.step) % data.Stream[iStream].length] * globalMul;
				//console.log( data.Stream[iStream][parseInt(this.step) % data.Stream[iStream].length], "*", globalMul, "=", newValue);
				this.jointLocalRots[jointId][coord] =  newValue * 360 / fp14_conv;
			}
		}
		if (changed) {
			this.recalcMatrices();
		}
		gr_instance.requireRedraw = true;
    } else {
        console.error("incorrect animation typeid");
    }
	
	this.step += 0.25;
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
    let newTime = this.time + dt;
    if (!this.enabled) {
        return;
    }

    let stateDesc = this.act.StateDescrs[this.stateIndex];
    let dataType = this.anim.DataTypes[this.stateIndex];

    if (dataType.TypeId == 8) {
        let layer = this.material.layers[dataType.Param1 & 0x7f];

        for (let iData in stateDesc.Data) {
            let data = stateDesc.Data[iData];

            let floatStep = newTime / stateDesc.ImportantFloat;

            let step = Math.trunc(floatStep) % (data.Dm.DatasCount1 - 1);
            let nextStep = (step + 1) % data.Dm.DatasCount1;

            let blendFactor2 = floatStep - Math.floor(floatStep);
            let blendFactor1 = (1 - blendFactor2);

            for (let key in data.Stream) {
                let stream = data.Stream[key];
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
