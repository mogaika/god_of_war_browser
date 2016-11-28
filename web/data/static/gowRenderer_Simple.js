function grRenderChain_Simple(ctrl) {
	console.log("grRenderChain_Simple new", ctrl);

	this.program = ctrl.downloadProgram("/static/simple.vs", "/static/simple.fs");

    gl.useProgram(this.program);

    this.aVertexPos = gl.getAttribLocation(this.program, "aVertexPos");
    this.umProjectionView = gl.getUniformLocation(this.program, "umProjectionView");
    this.umModelTransform = gl.getUniformLocation(this.program, "umModelTransform");

	gl.enableVertexAttribArray(this.aVertexPos);
	
	var vertexData = [
		-1000,0,0,
		1000,0,0,
		0,-1000,0,
		0,1000,0,
		0,0,-1000,
		0,0,1000,
	]
	var indexData = [
		0,1, 2,3, 4,5,
	]
	
	this.pivot = new grMesh(vertexData, indexData);
	
	gl.clearColor(0.2, 0.2, 0.2, 1);
	gl.clearDepth(1.0);
	gl.depthFunc(gl.LEQUAL);
	gl.disable(gl.BLEND);
	gl.depthMask(true);
	gl.enable(gl.DEPTH_TEST);
}

grRenderChain_Simple.prototype.free = function(ctrl) {
	console.log("grRenderChain_Simple free");
}

grRenderChain_Simple.prototype.drawMesh = function(mesh, primitive) {
	gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferVertex);
	gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.bufferIndex);

	gl.enableVertexAttribArray(this.aVertexPos);
	gl.vertexAttribPointer(this.aVertexPos, 3, gl.FLOAT, false, 0, 0);

	gl.drawElements(primitive, mesh.indexesCount, mesh.bufferIndexType, 0);
}

grRenderChain_Simple.prototype.render = function(ctrl) {
	console.log("grRenderChain_Simple render");
	
	gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
	gl.uniformMatrix4fv(this.umProjectionView, false, ctrl.camera.getProjViewMatrix());

	gl.uniformMatrix4fv(this.umModelTransform, false, mat4.create());
	this.drawMesh(this.pivot, gl.LINES);
	
	var rendered_meshes = 0;
	
	for (var i in ctrl.models) {
		var mdl = ctrl.models[i];
		gl.uniformMatrix4fv(this.umModelTransform, false, mdl.matrix);
		
		for (var j in mdl.meshes) {
			this.drawMesh(mdl.meshes[j], gl.TRIANGLES);
			
			rendered_meshes += 1;
		}
	}
	
	console.log("stats:", rendered_meshes);	
}