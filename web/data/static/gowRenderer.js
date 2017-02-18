var gr_instance;
var gl;

function grMesh(vertexArray, indexArray, primitive) {
	this.refs = 0;
	this.indexesCount = indexArray.length;
	this.primitive = (!!primitive) ? primitive : gl.TRIANGLES;
	this.isDepthTested = true;
	this.hasAlpha = false;
	this.isVisible = true;
	
	// construct array of unique indexes
	this.usedIndexes = [];
    for (var i in indexArray) {
        if (this.usedIndexes.indexOf(indexArray[i]) === -1) {
            this.usedIndexes.push(indexArray[i]);
		}
	}
	
	this.bufferVertex = gl.createBuffer();
	gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferVertex);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(vertexArray), gl.STATIC_DRAW);
	
	this.bufferIndex = gl.createBuffer();
	gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, this.bufferIndex);
	this.bufferIndexType = (indexArray.length > 255) ? gl.UNSIGNED_BYTE : gl.UNSIGNED_SHORT;
    gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, (this.bufferIndexType === gl.UNSIGNED_SHORT) ? (new Uint16Array(indexArray)) : (new Uint8Array(indexArray)), gl.STATIC_DRAW);
    
	this.bufferBlendColor = undefined;
	this.bufferUV = undefined;
	this.bufferNormals = undefined;
	this.bufferJointIds = undefined;
	this.jointMapping = undefined;
	this.materialIndex = undefined;
}

grMesh.prototype.setVisible = function(visible) {
	this.isVisible = visible;
}

grMesh.prototype.setDepthTest = function(isDepthTested) {
	this.isDepthTested = isDepthTested;
}

grMesh.prototype.setBlendColors = function(data) {
	if (!this.bufferBlendColor) {
		this.bufferBlendColor = gl.createBuffer();
	}
	this.hasAlpha = false;
	
	for (var i in this.usedIndexes) {
		if (data[this.usedIndexes[i] * 4 + 3] < 127) {
			this.hasAlpha = true;
			//console.log("mesh hasAlpha detected from ", this, "by ", i, "index");
			break;
		}
	}
	
	gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferBlendColor);
	gl.bufferData(gl.ARRAY_BUFFER, new Uint8Array(data), gl.STATIC_DRAW);
}

grMesh.prototype.setUVs = function(data, materialIndex) {
	if (!this.bufferUV) {
		this.bufferUV = gl.createBuffer();
	}
	gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferUV);
	gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data), gl.STATIC_DRAW);
	this.materialIndex = materialIndex;
}

grMesh.prototype.setNormals = function(data) {
	if (!this.bufferNormals) {
		this.bufferNormals = gl.createBuffer();
	}
	gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferNormals);
	gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data), gl.STATIC_DRAW);
}

grMesh.prototype.setJointIds = function(jointIds, jointMapping) {
	if (!this.bufferJointIds) {
		this.bufferJointIds = gl.createBuffer();
	}
	this.jointMapping = jointMapping;
	gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferJointIds);
	gl.bufferData(gl.ARRAY_BUFFER, new Uint16Array(jointIds), gl.STATIC_DRAW);
}

grMesh.prototype.free = function() {
    if (this.bufferVertex) gl.deleteBuffer(this.bufferVertex);
    if (this.bufferIndex) gl.deleteBuffer(this.bufferIndex);
    if (this.bufferBlendColor) gl.deleteBuffer(this.bufferBlendColor);
    if (this.bufferUV) gl.deleteBuffer(this.bufferUV);
	if (this.bufferNormals) gl.deleteBuffer(this.bufferNormals);
	if (this.bufferJointIds) gl.deleteBuffer(this.bufferJointIds);
}

grMesh.prototype.claim = function() { this.refs ++; }
grMesh.prototype.unclaim = function() { if (--this.refs === 0) this.free(); }

function grModel() {
	this.visible = true;
	this.meshes = [];
	this.materials = [];
	this.skeleton = undefined;
	this.matrix = mat4.create();
	this.type = undefined;
	this.exclusiveMesh = undefined;
}

grModel.prototype.showExclusiveMesh = function(mesh) {
	this.exclusiveMesh = mesh;
}

grModel.prototype.setType = function(type) {
	this.type = type;
}

grModel.prototype.addMesh = function(mesh) {
	mesh.claim();
	this.meshes.push(mesh);
}

grModel.prototype.addMaterial = function(material) {
	material.claim();
	this.materials.push(material);
}

function mshFromSklt(sklt, key="OurJointToIdleMat") {
	var vrtxs = [];
	var indxs = [];
	var clrs = [];
	
	vrtxs.length = 3 * sklt.length;
	clrs.length = 4 * sklt.length;
	
	for (var i in sklt) {
		var joint = sklt[i];
		vrtxs[i*3+0] = joint[key][12];
		vrtxs[i*3+1] = joint[key][13];
		vrtxs[i*3+2] = joint[key][14];
		clrs[i*4+0] = (i % 8) * 15;
		clrs[i*4+1] = ((i / 8) % 8) * 15;
		clrs[i*4+2] = ((i / 64) % 8) * 15;
		clrs[i*4+3] = 127;
		
		if (joint.Parent >= 0) {
			indxs.push(joint.Parent);
			indxs.push(i);
		}
	}
	
	var sklMesh = new grMesh(vrtxs, indxs, gl.LINES);
	sklMesh.setDepthTest(false);
	sklMesh.setBlendColors(clrs);
	return sklMesh;
}

