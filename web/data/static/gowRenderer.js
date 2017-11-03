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
    this.bufferIndexType = (indexArray.length > 254) ? gl.UNSIGNED_SHORT : gl.UNSIGNED_BYTE;
    gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, (this.bufferIndexType === gl.UNSIGNED_SHORT) ? (new Uint16Array(indexArray)) : (new Uint8Array(indexArray)), gl.STATIC_DRAW);

    this.bufferBlendColor = undefined;
    this.bufferUV = undefined;
    this.bufferNormals = undefined;
    this.bufferJointIds = undefined;
    this.bufferJointIds2 = undefined;
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
            break;
        }
    }

    gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferBlendColor);
    gl.bufferData(gl.ARRAY_BUFFER, new Uint8Array(data), gl.STATIC_DRAW);
}

grMesh.prototype.setUVs = function(data) {
    if (!this.bufferUV) {
        this.bufferUV = gl.createBuffer();
    }
    gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferUV);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data), gl.STATIC_DRAW);
}

grMesh.prototype.setMaterialID = function(materialIndex) {
    this.materialIndex = materialIndex;
}

grMesh.prototype.setNormals = function(data) {
    if (!this.bufferNormals) {
        this.bufferNormals = gl.createBuffer();
    }
    gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferNormals);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data), gl.STATIC_DRAW);
}

grMesh.prototype.setJointIds = function(jointMapping, jointIds, jointIds2) {
    this.jointMapping = jointMapping;
    if (!this.bufferJointIds) {
        this.bufferJointIds = gl.createBuffer();
    }
    gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferJointIds);
    gl.bufferData(gl.ARRAY_BUFFER, new Uint8Array(jointIds), gl.STATIC_DRAW);

    if (jointIds2 != undefined) {
        if (!this.bufferJointIds2) {
            this.bufferJointIds2 = gl.createBuffer();
        }
        gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferJointIds2);
        gl.bufferData(gl.ARRAY_BUFFER, new Uint8Array(jointIds2), gl.STATIC_DRAW);
    }
}

grMesh.prototype.free = function() {
    if (this.bufferVertex) gl.deleteBuffer(this.bufferVertex);
    if (this.bufferIndex) gl.deleteBuffer(this.bufferIndex);
    if (this.bufferBlendColor) gl.deleteBuffer(this.bufferBlendColor);
    if (this.bufferUV) gl.deleteBuffer(this.bufferUV);
    if (this.bufferNormals) gl.deleteBuffer(this.bufferNormals);
    if (this.bufferJointIds) gl.deleteBuffer(this.bufferJointIds);
}

grMesh.prototype.claim = function() {
    this.refs++;
}
grMesh.prototype.unclaim = function() {
    if (--this.refs === 0) this.free();
}

function grModel() {
    this.visible = true;
    this.meshes = [];
    this.materials = [];
    this.skeleton = undefined;
    this.matrix = mat4.create();
    this.type = undefined;
    this.exclusiveMeshes = undefined;
}

