'use strict';
var gl;
var models = [];
var viewW, viewH;

var shaderFs = `
void main(void) {
    gl_FragColor = vec4(0.0, 1.0, 0.0, 1.0);
}
`;

var shaderVs = `
attribute vec3 aVertexPosition;

uniform mat4 uModelMatrix;
uniform mat4 uProjectionViewMatrix;

void main(void) {
    gl_Position = uProjectionViewMatrix * uModelMatrix * vec4(aVertexPosition, 1.0);
}
`;

window.requestAnimFrame = (function() {
    return window.requestAnimationFrame ||
        window.webkitRequestAnimationFrame ||
        window.mozRequestAnimationFrame ||
        window.oRequestAnimationFrame ||
        window.msRequestAnimationFrame ||
        function(callback, element) {
            window.setTimeout(callback, 1000/60);
        };
})();

function Mesh() {
    this.objects = [];
}
Mesh.prototype.add = function(meshObject) {
    this.objects.push(meshObject);
}
Mesh.prototype.free = function() {
    for (var i in this.objects) {
        this.objects[i].free();
    }
}

function MeshObject(vertexData, indexData) {
    this.countVertexes = vertexData.length / 3;
    this.countIndexes = indexData.length;
    
    if (vertexData.countVertexes > 255) {
        console.warn('8 bit indexes not support more than 255 vertexes');
    }
    
    this.pVertexBuffer = gl.createBuffer();
    this.pIndexBuffer = gl.createBuffer();
    
    gl.bindBuffer(gl.ARRAY_BUFFER, this.pVertexBuffer);
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, this.pIndexBuffer);
    
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(vertexData), gl.STATIC_DRAW);
    gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, new Uint8Array(indexData), gl.STATIC_DRAW);
}
MeshObject.prototype.free = function() {
    if (this.pBufVertex != null) gl.deleteBuffer(this.pBufVertex);
    if (this.pIndexBuffer != null) gl.deleteBuffer(this.pIndexBuffer);   
}

var current_models = [];
function Model(mesh) {
    this.pMatrix = mat4.create();
    this.pMesh = mesh;
    
    current_models.push(this);
}
Model.prototype.free = function() {
    current_models.splice(current_models.indexOf(this), 1);
    this.pMesh.free();
}
Model.prototype.setMatrix = function(mat) {
    mat4.copy(this.pMatrix, mat);
}

function redraw3d() {
    requestAnimFrame(drawScene);
}

function reset3d() {
    for (var i in current_models)
        current_models[i].free();
    current_models.length = 0;
}

function init3d(canvas) {
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
        return;
    } else {
        initShaders();
        
        var viewportSet = function() {
            viewW = gl.canvas.clientWidth;
            viewH = gl.canvas.clientHeight;
            gl.canvas.width = viewW;
            gl.canvas.height = viewH;
            gl.viewport(0,0, viewW, viewH);            
            redraw3d();
        };
        
        gl.clearColor(1.0, 0, 1.0, 1);
        gl.enable(gl.DEPTH_TEST);
        gl.clearDepth(1.0);
        gl.depthFunc(gl.LEQUAL);
        
        $(window).resize(viewportSet);
        
        viewportSet();
    }
}

function getShader(text, isFragment) {
    var shader = gl.createShader(isFragment ? gl.FRAGMENT_SHADER : gl.VERTEX_SHADER);
    gl.shaderSource(shader, text);
    gl.compileShader(shader);
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
        console.log(text, gl.getShaderInfoLog(shader));
        return null;
    }
    return shader;
}

var shaderProgram;

function initShaders() {
    var fragmentShader = getShader(shaderFs, true);
    var vertexShader = getShader(shaderVs, false);

    shaderProgram = gl.createProgram();
    gl.attachShader(shaderProgram, vertexShader);
    gl.attachShader(shaderProgram, fragmentShader);
    gl.linkProgram(shaderProgram);

    if (!gl.getProgramParameter(shaderProgram, gl.LINK_STATUS)) {
        console.log("Could not initialise shaders");
    }

    gl.useProgram(shaderProgram);

    shaderProgram.vertexPositionAttribute = gl.getAttribLocation(shaderProgram, "aVertexPosition");
    gl.enableVertexAttribArray(shaderProgram.vertexPositionAttribute);

    shaderProgram.mProjectionView = gl.getUniformLocation(shaderProgram, "uProjectionViewMatrix");
    shaderProgram.mModel = gl.getUniformLocation(shaderProgram, "uModelMatrix");
}

function degToRad(degrees) {
    return degrees * Math.PI / 180;
}

function drawScene() {
    if (!data3d.is(':visible'))
        return;
    
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

    var projMatrix = mat4.create();
    mat4.perspective(projMatrix, degToRad(55.0), viewW/viewH, 1.0, 1000.0);
    
    
    var viewMatrix = mat4.create();

    mat4.translate(viewMatrix, viewMatrix, [0.0, 0.0, -200.0]);
    mat4.rotate(viewMatrix, viewMatrix, degToRad(45.0), [1, 0, 0]);
    mat4.rotate(viewMatrix, viewMatrix, degToRad(45.0), [0, 1, 0]);
    
    var projViewMatrix = mat4.create();
    mat4.mul(projViewMatrix, projMatrix, viewMatrix);
    
    gl.uniformMatrix4fv(shaderProgram.mProjectionView, false, projViewMatrix);
    
    var stat_vert = 0;
    var stat_index = 0;
    var stat_tria = 0;
    
    for (var i in current_models) {
        var mdl = current_models[i];
        
        gl.uniformMatrix4fv(shaderProgram.mModel, false, mdl.pMatrix);
        for (var j in mdl.pMesh.objects) {
            var mesh = mdl.pMesh.objects[j];
            
            gl.bindBuffer(gl.ARRAY_BUFFER, mesh.pVertexBuffer);
            gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.pIndexBuffer);
            
            gl.vertexAttribPointer(shaderProgram.vertexPositionAttribute,
                                   3, gl.FLOAT, false, 0, 0);
            
            gl.drawElements(gl.TRIANGLES, mesh.countIndexes, gl.UNSIGNED_BYTE, 0);
            
            stat_vert += mesh.countVertexes;
            stat_index += mesh.countIndexes;
            stat_tria += mesh.countIndexes / 3;
        }
    }
    
    //console.log('verts:' + stat_vert, 'index:' + stat_index, 'tria:' + stat_tria);
}

