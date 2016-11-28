function grRenderChain_Simple(ctrl) {
	this.vertexShader = ctrl.downloadShader("/static/simple.vs", false);
	this.fragmentShader = ctrl.downloadShader("/static/simple.fs", true);
	this.program = ctrl.createProgram(this.vertexShader, this.fragmentShader);

    gl.useProgram(this.program);

    this.aVertexPos = gl.getAttribLocation(this.program, "aVertexPos");
    this.umProjectionView = gl.getUniformLocation(this.program, "umProjectionView");
    this.umModelTransform = gl.getUniformLocation(this.program, "umModelTransform");

	gl.enableVertexAttribArray(this.aVertexPos);
		
	gl.clearColor(0.2, 0.2, 0.2, 1);
	gl.clearDepth(1.0);
	gl.depthFunc(gl.LEQUAL);
	gl.disable(gl.BLEND);
	gl.depthMask(true);
	gl.enable(gl.DEPTH_TEST);
}

grRenderChain_Simple.prototype.free = function(ctrl) {
	console.log("grRenderChain_Simple free");
	gl.disableVertexAttribArray(this.aVertexPos);
	gl.deleteProgram(this.program);
	gl.deleteShader(this.vertexShader);
	gl.deleteShader(this.fragmentShader);
}

grRenderChain_Simple.prototype.drawMesh = function(mesh) {
	gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferVertex);
	gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.bufferIndex);

	gl.enableVertexAttribArray(this.aVertexPos);
	gl.vertexAttribPointer(this.aVertexPos, 3, gl.FLOAT, false, 0, 0);

	gl.drawElements(mesh.primitive, mesh.indexesCount, mesh.bufferIndexType, 0);
}

grRenderChain_Simple.prototype.render = function(ctrl) {
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