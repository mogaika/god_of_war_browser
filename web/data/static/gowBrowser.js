'use strict';

var viewPack, viewTree, viewSummary, view3d;
var dataPack, dataTree, dataSummary, data3d;
var defferedLoadingWad;
var defferedLoadingWadNode;

String.prototype.replaceAll = function(search, replace) {
    if (replace === undefined) {
        return this.toString();
    }
    return this.replace(new RegExp('[' + search + ']', 'g'), replace);
};

function getActionLinkForWadNode(wad, nodeid, action, params = '') {
	return '/action/' + wad +'/' + nodeid + '/' + action + '?' + params;
}

function treeInputFilterHandler() {
	var filterText = $(this).val().toLowerCase();
	$(this).parent().find("div li label").each(function(a1, a2, a3) {
		var p = $(this).parent();
		if ($(this).text().toLowerCase().includes(filterText)) {
			p.show();
		} else {
			p.hide();
		}
	});
};

function set3dVisible(show) {
    if (show) {
        view3d.show();
        viewSummary.attr('style', '')
    } else {
        view3d.hide();
        viewSummary.attr('style', 'flex-grow:1;')
    }
}

function setTitle(viewHeh, title) {
    $(viewHeh).children(".view-item-title").text(title);
}

function setLocation(title, hash) {
	$("head title").text(title);
	if (window.history.pushState) {
		window.history.pushState(null, title, hash);
	} else {
		window.location.hash = hash;
	}
}

function packLoad() {
    dataPack.empty();
    $.getJSON('/json/pack', function(files) {
        var list = $('<ol>');
        for (var i in files) {
            var fileName = files[i];
            list.append($('<li>')
                    .attr('filename', fileName)
                    .append($('<label>').append(fileName))
                    .append($('<a download>')
                            .addClass('button-dump')
                            .attr('href', '/dump/pack/' + fileName)) );
        }
        dataPack.append(list);

		if (defferedLoadingWad) {
			packLoadFile(defferedLoadingWad);
		}

        $('#view-pack ol li label').click(function(ev) {
            packLoadFile($(this).parent().attr('filename'));
        });

		$('#view-pack-filter').trigger('input');

        console.log('pack loaded');
    })
}

function packLoadFile(filename) {
    dataTree.empty();
    $.getJSON('/json/pack/' + filename, function(data) {
        var ext = filename.slice(-3).toLowerCase();
        switch (ext) {
            case 'wad':
                treeLoadWad(filename, data);  
                break;
			case 'vag':
			case 'vpk':
				treeLoadVagVpk(data, filename);
				break;
            default:
                dataTree.append(JSON.stringify(data, undefined, 2).replaceAll('\n', '<br>'));
                break;
        }
        console.log('pack file ' + filename + ' loaded');
    });
}

function treeLoadVagVpk(data, filename) {
	set3dVisible(false);
	setTitle(viewTree, filename);
	var list = $("<ul>");
	var wavPath = '/dump/pack/' + filename + '/wav';
	
	list.append($("<li>").append("SampleRate: " + data.SampleRate));
	list.append($("<li>").append("Channels: " + data.Channels));
	list.append($("<li>").append($("<a>").attr("href", wavPath).append("Download WAV")));
	dataTree.append(list)
	
	dataTree.append($("<audio controls autoplay>").append($("<source>").attr("src", wavPath)));
	
    setLocation(filename, '#/' + filename);
}