grModel.prototype.loadSkeleton = function(sklt) {
	this.addMesh(mshFromSklt(sklt));
	this.skeleton = [];

	for (var i in sklt) {
		this.skeleton.push(new Float32Array(sklt[i].RenderMat));
	}
}

grModel.prototype.free = function() {
	for (var i in this.meshes) { this.meshes[i].unclaim(); }
	for (var i in this.materials) { this.materials[i].unclaim(); }
}

function grTexture__handleLoading(img, txr) {
	if (!txr.loaded) {
		txr.txr = gl.createTexture();
	}
	gl.bindTexture(gl.TEXTURE_2D, txr.txr);
	gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img);
	gl.generateMipmap(gl.TEXTURE_2D);
	gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
	gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
	gl.bindTexture(gl.TEXTURE_2D, null);
	txr.loaded = true;
}

function grTexture(src, wait = false) {
	this.loaded = wait;
	this.txr = undefined;
	
	if (wait) {
		this.txr = gl.createTexture();
		// pink placeholder
		gl.bindTexture(gl.TEXTURE_2D, this.txr);
		gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, new Uint8Array([255, 0, 255, 128]));
		gl.bindTexture(gl.TEXTURE_2D, null);
	}
	
    var img = new Image();
    img.src = src;
	var _this = this;
    img.onload = function() {
		grTexture__handleLoading(img, _this);
		gr_instance.requestRedraw();
    };
}

grTexture.prototype.free = function() { if (this.txr) gl.deleteTexture(this.txr); }
grTexture.prototype.get = function() { return this.loaded ? this.txr : gr_instance.emptyTexture.get(); }

function grMaterial() {
	this.refs = 0;
	this.color = [1,1,1,1];
	this.textureDiffuse = undefined;
	this.hasAlpha = false;
}

grMaterial.prototype.setColor = function(color) {
	this.color = color;
}

grMaterial.prototype.setDiffuse = function(txtr) {
	this.textureDiffuse = txtr;
}

grMaterial.prototype.setHasAlphaAttribute = function(alpha=true) {
	this.hasAlpha = alpha;
}

grMaterial.prototype.claim = function() { this.refs ++; }
grMaterial.prototype.unclaim = function() { if (--this.refs == 0) { this.free(); }}

grMaterial.prototype.free = function() {
	if (this.textureDiffuse) { this.textureDiffuse.free(); }
}

function grCameraTargeted() {
	this.farPlane = 50000.0;
	this.nearPlane = 1.0;
	this.fow = 55.0;
	this.target = [0, 0, 0];
	this.distance = 100.0;
	this.rotation = [15.0, 45.0*3, 0];
}

grCameraTargeted.prototype.getProjectionMatrix = function() {
	return mat4.perspective(mat4.create(), glMatrix.toRadian(this.fow), gr_instance.rectX/gr_instance.rectY, this.nearPlane,this.farPlane);
}

grCameraTargeted.prototype.getViewMatrix = function() {
	var viewMatrix = mat4.create();
    mat4.translate(viewMatrix, viewMatrix, [0.0, 0.0, -this.distance]);
    mat4.rotate(viewMatrix, viewMatrix, glMatrix.toRadian(this.rotation[0]), [1, 0, 0]);
    mat4.rotate(viewMatrix, viewMatrix, glMatrix.toRadian(this.rotation[1]), [0, 1, 0]);
	mat4.rotate(viewMatrix, viewMatrix, glMatrix.toRadian(this.rotation[2]), [0, 0, 1]);
	return mat4.translate(viewMatrix, viewMatrix, [-this.target[0], -this.target[1], -this.target[2]]);
}

grCameraTargeted.prototype.getProjViewMatrix = function() {
    return mat4.mul(mat4.create(), this.getProjectionMatrix(), this.getViewMatrix());
}

grCameraTargeted.prototype.onMouseWheel = function(delta) {
	var resizeDelta = Math.sqrt(this.distance) * delta * 0.01;
	this.distance -= resizeDelta * 2;
	if (this.distance < 1.0)
		this.distance = 1.0;
	gr_instance.requestRedraw();
}

grCameraTargeted.prototype.onMouseMove = function(btns, moveDelta) {
	if (btns[0]) {
		this.rotation[1] -= moveDelta[0] * 0.2;
		this.rotation[0] -= moveDelta[1] * 0.2;
	}
	if (btns[1]) {
		this.target[0] +=  moveDelta[0] * this.distance * 0.01;
		this.target[2] +=  moveDelta[1] * this.distance * 0.01;
	}
	gr_instance.requestRedraw();
}

