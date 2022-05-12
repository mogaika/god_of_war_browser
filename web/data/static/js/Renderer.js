let gr_instance;
let gl;


class RenderModel extends Claimable {
    constructor() {
        super();
        this.visible = true;
        this.meshes = new ClaimedPool();
        this.materials = new ClaimedPool();
        this.mask = 0;
        this.type = undefined;
        this.exclusiveMeshes = undefined;
    }

    showExclusiveMeshes(meshes) {
        this.exclusiveMeshes = meshes;
        gr_instance.flushScene();
    }

    setType(type) {
        this.type = type;
    }
    addMaterial(material) {
        this.materials.insert(material);
    }
    addMesh(mesh) {
        this.meshes.insert(mesh);
    }
    setMaskBit(bitIndex) {
        this.mask |= 1 << bitIndex;
    }

    _free() {
        this.meshes.removeAll();
        this.materials.removeAll();
        gr_instance.flushScene();
        super._free();
    }
}

class RenderMesh extends Claimable {
    constructor(vertexArray, indexArray, primitive = gl.TRIANGLES) {
        super();
        this.indexesCount = indexArray.length;
        this.primitive = primitive;
        this.isDepthTested = true;
        this.hasAlpha = false;
        this.isVisible = true;
        this.ps3static = false;
        this.useBindToJoin = false;
        this.layer = undefined;
        this.mask = 0;
        this.meta = {};

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
        this.bufferIndexType = (vertexArray.length > 254) ? gl.UNSIGNED_SHORT : gl.UNSIGNED_BYTE;
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

    setVisible(visible = true) {
        this.isVisible = !!visible;
    }

    setDepthTest(isDepthTested) {
        this.isDepthTested = isDepthTested;
    }

    setBlendColors(data) {
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

    setUVs(data) {
        if (!this.bufferUV) {
            this.bufferUV = gl.createBuffer();
        }
        gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferUV);
        gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data), gl.STATIC_DRAW);
    }

    setMaterialID(materialIndex) {
        this.materialIndex = materialIndex;
    }

    setNormals(data) {
        if (!this.bufferNormals) {
            this.bufferNormals = gl.createBuffer();
        }
        gl.bindBuffer(gl.ARRAY_BUFFER, this.bufferNormals);
        gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data), gl.STATIC_DRAW);
    }

    setJointIds(jointMapping, jointIds1, jointIds2, weights) {
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

    setps3static(yes = true) {
        this.ps3static = !!yes;
    }

    setUseBindToJoin(yes = true) {
        this.useBindToJoin = yes;
    }

    setLayer(layer) {
        this.layer = layer;
    }

    setMaskBit(bitIndex) {
        this.mask |= 1 << bitIndex;
    }

    _free() {
        if (this.bufferVertex) gl.deleteBuffer(this.bufferVertex);
        if (this.bufferIndex) gl.deleteBuffer(this.bufferIndex);
        if (this.bufferBlendColor) gl.deleteBuffer(this.bufferBlendColor);
        if (this.bufferUV) gl.deleteBuffer(this.bufferUV);
        if (this.bufferNormals) gl.deleteBuffer(this.bufferNormals);
        if (this.bufferJointIds1) gl.deleteBuffer(this.bufferJointIds1);
        if (this.bufferJointIds2) gl.deleteBuffer(this.bufferJointIds2);
        if (this.bufferWeights) gl.deleteBuffer(this.bufferWeights);
        gr_instance.flushScene();
        super._free();
    }
}

class RenderTexture extends Claimable {
    constructor(src, wait = false) {
        super();
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
        let texture = this;
        img.onload = function() {
            texture.onImageLoad(this);
        };
    }

    onImageLoad(img) {
        if (!this.loaded) {
            this.txr = gl.createTexture();
        }
        gl.bindTexture(gl.TEXTURE_2D, this.txr);
        gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img);
        gl.generateMipmap(gl.TEXTURE_2D);
        this.loaded = true;
        this.applyTexParameters();
        gr_instance.requestRedraw();
    }

    applyTexParameters() {
        if (!this.loaded) {
            return;
        }
        gl.bindTexture(gl.TEXTURE_2D, this.txr);
        if (gr_instance.glExtFilterAnisotropic) {
            let maxAnisotropy = gl.getParameter(gr_instance.glExtFilterAnisotropic.MAX_TEXTURE_MAX_ANISOTROPY_EXT);
            gl.texParameterf(gl.TEXTURE_2D, gr_instance.glExtFilterAnisotropic.TEXTURE_MAX_ANISOTROPY_EXT, maxAnisotropy);
        }
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

    markAsFontTexture() {
        this.isFontTexture = true;
        this.applyTexParameters();
    }

    _free() {
        if (this.txr) {
            gl.deleteTexture(this.txr);
        }
        super._free();
    }

    get glTexture() {
        return this.loaded ? this.txr : gr_instance.emptyTexture.glTexture;
    }
}

