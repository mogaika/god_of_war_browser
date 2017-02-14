function grRenderChain_SkinnedTextured(ctrl) {
	this.vertexShader = ctrl.downloadShader("/static/gowRenderers/SkinnedTextured.vs", false);
	this.fragmentShader = ctrl.downloadShader("/static/gowRenderers/SkinnedTextured.fs", true);
	this.program = ctrl.createProgram(this.vertexShader, this.fragmentShader);
	
    gl.useProgram(this.program);

    this.aVertexPos = gl.getAttribLocation(this.program, "aVertexPos");
	this.aVertexColor = gl.getAttribLocation(this.program, "aVertexColor");
	this.aVertexUV = gl.getAttribLocation(this.program, "aVertexUV");
	this.aVertexJointID = gl.getAttribLocation(this.program, "aVertexJointID");

    this.umProjectionView = gl.getUniformLocation(this.program, "umProjectionView");
    this.umModelTransform = gl.getUniformLocation(this.program, "umModelTransform");
	this.uMaterialColor = gl.getUniformLocation(this.program, "uMaterialColor");
	this.uMaterialDiffuseSampler = gl.getUniformLocation(this.program, "uMaterialDiffuseSampler");
	this.uUseMaterialDiffuseSampler = gl.getUniformLocation(this.program, "uUseMaterialDiffuseSampler");
	this.uUseVertexColor = gl.getUniformLocation(this.program, "uUseVertexColor");
	this.uUseModelTransform = gl.getUniformLocation(this.program, "uUseModelTransform");	
	
	this.umJoints = [];
	for (var i = 0; i < 120; i+=1) {
		this.umJoints.push(gl.getUniformLocation(this.program, "umJoints[" + i + "]"));
	}
	this.uUseJoints = gl.getUniformLocation(this.program, "uUseJoints");

	gl.enableVertexAttribArray(this.aVertexPos);
	gl.enableVertexAttribArray(this.aVertexColor);
	
	gl.uniform1i(this.uMaterialDiffuseSampler, 0);
	gl.uniform1i(this.uUseMaterialDiffuseSampler, 0);
	gl.uniform1i(this.uUseJoints, 0);
	
	gl.clearColor(0.2, 0.2, 0.2, 1);
	gl.clearDepth(1.0);
	gl.depthFunc(gl.LEQUAL);
	gl.disable(gl.BLEND);
	gl.depthMask(true);
	gl.enable(gl.DEPTH_TEST);
	gl.disable(gl.CULL_FACE);
	
	console.log(this);
}

grRenderChain_SkinnedTextured.prototype.free = function(ctrl) {
	// TODO :add missed fields
	gl.disableVertexAttribArray(this.aVertexPos);
	gl.disableVertexAttribArray(this.aVertexColor);
	gl.disableVertexAttribArray(this.aVertexUV);
	gl.deleteProgram(this.program);
	gl.deleteShader(this.vertexShader);
	gl.deleteShader(this.fragmentShader);
}

grRenderChain_SkinnedTextured.prototype.drawMesh = function(mesh, hasTexture = false, hasJoints = false) {
	gl.enableVertexAttribArray(this.aVertexPos);
	gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferVertex);
	gl.vertexAttribPointer(this.aVertexPos, 3, gl.FLOAT, false, 0, 0);

	if (mesh.bufferBlendColor) {
		gl.uniform1i(this.uUseVertexColor, 1);
		gl.enableVertexAttribArray(this.aVertexColor);
		gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferBlendColor);
		gl.vertexAttribPointer(this.aVertexColor, 4, gl.UNSIGNED_BYTE, true, 0, 0);
	} else {
		gl.uniform1i(this.uUseVertexColor, 0);
		gl.disableVertexAttribArray(this.aVertexColor);
	}
	
	if (mesh.bufferUV && hasTexture) {
		gl.uniform1i(this.uUseMaterialDiffuseSampler, 1);
		gl.enableVertexAttribArray(this.aVertexUV);
		gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferUV);
		gl.vertexAttribPointer(this.aVertexUV, 2, gl.FLOAT, false, 0, 0);
	} else {
		gl.uniform1i(this.uUseMaterialDiffuseSampler, 0);
		gl.disableVertexAttribArray(this.aVertexUV);
	}
	
	if (mesh.bufferJointIds && hasJoints) {
		gl.uniform1i(this.uUseJoints, 1);
		gl.enableVertexAttribArray(this.aVertexJointID);
		gl.bindBuffer(gl.ARRAY_BUFFER, mesh.bufferJointIds);
		gl.vertexAttribPointer(this.aVertexJointID, 1, gl.BYTE, false, 0, 0);
	} else {
		if (!mesh.bufferJointIds && hasJoints) {
			console.warn("has joints but without jointIdsBuffer", mesh);
		}
		gl.uniform1i(this.uUseJoints, 0);
		gl.disableVertexAttribArray(this.aVertexJointID);
	}

	gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, mesh.bufferIndex);
	gl.drawElements(mesh.primitive, mesh.indexesCount, mesh.bufferIndexType, 0);
}