function treeLoadWad(wadName, data) {
	if (defferedLoadingWadNode) {
		treeLoadWadNode(wadName, parseInt(defferedLoadingWadNode));
		defferedLoadingWadNode = undefined;
	} else {
		setLocation(wadName, '#/' + wadName);
	}
	
	var addNodes = function(nodes) {
        var ol = $('<ol>');
        for (var sn in nodes) {
            var node = data.Nodes[nodes[sn]];
            var li = $('<li>')
                    .attr('nodeid', node.Tag.Id)
                    .attr('nodename', node.Tag.Name)
					.attr('nodetag', node.Tag.Tag)
                    .append($('<label>').append(("0000" + node.Tag.Id).substr(-4,4) + '.' + node.Tag.Name));

			if (node.Tag.Tag == 0x1e) {
				if (node.Tag.Size == 0) {
					li.addClass('wad-node-link');
				} else {
					li.addClass('wad-node-data');
				}
			} else {
				li.append(' [' + node.Tag.Tag + ']');
			}
            
			li.append($('<a download>')
                    .addClass('button-dump')
                    .attr('href', '/dump/pack/' + wadName + '/' + node.Tag.Id))

            if (node.SubGroupNodes) {
                li.append(addNodes(node.SubGroupNodes));
            }
            ol.append(li);
        }
        return ol;
    }
    
    setTitle(viewTree, wadName);
    
    if (data.Roots) {
        dataTree.append(addNodes(data.Roots));
	}
	
	$('#view-item-filter').trigger('input');
    $('#view-tree ol li label').click(function(ev) {
        var node_element = $(this).parent();
        treeLoadWadNode(wadName, parseInt(node_element.attr('nodeid')));
    });
}

function hexdump(buffer, blockSize) {
	var table = $('<table>');
	blockSize = blockSize || 16;
	var lines = [];
	var hex = "0123456789ABCDEF";
	var blocks = Math.ceil(buffer.length/blockSize);
	for (var iBlock = 0; iBlock < blocks; iBlock += 1)	 {
		var blockPos = iBlock * blockSize;
		
		var line = '';
		var chars = '';
		for (var j = 0; j < Math.min(blockSize, buffer.length - blockPos); j += 1) {
			var code = buffer[blockPos + j];
			line += ' ' + hex[(0xF0 & code) >> 4] + hex[0x0F & code];
			chars += (code > 0x20 && code < 0x80) ? String.fromCharCode(code) : '.';
		}
		
		var tr = $('<tr>');
		tr.append($('<td>').append(("000000" + blockPos.toString(16)).slice(-6)));
		tr.append($('<td>').append(line));
		tr.append($('<td>').append(chars));
		table.append(tr);
	}	
	return table;
}

function treeLoadWadNode(wad, nodeid) {
    dataSummary.empty();
    $.getJSON('/json/pack/' + wad +'/' + nodeid, function(resp) {
        var data = resp.Data;
        var tag = resp.Tag;
		
		var needHexDump = false;
		var needMarshalDump = false;
		
        if (resp.error) {
            set3dVisible(false);
            setTitle(viewSummary, 'Error');
            dataSummary.append(resp.error);
			needHexDump = true;
        } else {
            setTitle(viewSummary, tag.Name);
			setLocation(wad + " => " + tag.Name, '#/' + wad + '/' + nodeid);
			
			if (tag.Tag == 0x1e) {
	            switch (resp.ServerId) {
					case 0x00000018: // sbk blk
					case 0x00040018: // sbk vag
						summaryLoadWadSbk(data, wad, nodeid);
						needMarshalDump = true;
						break;
	                case 0x00000007: // txr
	                    summaryLoadWadTxr(data);
	                    break;
	                case 0x00000008: // material
	                    summaryLoadWadMat(data);
	                    break;
					case 0x00000011: // collision
						gr_instance.destroyModels();
					    set3dVisible(true);
					    
						var mdl = new grModel();
						loadCollisionFromAjax(mdl, data);
						
						gr_instance.models.push(mdl);
					    gr_instance.requestRedraw();
						break;
	                case 0x0001000f: // mesh
	                    summaryLoadWadMesh(data, wad, nodeid);
	                    break;
	                case 0x0002000f: // mdl
	                    summaryLoadWadMdl(data, wad, nodeid);
	                    break;
	                case 0x00040001: // obj
	                    summaryLoadWadObj(data);
	                    break;
	                case 0x80000001: // cxt
	                    summaryLoadWadCxt(data);
	                    break;
	                case 0x00020001: // gameObject
	                    summaryLoadWadGameObject(data);
	                    break;
					case 0x00010004: // script
						summaryLoadWadScript(data);
						needMarshalDump = true;
						needHexDump = true;
						break;
	                case 0x0000000c: // gfx pal
	                default:
	                    set3dVisible(false);
						needMarshalDump = true;
						needHexDump = true;
	                    break;
	            }
			} else if (tag.Tag == 112) {
				summaryLoadWadGeomShape(data);
			} else {
				needHexDump = true;
			}
            console.log('wad ' + wad + ' file (' + tag.Name + ')' + tag.Id + ' loaded. serverid:' + resp.ServerId);
        }

		if (needMarshalDump) {
			dataSummary.append($("<pre>").append(JSON.stringify(data, null, "  ").replaceAll('\n', '<br>')));
		}

		if (needHexDump) {
			$.ajax({
				url: '/dump/pack/' + wad +'/' + nodeid,
				type: 'GET',
				dataType: 'binary',
				processData: false,
				success: function(blob) {
					var fileReader = new FileReader();
					fileReader.onload = function() {
						var arr = new Uint8Array(this.result);
						dataSummary.append($("<h5>").append('File size:' + arr.length));
						dataSummary.append(hexdump(arr));
					};
					fileReader.readAsArrayBuffer(blob);
				}
			});
		}
    });
}

