'use strict';

var viewPack, viewTree, viewSummary, view3d;
var dataPack, dataTree, dataSummary, data3d;
var defferedLoadingWad;
var defferedLoadingWadNode;

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
    $.getJSON('/json/pack', function(data) {
        var list = $('<ol>');
        for (var i in data.Files) {
            var fileName = data.Files[i].Name;
			if (fileName.endsWith("WAD")) {
	            list.append($('<li>')
	                    .attr('filename', fileName)
	                    .append($('<label>').append(fileName))
	                    .append($('<a download>')
	                            .addClass('button-dump')
	                            .attr('href', '/dump/pack/' + fileName)) );
			}
        }
        dataPack.append(list);
        
		if (defferedLoadingWad) {
			packLoadFile(defferedLoadingWad);
		}
		
        $('#view-pack ol li label').click(function(ev) {
            packLoadFile($(this).parent().attr('filename'));
        });

        console.log('pack loaded');
    })
}

function packLoadFile(filename) {
    dataTree.empty();
    $.getJSON('/json/pack/' + filename, function(data) {
        var ext = filename.slice(-3).toLowerCase();
        switch (ext) {
            case 'wad':
                treeLoadWad(data);  
                break;
            default:
                dataTree.append(JSON.stringify(data, undefined, 2).replace('\n', '<br>'));
                break;
        }
        console.log('pack file ' + filename + ' loaded');
    });
}

