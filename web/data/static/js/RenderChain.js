var gowBlend = 0.5;

function grRenderChain_SkinnedTexturedFlash(node, mesh) {
    this.node = node;
    this.mesh = mesh;
    this.meshInstanceId = 0;
    this.material = undefined;
    this.diffuselayer = undefined;
    this.envmaplayer = undefined;
}

function grRenderChain_SkinnedTexturedFlashesBatch() {
    this.normalFlashes = [];
    this.additiveFlashes = [];
}

grRenderChain_SkinnedTexturedFlashesBatch.prototype.clear = function() {
    this.normalFlashes.length = 0;
    this.additiveFlashes.length = 0;
}

grRenderChain_SkinnedTexturedFlashesBatch.prototype.addFlash = function(flash, method) {
    switch (method) {
        case 0:
            this.normalFlashes.push(flash);
            break;
        case 1:
            this.additiveFlashes.push(flash);
            break;
            // default: console.log("missed flash target for method ", method, flash); break;
    }
}

function grRenderChain_SkinnedTextured(ctrl) {
    this.skyBatch = new grRenderChain_SkinnedTexturedFlashesBatch();
    this.normalBatch = new grRenderChain_SkinnedTexturedFlashesBatch();
    this.textBatch = [];
    this.needRebuildScene = true;
    this.isSkyRendering = false;

    this.vertexShader = ctrl.downloadShader("/static/shaders/SkinnedTextured.vs", false);
    this.fragmentShader = ctrl.downloadShader("/static/shaders/SkinnedTextured.fs", true);
    this.program = ctrl.createProgram(this.vertexShader, this.fragmentShader);

    gl.useProgram(this.program);

    this.aVertexPos = gl.getAttribLocation(this.program, "aVertexPos");
    this.aVertexColor = gl.getAttribLocation(this.program, "aVertexColor");
    this.aVertexUV = gl.getAttribLocation(this.program, "aVertexUV");
    this.aVertexJointID = gl.getAttribLocation(this.program, "aVertexJointID");
    this.aVertexWeight = gl.getAttribLocation(this.program, "aVertexWeight");

    this.umProjection = gl.getUniformLocation(this.program, "umProjection");
    this.umView = gl.getUniformLocation(this.program, "umView");
    this.umModelTransform = gl.getUniformLocation(this.program, "umModelTransform");
    this.uMaterialColor = gl.getUniformLocation(this.program, "uMaterialColor");
    this.uLayerColor = gl.getUniformLocation(this.program, "uLayerColor");
    this.uLayerOffset = gl.getUniformLocation(this.program, "uLayerOffset");
    this.uLayerDiffuseSampler = gl.getUniformLocation(this.program, "uLayerDiffuseSampler");
    this.uLayerEnvmapSampler = gl.getUniformLocation(this.program, "uLayerEnvmapSampler");
    this.uUseLayerDiffuseSampler = gl.getUniformLocation(this.program, "uUseLayerDiffuseSampler");
    this.uUseEnvmapSampler = gl.getUniformLocation(this.program, "uUseEnvmapSampler");
    this.uUseVertexColor = gl.getUniformLocation(this.program, "uUseVertexColor");

    this.maxJoints = 180;
    this.umJoints = [];
    for (let i = 0; i < this.maxJoints; i += 1) {
        this.umJoints.push(gl.getUniformLocation(this.program, "umJoints[" + i + "]"));
    }
    this.uJointsWidth = gl.getUniformLocation(this.program, "uJointsWidth");

    gl.enableVertexAttribArray(this.aVertexPos);
    gl.enableVertexAttribArray(this.aVertexColor);

    gl.vertexAttrib4f(this.aVertexJointID, 0, 0, 0, 0);
    gl.vertexAttrib4f(this.aVertexWeight, 1, 0, 0, 0);

    gl.uniform1i(this.uLayerDiffuseSampler, 0);
    gl.uniform1i(this.uLayerEnvmapSampler, 1);
    gl.uniform1i(this.uUseLayerDiffuseSampler, 0);
    gl.uniform1i(this.uUseEnvmapSampler, 0);
    gl.uniform1i(this.uJointsWidth, 0);

    gl.clearColor(0.25, 0.25, 0.25, 1.0);
    gl.clearDepth(1.0);
    gl.depthFunc(gl.LEQUAL);
    gl.disable(gl.BLEND);
    gl.depthMask(true);
    gl.enable(gl.DEPTH_TEST);
}

