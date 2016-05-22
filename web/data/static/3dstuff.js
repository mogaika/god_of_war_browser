'use strict';
var gl;
var gMatrix = mat4.create();

window.requestAnimFrame = (function() {
    return window.requestAnimationFrame ||
        window.webkitRequestAnimationFrame ||
        window.mozRequestAnimationFrame ||
        window.oRequestAnimationFrame ||
        window.msRequestAnimationFrame ||
        function(/* function FrameRequestCallback */ callback, /* DOMElement Element */ element) {
            window.setTimeout(callback, 1000/60);
        };
})();

function MeshObject(texture) {
    this.pBufVertex = null;
    this.pBufIndex = null;
    this.pBufColor = null;
    this.pBufNormal = null;
    this.pBufTexture = null;
    this.pTextureId = texture;
}

MeshObject.prototype.loadVertexData = function(vertex, index) {
    this.pBufVertex = gl.createBuffer();
    this.pBufIndex = gl.createBuffer();
}

MeshObject.prototype.loadColorData = function(color) {
    this.pBufColor = gl.createBuffer();
}

MeshObject.prototype.loadNormalData = function(normal) {
    this.pBufNormal = gl.createBuffer();
}

MeshObject.prototype.loadTextureData = function(texture) {
    this.pBufTexture = gl.createBuffer();
}

MeshObject.prototype.free = function() {
    if (this.pBufVertex != null) {
        gl.deleteBuffer(this.pBufVertex);
        gl.deleteBuffer(this.pBufIndex);
    }
    if (this.pBufColor != null) { gl.deleteBuffer(this.pBufColor); }
    if (this.pBufNormal != null) { gl.deleteBuffer(this.pBufNormal); }
    if (this.pBufTexture != null) { gl.deleteBuffer(this.pBufTexture); }
}

function Mesh() {
    this.objects = [];
}

Mesh.prototype.free = function() {
    for (i in this.objects) {
        this.objects[i].free();
    }
}

function Model(mesh, bones) {
    this.pMatrix = mat4.create();
    this.pMesh = mesh;
}

function initGL(canvas) {
    var names = ["webgl", "experimental-webgl", "webkit-3d", "moz-webgl"];
    for (var i in names) {
        try {
            gl = canvas[0].getContext(names[i]);
        } catch(e) {}
        if (gl) {
            break;
        }
    }
    if (!gl) {
        alert("Could not initialise WebGL");
    } else {
        $(window).resize(function(){0
            gl.viewportWidth = canvas.width();
            gl.viewportHeight = canvas.height();
            gl.viewport(0,0, canvas.width(), canvas.height());
        });
        gl.viewportWidth = canvas.width();
        gl.viewportHeight = canvas.height();
        gl.viewport(0,0, canvas.width(), canvas.height());
    }
    
    requestAnimFrame(drawScene);
}

function getShader(id) {
    var shaderScript = document.getElementById(id);
    if (!shaderScript) {
        return null;
    }

    var str = "";
    var k = shaderScript.firstChild;
    while (k) {
        if (k.nodeType == 3) {
            str += k.textContent;
        }
        k = k.nextSibling;
    }

    var shader;
    if (shaderScript.type == "x-shader/x-fragment") {
        shader = gl.createShader(gl.FRAGMENT_SHADER);
    } else if (shaderScript.type == "x-shader/x-vertex") {
        shader = gl.createShader(gl.VERTEX_SHADER);
    } else {
        return null;
    }

    gl.shaderSource(shader, str);
    gl.compileShader(shader);
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
        alert(gl.getShaderInfoLog(shader));
        return null;
    }
    return shader;
}

var shaderProgram;

function initShaders() {
    var fragmentShader = getShader(gl, "shader-fs");
    var vertexShader = getShader(gl, "shader-vs");

    shaderProgram = gl.createProgram();
    gl.attachShader(shaderProgram, vertexShader);
    gl.attachShader(shaderProgram, fragmentShader);
    gl.linkProgram(shaderProgram);

    if (!gl.getProgramParameter(shaderProgram, gl.LINK_STATUS)) {
        alert("Could not initialise shaders");
    }

    gl.useProgram(shaderProgram);

    shaderProgram.vertexPositionAttribute = gl.getAttribLocation(shaderProgram, "aVertexPosition");
    gl.enableVertexAttribArray(shaderProgram.vertexPositionAttribute);

    shaderProgram.pMatrixUniform = gl.getUniformLocation(shaderProgram, "uPMatrix");
    shaderProgram.mvMatrixUniform = gl.getUniformLocation(shaderProgram, "uMVMatrix");
}

function setMatrixUniforms() {
    gl.uniformMatrix4fv(shaderProgram.pMatrixUniform, false, pMatrix);
    gl.uniformMatrix4fv(shaderProgram.mvMatrixUniform, false, mvMatrix);
}

function degToRad(degrees) {
    return degrees * Math.PI / 180;
}

function drawScene() {
    gl.viewport(0, 0, gl.viewportWidth, gl.viewportHeight);
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

    mat4.perspective(45, gl.viewportWidth / gl.viewportHeight, 0.1, 5000.0, gMatrix);

    mat4.identity(gMatrix);

    mat4.translate(gMatrix, [0.0, 0.0, 1000.0]);

    mat4.rotate(gMatrix, degToRad(45.0), [1, 0, 0]);
    mat4.rotate(gMatrix, degToRad(45.0), [0, 1, 0]);

    /*
    gl.bindBuffer(gl.ARRAY_BUFFER, cubeVertexPositionBuffer);
    gl.vertexAttribPointer(shaderProgram.vertexPositionAttribute, cubeVertexPositionBuffer.itemSize, gl.FLOAT, false, 0, 0);

    gl.bindBuffer(gl.ARRAY_BUFFER, cubeVertexNormalBuffer);
    gl.vertexAttribPointer(shaderProgram.vertexNormalAttribute, cubeVertexNormalBuffer.itemSize, gl.FLOAT, false, 0, 0);

    gl.bindBuffer(gl.ARRAY_BUFFER, cubeVertexTextureCoordBuffer);
    gl.vertexAttribPointer(shaderProgram.textureCoordAttribute, cubeVertexTextureCoordBuffer.itemSize, gl.FLOAT, false, 0, 0);

    gl.activeTexture(gl.TEXTURE0);
    gl.bindTexture(gl.TEXTURE_2D, glassTexture);
    gl.uniform1i(shaderProgram.samplerUniform, 0);

    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, cubeVertexIndexBuffer);
    setMatrixUniforms();
    gl.drawElements(gl.TRIANGLES, cubeVertexIndexBuffer.numItems, gl.UNSIGNED_SHORT, 0);
    */
}

