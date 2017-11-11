function grRenderChain_Textured(ctrl) {
    this.vertexShader = ctrl.downloadShader("/static/gowRenderers/Textured.vs", false);
    this.fragmentShader = ctrl.downloadShader("/static/gowRenderers/Textured.fs", true);
    this.program = ctrl.createProgram(this.vertexShader, this.fragmentShader);

    gl.useProgram(this.program);

    this.aVertexPos = gl.getAttribLocation(this.program, "aVertexPos");
    this.aVertexColor = gl.getAttribLocation(this.program, "aVertexColor");
    this.aVertexUV = gl.getAttribLocation(this.program, "aVertexUV");

    this.umProjectionView = gl.getUniformLocation(this.program, "umProjectionView");
    this.umModelTransform = gl.getUniformLocation(this.program, "umModelTransform");
    this.uMaterialColor = gl.getUniformLocation(this.program, "uMaterialColor");
    this.uMaterialDiffuseSampler = gl.getUniformLocation(this.program, "uMaterialDiffuseSampler");
    this.uUseMaterialDiffuseSampler = gl.getUniformLocation(this.program, "uUseMaterialDiffuseSampler");

    gl.enableVertexAttribArray(this.aVertexPos);
    gl.enableVertexAttribArray(this.aVertexColor);

    gl.uniform1i(this.uMaterialDiffuseSampler, 0);
    gl.uniform1i(this.uUseMaterialDiffuseSampler, 0);

    gl.clearColor(0.2, 0.2, 0.2, 1);
    gl.clearDepth(1.0);
    gl.depthFunc(gl.LEQUAL);
    gl.disable(gl.BLEND);
    gl.depthMask(true);
    gl.enable(gl.DEPTH_TEST);
}

grRenderChain_Textured.prototype.free = function(ctrl) {
    gl.disableVertexAttribArray(this.aVertexPos);
    gl.disableVertexAttribArray(this.aVertexColor);
    gl.disableVertexAttribArray(this.aVertexUV);
    gl.deleteProgram(this.program);
    gl.deleteShader(this.vertexShader);
    gl.deleteShader(this.fragmentShader);
}

grRenderChain_Textured.prototype.drawMesh = function(mesh, hasTexture = false) {
    gl.enableVertexAttribArray(this.aVertexPos);
    gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferVertex);
    gl.vertexAttribPointer(this.aVertexPos, 3, gl.FLOAT, false, 0, 0);

    if (mesh.bufferBlendColor) {
        gl.enableVertexAttribArray(this.aVertexColor);
        gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferBlendColor);
        gl.vertexAttribPointer(this.aVertexColor, 4, gl.UNSIGNED_BYTE, true, 0, 0);
    } else {
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

    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.bufferIndex);
    gl.drawElements(mesh.primitive, mesh.indexesCount, mesh.bufferIndexType, 0);
}

grRenderChain_Textured.prototype.render = function(ctrl) {
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    gl.uniformMatrix4fv(this.umProjectionView, false, ctrl.camera.getProjViewMatrix());

    gl.uniformMatrix4fv(this.umModelTransform, false, mat4.create());

    gl.activeTexture(gl.TEXTURE0);

    var rendered_meshes = 0;
    var mdls = [].concat(ctrl.models).concat(ctrl.helpers);

    for (var i in mdls) {
        var mdl = mdls[i];
        var lastMatIndex = -1;
        if (mdl.visible) {
            gl.uniformMatrix4fv(this.umModelTransform, false, mdl.matrix);

            for (var j in mdl.meshes) {
                var mesh = mdl.meshes[j];
                var hasTxr = true;

                if (mesh.isDepthTested) {
                    gl.enable(gl.DEPTH_TEST);
                } else {
                    gl.disable(gl.DEPTH_TEST);
                }
                gl.activeTexture(gl.TEXTURE0);
                if (mesh.materialIndex != undefined && mdl.materials && mdl.materials.length) {
                    var mat = mdl.materials[mesh.materialIndex];
                    gl.bindTexture(gl.TEXTURE_2D, mat.textureDiffuse.get());
                    gl.uniform4f(this.uMaterialColor, mat.color[0], mat.color[1], mat.color[2], mat.color[3]);
                } else {
                    gl.bindTexture(gl.TEXTURE_2D, ctrl.emptyTexture.get());
                    hasTxr = false;
                    gl.uniform4f(this.uMaterialColor, 255, 255, 255, 255);
                }

                this.drawMesh(mesh, hasTxr);

                rendered_meshes += 1;
            }
        }
    }

    console.log("stats:", rendered_meshes);
}