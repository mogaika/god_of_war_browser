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

function grHelper_Cube(x, y, z, size) {
	if (size == undefined) {
		size = 50;
	}
	var vertexData = [
		x-size,y-size,z-size, x+size,y-size,z-size, x+size,y+size,z-size, x-size,y+size,z-size,
		x-size,y-size,z+size, x+size,y-size,z+size, x+size,y+size,z+size, x-size,y+size,z+size,
		x-size,y-size,z-size, x-size,y+size,z-size, x-size,y+size,z+size, x-size,y-size,z+size,
		x+size,y-size,z-size, x+size,y+size,z-size, x+size,y+size,z+size, x+size,y-size,z+size,
		x-size,y-size,z-size, x-size,y-size,z+size, x+size,y-size,z+size, x+size,y-size,z-size,
		x-size,y+size,z-size, x-size,y+size,z+size, x+size,y+size,z+size, x+size,y+size,z-size, 
	]
	var indexData = [
		0,1,2, 0,2,3, 4,5,6, 4,6,7,
		8,9,10, 8,10,11, 12,13,14, 12,14,15,
		16,17,18, 16,18,19, 20,21,22, 20,22,23 
	]

	return new grMesh(vertexData, indexData, gl.TRIANGLES)
}

function grHelper_Pivot(size) {	
	var mdl = new grModel();
	mdl.addMesh(grHelper_PivotMesh(size));
	return mdl;	
}