function parseMeshPart(object, block) {
	var m_vertexes = [];
	var m_indexes = [];
	var m_colors;
	var m_textures;
	var m_normals;

	m_vertexes.length = block.Trias.X.length * 3;
	
	for (var i in block.Trias.X) {
		var j = i * 3;
		m_vertexes[j] = block.Trias.X[i];
		m_vertexes[j+1] = block.Trias.Y[i];
		m_vertexes[j+2] = block.Trias.Z[i];
		if (!block.Trias.Skip[i]) {
			m_indexes.push(i-1);
			m_indexes.push(i-2);
			m_indexes.push(i-0);
		}
	}
	
	var mesh = new grMesh(m_vertexes, m_indexes);
	
	if (block.Blend.R && block.Blend.R.length) {
		var m_colors = [];
		m_colors.length = block.Blend.R.length * 4;
		for (var i in block.Blend.R) {
			var j = i * 4;
			m_colors[j] = block.Blend.R[i];
			m_colors[j+1] = block.Blend.G[i];
			m_colors[j+2] = block.Blend.B[i];
			m_colors[j+3] = block.Blend.A[i];
		}
		
		mesh.setBlendColors(m_colors);
	}
	
	mesh.setMaterialID(object.MaterialId);
	
	if (block.Uvs.U && block.Uvs.U.length) {
		m_textures = [];
		m_textures.length = block.Uvs.U.length * 2;
			
		for (var i in block.Uvs.U) {
			var j = i * 2;
			m_textures[j] = block.Uvs.U[i];
			m_textures[j+1] = block.Uvs.V[i];
		}
		mesh.setUVs(m_textures);
	}
	
	if (block.Norms.X && block.Norms.X.length) {
		m_normals = [];
		m_normals.length = block.Norms.X.length * 3;
			
		for (var i in block.Norms.X) {
			var j = i * 3;
			m_normals[j] = block.Norms.X[i];
			m_normals[j+1] = block.Norms.Y[i];
			m_normals[j+2] = block.Norms.Z[i];
		}
		
		mesh.setNormals(m_normals);							
	}
	
	if (!!block.Joints && block.Joints.length && !!object.JointMapper && object.JointMapper.length) {
		//console.log(block.Joints, block.Joints2, object.JointMapper);
		var joints2 = undefined;
		if (block.Joints2) joints2 = block.Joints2;
		mesh.setJointIds(object.JointMapper, block.Joints, joints2);
	}
	
	return mesh;
}