function grController(viewDomObject) {
	var canvas = $(view).find('canvas');
	var contextNames = ["webgl", "experimental-webgl", "webkit-3d", "moz-webgl"];
    for (var i in contextNames) {
        try {
            gl = canvas[0].getContext(contextNames[i]);
        } catch(e) {};
        if (gl) break;
    }
    if (!gl) {
        console.error("Could not initialise WebGL");
		alert("Could not initialize WebGL");
        return;
    }
	
	this.renderChain = undefined;
	this.models = [];
	this.helpers = [
		grHelper_Pivot(),
	];
	this.camera = new grCameraTargeted();
	this.rectX = gl.canvas.width;
	this.rectY = gl.canvas.height;
	this.mouseDown = [false, false];
	this.mouseLastPos = [0, 0];
	this.emptyTexture = new grTexture("/static/emptytexture.png", true);
	
	window.requestAnimationFrame = (function() {
		return window.requestAnimationFrame ||
			window.webkitRequestAnimationFrame ||
			window.mozRequestAnimationFrame ||
			window.oRequestAnimationFrame ||
			window.msRequestAnimationFrame;
	})();
	
	canvas.mousewheel(function(event) {
		gr_instance.camera.onMouseWheel(event.deltaY * event.deltaFactor);
		event.stopPropagation();
        event.preventDefault();
    }).mousedown(function(event) {
		if (event.button < 2) {
			gr_instance.mouseDown[event.button] = true;
			gr_instance.mouseLastPos = [event.clientX, event.clientY];
			event.stopPropagation();
			event.preventDefault();
		}
    });
    
    $(window).mouseup(function(event) {
		if (event.button < 2) {
			gr_instance.mouseDown[event.button] = false;
			event.stopPropagation();
            event.preventDefault();
		}
    }).mousemove(function(event) {
        if (gr_instance.mouseDown.reduce((a,b)=>(a|b),false)) {
			var lastPos = gr_instance.mouseLastPos;
			var posDiff = [lastPos[0] - event.clientX, lastPos[1] - event.clientY];
			gr_instance.camera.onMouseMove(gr_instance.mouseDown, posDiff);
			gr_instance.mouseLastPos = [event.clientX, event.clientY];
            event.stopPropagation();
            event.preventDefault();
        }
    });

	$(window).resize(this._onResize);
}

function gwInitRenderer(viewDomObject) {
	if (gr_instance) {
		console.warn("grController already created", gr_instance);
		return;
	}
	gr_instance = new grController(viewDomObject);
	gr_instance.changeRenderChain(grRenderChain_SkinnedTextured);
	gr_instance.onResize();
	gr_instance.requestRedraw();
}

grController.prototype.changeRenderChain = function(chainType) {
	if (this.renderChain)
		this.renderChain.free();
	this.renderChain = new chainType(this);
}

grController.prototype.onResize = function() {
	gl.canvas.width = this.rectX = gl.canvas.clientWidth;
	gl.canvas.height = this.rectY = gl.canvas.clientHeight;
	gl.viewport(0, 0, gl.drawingBufferWidth, gl.drawingBufferHeight);
}
grController.prototype._onResize = function() { gr_instance.onResize(); gr_instance.requestRedraw(); }

grController.prototype.render = function() {
	var glError = gl.getError();
	if (glError !== gl.NONE) {
		console.warn("pre-draw gl.getError()", glError);
	}
	
	this.renderChain.render(this);
	
	if ((glError = gl.getError()) !== gl.NONE) {
		console.warn("post-draw gl.getError()", glError);
	}
}

grController.prototype.requestRedraw = function() {
	requestAnimationFrame(function() { gr_instance.render(); });
}

grController.prototype.destroyModels = function() {
	for (var i in this.models) {
		this.models[i].free(this);
	}
	this.models.length = 0;
}

grController.prototype.downloadFile = function(link, async) {
	var txt;
	$.ajax({
		url: link,
		async: !!async,
		success: function(data) {
			txt = data;
		}
	});
	return txt;
}

grController.prototype.createShader = function(text, isFragment) {
    var shader = gl.createShader(isFragment ? gl.FRAGMENT_SHADER : gl.VERTEX_SHADER);
    gl.shaderSource(shader, text);
    gl.compileShader(shader);
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
        console.warn(text, gl.getShaderInfoLog(shader));
        return;
    }
    return shader;
}

grController.prototype.createProgram = function(vertexShader, fragmentShader) {
    var shaderProgram = gl.createProgram();
    gl.attachShader(shaderProgram, vertexShader);
    gl.attachShader(shaderProgram, fragmentShader);
    gl.linkProgram(shaderProgram);

    if (!gl.getProgramParameter(shaderProgram, gl.LINK_STATUS)) {
        console.warn("Could not initialise shaders");
		return;
    }
	return shaderProgram;
}

grController.prototype.downloadShader = function(link, isFragment) {
	var text = this.downloadFile(link, false);
	if (text)
		return this.createShader(text, isFragment);
}

