function grRenderChain_SkinnedTexturedFlash(model, mesh) {
    this.model = model;
    this.mesh = mesh;
    this.material = undefined;
    this.layer = undefined;
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
    this.needRebuildScene = true;
    this.skyBatchMatrix = undefined;

    this.vertexShader = ctrl.downloadShader("/static/gowRenderers/SkinnedTextured.vs", false);
    this.fragmentShader = ctrl.downloadShader("/static/gowRenderers/SkinnedTextured.fs", true);
    this.program = ctrl.createProgram(this.vertexShader, this.fragmentShader);

    gl.useProgram(this.program);

    this.aVertexPos = gl.getAttribLocation(this.program, "aVertexPos");
    this.aVertexColor = gl.getAttribLocation(this.program, "aVertexColor");
    this.aVertexUV = gl.getAttribLocation(this.program, "aVertexUV");
    this.aVertexJointID = gl.getAttribLocation(this.program, "aVertexJointID");
    this.aVertexJointID2 = gl.getAttribLocation(this.program, "aVertexJointID2");

    this.umProjectionView = gl.getUniformLocation(this.program, "umProjectionView");
    this.umModelTransform = gl.getUniformLocation(this.program, "umModelTransform");
    this.uMaterialColor = gl.getUniformLocation(this.program, "uMaterialColor");
    this.uLayerColor = gl.getUniformLocation(this.program, "uLayerColor");
    this.uLayerOffset = gl.getUniformLocation(this.program, "uLayerOffset");
    this.uLayerDiffuseSampler = gl.getUniformLocation(this.program, "uLayerDiffuseSampler");
    this.uUseLayerDiffuseSampler = gl.getUniformLocation(this.program, "uUseLayerDiffuseSampler");
    this.uUseVertexColor = gl.getUniformLocation(this.program, "uUseVertexColor");
    this.uUseModelTransform = gl.getUniformLocation(this.program, "uUseModelTransform");
    this.uOnlyOpaqueRender = gl.getUniformLocation(this.program, "onlyOpaqueRender");

    this.umJoints = [];
    for (var i = 0; i < 12; i += 1) {
        this.umJoints.push(gl.getUniformLocation(this.program, "umJoints[" + i + "]"));
    }
    this.uUseJoints = gl.getUniformLocation(this.program, "uUseJoints");

    gl.enableVertexAttribArray(this.aVertexPos);
    gl.enableVertexAttribArray(this.aVertexColor);

    gl.uniform1i(this.uLayerDiffuseSampler, 0);
    gl.uniform1i(this.uUseLayerDiffuseSampler, 0);
    gl.uniform1i(this.uUseJoints, 0);

    gl.clearColor(0.25, 0.25, 0.25, 1.0);
    gl.clearDepth(1.0);
    gl.depthFunc(gl.LEQUAL);
    gl.disable(gl.BLEND);
    gl.depthMask(true);
    gl.enable(gl.DEPTH_TEST);
    gl.disable(gl.CULL_FACE);
}

grRenderChain_SkinnedTextured.prototype.free = function(ctrl) {
    // TODO :add missed fields
    gl.disableVertexAttribArray(this.aVertexPos);
    gl.disableVertexAttribArray(this.aVertexColor);
    gl.disableVertexAttribArray(this.aVertexUV);
    gl.disableVertexAttribArray(this.aVertexJointID);
    gl.disableVertexAttribArray(this.aVertexJointID2);
    gl.deleteProgram(this.program);
    gl.deleteShader(this.vertexShader);
    gl.deleteShader(this.fragmentShader);
}

grRenderChain_SkinnedTextured.prototype.drawMesh = function(mesh, hasTexture = false, hasJoints = false) {
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

    if (mesh.bufferUV && hasTexture) {
        gl.uniform1i(this.uUseLayerDiffuseSampler, 1);
        gl.enableVertexAttribArray(this.aVertexUV);
        gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferUV);
        gl.vertexAttribPointer(this.aVertexUV, 2, gl.FLOAT, false, 0, 0);
    } else {
        gl.uniform1i(this.uUseLayerDiffuseSampler, 0);
        gl.disableVertexAttribArray(this.aVertexUV);
    }

    if (mesh.bufferJointIds && hasJoints) {
        gl.uniform1i(this.uUseJoints, 1);
        gl.enableVertexAttribArray(this.aVertexJointID);
        gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferJointIds);
        gl.vertexAttribPointer(this.aVertexJointID, 1, gl.BYTE, false, 0, 0);

        gl.enableVertexAttribArray(this.aVertexJointID2);
        if (mesh.bufferJointIds2) {
            gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferJointIds2);
        } else {
            gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferJointIds);
        }
        gl.vertexAttribPointer(this.aVertexJointID2, 1, gl.BYTE, false, 0, 0);
    } else {
        // TODO : restore warn
        //if (hasJoints) {
        //    console.warn("has joints but without jointIdsBuffer", mesh);
        //}
        gl.uniform1i(this.uUseJoints, 0);
        gl.disableVertexAttribArray(this.aVertexJointID);
        gl.disableVertexAttribArray(this.aVertexJointID2);
    }

    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.bufferIndex);
    gl.drawElements(mesh.primitive, mesh.indexesCount, mesh.bufferIndexType, 0);
}

