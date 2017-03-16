function grHelper_PivotMesh(size) {	
	if (size == undefined) {
		size = 1000;
	}
	var vertexData = [
		size,0,0,
		-size,0,0,
		0,size,0,
		0,-size,0,
		0,0,size,
		0,0,-size,
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
	return mesh;
}


function grHelper_Pivot(size) {	
	var mdl = new grModel();
	mdl.addMesh(grHelper_PivotMesh(size));
	return mdl;	
}
