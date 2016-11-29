function grHelper_Pivot(ctrl) {	
	var vertexData = [
		1000,0,0,
		-1000,0,0,
		0,1000,0,
		0,-1000,0,
		0,0,1000,
		0,0,-1000,
	]
	var colorData = [
		0xff, 0x00, 0x00, 0xff,
		0x00, 0x00, 0x00, 0xff,
		0x00, 0xff, 0x00, 0xff,
		0x00, 0x00, 0x00, 0xff,
		0x00, 0x00, 0xff, 0xff,
		0x00, 0x00, 0x00, 0xff,
	]
	var indexData = [
		0,1, 2,3, 4,5,
	]

	var mesh = new grMesh(vertexData, indexData, gl.LINES)
	mesh.setBlendColors(colorData);

	var mdl = new grModel();
	mdl.addMesh(mesh);
	return mdl;	
}