grRenderChain_SkinnedTextured.prototype.free = function(ctrl) {
    // TODO :add missed fields
    gl.disableVertexAttribArray(this.aVertexPos);
    gl.disableVertexAttribArray(this.aVertexColor);
    gl.disableVertexAttribArray(this.aVertexUV);
    gl.disableVertexAttribArray(this.aVertexJointID);
    gl.disableVertexAttribArray(this.aVertexWeight);
    gl.deleteProgram(this.program);
    gl.deleteShader(this.vertexShader);
    gl.deleteShader(this.fragmentShader);
}

grRenderChain_SkinnedTextured.prototype.drawMesh = function(mesh, hasTexture = false, hasJoints = false, hasEnvMap = false) {
    gl.enableVertexAttribArray(this.aVertexPos);
    gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferVertex);
    gl.vertexAttribPointer(this.aVertexPos, 3, gl.FLOAT, false, 0, 0);

    if (mesh.bufferBlendColor) {
        gl.uniform1i(this.uUseVertexColor, 1);
        gl.enableVertexAttribArray(this.aVertexColor);
        gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferBlendColor);
        gl.vertexAttribPointer(this.aVertexColor, 4, gl.UNSIGNED_BYTE, true, 0, 0);
    } else {
        gl.uniform1i(this.uUseVertexColor, 0);
        gl.disableVertexAttribArray(this.aVertexColor);
    }

    if (hasTexture) {
        gl.uniform1i(this.uUseLayerDiffuseSampler, 1);
    } else {
        gl.uniform1i(this.uUseLayerDiffuseSampler, 0);
    }

    if (mesh.bufferUV) {
        gl.enableVertexAttribArray(this.aVertexUV);
        gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferUV);
        gl.vertexAttribPointer(this.aVertexUV, 2, gl.FLOAT, false, 0, 0);
    } else {
        gl.disableVertexAttribArray(this.aVertexUV);
    }

    if (hasEnvMap) {
        gl.uniform1i(this.uUseEnvmapSampler, 1);
    } else {
        gl.uniform1i(this.uUseEnvmapSampler, 0);
    }

    if (mesh.jointWidth && hasJoints && mesh.bufferJointIds && mesh.bufferWeights) {
        gl.uniform1i(this.uJointsWidth, mesh.jointWidth);

        gl.enableVertexAttribArray(this.aVertexJointID);
        gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferJointIds);
        gl.vertexAttribPointer(this.aVertexJointID, mesh.jointWidth, gl.UNSIGNED_BYTE, false, 0, 0);

        gl.enableVertexAttribArray(this.aVertexWeight);
        gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferWeights);
        gl.vertexAttribPointer(this.aVertexWeight, mesh.jointWidth, gl.FLOAT, false, 0, 0);
    } else {
        gl.uniform1i(this.uJointsWidth, 0);
        gl.disableVertexAttribArray(this.aVertexJointID);
        gl.disableVertexAttribArray(this.aVertexWeight);
    }

    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.bufferIndex);
    gl.drawElements(mesh.primitive, mesh.indexesCount, mesh.bufferIndexType, 0);
}