grModel.prototype.showExclusiveMeshes = function(meshes) {
    this.exclusiveMeshes = meshes;
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

function grModel__mshFromSklt(sklt, key = "OurJointToIdleMat") {
    var vrtxs = [];
    var indxs = [];
    var clrs = [];

    vrtxs.length = 3 * sklt.length;
    clrs.length = 4 * sklt.length;

    for (var i in sklt) {
        var joint = sklt[i];
        vrtxs[i * 3 + 0] = joint[key][12];
        vrtxs[i * 3 + 1] = joint[key][13];
        vrtxs[i * 3 + 2] = joint[key][14];
        clrs[i * 4 + 0] = (i % 8) * 15;
        clrs[i * 4 + 1] = ((i / 8) % 8) * 15;
        clrs[i * 4 + 2] = ((i / 64) % 8) * 15;
        clrs[i * 4 + 3] = 127;

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
    //this.addMesh(grModel__mshFromSklt(sklt));
    this.skeleton = [];

    for (var i in sklt) {
        this.skeleton.push(new Float32Array(sklt[i].RenderMat));
    }
}

grModel.prototype.free = function() {
    for (var i in this.meshes) {
        this.meshes[i].unclaim();
    }
    for (var i in this.materials) {
        this.materials[i].unclaim();
    }
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
    };
}
grTexture.prototype.markAsFontTexture = function() {
	gl.bindTexture(gl.TEXTURE_2D, this.txr);
	gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
	gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
	gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
	gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
}
grTexture.prototype.free = function() {
    if (this.txr) gl.deleteTexture(this.txr);
}
grTexture.prototype.get = function() {
    return this.loaded ? this.txr : gr_instance.emptyTexture.get();
}

function grMaterial() {
    this.refs = 0;
    this.color = [1, 1, 1, 1];
    this.textureDiffuse = undefined;
    this.hasAlpha = false;
    this.method = 0;
}

grMaterial.prototype.setColor = function(color) {
    this.color = color;
}

grMaterial.prototype.setDiffuse = function(txtr) {
    this.textureDiffuse = txtr;
}

grMaterial.prototype.setHasAlphaAttribute = function(alpha = true) {
    this.hasAlpha = alpha;
}

grMaterial.prototype.setMethodNormal = function() {
    this.method = 0;
}
grMaterial.prototype.setMethodAdditive = function() {
    this.method = 1;
}
grMaterial.prototype.setMethodSubstract = function() {
    this.method = 2;
}
grMaterial.prototype.setMethodUnknown = function() {
    this.method = 3;
}
grMaterial.prototype.claim = function() {
    this.refs++;
}
grMaterial.prototype.unclaim = function() {
    if (--this.refs == 0) {
        this.free();
    }
}

grMaterial.prototype.free = function() {
    if (this.textureDiffuse) {
        this.textureDiffuse.free();
    }
}

function grTextMesh(text=undefined, x=0, y=0, z=0, is3d=false) {
	this.refs = 0;
	this.position = [x, y, z];
	this.is3d = is3d;
	this.color = [1, 1, 1, 1];
	this.indexesCount = 0;
	this.bufferIndexType = undefined;
	this.bufferVertex = gl.createBuffer();
	this.bufferIndex = gl.createBuffer();
	this.bufferUV = gl.createBuffer();
	this.setText(text);
}
grTextMesh.prototype.set3d = function(is3d) {
	this.is3d = is3d;
}
grTextMesh.prototype.setColor = function(r, g, b, a) {
	this.color = [r, g, b, a];
}
grTextMesh.prototype.setPosition = function(x, y, z) {
    this.position = [x, y, z];
}
grTextMesh.prototype.setText = function(text, charBoxSize=9) {
	if (text == undefined) {
		this.bufferIndexType = undefined;
		
	}
	var vrtxs = [];
	var uvs = [];
	var indxs = [];
	var chsz = 1/16; // char size in texture units
	var x = 0;
	var y = 0;
		
	for (var i = 0; i < text.length; i++) {
		var char = text.charCodeAt(i);
		if (char > 0xff) {
			char = 182; // ¶
		}

		vrtxs.push(x);             vrtxs.push(y);
		vrtxs.push(x+charBoxSize); vrtxs.push(y);
		vrtxs.push(x);             vrtxs.push(y+charBoxSize);
		vrtxs.push(x+charBoxSize); vrtxs.push(y+charBoxSize);
		x += charBoxSize;
	
		var tx = Math.floor(char % 16) / 16;
		var ty = Math.floor(char / 16) / 16;
		uvs.push(tx);      uvs.push(ty+chsz);
		uvs.push(tx+chsz); uvs.push(ty+chsz);
		uvs.push(tx);      uvs.push(ty);
		uvs.push(tx+chsz); uvs.push(ty);
		
		var idx = i * 4;
		indxs.push(idx);   indxs.push(idx+1); indxs.push(idx+2);
		indxs.push(idx+1); indxs.push(idx+2); indxs.push(idx+3);
	}
	
    gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferVertex);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(vrtxs), gl.STATIC_DRAW);

	gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferUV);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(uvs), gl.STATIC_DRAW);

	this.bufferIndexType = (vrtxs.length > 254) ? gl.UNSIGNED_SHORT : gl.UNSIGNED_BYTE;
	this.indexesCount = indxs.length;
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, this.bufferIndex);
    gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, (this.bufferIndexType === gl.UNSIGNED_SHORT) ? (new Uint16Array(indxs)) : (new Uint8Array(indxs)), gl.STATIC_DRAW);
	
}
grTextMesh.prototype.claim = function() {
    this.refs++;
}
grTextMesh.prototype.unclaim = function() {
    if (--this.refs == 0) {
        this.free();
    }
}
grTextMesh.prototype.free = function() {
	gl.deleteBuffer(this.bufferVertex);
	gl.deleteBuffer(this.bufferIndex);
	gl.deleteBuffer(this.bufferUV);
}