grRenderChain_SkinnedTextured.prototype.renderFlashesArray = function(ctrl, flashesBatch, useSkelet) {
    var mdl = -1;
    var material = -1;
    var mesh = -1;
    var layer = -1;
    var texture = -1;
    var hasTxr = false;
    var hasSkelet = false;
    for (var iFlash in flashesBatch) {
        var flash = flashesBatch[iFlash];

        if (flash.model != mdl) {
            mdl = flash.model;
            if (mdl) {
                gl.uniform1i(this.uUseModelTransform, 1);
                gl.uniformMatrix4fv(this.umModelTransform, false, mdl.matrix);
            } else {
                gl.uniform1i(this.uUseModelTransform, 0);
            }
        }

        if (flash.mesh != mesh) {
            mesh = flash.mesh;
            if (mesh.isDepthTested) {
                gl.enable(gl.DEPTH_TEST);
            } else {
                gl.disable(gl.DEPTH_TEST);
            }

            if (mesh.jointMapping && useSkelet && mdl && mdl.skeleton) {
                for (var i in mesh.jointMapping) {
                    if (i >= 12) {
                        console.warn("jointMap array in shader is overflowed", mesh.jointMapping);
                    }
                    var jointId = mesh.jointMapping[i];
                    if (jointId >= mdl.skeleton.length) {
                        console.warn("joint mapping out of index. jointMapping[" + i + "]=" + jointId + " >= " + mdl.skeleton.length);
                    } else {
                        gl.uniformMatrix4fv(this.umJoints[i], false, mdl.skeleton[jointId]);
                    }
                }
                hasSkelet = true;
            }
        }

        if (flash.material != material) {
            material = flash.material;
            if (material != undefined) {
                gl.uniform4f(this.uMaterialColor, material.color[0], material.color[1], material.color[2], material.color[3]);
            } else {
                gl.uniform4f(this.uMaterialColor, 1.0, 1.0, 1.0, 1.0);
            }
        }

        if (flash.layer != layer) {
            layer = flash.layer;
            if (layer != undefined) {
                gl.uniform4f(this.uLayerColor, layer.color[0], layer.color[1], layer.color[2], layer.color[3]);
                gl.uniform2f(this.uLayerOffset, layer.uvoffset[0], layer.uvoffset[1]);
            } else {
                gl.uniform4f(this.uLayerColor, 1.0, 1.0, 1.0, 1.0);
                gl.uniform2f(this.uLayerOffset, 0.0, 0.0);
            }
        }

        if (layer != undefined && layer.textures && layer.textures.length && layer.textures[layer.textureIndex]) {
            gl.bindTexture(gl.TEXTURE_2D, layer.textures[layer.textureIndex].get());
            hasTxr = true;
        }

        this.drawMesh(mesh, hasTxr, hasSkelet);
    }
    return flashesBatch.length;
}

grRenderChain_SkinnedTextured.prototype.renderFlashesBatch = function(ctrl, flashesBatch, useSkelet = true) {
    gl.blendEquation(gl.FUNC_ADD);
    gl.depthMask(true);
    gl.enable(gl.BLEND);
    gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE);
    gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE);

    //gl.uniform1i(this.uOnlyOpaqueRender, 1);
    var cnt = this.renderFlashesArray(ctrl, flashesBatch.normalFlashes, useSkelet);
    //gl.uniform1i(this.uOnlyOpaqueRender, 0);

    gl.blendEquation(gl.FUNC_ADD);
    gl.depthMask(false);
    gl.enable(gl.BLEND);
    gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE, gl.ONE, gl.ONE);

    cnt += this.renderFlashesArray(ctrl, flashesBatch.additiveFlashes, useSkelet);
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
    gl.disableVertexAttribArray(this.aVertexJointID2);

    gl.uniformMatrix4fv(this.umProjectionView, false, ctrl.orthoMatrix);
    gl.uniform1i(this.uUseLayerDiffuseSampler, 1);
    gl.uniform1i(this.uUseVertexColor, 0);
    gl.uniform1i(this.uUseJoints, 0);
    gl.uniform1i(this.uOnlyOpaqueRender, 0);

    gl.activeTexture(gl.TEXTURE0);
    gl.bindTexture(gl.TEXTURE_2D, ctrl.fontTexture.get());
    gl.uniform4f(this.uLayerColor, 1.0, 1.0, 1.0, 1.0);
    gl.uniform2f(this.uLayerOffset, 0.0, 0.0);
    for (var i = 0; i < ctrl.texts.length; i++) {
        var text = ctrl.texts[i];

        gl.enableVertexAttribArray(this.aVertexPos);
        gl.bindBuffer(gl.ARRAY_BUFFER, text.bufferVertex);
        gl.vertexAttribPointer(this.aVertexPos, 2, gl.FLOAT, false, 0, 0);

        gl.enableVertexAttribArray(this.aVertexUV);
        gl.bindBuffer(gl.ARRAY_BUFFER, text.bufferUV);
        gl.vertexAttribPointer(this.aVertexUV, 2, gl.FLOAT, false, 0, 0);

        var isVisible = true;
        var mat = mat4.identity(mat4.create());
        if (text.is3d) {
            var pos3d = vec3.fromValues(text.position[0], text.position[1], text.position[2]);
            var pos2d = vec3.transformMat4(vec3.create(), pos3d, ctrl.camera.getProjViewMatrix());
            if (pos2d[2] < 1) {
                var pos = [(pos2d[0] + 1) * 0.5 * gl.drawingBufferWidth, (pos2d[1] + 1) * 0.5 * gl.drawingBufferHeight, 0];
                mat = mat4.translate(mat4.create(), mat, pos);
            } else {
                isVisible = false;
            }
        } else {
            mat = mat4.translate(mat4.create(), mat, text.position);
        }

        if (isVisible) {
            gl.uniformMatrix4fv(this.umModelTransform, false, mat);
            gl.uniform4f(this.uMaterialColor, text.color[0], text.color[1], text.color[2], text.color[3]);
            gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, text.bufferIndex);
            gl.drawElements(gl.TRIANGLES, text.indexesCount, text.bufferIndexType, 0);
        }
    }
}

