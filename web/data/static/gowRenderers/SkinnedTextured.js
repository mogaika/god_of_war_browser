function grRenderChain_SkinnedTextured(ctrl) {
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
    this.uMaterialDiffuseSampler = gl.getUniformLocation(this.program, "uMaterialDiffuseSampler");
    this.uUseMaterialDiffuseSampler = gl.getUniformLocation(this.program, "uUseMaterialDiffuseSampler");
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

    gl.uniform1i(this.uMaterialDiffuseSampler, 0);
    gl.uniform1i(this.uUseMaterialDiffuseSampler, 0);
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
        gl.uniform1i(this.uUseMaterialDiffuseSampler, 1);
        gl.enableVertexAttribArray(this.aVertexUV);
        gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferUV);
        gl.vertexAttribPointer(this.aVertexUV, 2, gl.FLOAT, false, 0, 0);
    } else {
        gl.uniform1i(this.uUseMaterialDiffuseSampler, 0);
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
        if (hasJoints) {
            console.warn("has joints but without jointIdsBuffer", mesh);
        }
        gl.uniform1i(this.uUseJoints, 0);
        gl.disableVertexAttribArray(this.aVertexJointID);
        gl.disableVertexAttribArray(this.aVertexJointID2);
    }

    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.bufferIndex);
    gl.drawElements(mesh.primitive, mesh.indexesCount, mesh.bufferIndexType, 0);
}

grRenderChain_SkinnedTextured.prototype.renderMesh = function(ctrl, mdl, mesh, hasSkelet, filterFunc, usedJoints) {
    if ((filterFunc && !filterFunc(mdl, mesh)) || !mesh.isVisible)
        return false;

    if (mesh.isDepthTested) {
        gl.enable(gl.DEPTH_TEST);
    } else {
        gl.disable(gl.DEPTH_TEST);
    }

    if (hasSkelet && mesh.jointMapping) {
        for (var i in mesh.jointMapping) {
            if (i >= 12) {
                console.warn("jointMap array in shader is overflowed", mesh.jointMapping)
            }
            var jointId = mesh.jointMapping[i];
            if (jointId >= mdl.skeleton.length) {
                console.warn("joint mapping out of index. jointMapping[" + i + "]=" + jointId + " >= " + mdl.skeleton.length);
            } else {
                gl.uniformMatrix4fv(this.umJoints[i], false, mdl.skeleton[jointId]);
            }
        }
    } else {
        hasSkelet = false;
    }

    var hasTxr = false;
    if (mesh.materialIndex != undefined && mdl.materials && mesh.materialIndex < mdl.materials.length) {
        var mat = mdl.materials[mesh.materialIndex];
        if (mat.textureDiffuse != undefined) {
            gl.bindTexture(gl.TEXTURE_2D, mat.textureDiffuse.get());
            hasTxr = true;
        }
        gl.uniform4f(this.uMaterialColor, mat.color[0], mat.color[1], mat.color[2], mat.color[3]);
    } else {
        gl.uniform4f(this.uMaterialColor, 1.0, 1.0, 1.0, 1.0);
    }

    this.drawMesh(mesh, hasTxr, hasSkelet);

    return true;
}

grRenderChain_SkinnedTextured.prototype.renderModel = function(ctrl, mdl, useSkelet, filterFunc) {
    var lastMatIndex = -1;
    var cnt = 0;
    if (mdl.visible) {
        gl.uniformMatrix4fv(this.umModelTransform, false, mdl.matrix);

        var hasSkelet = useSkelet && !!mdl.skeleton;
        var usedJoints = [];
        if (mdl.exclusiveMeshes != undefined) {
            for (var j in mdl.exclusiveMeshes) {
                if (this.renderMesh(ctrl, mdl, mdl.exclusiveMeshes[j], hasSkelet, filterFunc, usedJoints)) {
                    cnt++;
                }
            }
        } else {
            for (var j in mdl.meshes) {
                if (this.renderMesh(ctrl, mdl, mdl.meshes[j], hasSkelet, filterFunc, usedJoints)) {
                    cnt++;
                }
            }
        }
        //console.log("Used joints", usedJoints.sort(function(a, b){return a - b;}));
    }
    return cnt;
}

