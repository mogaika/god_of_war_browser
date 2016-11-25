'use strict';
var gl;

var models = [];
var textureIdMap = new Map();

var viewW, viewH;
var viewDist = 100.0;
var viewRotX = 45.0, viewRotY = 45.0;
var viewMouseDown = false;
var viewMouseX, viewMouseY;

var shaderBoolColor = false, shaderBoolUV = false;

var shaderFs = `
uniform sampler2D uSampler;
uniform bool bUseVertexUV;

varying lowp vec4 vColor;
varying mediump vec2 vTextureVU;

void main(void) {
    if (bUseVertexUV) {
        gl_FragColor = vColor * texture2D(uSampler, vTextureVU);
    } else {
        gl_FragColor = vColor;
    }
}
`;

var shaderVs = `
attribute highp vec3 aVertexPosition;
attribute highp vec4 aVertexColor;
attribute mediump vec2 aVertexUV;

uniform highp mat4 uModelMatrix;
uniform highp mat4 uProjectionViewMatrix;

uniform bool bUseVertexUV;
uniform bool bUseVertexColor;

varying lowp vec4 vColor;
varying mediump vec2 vTextureVU;

void main(void) {
    gl_Position = uProjectionViewMatrix * uModelMatrix * vec4(aVertexPosition, 1.0);
    if (bUseVertexColor) {
        vColor = aVertexColor * (256.0 / 128.0);
    } else {
        vColor = vec4(1.0);
    }
    if (bUseVertexUV) {
        vTextureVU = aVertexUV;
    }
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

function Texture(uid, src, haveTransparent=true) {
    this.ref = 0;
	this.refs = [];
    this.uid = uid;
    var gltex = gl.createTexture();
    this.pTexture = gltex;
	this.isHaveTransparentPixel = haveTransparent;
    
    gl.bindTexture(gl.TEXTURE_2D, this.pTexture);
    // pink placeholder
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE,
              new Uint8Array([255, 0, 255, 128]));
    
    var image = new Image();
    image.src = src;
    image.onload = function() {
        gl.bindTexture(gl.TEXTURE_2D, gltex);
        gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, image);
        gl.generateMipmap(gl.TEXTURE_2D);
        redraw3d();
    };

    textureIdMap.set(uid, this);
}
Texture.prototype.refInc = function(refobj) {
    this.refs.push(refobj);
}
Texture.prototype.refDec = function(refobj) {
	this.refs.splice(this.refs.indexOf(refobj), 1);
    if (this.refs.length === 0) {
        this.free();
    }
}
Texture.prototype.free = function() {
    console.info('free texture ' + this.uid);
    textureIdMap.set(this.uid, undefined);
    gl.deleteTexture(this.pTexture);
}
var latest_binded_texture = null;
Texture.prototype.bind = function() {
	if (this != latest_binded_texture) {
		gl.bindTexture(gl.TEXTURE_2D, this.pTexture);
		latest_binded_texture = this;
	}
}

function LoadTexture(uid, src, isHaveTransparent=true) {
    var ex = textureIdMap.get(uid);
    if (!ex) {
        ex = new Texture(uid, src, isHaveTransparent);
    }
    return ex;
}

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

function MeshObject(vertexData, indexData, colorData, texture, textureData, hasTransparentVertexes) {
    this.countVertexes = vertexData.length / 3;
    this.countIndexes = indexData.length;
	this.hasTransparentVertexes = hasTransparentVertexes;
    
    if (vertexData.countVertexes > 255) {
        console.warn('8 bit indexes not support more than 255 vertexes');
    }
    
    this.pVertexBuffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, this.pVertexBuffer);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(vertexData), gl.STATIC_DRAW);
    
    this.pIndexBuffer = gl.createBuffer();
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, this.pIndexBuffer);
    gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, new Uint8Array(indexData), gl.STATIC_DRAW);
    
    if (colorData) {
        this.pColorBuffer = gl.createBuffer();
        gl.bindBuffer(gl.ARRAY_BUFFER, this.pColorBuffer);
        gl.bufferData(gl.ARRAY_BUFFER, new Uint8Array(colorData), gl.STATIC_DRAW);
        
        this.countColors = colorData.length / 4;
        if (this.countColors != this.countVertexes) 
            console.error('color count mismath ' + this.countColors + ' != ' + this.countVertexes);
    } else {
        this.pColorBuffer = null;
        this.countColors = 0;
    }
    
    if (texture && textureData) {
        this.countTextures = textureData.length / 2;
        if (this.countTextures != this.countVertexes) {
            console.error('texture count mismath ' + this.countTextures + ' != ' + this.countVertexes);
			console.warn('ignore texture data because texture count mismath');
		} else {
			this.pTextureBuffer = gl.createBuffer();
	        gl.bindBuffer(gl.ARRAY_BUFFER, this.pTextureBuffer);
	        gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(textureData), gl.STATIC_DRAW);

	        this.texture = texture;
	        this.texture.refInc(this);
		}
    } else {
        this.pTextureBuffer = null;
        this.countTextures = 0;
        this.texture = null;
    }
}
MeshObject.prototype.free = function() {
    if (this.pBufVertex) gl.deleteBuffer(this.pBufVertex);
    if (this.pIndexBufferl) gl.deleteBuffer(this.pIndexBuffer);
    if (this.pColorBuffer) gl.deleteBuffer(this.pColorBuffer);
    if (this.pTextureBuffer) gl.deleteBuffer(this.pTextureBuffer);
    if (this.texture) this.texture.refDec(this);
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
        
        gl.clearColor(0.2, 0.2, 0.2, 1);
        gl.enable(gl.DEPTH_TEST);
        gl.clearDepth(1.0);
        gl.depthFunc(gl.LEQUAL);

        $(window).resize(viewportSet);
        
        viewportSet();
    }
    
    canvas.mousewheel(function(event) {
        var resizeDelta = Math.sqrt(viewDist) * event.deltaY * event.deltaFactor * 0.01;
        viewDist -= resizeDelta * 2;
        if (viewDist < 1.0)
            viewDist = 1.0;
        redraw3d();
        
        event.stopPropagation();
        event.preventDefault();
    }).mousedown(function(event) {
        if (event.button === 0) {
            viewMouseDown = true;
            viewMouseX = event.clientX;
            viewMouseY = event.clientY;
            
            event.stopPropagation();
            event.preventDefault();
        }
    });
    
    $(window).mouseup(function(event) {
        if (event.button === 0) {
            viewMouseDown = false;
            
            event.stopPropagation();
            event.preventDefault();
        }
    }).mousemove(function(event) {
        if (viewMouseDown) {
            viewRotY += (event.clientX - viewMouseX) * 0.2;
            viewRotX += (event.clientY - viewMouseY) * 0.2;
            
            viewMouseX = event.clientX;
            viewMouseY = event.clientY;
            redraw3d();
            
            event.stopPropagation();
            event.preventDefault();
        }
    });
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
    shaderProgram.vertexColorAttribute = gl.getAttribLocation(shaderProgram, "aVertexColor");
    shaderProgram.vertexUVAttribute = gl.getAttribLocation(shaderProgram, "aVertexUV");

    shaderProgram.mProjectionView = gl.getUniformLocation(shaderProgram, "uProjectionViewMatrix");
    shaderProgram.mModel = gl.getUniformLocation(shaderProgram, "uModelMatrix");
    shaderProgram.uSampler = gl.getUniformLocation(shaderProgram, "uSampler");
    shaderProgram.bVertexColor = gl.getUniformLocation(shaderProgram, "bUseVertexColor");
    shaderProgram.bVertexUV = gl.getUniformLocation(shaderProgram, "bUseVertexUV");
    
    gl.disableVertexAttribArray(shaderProgram.vertexColorAttribute);
    gl.disableVertexAttribArray(shaderProgram.vertexUVAttribute);
    gl.uniform1i(shaderProgram.uSampler, 0);
    gl.uniform1i(shaderProgram.bVertexColor, 0);
    gl.uniform1i(shaderProgram.bVertexUV, 0);
    shaderBoolColor = false;
    shaderBoolUV = false;
}

var passes = [true, true, true];
var additive_pass = false;

function drawScene() {
    if (!data3d.is(':visible'))
        return;
    
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

    var projMatrix = mat4.create();
    mat4.perspective(projMatrix, glMatrix.toRadian(45.0), viewW/viewH, 1.0, 8000.0);
    
    var viewMatrix = mat4.create();

    mat4.translate(viewMatrix, viewMatrix, [0.0, 0.0, -viewDist]);
    mat4.rotate(viewMatrix, viewMatrix, glMatrix.toRadian(viewRotX), [1, 0, 0]);
    mat4.rotate(viewMatrix, viewMatrix, glMatrix.toRadian(viewRotY), [0, 1, 0]);
    
    var projViewMatrix = mat4.create();
    mat4.mul(projViewMatrix, projMatrix, viewMatrix);
    
    gl.uniformMatrix4fv(shaderProgram.mProjectionView, false, projViewMatrix);
    
    function drawModelsArray(models, stat, meshFilter) {
		var stat_vert = 0;
	    var stat_index = 0;
	    var stat_tria = 0;
		for (var i in models) {
			var mdl = models[i];

			gl.uniformMatrix4fv(shaderProgram.mModel, false, mdl.pMatrix);
	        for (var j in mdl.pMesh.objects) {
	            var mesh = mdl.pMesh.objects[j];

				if (!meshFilter(mesh)) {
					continue;
				}

	            gl.enableVertexAttribArray(shaderProgram.vertexPositionAttribute);
	            gl.bindBuffer(gl.ARRAY_BUFFER, mesh.pVertexBuffer);
	            gl.vertexAttribPointer(shaderProgram.vertexPositionAttribute, 3, gl.FLOAT, false, 0, 0);

	            if (mesh.pColorBuffer) {
	                gl.enableVertexAttribArray(shaderProgram.vertexColorAttribute);
	                gl.bindBuffer(gl.ARRAY_BUFFER, mesh.pColorBuffer);
	                gl.vertexAttribPointer(shaderProgram.vertexColorAttribute, 4, gl.UNSIGNED_BYTE, true, 0, 0);
	                if (shaderBoolColor === false) {
	                    gl.uniform1i(shaderProgram.bVertexColor, 1);
	                    shaderBoolColor = true;
	                }
	            } else {
	                gl.disableVertexAttribArray(shaderProgram.vertexColorAttribute);
	                if (shaderBoolColor === true) {
	                    gl.uniform1i(shaderProgram.bVertexColor, 0);
	                    shaderBoolColor = false;
	                }
	            }

	            if (mesh.texture && mesh.pTextureBuffer) {
	                gl.enableVertexAttribArray(shaderProgram.vertexUVAttribute);
	                gl.bindBuffer(gl.ARRAY_BUFFER, mesh.pTextureBuffer);
	                gl.vertexAttribPointer(shaderProgram.vertexUVAttribute, 2, gl.FLOAT, false, 0, 0);

					mesh.texture.bind();

	                if (shaderBoolUV === false) {
	                    gl.activeTexture(gl.TEXTURE0);
	                    gl.uniform1i(shaderProgram.bVertexUV, 1);
	                    shaderBoolUV = true;
	                }
	            } else {
	                gl.disableVertexAttribArray(shaderProgram.vertexUVAttribute);
	                if (shaderBoolUV === true) {
	                    gl.uniform1i(shaderProgram.bVertexUV, 0);
	                    shaderBoolUV = false;
	                }
	            }

	            gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.pIndexBuffer);

	            gl.drawElements(gl.TRIANGLES, mesh.countIndexes, gl.UNSIGNED_BYTE, 0);

	            stat_vert += mesh.countVertexes;
	            stat_index += mesh.countIndexes;
	            stat_tria += mesh.countIndexes / 3;
	        }
	    }
		stat.vert += stat_vert
		stat.index += stat_index
		stat.tria += stat_tria
	}

	var stat_pass1 = {'vert':0, 'index':0, 'tria':0};

	gl.depthMask(true);
	if (additive_pass) {
		gl.enable(gl.BLEND);
		gl.blendFunc(gl.ONE, gl.ONE);
	} else {
		gl.disable(gl.BLEND);
	}

	if (passes[0]) {
		drawModelsArray(current_models, stat_pass1, function(mesh) {return !(mesh.texture && mesh.texture.isHaveTransparentPixel) && !mesh.hasTransparentVertexes;});
	}

	if (!additive_pass) {
		gl.enable(gl.BLEND);
		gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA);
	}

	var stat_pass2 = {'vert':0, 'index':0, 'tria':0};
	if (passes[1]) {
		drawModelsArray(current_models, stat_pass2, function(mesh) {return !!mesh.hasTransparentVertexes && !(mesh.texture && mesh.texture.isHaveTransparentPixel);});
	}

	var stat_pass3 = {'vert':0, 'index':0, 'tria':0};
	if (passes[2]) {
		drawModelsArray(current_models, stat_pass3, function(mesh) {return mesh.texture && !!mesh.texture.isHaveTransparentPixel;});
	}

	console.log('pass1', JSON.stringify(stat_pass1), 'pass2', JSON.stringify(stat_pass2), 'pass3', JSON.stringify(stat_pass3));
}