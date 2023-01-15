'use strict';

let ga_instance;

class AnimationManager {
    constructor() {
        this.lastUpdateTime = window.performance.now() / 1000.0;
        this.active = true;
        this.matLayerAnimations = [];
        this.matSheetAnimations = [];
        this.objSkeletAnimations = [];
        this.speed = 1.0;
    }

    update() {
        let currentTime = window.performance.now() / 1000.0;
        let dt = (currentTime - this.lastUpdateTime) * this.speed;

        if (this.active) {
            for (let i in this.matLayerAnimations) {
                this.matLayerAnimations[i].update(dt);
            }
            for (let i in this.matSheetAnimations) {
                this.matSheetAnimations[i].update(dt);
            }
            for (let i in this.objSkeletAnimations) {
                this.objSkeletAnimations[i].update(dt);
            }
        }

        this.lastUpdateTime = currentTime;
    }

    addAnimation(anim) {
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

    freeAnimation(anim) {
        switch (anim.type) {
            case 0: {
                let id = this.objSkeletAnimations.indexOf(anim);
                if (id >= 0) {
                    let a = this.objSkeletAnimations.splice(id, 1)[0];
                    a.reset();
                    a.recalcMatrices();
                    gr_instance.requireRedraw = true;
                }
            }
            break;
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
};

class AnimationBase extends Claimable {
    _free() {
        ga_instance.freeAnimation(this);
        super._free();
    }
}

class AnimationMatertialLayer extends AnimationBase {
    constructor(anim, act, stateIndex, material) {
        super();
        this.type = 8;
        this.enabled = false;
        this.anim = anim;
        this.act = act;
        this.stateIndex = stateIndex;
        this.material = material;
        this.time = 0.0;
        material.anims.push(this);
    }

    update(dt) {
        let newTime = this.time + dt;
        if (!this.enabled) {
            return;
        }

        let stateDesc = this.act.StateDescrs[this.stateIndex];
        let dataType = this.anim.DataTypes[this.stateIndex];

        if (dataType.TypeId == 8) {
            let layer = this.material.layers.get(dataType.Param1 & 0x7f);

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

    enable(enable = true) {
        this.enabled = !!enable;
    }
}

class AnimationObjectSkelet extends AnimationBase {
    constructor(anim, act, stateIndex, obj, treeNode) {
        super();
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

        treeNode.assignAnimation(this);
        this.treeNode = treeNode;

        this.reset();
        this.recalcMatrices();
    }

    recalcMatrices() {
        const obj = this.object;
        const treeNode = this.treeNode;

        for (let i = 0; i < obj.Joints.length; i++) {
            let joint = obj.Joints[i];

            const convv = 1.0 / (1 << 14);

            let localQ;
            if ((joint.Flags & 0x8000) != 0) {
                localQ = vec4.scale(vec4.create(), this.jointLocalRots[i], convv);
                localQ = quat.normalize(localQ, localQ);
            } else {
                let localE = vec4.scale(vec4.create(), this.jointLocalRots[i], convv * 360.0);
                localQ = quat.fromEuler(quat.create(), localE[0], localE[1], localE[2]);
            }

            let local = this.jointLocalMatrices[i];
            local = mat4.fromRotationTranslationScale(local, localQ, this.jointLocalPos[i], this.jointLocalScale[i]);

            //console.log(i, "o", quat.str(vec4.scale(vec4.create(), this.jointLocalRots[i], convv)));
            //console.log(i, "n", quat.str(quat.normalize(quat.create(), vec4.scale(vec4.create(), this.jointLocalRots[i], convv))));
            //console.log(i, "r", quat.str(mat4.getRotation(quat.create(), local)));
            //console.log(i, "strage shit", !!(joint.Flags & 0x8000), joint.Flags, this.jointLocalPos[i],  this.jointLocalScale[i]);

            this.jointLocalMatrices[i] = local;

            //let global = (joint.Parent >= 0) ? mat4.mul(this.jointGlobalMatrices[i], this.jointGlobalMatrices[joint.Parent], local) : local;
            //this.jointGlobalMatrices[i] = global;

            //let result = global;
            //if ((joint.Flags & 0x8) != 0) {
            // mat4.mul(result, obj.Matrixes2[joint.ExternalId], result);
            // console.warn("joint flags 0x8: ", joint.Name, joint);
            // console.log(joint.Name, obj.Matrixes2[joint.ExternalId]);
            // result = mat4.mul(result, result, obj.Matrixes2[joint.ExternalId]);
            //}

            //let resultInverted = result;
            //if (joint.IsSkinned) {
            //    let inverseBindMat = obj.Matrixes3[joint.InvId];
            //    resultInverted = mat4.mul(mat4.create(), result, inverseBindMat);
            //}
            treeNode.joints[i].setLocalMatrixWithoutUpdate(local);
        }

        treeNode.update();
    }

    reset() {
        this.time = 0;
        let obj = this.object;

        for (let i = 0; i < obj.Joints.length; i++) {
            this.jointLocalRots[i] = obj.Vectors5[i].slice();
            this.jointLocalPos[i] = obj.Vectors4[i].slice();
            this.jointLocalScale[i] = obj.Vectors6[i].slice();
            this.jointLocalMatrices[i] = mat4.create();
            this.jointGlobalMatrices[i] = mat4.create();
        }
        gr_instance.requireRedraw = true;
    }

    // return first sample index
    returnStreamDataIndex(manager, globalManager, animNextTime, frameTime) {
        const eps = 1.0 / (1024.0 * 16.0);
        // TODO: parse reverse animation situation if (animNextTime < animPrevTime)

        let globalFramesCount = globalManager.Count + globalManager.Offset + globalManager.DatasCount3 - 1;
        let animFrame = animNextTime / frameTime;
        if ((animFrame + eps) > globalFramesCount ||
            (animFrame - eps) < manager.Offset) {
            return undefined;
        }

        let streamSampleIdx = animFrame - manager.Offset;
        if (streamSampleIdx >= manager.Count + manager.DatasCount3) {
            return undefined;
        } else if (streamSampleIdx > manager.Count - 1) {
            return manager.Count - 1;
        } else if (streamSampleIdx < 0) {
            return 0;
        } else {
            return streamSampleIdx;
        }
    }

    handleSkinningStream(stream, globalManager, prevTime, newTime, frameTime, jointLocalRots, l = false) {
        const eps = 1.0 / (1024.0 * 16.0);
        // TODO: parse reverse animation situation if (animNextTime < animPrevTime)

        let newSamplePos = this.returnStreamDataIndex(stream.Manager, globalManager, newTime, frameTime);
        if (newSamplePos === undefined) {
            return false;
        }
        let newSampleIndex = Math.floor(newSamplePos);
        let newSampleOffset = newSamplePos - newSampleIndex;

        let changed = false;

        //console.info(newSamplePos, stream.Samples.hasOwnProperty(-100));
        if (stream.Samples.hasOwnProperty(-100)) {
            // additive change
            let newValueMultiplyer = newSampleOffset; // by default add as much as we shoul

            let prevValueMultiplyer, prevSampleIndex;
            let prevSampleOffset;
            let prevSamplePos = this.returnStreamDataIndex(stream.Manager, globalManager, prevTime, frameTime);
            if (prevSamplePos !== undefined) {
                // if not first frame in batch
                prevSampleIndex = Math.floor(prevSamplePos);
                prevSampleOffset = prevSamplePos - prevSampleIndex;
                if (prevSampleIndex == newSampleIndex) {
                    // we in the same sample, compensate prev time this sample was played
                    newValueMultiplyer -= prevSampleOffset;
                } else {
                    // we need to done prev sample addition first
                    prevValueMultiplyer = 1 - prevSampleOffset;
                    if (prevValueMultiplyer < eps) {
                        prevValueMultiplyer = undefined;
                    }
                }
            }

            for (let iStream in stream.Samples) {
                if (iStream < 0) {
                    continue;
                }
                changed = true;
                let jointId = parseInt(iStream / 4);
                let coord = parseInt(iStream) % 4;

                let value = jointLocalRots[jointId][coord];
                let prevSampleValue;

                if (prevValueMultiplyer !== undefined) {
                    prevSampleValue = stream.Samples[iStream][prevSampleIndex];
                    value += prevSampleValue * prevValueMultiplyer;
                }

                // TODO: add cycle that will add missed samples values (fast forward case)
                let newSampleValue = stream.Samples[iStream][newSampleIndex];
                value += newSampleValue * newValueMultiplyer;

                let prevVal = jointLocalRots[jointId][coord];
                jointLocalRots[jointId][coord] = value;

                //if (!l) jointLocalRots[jointId][coord] = value;
                //if (l) console.log("add", prevVal, coord, value);
            }
        } else {
            // exact change
            let nextSampleIndex = newSampleIndex + 1;
            for (let iStream in stream.Samples) {
                changed = true;
                let jointId = parseInt(iStream / 4);
                let coord = parseInt(iStream) % 4;

                let newSampleValue = stream.Samples[iStream][newSampleIndex];
                let value;
                if (newSampleOffset < eps) {
                    value = newSampleValue;
                } else {
                    let nextSampleValue = stream.Samples[iStream][nextSampleIndex];
                    value = newSampleValue + (nextSampleValue - newSampleValue) * newSampleOffset;
                }
                let prevVal = jointLocalRots[jointId][coord];
                jointLocalRots[jointId][coord] = value;
                //if (l) { console.log("raw", prevVal, coord); }
                //console.log(jointId, coord,
                //		"pv", prevVal, "nv", value, "rnv", jointLocalRots[jointId][coord]);	
            }
        }
        return changed;
    }

    update(dt) {
        let newTime = this.time + dt;
        if (!this.enabled) {
            return;
        }

        let stateDesc = this.act.StateDescrs[this.stateIndex];
        let dataType = this.anim.DataTypes[this.stateIndex];

        if (dataType.TypeId == 0) {
            let skeleton = this.object.Joints;
            let changed = false;

            for (let iData in stateDesc.Data) {
                let data = stateDesc.Data[iData];
                let globalRotationStream = data.RotationStream;
                let globalPositionStream = data.PositionStream;
                let globalScaleStream = data.ScaleStream;

                if (globalRotationStream.Manager.Count) {
                    let stream = globalRotationStream;
                    if (this.handleSkinningStream(stream, stream.Manager, this.time, newTime, stateDesc.FrameTime, this.jointLocalRots)) {
                        changed = true;
                    }
                } else {
                    for (let iAdditiveSample in data.RotationSubStreamsAdd) {
                        let stream = data.RotationSubStreamsAdd[iAdditiveSample];
                        if (this.handleSkinningStream(stream, globalRotationStream.Manager, this.time, newTime, stateDesc.FrameTime, this.jointLocalRots)) {
                            changed = true;
                        }
                    }

                    for (let iRoughSample in data.RotationSubStreamsRough) {
                        let stream = data.RotationSubStreamsRough[iRoughSample];
                        if (this.handleSkinningStream(stream, globalRotationStream.Manager, this.time, newTime, stateDesc.FrameTime, this.jointLocalRots)) {
                            changed = true;
                        }
                    }
                }

                if (globalPositionStream.Manager.Count) {
                    let stream = globalPositionStream;
                    if (this.handleSkinningStream(stream, stream.Manager, this.time, newTime, stateDesc.FrameTime, this.jointLocalPos, true)) {
                        changed = true;
                    }
                } else {
                    for (let iAdditiveSample in data.PositionSubStreamsAdd) {
                        let stream = data.PositionSubStreamsAdd[iAdditiveSample];
                        if (this.handleSkinningStream(stream, globalPositionStream.Manager, this.time, newTime, stateDesc.FrameTime, this.jointLocalPos, true)) {
                            changed = true;
                        }
                    }

                    for (let iRoughSample in data.PositionSubStreamsRough) {
                        let stream = data.PositionSubStreamsRough[iRoughSample];
                        if (this.handleSkinningStream(stream, globalPositionStream.Manager, this.time, newTime, stateDesc.FrameTime, this.jointLocalPos, true)) {
                            changed = true;
                        }
                    }
                }

                if (globalScaleStream.Manager.Count) {
                    let stream = globalScaleStream;
                    if (this.handleSkinningStream(stream, stream.Manager, this.time, newTime, stateDesc.FrameTime, this.jointLocalScale, true)) {
                        changed = true;
                    }
                } else {
                    
                    for (let iAdditiveSample in data.ScaleSubStreamsAdd) {
                        let stream = data.ScaleSubStreamsAdd[iAdditiveSample];
                        if (this.handleSkinningStream(stream, globalScaleStream.Manager, this.time, newTime, stateDesc.FrameTime, this.jointLocalPos, true)) {
                            changed = true;
                        }
                    }
                    

                    for (let iRoughSample in data.ScaleSubStreamsRough) {
                        let stream = data.ScaleSubStreamsRough[iRoughSample];
                        if (this.handleSkinningStream(stream, globalScaleStream.Manager, this.time, newTime, stateDesc.FrameTime, this.jointLocalScale, true)) {
                            changed = true;
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
}

class AnimationMaterialSheet extends AnimationBase {
    constructor(anim, act, stateIndex, material) {
        super();
        this.type = 9;
        this.anim = anim;
        this.act = act;
        this.stateIndex = stateIndex;
        this.material = material;
        this.time = 0.0;
        this.step = 0;
        material.anims.push(this);
    }

    update(dt) {
        var newTime = this.time + dt * 0.5;
        var stateDesc = this.act.StateDescrs[this.stateIndex];
        var dataType = this.anim.DataTypes[this.stateIndex];

        if (dataType.TypeId == 9) {
            var floatStep = newTime / stateDesc.FrameTime;
            var step = Math.trunc(floatStep) % stateDesc.Data.length;
            if (step != this.step) {
                gr_instance.requireRedraw = true;
                this.material.layers.get(0).setTextureIndex(stateDesc.Data[step]);
            }
        } else {
            log.error("incorrect animation typeid");
        }

        this.time = newTime;
    }
}

function gaInit() {
    if (ga_instance) {
        console.warn("AnimationManager already created", ga_instance);
        return;
    }
    ga_instance = new AnimationManager();
}