grRenderChain_SkinnedTextured.prototype.renderFlashesArray = function(ctrl, flashesBatch) {
    let model = -1;
    let material = -1;
    let mesh = -1;
    let layer = -1;
    let envmaplayer = -1;
    let mask = ~ctrl.filterMask;

    for (const flash of flashesBatch) {
        let hasEnvMap = false;
        let hasJoints = false;
        let hasTexture = false;

        const node = flash.node;

        if (model !== node.model) {
            model = node.model;
        }

        if (model.mask & mask) {
            continue;
        }

        if (mesh != flash.mesh) {
            mesh = flash.mesh;

            if (mesh.mask & mask) {
                continue;
            }

            if (mesh.isDepthTested) {
                gl.enable(gl.DEPTH_TEST);
            } else {
                gl.disable(gl.DEPTH_TEST);
            }

            const parent = node.parent;
            if (parent && parent.constructor == ObjectTreeNodeSkinned) {
                jointsSkinningNode = parent;

                const joints = parent.joints;

                for (const jointId in joints) {
                    if (jointId > this.maxJoints) {
                        console.warn("too much joints for mesh", joints.length);
                        break;
                    }
                    const joint = joints[jointId];
                    let matrix = mesh.useBindToJoin ? joint.renderMatrix : joint.globalMatrix;
                    if (this.isSkyRendering) {
                        // console.log("sky rendering", matrix);
                        matrix = mat4.fromQuat(matrix, mat4.getRotation(quat.create(), matrix));
                        matrix = mat4.create();
                        // console.log(jointId, i, matrix);
                    }

                    gl.uniformMatrix4fv(this.umJoints[jointId], false, matrix);
                }
                hasJoints = true;
                if (mesh.ps3static) {
                    //if (mesh.jointMapping.length > 1) {
                    // console.error("type ps3static problem with joint mapping", mesh);
                    //}
                    // gl.uniform1i(this.uUseRootJointScaleOnly, 1);
                    //gl.uniformMatrix4fv(this.umModelTransform, false, node.parent.globalMatrix);

                    hasJoints = false;
                    
                    if (joints) {
                        const thisJointRoot = joints[0];
                        const scaleVec = mat4.getScaling(vec3.create(), thisJointRoot.globalMatrix);
                        const scaleMat = mat4.fromScaling(mat4.create(), scaleVec);
                        gl.uniformMatrix4fv(this.umModelTransform, false, scaleMat);

                        // console.log(node, scaleVec);
                    }
                }
            } else {
                gl.uniformMatrix4fv(this.umModelTransform, false, node.globalMatrix);
            }
        }

        if (material != flash.material) {
            material = flash.material;
            if (material) {
                gl.uniform4f(this.uMaterialColor, material.color[0], material.color[1], material.color[2], material.color[3]);
            } else {
                gl.uniform4f(this.uMaterialColor, 1.0, 1.0, 1.0, 1.0);
            }
        }

        if (layer != flash.diffuselayer) {
            layer = flash.diffuselayer;
            if (layer) {
                gl.uniform4f(this.uLayerColor, layer.color[0], layer.color[1], layer.color[2], layer.color[3]);
                gl.uniform2f(this.uLayerOffset, layer.uvoffset[0], layer.uvoffset[1]);
            } else {
                gl.uniform4f(this.uLayerColor, 1.0, 1.0, 1.0, 1.0);
                gl.uniform2f(this.uLayerOffset, 0.0, 0.0);
            }
        }
        if (layer && layer.textures.length) {
            gl.activeTexture(gl.TEXTURE0);
            gl.bindTexture(gl.TEXTURE_2D, layer.textures.get(layer.textureIndex).glTexture);
            hasTexture = true;
        }

        if (envmaplayer != flash.envmaplayer) {
            envmaplayer = flash.envmaplayer;
        }
        if (envmaplayer && envmaplayer.textures.length) {
            gl.activeTexture(gl.TEXTURE1);
            gl.bindTexture(gl.TEXTURE_2D, envmaplayer.textures.get(envmaplayer.textureIndex).glTexture);
            hasEnvMap = true;
        }

        this.drawMesh(mesh, hasTexture, hasJoints, hasEnvMap);
    }
    return flashesBatch.length;
}

grRenderChain_SkinnedTextured.prototype.renderFlashesBatch = function(ctrl, flashesBatch) {
    gl.blendEquation(gl.FUNC_ADD);
    gl.depthMask(true);
    gl.enable(gl.BLEND);
    gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE);

    let cnt = this.renderFlashesArray(ctrl, flashesBatch.normalFlashes);

    gl.blendEquation(gl.FUNC_ADD);
    gl.depthMask(false);
    gl.enable(gl.BLEND);
    gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE, gl.ONE, gl.ONE);

    cnt += this.renderFlashesArray(ctrl, flashesBatch.additiveFlashes);
    gl.depthMask(true);
    return cnt;
}

