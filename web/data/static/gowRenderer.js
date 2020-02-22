let gr_instance;
let gl;

function grMesh(vertexArray, indexArray, primitive) {
    this.refs = 0;
    this.indexesCount = indexArray.length;
    this.primitive = (!!primitive) ? primitive : gl.TRIANGLES;
    this.isDepthTested = true;
    this.hasAlpha = false;
    this.isVisible = true;
    this.ps3static = false;
    this.useBindToJoin = false;
    this.layer = undefined;
    this.mask = 0;

    // construct array of unique indexes
    this.usedIndexes = [];
    for (let i in indexArray) {
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
    this.bufferJointIds1 = undefined;
    this.bufferJointIds2 = undefined;
    this.bufferWeights = undefined;
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

    for (let i in this.usedIndexes) {
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

grMesh.prototype.setJointIds = function(jointMapping, jointIds1, jointIds2, weights) {
    this.jointMapping = jointMapping;
    if (!this.bufferJointIds1) {
        this.bufferJointIds1 = gl.createBuffer();
    }
    gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferJointIds1);
    gl.bufferData(gl.ARRAY_BUFFER, new Uint8Array(jointIds1), gl.STATIC_DRAW);

    if (jointIds2 !== undefined) {
        if (!this.bufferJointIds2) {
            this.bufferJointIds2 = gl.createBuffer();
        }
        gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferJointIds2);
        gl.bufferData(gl.ARRAY_BUFFER, new Uint8Array(jointIds2), gl.STATIC_DRAW);
    }
    
    if (weights !== undefined) {
    	if (!this.bufferWeights) {
            this.bufferWeights = gl.createBuffer();
        }
        gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferWeights);
        gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(weights), gl.STATIC_DRAW);
    }
}

grMesh.prototype.setps3static = function(yes) {
    this.ps3static = !!yes;
}

grMesh.prototype.setUseBindToJoin = function(yes) {
    this.useBindToJoin = yes;
}

grMesh.prototype.setLayer = function(layer) {
    this.layer = layer;
}

grMesh.prototype.setMaskBit = function(bitIndex) {
    this.mask |= 1 << bitIndex;
}

grMesh.prototype.free = function() {
    if (this.bufferVertex) gl.deleteBuffer(this.bufferVertex);
    if (this.bufferIndex) gl.deleteBuffer(this.bufferIndex);
    if (this.bufferBlendColor) gl.deleteBuffer(this.bufferBlendColor);
    if (this.bufferUV) gl.deleteBuffer(this.bufferUV);
    if (this.bufferNormals) gl.deleteBuffer(this.bufferNormals);
    if (this.bufferJointIds1) gl.deleteBuffer(this.bufferJointIds1);
    if (this.bufferJointIds2) gl.deleteBuffer(this.bufferJointIds2);
    if (this.bufferWeights) gl.deleteBuffer(this.bufferWeights);
    gr_instance.flushScene();
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
    this.matrices = undefined; // non multiplied by bind pose matrix
    this.matricesInverted = undefined; // multiplied by bind pose matrix
    this.matrix = mat4.create();
    this.mask = 0;
    this.type = undefined;
    this.animation = undefined;
    this.exclusiveMeshes = undefined;
    this.attachedTexts = [];
}

grModel.prototype.showExclusiveMeshes = function(meshes) {
    this.exclusiveMeshes = meshes;
    gr_instance.flushScene();
}

grModel.prototype.setType = function(type) {
    this.type = type;
}

grModel.prototype.addMesh = function(mesh) {
    mesh.claim();
    this.meshes.push(mesh);
}

grModel.prototype.addText = function(textMesh, jointId) {
    if (!textMesh.is3d) {
        console.error("Can't use 2d text mesh", textMesh, jointId, this);
        return;
    }
    textMesh.ownerModel = this;
    textMesh.jointId = jointId;
    textMesh.claim();
    this.attachedTexts.push(textMesh);
}

grModel.prototype.addMaterial = function(material) {
    material.claim();
    this.materials.push(material);
}

grModel.prototype.setMaskBit = function(bitIndex) {
    this.mask |= 1 << bitIndex;
}

grModel.prototype._labelsFromSklt = function(sklt) {
    for (let i in sklt) {
        let currentJoint = sklt[i];
        let jointText = new grTextMesh(i, 0, 0, 0, true, 10);
        this.addText(jointText, parseInt(i));
        jointText.setOffset(-0.5, -0.5);
        jointText.setColor(1.0, 1.0, 1.0, 0.3);
        jointText.setMaskBit(1);
        this.addText(jointText, parseInt(i));
        gr_instance.texts.push(jointText);
    }
}