function grCameraTargeted() {
    this.farPlane = 50000.0;
    this.nearPlane = 1.0;
    this.fow = 55.0;
	this.target = [0, 0, 0];
    this.distance = 100.0;
    this.rotation = [15.0, 45.0 * 3, 0];
}

grCameraTargeted.prototype.getProjectionMatrix = function() {
    return mat4.perspective(mat4.create(), glMatrix.toRadian(this.fow), gr_instance.rectX / gr_instance.rectY, this.nearPlane, this.farPlane);
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
        this.rotation[1] += moveDelta[0] * 0.2;
        this.rotation[0] += moveDelta[1] * 0.2;
    }
    if (btns[1]) {
        this.target[0] += moveDelta[0] * this.distance * 0.01;
        this.target[2] += moveDelta[1] * this.distance * 0.01;
    }
    gr_instance.requestRedraw();
}
function grController(viewDomObject) {
    var canvas = $(view).find('canvas');
    var contextNames = ["webgl", "experimental-webgl", "webkit-3d", "moz-webgl"];
    for (var i in contextNames) {
        try {
            gl = canvas[0].getContext(contextNames[i]);
        } catch (e) {};
        if (gl) break;
    }
    if (!gl) {
        console.error("Could not initialise WebGL");
        alert("Could not initialize WebGL");
        return;
    }

    this.renderChain = undefined;
    this.models = [];
	this.texts = [];
    this.helpers = [
        grHelper_Pivot(),
    ];
    this.camera = new grCameraTargeted();
	this.orthoMatrix = mat4.create();
    this.rectX = gl.canvas.width;
    this.rectY = gl.canvas.height;
    this.mouseDown = [false, false];
    this.emptyTexture = new grTexture("/static/emptytexture.png", true);
	this.fontTexture = new grTexture("/static/font2.png", true);
	this.fontTexture.markAsFontTexture();

	canvas.mousewheel(function(event) {
        gr_instance.camera.onMouseWheel(event.deltaY * event.deltaFactor);
        event.stopPropagation();
        event.preventDefault();
	}).mousedown(function(event) {
        if (event.button < 2) {
			gr_instance.mouseDown[event.button] = true;
            event.stopPropagation();
            event.preventDefault();
			if (event.button === 0) {
				this.requestPointerLock();
			}
        }
    })
	
	canvas[0].addEventListener('webglcontextlost', function(e) {
        console.log("webgl context lost", e);
    });
	canvas[0].addEventListener('webglcontextrestored', function(e) {
        console.log("webgl context restored", e);
    });

    $(document).mouseup(function(event) {
		 if (event.button < 2) {
			gr_instance.mouseDown[event.button] = false;
            event.stopPropagation();
            event.preventDefault();
			if (event.button === 0) {
				document.exitPointerLock();
			}
        }
    }).mousemove(function(event) {
        if (gr_instance.mouseDown.reduce((a, b) => (a | b), false)) {
			var posDiff = [event.originalEvent.movementX, event.originalEvent.movementY];
            gr_instance.camera.onMouseMove(gr_instance.mouseDown, posDiff);
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
	this.orthoMatrix = mat4.ortho(this.orthoMatrix, 0, this.rectX, 0, this.rectY, -1, 1);
}
grController.prototype._onResize = function() {
    gr_instance.onResize();
    gr_instance.requestRedraw();
}

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
    requestAnimationFrame(function() {
        gr_instance.render();
    });
}

grController.prototype.destroyModels = function() {
    for (var i in this.models) {
        this.models[i].free(this);
    }
    this.models.length = 0;
}

grController.prototype.destroyTexts = function() {
    for (var i in this.texts) {
        this.texts[i].free(this);
    }
    this.texts.length = 0;
}

grController.prototype.cleanup = function() {
    this.destroyTexts();
	this.destroyModels();
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