grRenderChain_SkinnedTextured.prototype.renderText = function(ctrl) {
    gl.enable(gl.BLEND);
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);
    gl.depthMask(false);
    gl.disable(gl.DEPTH_TEST);

    gl.disableVertexAttribArray(this.aVertexColor);
    gl.disableVertexAttribArray(this.aVertexJointID);
    gl.disableVertexAttribArray(this.aVertexWeight);

    gl.uniformMatrix4fv(this.umView, false, mat4.create());
    gl.uniformMatrix4fv(this.umProjection, false, ctrl.orthoMatrix);
    gl.uniform1i(this.uUseLayerDiffuseSampler, 1);
    gl.uniform1i(this.uUseEnvmapSampler, 0);
    gl.uniform1i(this.uUseVertexColor, 0);
    gl.uniform1i(this.uJointsWidth, 0);

    gl.activeTexture(gl.TEXTURE0);
    gl.bindTexture(gl.TEXTURE_2D, ctrl.fontTexture.glTexture);
    gl.uniform4f(this.uLayerColor, 1.0, 1.0, 1.0, 1.0);
    gl.uniform2f(this.uLayerOffset, 0.0, 0.0);

    let mask = ~ctrl.filterMask;

    let projViewMat = ctrl.camera.getProjViewMatrix();

    for (let i = 0; i < this.textBatch.length; i++) {
        const node = this.textBatch[i];
        let text = node.model;

        if (text.mask & mask) {
            continue;
        }

        gl.enableVertexAttribArray(this.aVertexPos);
        gl.bindBuffer(gl.ARRAY_BUFFER, text.bufferVertex);
        gl.vertexAttribPointer(this.aVertexPos, 2, gl.FLOAT, false, 0, 0);

        gl.enableVertexAttribArray(this.aVertexUV);
        gl.bindBuffer(gl.ARRAY_BUFFER, text.bufferUV);
        gl.vertexAttribPointer(this.aVertexUV, 2, gl.FLOAT, false, 0, 0);

        let isVisible = true;
        let mat;

        if (text.is3d) {
            let pos3d = mat4.getTranslation(vec3.create(), node.globalMatrix);
            let pos2d = vec3.transformMat4(vec3.create(), pos3d, projViewMat);
            if (pos2d[2] < 1) {
                const pos = [
                    (pos2d[0] + 1) * 0.5 * gl.drawingBufferWidth,
                    (pos2d[1] + 1) * 0.5 * gl.drawingBufferHeight,
                    0
                ];
                mat = mat4.fromTranslation(mat4.create(), pos);
            } else {
                isVisible = false;
            }
        } else {
            mat = node.globalMatrix;
        }
        let globOffset = text.getGlobalOffset();

        if (!isVisible) {
            continue;
        }

        mat = mat4.translate(mat, mat, [globOffset[0], globOffset[1], 0.0]);
        gl.uniformMatrix4fv(this.umModelTransform, false, mat);
        gl.uniform4f(this.uMaterialColor, text.color[0], text.color[1], text.color[2], text.color[3]);
        gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, text.bufferIndex);
        gl.drawElements(gl.TRIANGLES, text.indexesCount, text.bufferIndexType, 0);
    }
}

grRenderChain_SkinnedTextured.prototype.render = function(ctrl) {
    let wasRebuilded = needRebuildScene;
    if (needRebuildScene) {
        this.rebuildScene(ctrl);
    }
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    gl.uniformMatrix4fv(this.umModelTransform, false, mat4.create());
    gl.activeTexture(gl.TEXTURE0);

    let cnt = 0;

    if (ctrl.cull) {
        gl.enable(gl.CULL_FACE);
    } else {
        gl.disable(gl.CULL_FACE);
    }

    gl.uniformMatrix4fv(this.umProjection, false, ctrl.camera.getProjectionMatrix());
    gl.uniformMatrix4fv(this.umView, false, mat4.fromQuat(mat4.create(), mat4.getRotation(quat.create(), ctrl.camera.getViewMatrix())));

    this.isSkyRendering = true;
    cnt += this.renderFlashesBatch(ctrl, this.skyBatch, false);
    this.isSkyRendering = false;

    gl.clear(gl.DEPTH_BUFFER_BIT);

    // render models
    gl.uniformMatrix4fv(this.umView, false, ctrl.camera.getViewMatrix());
    cnt += this.renderFlashesBatch(ctrl, this.normalBatch, true);

    // console.info("total", cnt,
    //     "sky", this.skyBatch.normalFlashes.length + this.skyBatch.additiveFlashes.length,
    //     "normal", this.skyBatch.normalFlashes.length + this.normalBatch.normalFlashes.length,
    //     "additive", this.skyBatch.additiveFlashes.length + this.normalBatch.additiveFlashes.length,
    //     "was rebuilded", wasRebuilded);

    gl.disable(gl.CULL_FACE);
    this.renderText(ctrl);
}

