'use strict';

let ga_instance;

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

            let floatStep = newTime / stateDesc.FrameTime;

            let step = Math.trunc(floatStep) % (data.Stream.Manager.Count - 1);
            let nextStep = (step + 1) % data.Stream.Manager.Count;

            let blendFactor2 = floatStep - Math.floor(floatStep);
            let blendFactor1 = (1 - blendFactor2);

            for (let key in data.Stream.Samples) {
                let stream = data.Stream.Samples[key];
				layer.uvoffset[key] = stream[step] * blendFactor1 + stream[nextStep] * blendFactor2;
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
    this.time = 0;
	this.enabled = true;

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
		
		//const convv = (360.0 / (1<<14));
		const convv = (1.0 / 16384.0);

		let localQ;	
		if ((joint.Flags & 0x8000) != 0) {
			localQ = vec4.scale(vec4.create(), this.jointLocalRots[i], convv);
			localQ = quat.normalize(localQ, localQ);
		} else {
			let localE = vec4.scale(vec4.create(), this.jointLocalRots[i], convv*360.0);
			localQ = quat.fromEuler(quat.create(), localE[0], localE[1], localE[2]);
		}
		
		let local = this.jointLocalMatrices[i];
		local = mat4.fromRotationTranslationScale(local, localQ, this.jointLocalPos[i], this.jointLocalScale[i]);

		//console.log(i, "o", quat.str(vec4.scale(vec4.create(), this.jointLocalRots[i], convv)));
		//console.log(i, "n", quat.str(quat.normalize(quat.create(), vec4.scale(vec4.create(), this.jointLocalRots[i], convv))));
		//console.log(i, "r", quat.str(mat4.getRotation(quat.create(), local)));
		//console.log(i, "strage shit", !!(joint.Flags & 0x8000), joint.Flags, this.jointLocalPos[i],  this.jointLocalScale[i]);

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
	this.time = 0;
	let obj = this.object;
	
	for (let i = 0; i < obj.Joints.length; i++) {
		this.jointLocalRots[i] = obj.Vectors5[i].slice();
		this.jointLocalPos[i] = obj.Vectors4[i].slice();
		this.jointLocalScale[i] = obj.Vectors6[i].slice();
		this.jointLocalMatrices[i] = mat4.create();
		this.jointGlobalMatrices[i] = mat4.create();
	}
}


// return first sample index
function animationReturnStreamData(manager, globalManager, animNextTime, frameTime) {
	const eps = 1.0/(1024.0 * 16.0);
	
	// TODO: parse reverse animation situation if (animNextTime < animPrevTime)
	
	let globalFramesCount = globalManager.Count + globalManager.Offset + globalManager.DatasCount3 - 1;
	let animFrameTime = animNextTime / frameTime;
	if ((animFrameTime + eps) > globalFramesCount ||
	    (animFrameTime - eps) < manager.Offset){
		return undefined;
	}
	
	let streamSampleIdx = animFrameTime - manager.Offset;
	if (animFrameTime > manager.Count-1) {
		animFrameTime = manager.Count-1;
	} else if (animFrameTime < 0) {
		animFrameTime = 0;
	}
	return animFrameTime;
}

gaObjSkeletAnimation.prototype.update = function(dt, currentTime) {
	let newTime = this.time + dt;
    if (!this.enabled) {
        return;
    }

    let stateDesc = this.act.StateDescrs[this.stateIndex];
    let dataType = this.anim.DataTypes[this.stateIndex];
	
	if (dataType.TypeId == 0) {
		let mdl = this.mdl;
		let skeleton = this.object.Joints;
		let changed = false;
		
		for (let iData in stateDesc.Data) {
			let data = stateDesc.Data[iData];
			let globalStream = data.Stream;

			if (globalStream.Manager.Count) {
				let stream = globalStream;
				let samplePos = animationReturnStreamData(stream.Manager, stream.Manager, newTime, stateDesc.FrameTime);
				if (samplePos == undefined) { continue; }

				let sampleNextIndex = Math.ceil(samplePos);
				let samplePrevIndex = Math.floor(samplePos);
				let sampleBlendCoof = samplePos - samplePrevIndex;

				for (let iStream in stream.Samples) {
					changed = true;

					let jointId = parseInt(iStream / 4);
					let coord = parseInt(iStream) % 4;
					
					let newValue;
					let prevVal = data.Stream.Samples[iStream][samplePrevIndex];
					let nextVal = data.Stream.Samples[iStream][sampleNextIndex];
	
					if (sampleBlendCoof < 0.0001 || (sampleNextIndex == sampleNextIndex)) {
						newValue = prevVal;
					} else if (sampleBlendCoof > 0.9999) {
						newValue = nextVal;
					} else {
						newValue = (nextVal - prevVal) * sampleBlendCoof + prevVal;
					}

					this.jointLocalRots[jointId][coord] =  newValue;
					//console.log("u", this.jointLocalRots[jointId], coord, jointId, coord, newValue, newValue / (1<<14));
				}
			} else { // if (globalStream.Manager.Count) else
				for (let iAdditiveSample in data.SubStreamsAdd) {
					let stream = data.SubStreamsAdd[iAdditiveSample];
					
					let samplePos = animationReturnStreamData(stream.Manager, globalStream.Manager, newTime, stateDesc.FrameTime);
					if (samplePos == undefined) { continue; }
	
					let sampleNextIndex = Math.ceil(samplePos);
					let samplePrevIndex = Math.floor(samplePos);
					let sampleBlendCoof = samplePos - samplePrevIndex;
	
					let multiplyerArray = iStream[-100];
					for (let iStream in stream.Samples) {
						if (iStream < 0) { continue; }
						changed = true;
	
						let jointId = parseInt(iStream / 4);
						let coord = parseInt(iStream) % 4;
						
						let newValue;
						let prevVal = data.Stream.Samples[iStream][samplePrevIndex];
						let nextVal = data.Stream.Samples[iStream][sampleNextIndex];
		
						if (sampleBlendCoof < 0.001) {
							this.jointLocalRots[jointId][coord] =  this.jointLocalRots[jointId][coord] + multiplyerArray
						} 
	
						this.jointLocalRots[jointId][coord] =  newValue;
						//console.log("u", this.jointLocalRots[jointId], coord, jointId, coord, newValue, newValue / (1<<14));
					}
				}
			}
		}
		if (changed) {
			this.recalcMatrices();
			gr_instance.requireRedraw = true;
		}
    } else {
        console.error("incorrect animation typeid");
    }
	
    this.time = newTime;
	if (this.time >= this.act.Duration) {
		this.reset();
	}
}

function gaMatertialSheetAnimation(anim, act, stateIndex, material) {
    this.type = 9;
    this.anim = anim;
    this.act = act;
    this.stateIndex = stateIndex;
    this.material = material;
    this.time = 0.0;
    this.step = 0;
    material.anims.push(this);
}

gaMatertialSheetAnimation.prototype.update = function(dt, currentTime) {
    var newTime = this.time + dt * 0.5;
    var stateDesc = this.act.StateDescrs[this.stateIndex];
    var dataType = this.anim.DataTypes[this.stateIndex];

    if (dataType.TypeId == 9) {
        var floatStep = newTime / stateDesc.FrameTime;
        var step = Math.trunc(floatStep) % stateDesc.Data.length;
        if (step != this.step) {
            gr_instance.requireRedraw = true;
            this.material.layers[0].setTextureIndex(stateDesc.Data[step]);
        }
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
