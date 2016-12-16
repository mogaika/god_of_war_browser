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
                    .append($('<label>').append(("0000" + node.Id).substr(-4,4) + '.' + node.Name));
            
            if (node.IsLink) {
                // TODO: link visual
            } else {
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
    
    if (data.Roots)
        dataTree.append(addNodes(data.Roots));
    
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

function treeLoadWadNode(wad, nodeid) {
    dataSummary.empty();

    $.getJSON('/json/pack/' + wad +'/' + nodeid, function(resp) {
        var data = resp.Data;
        var node = resp.Node;

        if (resp.error) {
            set3dVisible(false);
            setTitle(viewSummary, 'Error');
            dataSummary.append(resp.error);
        } else {
            setTitle(viewSummary, node.Name);

			setLocation(wad + " => " + node.Name, '#/' + wad + '/' + nodeid);

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
                    dataSummary.append(JSON.stringify(data, undefined, 2).replace('\n', '<br>'));
                    break;
            }
            console.log('wad ' + wad + ' file (' + node.Name + ')' + node.Id + ' loaded. format:' + node.Format);
        }
    });
}

function loadMeshFromAjax(model, data) {
    for (var iPart in data.Parts) {
        var part = data.Parts[iPart];
        for (var iGroup in part.Groups) {
            var group = part.Groups[iGroup];
            for (var iObject in group.Objects) {
                var object = group.Objects[iObject];
                for (var iPacket in object.Packets) {
                    var packet = object.Packets[iPacket];
                    for (var iBlock in packet.Blocks) {
                        var block = packet.Blocks[iBlock];
                        
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
                        
                        if (block.Uvs.U && block.Uvs.U.length) {
                            m_textures = [];
                            m_textures.length = block.Uvs.U.length * 2;
                                
                            for (var i in block.Uvs.U) {
								var j = i * 2;
								m_textures[j] = block.Uvs.U[i];
								m_textures[j+1] = block.Uvs.V[i];
							}
							mesh.setUVs(m_textures, object.MaterialId);							
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
						
						if (block.Joints && block.Joints.length) {
							mesh.setJointIds(block.Joints);
						}
                        
                        model.addMesh(mesh);
                    }
                }
            }
        }
    }
}

function summaryLoadWadMesh(data) {
	gr_instance.destroyModels();
    set3dVisible(true);
    
	var mdl = new grModel();
	
	loadMeshFromAjax(mdl, data);
	
	gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function loadMdlFromAjax(mdl, data) {
	if (data.Meshes && data.Meshes.length) {
		loadMeshFromAjax(mdl, data.Meshes[0]);
	}

	for (var i in data.Materials) {
		var material = new grMaterial();
		
        var textures = data.Materials[i].Textures;
		var rawMat = data.Materials[i].Mat;
		if (rawMat && rawMat.Color) {
			material.setColor(rawMat.Color);
		}
        if (textures && textures.length && textures[0]) {
            var imgs = textures[0].Images;
            if (imgs && imgs.length && imgs[0]) {
				material.setDiffuse(new grTexture('data:image/png;base64,' + imgs[0].Image));
            }
        }
		
		mdl.addMaterial(material);
    }
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
    
	loadMdlFromAjax(mdl, data);
	
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
                case 'Floats':
                case 'Flags':
                    td.append(JSON.stringify(v, undefined, 2));
                    break;
                case 'Texture':
                    td.append(v);
                    if (v != '') {
                        var txrobj = data.Textures[l];
                        td.append('<br>').append(txrobj.Data.GfxName);
                        td.append('<br>').append(txrobj.Data.PalName);
                        td.append('<br>').append($('<img>').attr('src', 'data:image/png;base64,' + txrobj.Images[0].Image));
                    }
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

function loadObjFromAjax(mdl, data) {
	loadMdlFromAjax(mdl, data.Model);
	//mdl.matrix = mat4.clone(data.Data.Joints[0].ParentToJoint);
	mdl.loadSkeleton(data.Data.Joints);
}

function summaryLoadWadObj(data) {
	gr_instance.destroyModels();

    var jointsTable = $('<table>');

	$.each(data.Data.Joints, function(joint_id, joint) {
		var row = $('<tr>').append(
			$('<td>').append(joint.Id).attr("rowspan", 7*2)
		);
		
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

function summaryLoadWadCxt(data) {
	//console.log("cxt", data);
	gr_instance.destroyModels();
   	set3dVisible(true);
   
	for (var i in data.Instances) {
		var inst = data.Instances[i];
		var obj = data.Objects[inst.Object];
		//console.log(inst, obj);
		if (obj && obj.Model) {
			var mdl = new grModel();
			loadObjFromAjax(mdl, obj);
			var rot = quat.fromEuler(quat.create(), inst.Rotation[0], inst.Rotation[1], inst.Rotation[2]);
			var instMat = mat4.fromRotationTranslation(mat4.create(), rot, inst.Position1);
			//mdl.matrix = mat4.translate(mat4.create(), mdl.matrix, instMat);
			mdl.matrix = instMat;
			
			gr_instance.models.push(mdl);
		}
	}
	
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