grRenderChain_SkinnedTextured.prototype.createFlash = function(node, mesh) {
    return new grRenderChain_SkinnedTexturedFlash(node, mesh);
}

grRenderChain_SkinnedTextured.prototype.fillFlashesFromModelNode = function(flashesBatch, node) {
    const model = node.model;
    const meshes = (model.exclusiveMeshes != undefined) ? model.exclusiveMeshes : model.meshes.list;

    for (const mesh of meshes) {
        if (mesh.isVisible === false) {
            continue;
        }

        if (mesh.materialIndex !== undefined && mesh.materialIndex < model.materials.length) {
            let material = model.materials.get(mesh.materialIndex);

            let usualLayer = undefined;
            let strangeBlendedLayer = undefined;
            let additiveLayer = undefined;

            let flash = this.createFlash(node, mesh);

            const parseLayer = function(iLayer) {
                let layer = material.layers.get(iLayer);
                if (layer == undefined) {
                    usualLayer = layer;
                    return;
                }
                switch (layer.method) {
                    case 0:
                        usualLayer = layer;
                        break;
                    case 1:
                        additiveLayer = layer;
                        break;
                    case 3:
                        strangeBlendedLayer = layer;
                        break;
                    default:
                        // console.warn("unknown layer method " + layer.method, layer, material);
                        break;
                }
            }

            if (mesh.layer !== undefined) {
                parseLayer(mesh.layer);
                // console.log("layer overrided", mesh.layer);
            } else {
                // console.log("layer detection");
                for (let iLayer in material.layers.list) {
                    parseLayer(iLayer);
                }
            }

            if (additiveLayer !== undefined) {
                flash.diffuselayer = additiveLayer;
                flashesBatch.addFlash(flash, 1);
                // console.log("added additive");
            } else {
                if (strangeBlendedLayer != undefined) {
                    flash.diffuselayer = strangeBlendedLayer;
                    flash.envmaplayer = usualLayer;
                } else {
                    flash.diffuselayer = usualLayer;
                }
                flashesBatch.addFlash(flash, 0);
            }

            flash.material = material;
        } else {
            let flash = this.createFlash(node, mesh);
            flashesBatch.addFlash(flash, 0);
        }
    }
}

grRenderChain_SkinnedTextured.prototype.fillFlashesFromNode = function(node) {
    if (!node.isActive) {
        return;
    }
    if (node.constructor === ObjectTreeNode || node.constructor === ObjectTreeNodeSkinned || node.constructor === ObjectTreeNodeSkinJoint) {
        for (node of node.nodes) {
            this.fillFlashesFromNode(node);
        }
    } else if (node.constructor === ObjectTreeNodeModel) {
        const model = node.model;
        if (model.constructor == RenderTextMesh) {
            this.textBatch.push(node);
        } else {
            switch (model.type) {
                default: {
                    this.fillFlashesFromModelNode(this.normalBatch, node);
                }
                break;
                case "sky": {
                    this.fillFlashesFromModelNode(this.skyBatch, node);
                }
                break;
            }
        }
    }
}

grRenderChain_SkinnedTextured.prototype.rebuildScene = function(ctrl) {
    for (node of ctrl.nodes) {
        this.fillFlashesFromNode(node);
    }
    for (node of ctrl.persistentNodes) {
        this.fillFlashesFromNode(node);
    }

    // console.log("flashes rebuilded", this.normalBatch, this.skyBatch);
    needRebuildScene = false;
}

grRenderChain_SkinnedTextured.prototype.flushScene = function(ctrl) {
    this.normalBatch.clear();
    this.skyBatch.clear();
    this.textBatch.length = 0;
    needRebuildScene = true;
}