grRenderChain_SkinnedTextured.prototype.renderModels = function(ctrl, mdls, filterFunc, useSkelet = true) {
    var cnt = 0;
    for (var i in mdls) {
        var mdl = mdls[i];
        cnt += this.renderModel(ctrl, mdl, useSkelet, filterFunc)
    }
    return cnt;
}

function __mdl_mesh_normal_nonalpha_tester(mdl, mesh) {
    return mesh.materialIndex != undefined && mdl.materials.length != 0 && mdl.materials[mesh.materialIndex].method == 0 && !mdl.materials[mesh.materialIndex].hasAlpha;
}

function __mdl_mesh_normal_alpha_tester(mdl, mesh) {
    return mesh.materialIndex == undefined || mdl.materials.length == 0 ? true : mdl.materials[mesh.materialIndex].method == 0 && mdl.materials[mesh.materialIndex].hasAlpha;
}

function __mdl_mesh_additive_tester(mdl, mesh) {
    return mesh.materialIndex != undefined && mdl.materials.length != 0 && mdl.materials[mesh.materialIndex].method == 1;
}

function __mdl_mesh_undrawed_tester(mdl, mesh) {
    return mesh.materialIndex != undefined && mdl.materials.length != 0 && mdl.materials[mesh.materialIndex].method > 1;
}

grRenderChain_SkinnedTextured.prototype.renderCycle = function(ctrl, mdls, useSkelet = true) {
    if (mdls.length > 0) {
        gl.disable(gl.BLEND);
        gl.depthMask(true);

        gl.enable(gl.BLEND);
        gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE);

        //gl.uniform1i(this.uOnlyOpaqueRender, 1);
        var undrawedDrawed = this.renderModels(ctrl, mdls, __mdl_mesh_undrawed_tester, useSkelet);
        if (undrawedDrawed) {
            console.warn(undrawedDrawed + ' miss filter')
        }
        this.renderModels(ctrl, mdls, __mdl_mesh_normal_nonalpha_tester, useSkelet);


        gl.depthMask(true);
        //gl.enable(gl.BLEND);
        gl.uniform1i(this.uOnlyOpaqueRender, 0);

        gl.enable(gl.BLEND);

        //gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);
        //gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA);
        //gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE, gl.ONE, gl.ONE);
        //glBlendFunc(GL_ZERO,GL_SRC_COLOR);   
        //gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE);
        this.renderModels(ctrl, mdls, __mdl_mesh_normal_alpha_tester, useSkelet);

        gl.depthMask(false);

        //gl.blendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA);
        gl.blendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD);
        gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE, gl.ONE, gl.ONE);
        //gl.blendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA);
        //gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA);
        this.renderModels(ctrl, mdls, __mdl_mesh_additive_tester, useSkelet);

        gl.depthMask(true);
        //gl.uniform1i(this.uOnlyOpaqueRender, 1);
    }
}

grRenderChain_SkinnedTextured.prototype.render = function(ctrl) {
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    gl.uniformMatrix4fv(this.umModelTransform, false, mat4.create());
    gl.activeTexture(gl.TEXTURE0);

    var skies = ctrl.models.filter(function(m) {
        return m.type === "sky";
    });

    if (skies.length > 0) {
        if (skies.length > 1) {
            console.warn("Too many skies: ", skies);
        }

        gl.uniform1i(this.uUseModelTransform, 0);

        var sky = skies[0];
        var finalMat = mat4.mul(mat4.create(), ctrl.camera.getViewMatrix(), sky.matrix);
        var rot = mat4.getRotation(quat.create(), finalMat);
        var nullPosedMatrix = mat4.fromQuat(mat4.create(), rot);
        nullPosedMatrix = mat4.mul(mat4.create(), ctrl.camera.getProjectionMatrix(), nullPosedMatrix);

        gl.uniformMatrix4fv(this.umProjectionView, false, nullPosedMatrix);

        this.renderCycle(ctrl, [sky], false);
        //this.renderModels(ctrl, [sky], undefined, false);

        gl.clear(gl.DEPTH_BUFFER_BIT);
    }

    var mdls = [].concat(ctrl.helpers).concat(ctrl.models.filter(function(m) {
        return !m.type;
    }));
    if (mdls.length > 0) {
        gl.uniform1i(this.uUseModelTransform, 1);
        gl.uniformMatrix4fv(this.umProjectionView, false, ctrl.camera.getProjViewMatrix());

        this.renderCycle(ctrl, mdls);
    }
    console.info("redrawed");
}