function grModel__mshFromSklt(sklt, useLabels = true) {
    let meshes = [];
    let vrtxs = [];
    let indxs = [];
    let clrs = [];
    let joints = [];
    let jointsMap = [];

    for (let i in sklt) {
        let currentJoint = sklt[i];
        if (currentJoint.Parent < 0) {
            continue;
        }

        for (let joint of [currentJoint, sklt[currentJoint.Parent]]) {
            indxs.push(indxs.length);
            vrtxs.push(0);
            vrtxs.push(0);
            vrtxs.push(0);
            clrs.push((i % 8) * 15);
            clrs.push(((i / 8) % 8) * 15);
            clrs.push(((i / 64) % 8) * 15);
            clrs.push(127);
            let idx = jointsMap.indexOf(joint.Id);
            if (idx < 0) {
                joints.push(jointsMap.length);
                jointsMap.push(joint.Id);
            } else {
                joints.push(idx);
            }
        }

        if (jointsMap.length > 10 || (i == sklt.length - 1 && indxs.length > 0)) {
            let sklMesh = new grMesh(vrtxs, indxs, gl.LINES);
            sklMesh.setDepthTest(false);
            sklMesh.setBlendColors(clrs);
            sklMesh.setJointIds(jointsMap, joints, joints);
            sklMesh.setMaskBit(2);
            meshes.push(sklMesh);
            vrtxs = [];
            indxs = [];
            clrs = [];
            joints = [];
            jointsMap = [];
        }
    }

    return meshes;
}

grModel.prototype.loadSkeleton = function(sklt) {
    for (let m of grModel__mshFromSklt(sklt)) {
        this.addMesh(m);
    }
    this.matrices = [];
    this.matricesInverted = [];
    for (let i in sklt) {
        this.matrices.push(new Float32Array(sklt[i].OurJointToIdleMat));
        this.matricesInverted.push(new Float32Array(sklt[i].RenderMat));
    }
    this._labelsFromSklt(sklt);
}

grModel.prototype.setJointMatrix = function(nodeid, matrix, matrixInverted) {
    this.matrices[nodeid] = matrix;
    this.matricesInverted[nodeid] = matrixInverted;
}

grModel.prototype.free = function() {
    for (let i in this.attachedTexts) {
        this.attachedTexts[i].unclaim();

    }
    for (let i in this.meshes) {
        this.meshes[i].unclaim();
    }
    for (let i in this.materials) {
        this.materials[i].unclaim();
    }

    if (this.animation != undefined) {
        ga_instance.freeAnimation(this.animation);
    }
    gr_instance.flushScene();
}

function grTexture__handleLoading(img, txr) {
    if (!txr.loaded) {
        txr.txr = gl.createTexture();
    }
    gl.bindTexture(gl.TEXTURE_2D, txr.txr);
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img);
    gl.generateMipmap(gl.TEXTURE_2D);
    txr.loaded = true;

    gr_instance.requestRedraw();

    txr.applyTexParameters();
}

function grTexture(src, wait = false) {
    this.loaded = wait;
    this.txr = undefined;
    this.isFontTexture = false;

    if (wait) {
        this.txr = gl.createTexture();
        // pink placeholder
        gl.bindTexture(gl.TEXTURE_2D, this.txr);
        gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, new Uint8Array([255, 0, 255, 128]));
        gl.bindTexture(gl.TEXTURE_2D, null);
    }

    let img = new Image();
    img.src = src;
    let _this = this;
    img.onload = function() {
        grTexture__handleLoading(img, _this);
    };
}
grTexture.prototype.applyTexParameters = function() {
    if (!this.loaded) {
        return;
    }
    gl.bindTexture(gl.TEXTURE_2D, this.txr);
    if (this.isFontTexture) {
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
    } else {
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
    }
    gl.bindTexture(gl.TEXTURE_2D, null);
}
grTexture.prototype.markAsFontTexture = function() {
    this.isFontTexture = true;
    this.applyTexParameters();
}
grTexture.prototype.free = function() {
    if (this.txr) gl.deleteTexture(this.txr);
}
grTexture.prototype.get = function() {
    return this.loaded ? this.txr : gr_instance.emptyTexture.get();
}

function grMaterialLayer() {
    this.color = [1, 1, 1, 1];
    this.uvoffset = [0, 0];
    this.textureIndex = 0;
    this.textures = undefined;
    this.method = 0;
    this.hasAlpha = false;
}

grMaterialLayer.prototype.setColor = function(color) {
    this.color = color;
}

grMaterialLayer.prototype.setHasAlphaAttribute = function(alpha = true) {
    this.hasAlpha = alpha;
}

grMaterialLayer.prototype.setTextures = function(txtrs) {
    this.textures = txtrs;
}