class RenderMaterialLayer extends Claimable {
    constructor() {
        super();
        this.color = [1, 1, 1, 1];
        this.uvoffset = [0, 0];
        this.textureIndex = 0;
        this.textures = new ClaimedPool();
        this.method = 0;
        this.hasAlpha = false;
    }

    setColor(color) {
        this.color = color;
    }
    setHasAlphaAttribute(alpha = true) {
        this.hasAlpha = alpha;
    }
    setTextures(textures) {
        this.textures.removeAll();
        for (const txr of textures) {
            this.textures.insert(txr);
        }
    }
    setTextureIndex(index) {
        if (index > this.textures.length) {
            console.warn("trying to set texture index > textures count");
        }
        this.textureIndex = index % this.textures.length;
    }
    setMethodNormal() {
        this.method = 0;
    }
    setMethodAdditive() {
        this.method = 1;
    }
    setMethodSubstract() {
        this.method = 2;
    }
    setMethodUnknown() {
        this.method = 3;
    }

    _free() {
        this.textures.removeAll();
        super._free();
    }
}

class RenderMaterial extends Claimable {
    constructor() {
        super();
        this.color = [1, 1, 1, 1];
        this.layers = new ClaimedPool();
        this.anims = [];
    }

    addLayer(layer) {
        this.layers.insert(layer);
    }
    setColor(color) {
        this.color = color;
    }
    _free() {
        this.layers.removeAll();
        for (let i in this.anims) {
            ga_instance.freeAnimation(this.anims[i]);
        }
        gr_instance.flushScene();
        super._free();
    }
}