grRenderChain_SkinnedTextured.prototype.renderModels = function(ctrl, mdls, filterFunc, useSkelet = true) {
	var rendered_meshes = 0;
	for (var i in mdls) {
		var mdl = mdls[i];
		var lastMatIndex = -1;
		if (mdl.visible) {
			gl.uniformMatrix4fv(this.umModelTransform, false, mdl.matrix);

			var hasSkelet = useSkelet && !!mdl.skeleton;
			var usedJoints = [];
			for (var j in mdl.meshes) {
				var mesh = mdl.meshes[j];
				
				if (filterFunc && !filterFunc(mdl, mesh))
					continue;
				
				var hasTxr = true;
				
				if (mesh.isDepthTested) {
					gl.enable(gl.DEPTH_TEST);
				} else {
					gl.disable(gl.DEPTH_TEST);
				}
				
				if (hasSkelet && mesh.jointMapping) {
					for (var i in mesh.jointMapping) {
						if (i >= 12) {
							console.warn("jointMap array in shader is overflowed", mesh.jointMapping)
						}
						//console.log(mdl.skeleton[mesh.jointMapping[i]], mdl.skeleton, mesh.jointMapping[i]);
						var jointId = mesh.jointMapping[i];
						if (usedJoints.indexOf(jointId) < 0) {
							usedJoints.push(jointId);
						}
						if (jointId >= mdl.skeleton.length) {
							console.warn("joint mapping out of index. jointMapping[" + i + "]=" + jointId + " >= " + mdl.skeleton.length);
						} else {
							//console.info(i, mesh.jointMapping[i], mdl.skeleton[mesh.jointMapping[i]]);
							gl.uniformMatrix4fv(this.umJoints[i], false, mdl.skeleton[jointId]);
						}						
					}
					//console.info("used " + mesh.jointMapping.length + " joint slots");
				} else {
					hasSkelet = false;
				}
				//console.log(mesh, mesh.materialIndex, mdl.materials);
				if (mesh.materialIndex != undefined && mdl.materials && mdl.materials.length) {
					var mat = mdl.materials[mesh.materialIndex];
					gl.bindTexture(gl.TEXTURE_2D, mat.textureDiffuse.get());
					gl.uniform4f(this.uMaterialColor, mat.color[0], mat.color[1], mat.color[2], mat.color[3]);
				} else {
					gl.bindTexture(gl.TEXTURE_2D, ctrl.emptyTexture.get());
					hasTxr = false;
					gl.uniform4f(this.uMaterialColor, 255, 255, 255, 255);
				}
				
				this.drawMesh(mesh, hasTxr, hasSkelet);
				
				rendered_meshes += 1;
			}
			console.log("Used joints", usedJoints.sort(function(a, b){return a - b;}));
		}
	}
	return rendered_meshes;
}

function __mdl_mesh_has_alpha(mdl, mesh) {
	return (mesh.materialIndex && mesh.materialIndex < mdl.materials.length && mdl.materials[mesh.materialIndex].hasAlpha);	
}

grRenderChain_SkinnedTextured.prototype.render = function(ctrl) {
	gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

	gl.uniformMatrix4fv(this.umModelTransform, false, mat4.create());
	
	gl.activeTexture(gl.TEXTURE0);
	
	var rendered_meshes = 0;
	
	var skies = ctrl.models.filter(function(m){return m.type === "sky";});
	
	if (skies.length > 0) {
		if (skies.length > 1) {
			console.warn("Too many skies: ", skies);
		}
		
		gl.uniform1i(this.uUseModelTransform, 0);
		
		var sky = skies[0];
		var finalMat = mat4.mul(mat4.create(), ctrl.camera.getViewMatrix(), sky.matrix);
		var rot = mat4.getRotation(quat.create(), finalMat);
		var nullPosedMatrix = mat4.fromQuat(mat4.create(), rot);
		nullPosedMatrix = mat4.mul(mat4.create(), ctrl.camera.getProjectionMatrix(), nullPosedMatrix);
		
		gl.uniformMatrix4fv(this.umProjectionView, false, nullPosedMatrix);
		
		rendered_meshes += this.renderModels(ctrl, [sky], undefined, false);
	}
	
	var mdls = [].concat(ctrl.helpers).concat(ctrl.models.filter(function(m){return !m.type;}));
	if (mdls.length > 0) {
		gl.uniform1i(this.uUseModelTransform, 1);
		gl.uniformMatrix4fv(this.umProjectionView, false, ctrl.camera.getProjViewMatrix());
		
		gl.disable(gl.BLEND);
		gl.depthMask(true);
		//rendered_meshes += this.renderModels(ctrl, mdls, function(mdl, mesh) {return !__mdl_mesh_has_alpha(mdl, mesh);});
		rendered_meshes += this.renderModels(ctrl, mdls, function(){return true;});
		
		//console.log(mdls);
		
		//gl.enable(gl.BLEND);
		//gl.blendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA);
		//gl.depthMask(false);
		//rendered_meshes += this.renderModels(ctrl, mdls, __mdl_mesh_has_alpha);
	}
	
	console.log("stats:", rendered_meshes);	
}