grMaterialLayer.prototype.setTextureIndex = function(index) {
    this.textureIndex = index % this.textures.length;
    if (index > this.textures.length) {
        console.warn("trying to set texture index > textures count");
    }
}

grMaterialLayer.prototype.setMethodNormal = function() {
    this.method = 0;
}
grMaterialLayer.prototype.setMethodAdditive = function() {
    this.method = 1;
}
grMaterialLayer.prototype.setMethodSubstract = function() {
    this.method = 2;
}
grMaterialLayer.prototype.setMethodUnknown = function() {
    this.method = 3;
}

grMaterialLayer.prototype.free = function() {
    if (this.textures) {
        for (let i in this.textures) {
            this.textures[i].free();
        }
    }
}

function grMaterial() {
    this.refs = 0;
    this.color = [1, 1, 1, 1];
    this.layers = [];
    this.anims = [];
}
grMaterial.prototype.addLayer = function(layer) {
    this.layers.push(layer);
}
grMaterial.prototype.setColor = function(color) {
    this.color = color;
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
    for (let i in this.layers) {
        this.layers[i].free();
    }
    for (let i in this.anims) {
        ga_instance.freeAnimation(this.anims[i]);
    }
    gr_instance.flushScene();
}

function grTextMesh(text = undefined, x = 0, y = 0, z = 0, is3d = false, charSize = 9.0) {
    this.refs = 0;
    this.position = [x, y, z];
    this.offset = [0.0, 0.0];
    this.textLength = 0;
    this.is3d = is3d;
    this.charSize = charSize;
    this.color = [1, 1, 1, 1];
    this.indexesCount = 0;
    this.bufferIndexType = undefined;
    this.bufferVertex = gl.createBuffer();
    this.bufferIndex = gl.createBuffer();
    this.bufferUV = gl.createBuffer();
    this.setText(text);
    this.align = [0.0, 0.0];
    this.ownerModel = undefined;
    this.jointId = undefined;
    this.mask = 0;
}
grTextMesh.prototype.set3d = function(is3d) {
    this.is3d = is3d;
}
grTextMesh.prototype.setColor = function(r, g, b, a = 1.0) {
    this.color = [r, g, b, a];
}
grTextMesh.prototype.setOffset = function(x, y) {
    this.offset = [x, y];
}
grTextMesh.prototype.getGlobalOffset = function(x, y, z) {
    return [this.offset[0] * this.charSize, this.offset[1] * this.charSize];
}
grTextMesh.prototype.setPosition = function(x, y, z) {
    this.position = [x, y, z];
}
grTextMesh.prototype.setText = function(text) {
    if (text == undefined) {
        this.bufferIndexType = undefined;
    }

    this.textLength = text.length;
    let vrtxs = [];
    let uvs = [];
    let indxs = [];
    let charBoxSize = this.charSize;
    let chsz = 1 / 16; // char size in texture units
    let x = 0;
    let y = 0;

    for (let i = 0; i < text.length; i++) {
        let char = text.charCodeAt(i);
        if (char > 0xff) {
            char = 182; // ¶
        }

        vrtxs.push(x);
        vrtxs.push(y);
        vrtxs.push(x + charBoxSize);
        vrtxs.push(y);
        vrtxs.push(x);
        vrtxs.push(y + charBoxSize);
        vrtxs.push(x + charBoxSize);
        vrtxs.push(y + charBoxSize);
        x += charBoxSize;

        let tx = Math.floor(char % 16) / 16;
        let ty = Math.floor(char / 16) / 16;
        uvs.push(tx);
        uvs.push(ty + chsz);
        uvs.push(tx + chsz);
        uvs.push(ty + chsz);
        uvs.push(tx);
        uvs.push(ty);
        uvs.push(tx + chsz);
        uvs.push(ty);

        let idx = i * 4;
        indxs.push(idx);
        indxs.push(idx + 1);
        indxs.push(idx + 2);
        indxs.push(idx + 1);
        indxs.push(idx + 2);
        indxs.push(idx + 3);
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

grTextMesh.prototype.setMaskBit = function(bitIndex) {
    this.mask |= 1 << bitIndex;
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
    gr_instance.flushScene();
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
    let viewMatrix = mat4.create();
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
    let resizeDelta = Math.sqrt(this.distance) * delta * 0.01;
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

function grCameraInterface() {
    grCameraTargeted.call(this);
    this.rotation = [0, 0, 0];
    this.target = [0, 0, 0];
    this.distance = 100;
}
grCameraInterface.prototype = Object.create(grCameraTargeted.prototype);
grCameraInterface.prototype.constructor = grCameraInterface;
grCameraInterface.prototype.getProjectionMatrix = function() {
    let w = gr_instance.rectX * this.distance * 0.004;
    let h = gr_instance.rectY * this.distance * 0.004;
    return mat4.ortho(mat4.create(), -w, w, h, -h, this.nearPlane, this.farPlane);
}
grCameraInterface.prototype.onMouseMove = function(btns, moveDelta) {
    this.target[0] += moveDelta[0] * this.distance * 0.01;
    this.target[1] += moveDelta[1] * this.distance * 0.01;
    gr_instance.requestRedraw();
}

function grController(viewDomObject) {
    let canvas = $(viewDomObject).find('canvas');
    let contextNames = ["webgl", "experimental-webgl", "webkit-3d", "moz-webgl"];
    for (let i in contextNames) {
        try {
            gl = canvas[0].getContext(contextNames[i], {
                antialias: false
            });
        } catch (e) {};
        if (gl) break;
    }
    if (!gl) {
        console.error("Could not initialise WebGL");
        alert("Could not initialize WebGL");
        return;
    }

    this.requireRedraw = false;
    this.renderChain = undefined;
    this.models = [];
    this.texts = [];
    this.helpers = [
        grHelper_Pivot(),
    ];
    this.cameraModels = new grCameraTargeted();
    this.cameraInterface = new grCameraInterface();
    this.orthoMatrix = mat4.create();
    this.rectX = gl.canvas.width;
    this.rectY = gl.canvas.height;
    this.mouseDown = [false, false];
    this.emptyTexture = new grTexture("/static/emptytexture.png", true);
    this.fontTexture = new grTexture("/static/font2.png", true);
    this.fontTexture.markAsFontTexture();
    this.filterMask = 0;
    this.cull = false;

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
            let posDiff = [event.originalEvent.movementX, event.originalEvent.movementY];
            gr_instance.camera.onMouseMove(gr_instance.mouseDown, posDiff);
            event.stopPropagation();
            event.preventDefault();
        }
    });

    this.setInterfaceCameraMode();
    $(window).resize(this._onResize);
}
grController.prototype.setFilterMask = function(mask) {
    if (this.filterMask == mask) {
        return;
    }
    this.filterMask = mask;
    this.requestRedraw();
}
grController.prototype.flushScene = function() {
    this.renderChain.flushScene(this);
}
grController.prototype.setInterfaceCameraMode = function(is3d) {
    this.camera = (!!is3d) ? this.cameraInterface : this.cameraModels;
}
grController.prototype.changeRenderChain = function(chainType) {
    if (this.renderChain)
        this.renderChain.free();
    this.renderChain = new chainType(this);
    this.flushScene();
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
    if (this.requireRedraw) {
        let glError = gl.getError();
        if (glError !== gl.NONE) {
            console.warn("pre-draw gl.getError()", glError);
        }

        this.renderChain.render(this);

        if ((glError = gl.getError()) !== gl.NONE) {
            console.warn("post-draw gl.getError()", glError);
        }
        this.requireRedraw = false;
    }
}
grController.prototype.requestRedraw = function() {
    this.requireRedraw = true;
}
grController.prototype.animationFrameCallback = function() {
    if (typeof ga_instance !== "undefined" && ga_instance) {
        ga_instance.update();
    }
    this.render();
    this.initFrameCallback();
}
grController.prototype.initFrameCallback = function() {
    requestAnimationFrame(function() {
        gr_instance.animationFrameCallback();
    });
}
grController.prototype.destroyModels = function() {
    for (let i in this.models) {
        this.models[i].free(this);
    }
    this.models.length = 0;
    this.flushScene();
}
grController.prototype.destroyTexts = function() {
    for (let i in this.texts) {
        this.texts[i].free(this);
    }
    this.texts.length = 0;
    this.flushScene();
}
grController.prototype.cleanup = function() {
    this.destroyTexts();
    this.destroyModels();
}
grController.prototype.downloadFile = function(link, async) {
    let txt;
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
    let shader = gl.createShader(isFragment ? gl.FRAGMENT_SHADER : gl.VERTEX_SHADER);
    gl.shaderSource(shader, text);
    gl.compileShader(shader);
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
        console.warn(text, gl.getShaderInfoLog(shader));
        return;
    }
    return shader;
}
grController.prototype.createProgram = function(vertexShader, fragmentShader) {
    let shaderProgram = gl.createProgram();
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
    let text = this.downloadFile(link, false);
    if (text)
        return this.createShader(text, isFragment);
}

function gwInitRenderer(viewDomObject) {
    if (gr_instance) {
        console.warn("grController already created", gr_instance);
        return;
    }
    gr_instance = new grController(viewDomObject);
    gr_instance.changeRenderChain(grRenderChain_SkinnedTextured);
    gr_instance.onResize();
    gr_instance.initFrameCallback();
}