class RenderTextMesh extends Claimable {
    constructor(text, is3d = false, charSize = 9.0) {
        super();
        this.offset = [0.0, 0.0];
        this.is3d = is3d;
        this.charSize = charSize;
        this.color = [1, 1, 1, 1];
        this.indexesCount = 0;
        this.bufferIndexType = undefined;
        this.bufferVertex = gl.createBuffer();
        this.bufferIndex = gl.createBuffer();
        this.bufferUV = gl.createBuffer();
        this.align = [0.0, 0.0];
        this.mask = 0;

        this.setText(text);
    }
    set3d(is3d = true) {
        this.is3d = is3d;
    }
    setColor(r, g, b, a = 1.0) {
        this.color = [r, g, b, a];
    }
    setOffset(x, y) {
        this.offset = [x, y];
    }
    getGlobalOffset(x, y, z) {
        return [this.offset[0] * this.charSize, this.offset[1] * this.charSize];
    }
    setPosition(x, y, z) {
        this.position = [x, y, z];
    }
    setText(text) {
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

    setMaskBit(bitIndex) {
        this.mask |= 1 << bitIndex;
    }

    _free() {
        gl.deleteBuffer(this.bufferVertex);
        gl.deleteBuffer(this.bufferIndex);
        gl.deleteBuffer(this.bufferUV);
        gr_instance.flushScene();
        super._free();
    }
}

class RenderCameraTargeted {
    constructor() {
        this.farPlane = 50000.0;
        this.nearPlane = 1.0;
        this.fow = 55.0;
        this.target = [0, 0, 0];
        this.distance = 100.0;
        this.rotation = [15.0, 45.0 * 3, 0];
    }
    setTarget(target) {
        this.target = target;
        gr_instance.requestRedraw();
    }
    getProjectionMatrix() {
        return mat4.perspective(mat4.create(), glMatrix.toRadian(this.fow), gr_instance.rectX / gr_instance.rectY, this.nearPlane, this.farPlane);
    }
    getViewMatrix() {
        let viewMatrix = mat4.create();
        mat4.translate(viewMatrix, viewMatrix, [0.0, 0.0, -this.distance]);
        mat4.rotate(viewMatrix, viewMatrix, glMatrix.toRadian(this.rotation[0]), [1, 0, 0]);
        mat4.rotate(viewMatrix, viewMatrix, glMatrix.toRadian(this.rotation[1]), [0, 1, 0]);
        mat4.rotate(viewMatrix, viewMatrix, glMatrix.toRadian(this.rotation[2]), [0, 0, 1]);
        return mat4.translate(viewMatrix, viewMatrix, [-this.target[0], -this.target[1], -this.target[2]]);
    }
    moveForward(speed) {
        let m = mat4.create();
        mat4.rotate(m, m, glMatrix.toRadian(this.rotation[0]), [1, 0, 0]);
        mat4.rotate(m, m, glMatrix.toRadian(this.rotation[1]), [0, 1, 0]);
        mat4.rotate(m, m, glMatrix.toRadian(this.rotation[2]), [0, 0, 1]);
        mat4.invert(m, m);
        let v = vec3.fromValues(0, 0, 1);

        vec3.scale(v, v, -speed);
        vec3.transformMat4(v, v, m);
        vec3.add(this.target, this.target, v);

        gr_instance.requestRedraw();
    }
    getProjViewMatrix() {
        return mat4.mul(mat4.create(), this.getProjectionMatrix(), this.getViewMatrix());
    }
    onMouseWheel(delta) {
        let resizeDelta = Math.sqrt(this.distance) * delta * 0.01;
        this.distance -= resizeDelta * 2;
        if (this.distance < 1.0)
            this.distance = 1.0;
        gr_instance.requestRedraw();
    }
    onMouseMove(btns, moveDelta) {
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
}

class RenderCameraInterface extends RenderCameraTargeted {
    constructor() {
        super();
        this.rotation = [0, 0, 0];
        this.target = [0, 0, 0];
        this.distance = 100;
    }
    moveForward(speed) {}
    getProjectionMatrix() {
        let w = gr_instance.rectX * this.distance * 0.004;
        let h = gr_instance.rectY * this.distance * 0.004;
        return mat4.ortho(mat4.create(), -w, w, h, -h, this.nearPlane, this.farPlane);
    }
    onMouseMove(btns, moveDelta) {
        this.target[0] += moveDelta[0] * this.distance * 0.01;
        this.target[1] += moveDelta[1] * this.distance * 0.01;
        gr_instance.requestRedraw();
    }
}

class RenderController extends ObjectTreeNode {
    constructor(viewDomObject) {
        super("root");

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
        gr_instance = this;

        this.frameChecker = 0;
        this.requireRedraw = false;
        this.renderChain = undefined;
        this.moving = false;
        this.moveSpeed = 500.0;
        this.persistentNodes = [
            new ObjectTreeNodeModel("pivot", RenderHelper.Pivot()),
        ];
        this.resetCamera();
        this.orthoMatrix = mat4.create();
        this.rectX = gl.canvas.width;
        this.rectY = gl.canvas.height;
        this.mouseDown = [false, false];
        this.emptyTexture = new RenderTexture("/static/images/emptytexture.png", true);
        this.fontTexture = new RenderTexture("/static/images/font2.png", true);
        this.fontTexture.markAsFontTexture();
        this.filterMask = 0;
        this.cull = false;
        this.glExtFilterAnisotropic = gl.getExtension('EXT_texture_filter_anisotropic');
        this.movingForward = false;
        this.movingBackwards = false;

        let eventWithShift = false;
        canvas.mousewheel(function(event) {
            gr_instance.camera.onMouseWheel(event.deltaY * event.deltaFactor);
            event.stopPropagation();
            event.preventDefault();
        }).mousedown(function(event) {
            if (event.button < 2) {
                if (event.button == 0 && event.shiftKey) {
                    eventWithShift = true;
                    event.button = 1;
                } else {
                    eventWithShift = false;
                }
                gr_instance.mouseDown[event.button] = true;
                event.stopPropagation();
                event.preventDefault();
                if (event.button === 0) {
                    this.requestPointerLock();
                }
            }
        })

        $(document).mouseup(function(event) {
            if (event.button < 2) {
                gr_instance.mouseDown[event.button] = false;
                if (event.button == 0 && eventWithShift) {
                    gr_instance.mouseDown[1] = false;
                }

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
        }).keydown(function(event) {
            if (event.key == "w") {
                if (gr_instance.mouseDown.reduce((a, b) => (a | b), false)) {
                    gr_instance.movingForward = true;
                    gr_instance.requestRedraw();
                }
            } else if (event.key == "s") {
                if (gr_instance.mouseDown.reduce((a, b) => (a | b), false)) {
                    gr_instance.movingBackwards = true;
                    gr_instance.requestRedraw();
                }
            }
        }).keyup(function(event) {
            if (event.key == "w") {
                gr_instance.movingForward = false;
            } else if (event.key == "s") {
                gr_instance.movingBackwards = false;
            }
        });

        canvas[0].addEventListener('webglcontextlost', function(e) {
            console.warn("WebGL context lost", e);
        });
        canvas[0].addEventListener('webglcontextrestored', function(e) {
            console.log("WebGL context restored", e);
        });

        canvas.resize(this._onResize);
    }

    resetCamera() {
        let is2d = this.camera != this.cameraModels;

        this.cameraModels = new RenderCameraTargeted();
        this.cameraInterface = new RenderCameraInterface();
        this.setInterfaceCameraMode(is2d);
        this.requestRedraw();
    }
    setFilterMask(mask) {
        if (this.filterMask == mask) {
            return;
        }
        this.filterMask = mask;
        this.requestRedraw();
    }
    flushScene() {
        this.renderChain.flushScene(this);
    }
    setInterfaceCameraMode(is2d) {
        this.camera = (!!is2d) ? this.cameraInterface : this.cameraModels;
    }
    changeRenderChain(chainType) {
        if (this.renderChain)
            this.renderChain.free();
        this.renderChain = new chainType(this);
        this.flushScene();
    }
    onResize() {
        gl.canvas.width = this.rectX = gl.canvas.clientWidth;
        gl.canvas.height = this.rectY = gl.canvas.clientHeight;
        gl.viewport(0, 0, gl.drawingBufferWidth, gl.drawingBufferHeight);
        this.orthoMatrix = mat4.ortho(this.orthoMatrix, 0, this.rectX, 0, this.rectY, -1, 1);
    }
    _onResize() {
        gr_instance.onResize();
        gr_instance.requestRedraw();
    }
    render(dt) {
        if (gl.canvas.clientWidth !== this.rectX || gl.canvas.clientHeight !== this.rectY) {
            this.onResize();
        }
        if (this.movingForward != this.movingBackwards) {
            this.camera.moveForward(this.moveSpeed * dt * (this.movingForward ? 1 : -1));
        }
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
    requestRedraw() {
        this.requireRedraw = true;
    }

    animationFrameCallback() {
        let currentTime = window.performance.now() / 1000.0;
        if (this.lastUpdateTime === undefined) {
            this.lastUpdateTime = currentTime;
        }
        let dt = currentTime - this.lastUpdateTime;

        this.frameChecker = 0;
        if (typeof ga_instance !== "undefined" && ga_instance) {
            ga_instance.update();
        }
        this.render(dt);
        this.initFrameCallback();
        this.lastUpdateTime = currentTime;
    }
    initFrameCallback() {
        requestAnimationFrame(function() {
            gr_instance.animationFrameCallback();
        });
    }
    cleanup() {
        this._nodes.removeAll();
        this.flushScene();
    }
    downloadFile(link, async) {
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
    createShader(text, isFragment) {
        let shader = gl.createShader(isFragment ? gl.FRAGMENT_SHADER : gl.VERTEX_SHADER);
        gl.shaderSource(shader, text);
        gl.compileShader(shader);
        if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
            console.error(gl.getShaderInfoLog(shader));
            console.warn(text);
            return;
        }
        return shader;
    }
    createProgram(vertexShader, fragmentShader) {
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
    downloadShader(link, isFragment) {
        let text = this.downloadFile(link, false);
        if (text)
            return this.createShader(text, isFragment);
    }
}

function gwInitRenderer(viewDomObject) {
    if (gr_instance) {
        console.warn("RenderController already created", gr_instance);
        return;
    }
    gr_instance = new RenderController(viewDomObject);
    gr_instance.changeRenderChain(grRenderChain_SkinnedTextured);
    gr_instance.onResize();
    gr_instance.initFrameCallback();
}