function treeLoadWad(data) {
	var addNodes = function(nodes) {
        var ol = $('<ol>').attr('wadname', data.Name);
        for (var sn in nodes) {
            var node = data.Nodes[nodes[sn]];
            var li = $('<li>')
                    .attr('nodeid', node.Id)
                    .attr('nodeformat', node.Format)
                    .attr('nodename', node.Name)
					.attr('nodetag', node.Tag)
                    .append($('<label>').append(("0000" + node.Id).substr(-4,4) + '.' + node.Name));

            if (node.IsLink) {
				li.addClass('wad-node-link');
            } else {
				if (node.Tag == 0x1e) {
					li.addClass('wad-node-data');
				} else {
					li.append(' [' + node.Tag + ']');
				}
                li.append($('<a download>')
                        .addClass('button-dump')
                        .attr('href', '/dump/pack/' + data.Name + '/' + node.Id))
                if (node.SubNodes) {
                    li.append(addNodes(node.SubNodes));
                }
            }
            ol.append(li);
        }
        return ol;
    }
    
    setTitle(viewTree, data.Name);
    
    if (data.Roots) {
        dataTree.append(addNodes(data.Roots));
		
		if (!defferedLoadingWadNode) {
			set3dVisible(true);
			gr_instance.destroyModels();
			for (var sn in data.Roots) {
				var node = data.Nodes[data.Roots[sn]];
				if (node.Name.startsWith("CXT_") && node.Format == 0x80000001) {
					$.getJSON('/json/pack/' + data.Name +'/' + node.Id, function(resp) {
						var data = resp.Data;
						var node = resp.Node;
						if (node.Format == 0x80000001) {
							loadCxtFromAjax(data);
							gr_instance.requestRedraw();
						}
						console.log("loaded wad part: " + node.Name); 
					});
				}
			}
		}
		
	}
    
	if (defferedLoadingWadNode) {
		treeLoadWadNode(data.Name, parseInt(defferedLoadingWadNode));
	} else {
		setLocation(data.Name, '#/' + data.Name);
	}
	
    $('#view-tree ol li label').click(function(ev) {
        var node_element = $(this).parent();
        treeLoadWadNode(dataTree.children().attr('wadname'), parseInt(node_element.attr('nodeid')));
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
        var node = resp.Node;
		
		var needHexDump = false;
		
        if (resp.error) {
            set3dVisible(false);
            setTitle(viewSummary, 'Error');
            dataSummary.append(resp.error);
			needHexDump = true;
        } else {
            setTitle(viewSummary, node.Name);
			setLocation(wad + " => " + node.Name, '#/' + wad + '/' + nodeid);
			
			if (node.Tag == 0x1e) {
	            switch (node.Format) {
	                case 0x00000007: // txr
	                    summaryLoadWadTxr(data);
	                    break;
	                case 0x00000008: // material
	                    summaryLoadWadMat(data);
	                    break;
	                case 0x0001000f: // mesh
	                    summaryLoadWadMesh(data);
	                    break;
	                case 0x0002000f: // mdl
	                    summaryLoadWadMdl(data);
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
	                case 0x0000000c: // gfx pal
	                default:
	                    set3dVisible(false);
	                    dataSummary.append($("<pre>").append(JSON.stringify(data, null, "  ").replace('\n', '<br>')));
						needHexDump = true;
	                    break;
	            }
			} else {
				needHexDump = true;
			}
            console.log('wad ' + wad + ' file (' + node.Name + ')' + node.Id + ' loaded. format:' + node.Format);
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

function summaryLoadWadMesh(data) {
	gr_instance.destroyModels();
    set3dVisible(true);
    
	var mdl = new grModel();
	
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
		
		var textures = data.Materials[i].Textures;
		var rawMat = data.Materials[i].Mat;
		if (rawMat && rawMat.Color) {
			material.setColor(rawMat.Color);
		}
		if (rawMat.Layers && rawMat.Layers.length) {
			var zl = rawMat.Layers[0];
			if (zl.ParsedFlags.RenderingStrangeBlendedd === true) { material.setMethodUnknown(); }
			if (zl.ParsedFlags.RenderingSubstract === true) { material.setMethodSubstract(); }
			if (zl.ParsedFlags.RenderingUsual === true) { material.setMethodNormal(); }
			if (zl.ParsedFlags.RenderingAdditive === true) { material.setMethodAdditive(); }
		}
        if (textures && textures.length && textures[0]) {
            var imgs = textures[0].Images;
            if (imgs && imgs.length && imgs[0]) {
				material.setDiffuse(new grTexture('data:image/png;base64,' + imgs[0].Image));
				material.setHasAlphaAttribute(textures[0].HaveTransparent);
            }
        }
		mdl.addMaterial(material);
    }
	
	if (parseScripts) {
		for (var i in data.Scripts) {
			var scr = data.Scripts[i];
			switch (scr.TargetScript) {
				case "SCR_Sky":
					mdl.setType("sky");
					break;
				default:
					console.warn("Unknown SCR target: " + scr.TargetScript, data, mdl, scr);
					break;
			}
		}
	}
	return table;
}

function summaryLoadWadMdl(data) {
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
    
	var table = loadMdlFromAjax(mdl, data, false, true);
	dataSummary.append(table);
	
	gr_instance.models.push(mdl);    
    gr_instance.requestRedraw();
}

function summaryLoadWadTxr(data) {
    set3dVisible(false);
    var table = $('<table>');
    $.each(data.Data, function(k, val) {
        table.append($('<tr>')
            .append($('<td>').append(k))
            .append($('<td>').append(val)));
    });
	table.append($('<tr>')
        .append($('<td>').append('Have transparent'))
        .append($('<td>').append(data.HaveTransparent?"true":"false")));
    table.append($('<tr>')
        .append($('<td>').append('Used gfx'))
        .append($('<td>').append(data.UsedGfx)));
    table.append($('<tr>')
        .append($('<td>').append('Used pal'))
        .append($('<td>').append(data.UsedPal)));

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
                case 'Floats':
                    td.append(JSON.stringify(v, undefined, 2));
                    break;
                case 'Texture':
                    td.append(v);
                    if (v != '') {
                        var txrobj = data.Textures[l];
                        td.append(' \\ ' + txrobj.Data.GfxName + ' \\ ' + txrobj.Data.PalName);
                        td.append('<br>').append($('<img>').attr('src', 'data:image/png;base64,' + txrobj.Images[0].Image));
						td.append('<br>').append($('<img>').attr('src', 'data:image/png;base64,' + txrobj.Images[0].AlphaOnly));
						td.append($('<img>').attr('src', 'data:image/png;base64,' + txrobj.Images[0].ColorOnly));
                    }
                    break;
				case 'ParsedFlags':
					td.append(JSON.stringify(v, null, "  ").replace('\n', '<br>'));
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

function loadObjFromAjax(mdl, data, parseScripts=false) {
	loadMdlFromAjax(mdl, data.Model, parseScripts);
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
	
	if (data.Model) {
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
		if (obj && obj.Model) {
			var mdl = new grModel();
			loadObjFromAjax(mdl, obj, true);
			var rs = 180.0/Math.PI;
			var rot = quat.fromEuler(quat.create(), inst.Rotation[0]*rs, inst.Rotation[1]*rs, inst.Rotation[2]*rs);
			var instMat = mat4.fromRotationTranslation(mat4.create(), rot, inst.Position1);
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

$(document).ready(function(){
    viewPack = $('#view-pack');
    viewTree = $('#view-tree');
    viewSummary = $('#view-summary');
    view3d = $('#view-3d');
    
    dataPack = viewPack.children('.view-item-container');
    dataTree = viewTree.children('.view-item-container');
    dataSummary = viewSummary.children('.view-item-container');
    data3d = view3d.children('.view-item-container');
    
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