grRenderChain_SkinnedTextured.prototype.render = function(ctrl) {
    var wasRebuilded = needRebuildScene;
    if (needRebuildScene) {
        this.rebuildScene(ctrl);
    }
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    gl.uniformMatrix4fv(this.umModelTransform, false, mat4.create());
    gl.activeTexture(gl.TEXTURE0);

    var cnt = 0;

    // render sky
    if (this.skyBatchMatrix) {
        gl.uniform1i(this.uUseModelTransform, 0);
        var finalMat = mat4.mul(mat4.create(), ctrl.camera.getViewMatrix(), this.skyBatchMatrix);
        var rot = mat4.getRotation(quat.create(), finalMat);
        var nullPosedMatrix = mat4.fromQuat(mat4.create(), rot);
        nullPosedMatrix = mat4.mul(mat4.create(), ctrl.camera.getProjectionMatrix(), nullPosedMatrix);
        gl.uniformMatrix4fv(this.umProjectionView, false, nullPosedMatrix);
        cnt += this.renderFlashesBatch(ctrl, this.skyBatch, false);
        gl.clear(gl.DEPTH_BUFFER_BIT);
    }

    // render models
    gl.uniform1i(this.uUseModelTransform, 1);
    gl.uniformMatrix4fv(this.umProjectionView, false, ctrl.camera.getProjViewMatrix());
    cnt += this.renderFlashesBatch(ctrl, this.normalBatch, true);

    console.info("total", cnt,
        "normal", this.skyBatch.normalFlashes.length + this.normalBatch.normalFlashes.length,
        "additive", this.skyBatch.additiveFlashes.length + this.normalBatch.additiveFlashes.length,
        "was rebuilded", wasRebuilded);

    this.renderText(ctrl);
}

grRenderChain_SkinnedTextured.prototype.createFlash = function(mdl, mesh) {
    return new grRenderChain_SkinnedTexturedFlash(mdl, mesh);
}

grRenderChain_SkinnedTextured.prototype.fillFlashesFromModel = function(flashesBatch, mdl) {
    if (mdl.visible === false) {
        return;
    }
    var meshes = (mdl.exclusiveMeshes != undefined) ? mdl.exclusiveMeshes : mdl.meshes;

    for (var iMesh in meshes) {
        var mesh = meshes[iMesh];
        if (mesh.isVisible === false) {
            continue;
        }

        if (mesh.materialIndex != undefined && mdl.materials && mesh.materialIndex < mdl.materials.length) {
            var mat = mdl.materials[mesh.materialIndex];
            for (var iLayer in mat.layers) {
                var layer = mat.layers[iLayer];
                var flash = this.createFlash(mdl, mesh);
                flash.layer = layer;
                flash.material = mat;
                flashesBatch.addFlash(flash, layer.method);
            }
        } else {
            var flash = this.createFlash(mdl, mesh);
            flashesBatch.addFlash(flash, 0);
        }
    }
}

grRenderChain_SkinnedTextured.prototype.fillFlashesFromModels = function(models) {
    for (var iMdl in models) {
        var mdl = models[iMdl];
        if (!mdl.type) {
            this.fillFlashesFromModel(this.normalBatch, mdl);
        } else if (mdl.type == "sky") {
            this.fillFlashesFromModel(this.skyBatch, mdl);
            this.skyBatchMatrix = mdl.matrix;
        }
    }
}

grRenderChain_SkinnedTextured.prototype.rebuildScene = function(ctrl) {
    this.skyBatchMatrix = undefined;
    this.fillFlashesFromModels(ctrl.models);
    this.fillFlashesFromModels(ctrl.helpers);
    console.log("flashes rebuilded", this.normalBatch, this.skyBatch);
    needRebuildScene = false;
}

grRenderChain_SkinnedTextured.prototype.flushScene = function(ctrl) {
    this.normalBatch.clear();
    this.skyBatch.clear();
    needRebuildScene = true;
}