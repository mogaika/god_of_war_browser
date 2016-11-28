function grHelper_Pivot(ctrl) {	
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
	
	var mdl = new grModel();
	mdl.addMesh(new grMesh(vertexData, indexData, gl.LINES));
	return mdl;	
}