function loadMeshFromAjax(model, data, needTable = false) {
	var table = needTable ? $('<table>') : undefined;
	for (var iPart in data.Parts) {
        var part = data.Parts[iPart];
        for (var iGroup in part.Groups) {
            var group = part.Groups[iGroup];
            for (var iObject in group.Objects) {
                var object = group.Objects[iObject];

				//for (var iSkin in object.Blocks) {
					var iSkin = 0;
					var skin = object.Blocks[iSkin];
					var objName = "p" + iPart + "_g" + iGroup + "_o" + iObject + "_s" + iSkin;
					
					var meshes = [];
					for (var iBlock in skin) {
						var block = skin[iBlock];
						
						var mesh = parseMeshPart(object, block)
						meshes.push(mesh);
						model.addMesh(mesh);
					}

					if (table) {
						var label = $('<label>');
						var chbox = $('<input type="checkbox" checked>');
						var td = $('<td>').append(label);
						chbox.click(meshes, function(ev) {
							for (i in ev.data) {
								ev.data[i].setVisible(this.checked);
							}
							gr_instance.requestRedraw();
						});
						td.mouseenter([model, meshes],function(ev) {
							ev.data[0].showExclusiveMeshes(ev.data[1]);
							gr_instance.requestRedraw();
						}).mouseleave(model, function(ev, a) {
							ev.data.showExclusiveMeshes();
							gr_instance.requestRedraw();
						});
						label.append(chbox);
						label.append("o_" + objName);
						table.append($('<tr>').append(td));
						
						//var params = ("00000000000000000000000000000000" + object.Params[7].toString(2)).substr(-32);
						
						/*
						var params = '';
						for (var i in object.Params) {
							if (i % 2 == 0) {
								params += '</br>';
							}
							params += '0x' + object.Params[i].toString(0x10) + ',';
						}
						
						td.append(params);
						*/
					}
				//}
            }
        }
    }
	return table;
}

