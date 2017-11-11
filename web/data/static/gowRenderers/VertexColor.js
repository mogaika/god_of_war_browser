function grRenderChain_VertexColor(ctrl) {
    this.vertexShader = ctrl.downloadShader("/static/vertexColor.vs", false);
    this.fragmentShader = ctrl.downloadShader("/static/vertexColor.fs", true);
    this.program = ctrl.createProgram(this.vertexShader, this.fragmentShader);

    gl.useProgram(this.program);

    this.aVertexPos = gl.getAttribLocation(this.program, "aVertexPos");
    this.aVertexColor = gl.getAttribLocation(this.program, "aVertexColor");
    this.umProjectionView = gl.getUniformLocation(this.program, "umProjectionView");
    this.umModelTransform = gl.getUniformLocation(this.program, "umModelTransform");

    gl.enableVertexAttribArray(this.aVertexPos);
    gl.enableVertexAttribArray(this.aVertexColor);

    gl.clearColor(0.2, 0.2, 0.2, 1);
    gl.clearDepth(1.0);
    gl.depthFunc(gl.LEQUAL);
    gl.disable(gl.BLEND);
    gl.depthMask(true);
    gl.enable(gl.DEPTH_TEST);
}

grRenderChain_VertexColor.prototype.free = function(ctrl) {
    gl.disableVertexAttribArray(this.aVertexPos);
    gl.disableVertexAttribArray(this.aVertexColor);
    gl.deleteProgram(this.program);
    gl.deleteShader(this.vertexShader);
    gl.deleteShader(this.fragmentShader);
}

grRenderChain_VertexColor.prototype.drawMesh = function(mesh) {
    gl.enableVertexAttribArray(this.aVertexPos);
    gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferVertex);
    gl.vertexAttribPointer(this.aVertexPos, 3, gl.FLOAT, false, 0, 0);

    gl.enableVertexAttribArray(this.aVertexColor);
    gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferBlendColor);
    gl.vertexAttribPointer(this.aVertexColor, 4, gl.UNSIGNED_BYTE, true, 0, 0);

    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.bufferIndex);
    gl.drawElements(mesh.primitive, mesh.indexesCount, mesh.bufferIndexType, 0);
}

grRenderChain_VertexColor.prototype.render = function(ctrl) {
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    gl.uniformMatrix4fv(this.umProjectionView, false, ctrl.camera.getProjViewMatrix());

    gl.uniformMatrix4fv(this.umModelTransform, false, mat4.create());

    var rendered_meshes = 0;

    var mdls = [].concat(ctrl.models).concat(ctrl.helpers);

    for (var i in mdls) {
        var mdl = mdls[i];
        if (mdl.visible) {
            gl.uniformMatrix4fv(this.umModelTransform, false, mdl.matrix);

            for (var j in mdl.meshes) {
                this.drawMesh(mdl.meshes[j]);

                rendered_meshes += 1;
            }
        }
    }

    console.log("stats:", rendered_meshes);
}