function summaryLoadWadMesh(data, wad, nodeid) {
	gr_instance.destroyModels();
    set3dVisible(true);
   
	var mdl = new grModel();

	var dumplink = getActionLinkForWadNode(wad, nodeid, 'obj');
	dataSummary.append($('<a class="center">').attr('href', dumplink).append('Download .obj (xyz+norm+uv)'));

	var table = loadMeshFromAjax(mdl, data, true);
	dataSummary.append(table);
	
	gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function loadMdlFromAjax(mdl, data, parseScripts=false, needTable=false) {
	var table = undefined;
	if (data.Meshes && data.Meshes.length) {
		table = loadMeshFromAjax(mdl, data.Meshes[0], needTable);
	}

	for (var i in data.Materials) {
		var material = new grMaterial();
		
		var textures = data.Materials[i].TexturesBlended;
		var rawMat = data.Materials[i].Mat;
		if (rawMat && rawMat.Color) {
			material.setColor(rawMat.Color);
		}
		var layerId = undefined;
		if (rawMat.Layers && rawMat.Layers.length) {
			for (var i in rawMat.Layers) {
				layerId = i;
				if (rawMat.Layers[i].ParsedFlags.RenderingStrangeBlended === true) {
					break;
				}
			}
		}
		if (layerId !== undefined) {
			var zl = rawMat.Layers[layerId];
			if (zl.ParsedFlags.RenderingStrangeBlended === true) { material.setMethodUnknown(); }
			if (zl.ParsedFlags.RenderingSubstract === true) { material.setMethodSubstract(); }
			if (zl.ParsedFlags.RenderingUsual === true) { material.setMethodNormal(); }
			if (zl.ParsedFlags.RenderingAdditive === true) { material.setMethodAdditive(); }

	        if (textures && textures.length && textures[layerId]) {
	            var imgs = textures[layerId].Images;
	            if (imgs && imgs.length && imgs[0]) {
					var img = imgs[0].Image;
					if (rawMat.Layers[layerId].ParsedFlags.RenderingStrangeBlended === true) {
						console.log('COLORONLY ONE');
						img = imgs[0].ColorOnly;
					}
					material.setDiffuse(new grTexture('data:image/png;base64,' + img));
					material.setHasAlphaAttribute(textures[layerId].HaveTransparent);
	            }
	        }
		}
		mdl.addMaterial(material);
    }
	
	if (parseScripts) {
		console.log(data.Scripts);
		for (var i in data.Scripts) {
			var scr = data.Scripts[i];
			switch (scr.TargetName) {
				case "SCR_Sky":
					mdl.setType("sky");
					break;
				default:
					console.warn("Unknown SCR target: " + scr.TargetName, data, mdl, scr);
					break;
			}
		}
	}
	return table;
}

function summaryLoadWadMdl(data, wad, nodeid) {
	gr_instance.destroyModels();
    set3dVisible(true);
    
	var mdl = new grModel();
    	
    var table = $('<table>');
    if (data.Raw) {
        $.each(data.Raw, function(k, val) {
            switch (k) {
                case 'UnkFloats':
                case 'Someinfo':
                    val = JSON.stringify(val);
                    break;
                default:
                    break;
            }
            table.append($('<tr>').append($('<td>').append(k)));
            table.append($('<tr>').append($('<td>').append(val)));
        });
    }
    dataSummary.append(table);
    
	var dumplink = getActionLinkForWadNode(wad, nodeid, 'zip');
	dataSummary.append($('<a class="center">').attr('href', dumplink).append('Download .zip(obj+mtl+png)'));
		
	var table = loadMdlFromAjax(mdl, data, false, true);
	dataSummary.append(table);
	
	gr_instance.models.push(mdl);    
    gr_instance.requestRedraw();
}

function summaryLoadWadTxr(data) {
    set3dVisible(false);
    var table = $('<table>');
    $.each(data.Data, function(k, val) {
		if (k == 'Flags') {
			val = '0x' + val.toString(16);
		}
        table.append($('<tr>')
            .append($('<td>').append(k))
            .append($('<td>').append(val)));
    });
	
	table.append($('<tr>').append($('<td>').attr('colspan', 2).append('Parsed flags')));
	
	$.each(data, function(k, val) {
		if (k != 'Data' && k != 'Images' && k != 'Refs') {
			table.append($('<tr>')
		        .append($('<td>').append(k))
		        .append($('<td>').append(val.toString())));
		}
	});

    dataSummary.append(table);
    for (var i in data.Images) {
        var img = data.Images[i];
        dataSummary.append($('<img>')
				.addClass('no-interpolate')
                .attr('src', 'data:image/png;base64,' + img.Image)
                .attr('alt', 'gfx:' + img.Gfx + '  pal:' + img.Pal));
    }
}

function summaryLoadWadMat(data) {
    set3dVisible(false);
    var clr = data.Mat.Color;
    var clrBgAttr = 'background-color: rgb('+parseInt(clr[0]*255)+','+parseInt(clr[1]*255)+','+parseInt(clr[2]*255)+')';

    var table = $('<table>');
    table.append($('<tr>')
        .append($('<td>').append('Color'))
        .append($('<td>').attr('style', clrBgAttr).append(
            JSON.stringify(clr, undefined, 2)
        ))
    );

    for (var l in data.Mat.Layers) {
        var layer = data.Mat.Layers[l];
        var ltable = $('<table>')

        $.each(layer, function(k, v) {
            var td = $('<td>');
            switch (k) {
				case 'Flags':
					var str = '';
					for (var i in v) {
						str = str + '0x' + v[i].toString(0x10) + ', ';
					}
					td.append(str);
					break;
                case 'BlendColor':
					var r = Array(4);
					for (var i in data.Mat.Color) {
						r[i] = v[i] * data.Mat.Color[i];
					}
                    td.attr('style', 'background-color: rgb('+parseInt(r[0]*255)+','+parseInt(r[1]*255)+','+parseInt(r[2]*255)+')')
						.append(JSON.stringify(v, undefined, 2) + ';  result:' + JSON.stringify(r, undefined, 2));
                    break;
                case 'Texture':
                    td.append(v);
                    if (v != '') {
                        var txrobj = data.Textures[l];
						var txrblndobj = data.TexturesBlended[l];
                        td.append(' \\ ' + txrobj.Data.GfxName + ' \\ ' + txrobj.Data.PalName).append('<br>');
						td.append('Color + Alpha \\ Color only \\ Alpha(green=100%) ').append('<br>');
                        td.append($('<img>').attr('src', 'data:image/png;base64,' + txrobj.Images[0].Image));
						td.append($('<img>').attr('src', 'data:image/png;base64,' + txrobj.Images[0].ColorOnly));
						td.append($('<img>').attr('src', 'data:image/png;base64,' + txrobj.Images[0].AlphaOnly));
						td.append('<br>').append(' BLENDED Color + Alpha \\ BLENDED Color only ').append('<br>');
						td.append($('<img>').attr('src', 'data:image/png;base64,' + txrblndobj.Images[0].Image));
						td.append($('<img>').attr('src', 'data:image/png;base64,' + txrblndobj.Images[0].ColorOnly));
                    }
                    break;
				case 'ParsedFlags':
					td.append(JSON.stringify(v, null, "  ").replaceAll('\n', '<br>'));
					break;
                default:
                    td.append(v);
                    break;
            }
            ltable.append($('<tr>').append($('<td>').append(k)).append(td));
        });

        table.append($('<tr>')
            .append($('<td>').append('Layer ' + (l+1)))
            .append($('<td>').append(ltable))
        );
    };

    dataSummary.append(table);
}

function loadCollisionFromAjax(mdl, data) {
	if (data.ShapeName == "BallHull") {
		var vec = data.Shape.Vector;
		mdl.addMesh(grHelper_SphereLines(vec[0], vec[1], vec[2], vec[3]*2, 7, 7));
	}
}

function loadObjFromAjax(mdl, data, parseScripts=false) {
	if (data.Model) {
		loadMdlFromAjax(mdl, data.Model, parseScripts);
	} else if (data.Collision) {
		loadCollisionFromAjax(mdl, data.Collision);
	}	
	mdl.loadSkeleton(data.Data.Joints);
}

function summaryLoadWadObj(data) {
	gr_instance.destroyModels();

    var jointsTable = $('<table>');

	$.each(data.Data.Joints, function(joint_id, joint) {
		var row = $('<tr>').append($('<td>').attr('style','background-color:rgb('+
						parseInt((joint.Id % 8) * 15) + ',' +
						parseInt(((joint.Id / 8) % 8) * 15) + ',' +
						parseInt(((joint.Id / 64) % 8) * 15) + ');')
						.append(joint.Id).attr("rowspan", 7*2));
		
		for (var k in joint) {
			if (k === "Name"
				|| k === "IsSkinned"
				|| k === "OurJointToIdleMat"
				|| k === "ParentToJoint"
				|| k === "BindToJointMat"
				|| k === "RenderMat"
				|| k === "Parent") {
				row.append($('<td>').text(k));
				jointsTable.append(row);
				jointsTable.append($('<tr>').append($('<td>').text(JSON.stringify(joint[k]))));
				var row = $('<tr>');
			}
		}
		jointsTable.append(row);
	});
	dataSummary.append(jointsTable);
	
	if (data.Model || data.Collision) {
    	set3dVisible(true);
    
		var mdl = new grModel();
		loadObjFromAjax(mdl, data);
		
		gr_instance.models.push(mdl);    
	    gr_instance.requestRedraw();
	} else {
		set3dVisible(false);
	}
}


function summaryLoadWadGameObject(data) {
	gr_instance.destroyModels();
	set3dVisible(false);
    var table = $('<table>');
	for (var k in data) {
		table.append($('<tr>').append($('<td>').text(k)).append($('<td>').text(JSON.stringify(data[k]))));
	}
	dataSummary.append(table);
}

function loadCxtFromAjax(data, parseScripts=true) {
	for (var i in data.Instances) {
		var inst = data.Instances[i];
		var obj = data.Objects[inst.Object];

		var rs = 180.0/Math.PI;
		var rot = quat.fromEuler(quat.create(), inst.Rotation[0]*rs, inst.Rotation[1]*rs, inst.Rotation[2]*rs);
		
		//var instMat = mat4.fromTranslation(mat4.create(), inst.Position1);
		//instMat = mat4.mul(mat4.create(), instMat, mat4.fromQuat(mat4.create(), rot));

		// same as above
		var instMat = mat4.fromRotationTranslation(mat4.create(), rot, inst.Position1);

		//console.log(inst.Object, instMat);
		//if (obj && (obj.Model || (obj.Collision && inst.Object.includes("deathzone")))) {
		//if (obj && (obj.Model)) {
		if (obj && (obj.Model || obj.Collision)) {
			var mdl = new grModel();
			loadObjFromAjax(mdl, obj, parseScripts);
			mdl.matrix = instMat;
			gr_instance.models.push(mdl);
		}
	}
}

function summaryLoadWadCxt(data) {
	gr_instance.destroyModels();
   	set3dVisible(true);
   
	loadCxtFromAjax(data);
	
    gr_instance.requestRedraw();
}

function summaryLoadWadSbk(data, wad, nodeid) {
	set3dVisible(false);
	var list = $("<ul>");
	for (var i = 0; i < data.Sounds.length; i++) {
		var snd = data.Sounds[i];
		var link = '/action/' + wad +'/' + nodeid + '/';
		
		var getSndLink = function(type) {
			return getActionLinkForWadNode(wad, nodeid, type, 'snd='+snd.Name);
		};

		var vaglink = $("<a>").append(snd.Name).attr('href', getSndLink('vag'));
		var wavlink = $("<audio controls>").attr("preload","none").append($("<source>").attr("src", getSndLink('wav')));	

		var li = $("<li>").append(vaglink);

		if (data.IsVagFiles) {
			li.append("<br>").append(wavlink);
		}
		list.append(li);
	}
	dataSummary.append(list);
}

$(document).ready(function(){
    viewPack = $('#view-pack');
    viewTree = $('#view-tree');
    viewSummary = $('#view-summary');
    view3d = $('#view-3d');
    
    dataPack = viewPack.children('.view-item-container');
    dataTree = viewTree.children('.view-item-container');
    dataSummary = viewSummary.children('.view-item-container');
    data3d = view3d.children('.view-item-container');
    

	$('#view-pack-filter').on('input', treeInputFilterHandler).val('.wad');
	$('#view-item-filter').on('input', treeInputFilterHandler);

	
	var urlParts = decodeURI(window.location.hash).split("/");
	if (urlParts.length > 1) {
		if (urlParts[1].length > 0) {
			defferedLoadingWad = urlParts[1];
		}
	}
	if (urlParts.length > 2) {
		if (urlParts[2].length > 0) {
			defferedLoadingWadNode = urlParts[2];
		}
	}
	
    packLoad();
    
    gwInitRenderer(data3d);
});


function summaryLoadWadGeomShape(data) {
	gr_instance.destroyModels();
    set3dVisible(true);
    
	var m_vertexes = [];
	m_vertexes.length = data.Vertexes.length * 3;
	for (var i in data.Vertexes) {
		var j = i * 3;
		var v = data.Vertexes[i];
		m_vertexes[j] = v.Pos[0];
		m_vertexes[j+1] = v.Pos[1];
		m_vertexes[j+2] = v.Pos[2];
	}
	
	var m_indexes = [];
	m_indexes.length = data.Indexes.length * 3;
	for (var i in data.Indexes) {
		var j = i * 3;
		var v = data.Indexes[i];
		m_indexes[j] = v.Indexes[0];
		m_indexes[j+1] = v.Indexes[1];
		m_indexes[j+2] = v.Indexes[2];
	}
	
	var mdl = new grModel();	
	mdl.addMesh(new grMesh(m_vertexes, m_indexes));

	gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function summaryLoadWadScript(data) {
	gr_instance.destroyModels();
	
	dataSummary.append($("<h3>").append("Scirpt " + data.TargetName));
	
	if (data.TargetName == 'SCR_Entities') {
		for (var i in data.Data.Array) {
			var e = data.Data.Array[i];
			
			var ht = $("<table>").append($("<tr>").append($("<td>").attr("colspan", 2).append(e.Name)));
			for (var j in e) {
				var v = e[j];
				if (j == "Handlers") {
					for (var hi in v) {
						ht.append(
							$("<tr>").append($("<td>").append('Handler #' + hi))
								     .append($("<td>").append(v[hi].Decompiled.replaceAll('\n', '<br>'))));
					}
				} else {
					if (j == "Matrix") {
						v = JSON.stringify(v);
					}
					ht.append(
						$("<tr>").append($("<td>").append(j))
								 .append($("<td>").append(v)));
				}
			}
			
			dataSummary.append(ht);
		}
	}
	
	set3dVisible(false);
}