'use strict';

function treeLoadWadAsNodes(wadName, data) {
	if (defferedLoadingWadNode) {
        treeLoadWadNode(wadName, parseInt(defferedLoadingWadNode));
        defferedLoadingWadNode = undefined;
    } else {
        setLocation(wadName, '#/' + wadName);
    }
	dataTree.empty();

    var addNodes = function(nodes) {
        var ol = $('<ol>');
        for (var sn in nodes) {
            var node = data.Nodes[nodes[sn]];
            var li = $('<li>')
                .attr('nodeid', node.Tag.Id)
                .attr('nodename', node.Tag.Name)
                .attr('nodetag', node.Tag.Tag)
                .append($('<label>').append(("0000" + node.Tag.Id).substr(-4, 4) + '.' + node.Tag.Name));

            if (node.Tag.Tag == 30) {
                if (node.Tag.Size == 0) {
                    li.addClass('wad-node-link');
                } else {
                    li.addClass('wad-node-data');
                }
            } else {
                li.append(' [' + node.Tag.Tag + ']');
            }

			li.append($('<div>')
                .addClass('button-upload')
				.attr('title', 'Upload your version of wad tag data')
				.attr('href', '/upload/pack/' + wadName + '/' + node.Tag.Id)
				.click(uploadAjaxHandler));

            li.append($('<a>')
                .addClass('button-dump')
				.attr('title', 'Download wad tag data')
                .attr('href', '/dump/pack/' + wadName + '/' + node.Tag.Id))

            if (node.SubGroupNodes) {
                li.append(addNodes(node.SubGroupNodes));
            }
            ol.append(li);
        }
        return ol;
    }
	
    if (data.Roots) {
        dataTree.append(addNodes(data.Roots));
    }

    $('#view-item-filter').trigger('input');
    $('#view-tree ol li label').click(function(ev) {
        var node_element = $(this).parent();
        treeLoadWadNode(wadName, parseInt(node_element.attr('nodeid')));
    });
}

function treeLoadWadAsTags(wadName, data) {
	dataTree.empty();

	console.log(data);

	var ol = $('<ol>');
	for (var tagId in data.Tags) {
		var tag = data.Tags[tagId];
		var li = $('<li>')
            .attr('tagid', tag.Id)
            .attr('tagname', tag.Name)
            .attr('tagtag', tag.Tag)
		    .append($('<label>').append(("0000" + tag.Id).substr(-4, 4) + '.[' + ("000" + tag.Tag).substr(-3, 3) + ']' + tag.Name));
		
		if (tag.Tag == 30) {
			if (tag.Size == 0) {
	            li.addClass('wad-node-link');
	        } else {
	            li.addClass('wad-node-data');
	        }
		} else {
			if (tag.Size == 0) {
				li.addClass('wad-node-nodata');
			}
		}

		ol.append(li);
	}

	dataTree.append(ol);

    $('#view-item-filter').trigger('input');
    $('#view-tree ol li label').click(function(ev) {
        var node_element = $(this).parent();
        treeLoadWadTag(wadName, parseInt(node_element.attr('tagid')));
    });
}

function treeLoadWadNode(wad, tagid) {
    dataSummary.empty();

    $.getJSON('/json/pack/' + wad + '/' + tagid, function(resp) {
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
            setLocation(wad + " => " + tag.Name, '#/' + wad + '/' + tagid);

            if (tag.Tag == 0x1e) {
                switch (resp.ServerId) {
                    case 0x00000021: // flp
                        set3dVisible(false);
                        needMarshalDump = true;
                        break;
                    case 0x00000018: // sbk blk
                    case 0x00040018: // sbk vag
                        summaryLoadWadSbk(data, wad, tagid);
                        needMarshalDump = true;
                        break;
                    case 0x00000007: // txr
                        summaryLoadWadTxr(data, wad, tagid);
                        break;
                    case 0x00000008: // material
                        summaryLoadWadMat(data);
                        break;
                    case 0x00000011: // collision
                        gr_instance.cleanup();
                        set3dVisible(true);

                        var mdl = new grModel();
                        loadCollisionFromAjax(mdl, data);

                        gr_instance.models.push(mdl);
                        gr_instance.requestRedraw();
                        break;
                    case 0x0001000f: // mesh
                        summaryLoadWadMesh(data, wad, tagid);
                        break;
                    case 0x0002000f: // mdl
                        summaryLoadWadMdl(data, wad, tagid);
                        break;
                    case 0x00040001: // obj
                        summaryLoadWadObj(data, wad, tagid);
                        break;
                    case 0x80000001: // cxt
                        summaryLoadWadCxt(data, wad, tagid);
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
            displayResourceHexDump(wad, tagid);
        }
    });
}

function treeLoadWadTag(wad, tagid) {
    dataSummary.empty();
	gr_instance.cleanup();
    set3dVisible(false);
    displayResourceHexDump(wad, tagid);
}

function displayResourceHexDump(wad, tagid) {
	$.ajax({
       url: '/dump/pack/' + wad + '/' + tagid,
       type: 'GET',
       dataType: 'binary',
       processData: false,
       success: function(blob) {
           var fileReader = new FileReader();
           fileReader.onload = function() {
               var arr = new Uint8Array(this.result);
               dataSummary.append($("<h5>").append('Size in bytes:' + arr.length));
               dataSummary.append(hexdump(arr));
           };
           fileReader.readAsArrayBuffer(blob);
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
        m_vertexes[j + 1] = block.Trias.Y[i];
        m_vertexes[j + 2] = block.Trias.Z[i];
        if (!block.Trias.Skip[i]) {
            m_indexes.push(i - 1);
            m_indexes.push(i - 2);
            m_indexes.push(i - 0);
        }
    }

    var mesh = new grMesh(m_vertexes, m_indexes);

    if (block.Blend.R && block.Blend.R.length) {
        var m_colors = [];
        m_colors.length = block.Blend.R.length * 4;
        for (var i in block.Blend.R) {
            var j = i * 4;
            m_colors[j] = block.Blend.R[i];
            m_colors[j + 1] = block.Blend.G[i];
            m_colors[j + 2] = block.Blend.B[i];
            m_colors[j + 3] = block.Blend.A[i];
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
            m_textures[j + 1] = block.Uvs.V[i];
        }
        mesh.setUVs(m_textures);
    }

    if (block.Norms.X && block.Norms.X.length) {
        m_normals = [];
        m_normals.length = block.Norms.X.length * 3;

        for (var i in block.Norms.X) {
            var j = i * 3;
            m_normals[j] = block.Norms.X[i];
            m_normals[j + 1] = block.Norms.Y[i];
            m_normals[j + 2] = block.Norms.Z[i];
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
                    td.mouseenter([model, meshes], function(ev) {
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
    gr_instance.cleanup();
    set3dVisible(true);

    var mdl = new grModel();

    var dumplink = getActionLinkForWadNode(wad, nodeid, 'obj');
    dataSummary.append($('<a class="center">').attr('href', dumplink).append('Download .obj (xyz+norm+uv)'));

    var table = loadMeshFromAjax(mdl, data, true);
    dataSummary.append(table);

    gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function loadMdlFromAjax(mdl, data, parseScripts = false, needTable = false) {
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
            if (zl.ParsedFlags.RenderingStrangeBlended === true) {
                material.setMethodUnknown();
            }
            if (zl.ParsedFlags.RenderingSubstract === true) {
                material.setMethodSubstract();
            }
            if (zl.ParsedFlags.RenderingUsual === true) {
                material.setMethodNormal();
            }
            if (zl.ParsedFlags.RenderingAdditive === true) {
                material.setMethodAdditive();
            }

            if (textures && textures.length && textures[layerId]) {
                var imgs = textures[layerId].Images;
                if (imgs && imgs.length && imgs[0]) {
                    //var img = imgs[imgs.length-1].Image;
                    var img = imgs[0].Image;
                    if (rawMat.Layers[layerId].ParsedFlags.RenderingStrangeBlended === true) {
                        console.log('COLORONLY ONE');
                        img = img.ColorOnly;
                    }
                    material.setDiffuse(new grTexture('data:image/png;base64,' + img));
                    material.setHasAlphaAttribute(textures[layerId].HaveTransparent);
                }
            }
        }
        mdl.addMaterial(material);
    }

    if (parseScripts) {
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
    gr_instance.cleanup();
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

function summaryLoadWadTxr(data, wad, nodeid) {
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

    var form = $('<form action="' + getActionLinkForWadNode(wad, nodeid, 'upload') + '" method="post" enctype="multipart/form-data">');
    form.append($('<input type="file" name="img">'));
    var replaceBtn = $('<input type="button" value="Replace texture">')
    replaceBtn.click(function() {
        var form = $(this).parent();
        $.ajax({
            url: form.attr('action'),
            type: 'post',
            data: new FormData(form[0]),
            processData: false,
            contentType: false,
            success: function(a1) {
                if (a1 !== "") {
                    alert('Error replacing: ' + a1);
                } else {
                    alert('Success!');
                    window.location.reload();
                }
            }
        });
    });
    form.append(replaceBtn);

    dataSummary.append(form);
}

function summaryLoadWadMat(data) {
    set3dVisible(false);
    var clr = data.Mat.Color;
    var clrBgAttr = 'background-color: rgb(' + parseInt(clr[0] * 255) + ',' + parseInt(clr[1] * 255) + ',' + parseInt(clr[2] * 255) + ')';

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
                    td.attr('style', 'background-color: rgb(' + parseInt(r[0] * 255) + ',' + parseInt(r[1] * 255) + ',' + parseInt(r[2] * 255) + ')')
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
            .append($('<td>').append('Layer ' + (l + 1)))
            .append($('<td>').append(ltable))
        );
    };

    dataSummary.append(table);
}

function loadCollisionFromAjax(mdl, data) {
    if (data.ShapeName == "BallHull") {
        var vec = data.Shape.Vector;
        mdl.addMesh(grHelper_SphereLines(vec[0], vec[1], vec[2], vec[3] * 2, 7, 7));
    }
}

function loadObjFromAjax(mdl, data, matrix = undefined, parseScripts = false) {
    if (data.Model) {
        loadMdlFromAjax(mdl, data.Model, parseScripts);
    } else if (data.Collision) {
        loadCollisionFromAjax(mdl, data.Collision);
    }

	if (data.Script) {
		if (data.Script.TargetName == "SCR_Entities") {
			$.each(data.Script.Data.Array, function(entity_id, entity) {
				var objMat = new Float32Array(data.Data.Joints[0].RenderMat);
				var entityMat = new Float32Array(entity.Matrix);

				if (matrix) {
					// obj = obj * transformMat
					objMat = mat4.mul(mat4.create(), matrix, objMat);
				}
				// mat = obj * entity
				var mat = mat4.mul(mat4.create(), objMat, entityMat);

				var pos = mat4.getTranslation(vec3.create(), mat);

				var text3d = new grTextMesh(entity.Name, pos[0], pos[1], pos[2], true);
				
				var alpha = 1;
				switch (entity_id % 3) {
					case 0:
						text3d.setColor(1, 0, 0, alpha);
						break;
					case 1:
						text3d.setColor(0, 1, 0, alpha);
						break;
					case 2:
						text3d.setColor(1, 1, 0, alpha);
						break;
				}
				gr_instance.texts.push(text3d);
			});
		}
	}

    mdl.loadSkeleton(data.Data.Joints);
	if (matrix) {
		mdl.matrix = matrix;
	}
}

function summaryLoadWadObj(data, wad, nodeid) {
    gr_instance.cleanup();

    var dumplink = getActionLinkForWadNode(wad, nodeid, 'zip');
    dataSummary.append($('<a class="center">').attr('href', dumplink).append('Download .zip(obj+mtl+png)'));

    var jointsTable = $('<table>');

    $.each(data.Data.Joints, function(joint_id, joint) {
        var row = $('<tr>').append($('<td>').attr('style', 'background-color:rgb(' +
                parseInt((joint.Id % 8) * 15) + ',' +
                parseInt(((joint.Id / 8) % 8) * 15) + ',' +
                parseInt(((joint.Id / 64) % 8) * 15) + ');')
            .append(joint.Id).attr("rowspan", 7 * 2));

        for (var k in joint) {
            if (k === "Name" ||
                k === "IsSkinned" ||
                k === "OurJointToIdleMat" ||
                k === "ParentToJoint" ||
                k === "BindToJointMat" ||
                k === "RenderMat" ||
                k === "Parent") {
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
    gr_instance.cleanup();
    set3dVisible(false);
    var table = $('<table>');
    for (var k in data) {
        table.append($('<tr>').append($('<td>').text(k)).append($('<td>').text(JSON.stringify(data[k]))));
    }
    dataSummary.append(table);
}

function loadCxtFromAjax(data, parseScripts = true) {
    for (var i in data.Instances) {
        var inst = data.Instances[i];
        var obj = data.Objects[inst.Object];

        var rs = 180.0 / Math.PI;
        var rot = quat.fromEuler(quat.create(), inst.Rotation[0] * rs, inst.Rotation[1] * rs, inst.Rotation[2] * rs);

        //var instMat = mat4.fromTranslation(mat4.create(), inst.Position1);
        //instMat = mat4.mul(mat4.create(), instMat, mat4.fromQuat(mat4.create(), rot));

        // same as above
        var instMat = mat4.fromRotationTranslation(mat4.create(), rot, inst.Position1);

        //console.log(inst.Object, instMat);
        //if (obj && (obj.Model || (obj.Collision && inst.Object.includes("deathzone")))) {
        //if (obj && (obj.Model)) {
        if (obj && (obj.Model || obj.Collision)) {
            var mdl = new grModel();
            loadObjFromAjax(mdl, obj, instMat, parseScripts);
            gr_instance.models.push(mdl);
        }
    }
}

function summaryLoadWadCxt(data, wad, nodeid) {
    gr_instance.cleanup();
    
	if (data.Instances !== null && data.Instances.length) {
		set3dVisible(true);
		loadCxtFromAjax(data);
		gr_instance.requestRedraw();
	} else {
		set3dVisible(false);
    	displayResourceHexDump(wad, nodeid);
	}
}

function summaryLoadWadSbk(data, wad, nodeid) {
    set3dVisible(false);
    var list = $("<ul>");
    for (var i = 0; i < data.Sounds.length; i++) {
        var snd = data.Sounds[i];
        var link = '/action/' + wad + '/' + nodeid + '/';

        var getSndLink = function(type) {
            return getActionLinkForWadNode(wad, nodeid, type, 'snd=' + snd.Name);
        };

        var vaglink = $("<a>").append(snd.Name).attr('href', getSndLink('vag'));
        var wavlink = $("<audio controls>").attr("preload", "none").append($("<source>").attr("src", getSndLink('wav')));

        var li = $("<li>").append(vaglink);

        if (data.IsVagFiles) {
            li.append("<br>").append(wavlink);
        }
        list.append(li);
    }
    dataSummary.append(list);
}

function summaryLoadWadGeomShape(data) {
    gr_instance.cleanup();
    set3dVisible(true);

    var m_vertexes = [];
    m_vertexes.length = data.Vertexes.length * 3;
    for (var i in data.Vertexes) {
        var j = i * 3;
        var v = data.Vertexes[i];
        m_vertexes[j] = v.Pos[0];
        m_vertexes[j + 1] = v.Pos[1];
        m_vertexes[j + 2] = v.Pos[2];
    }

    var m_indexes = [];
    m_indexes.length = data.Indexes.length * 3;
    for (var i in data.Indexes) {
        var j = i * 3;
        var v = data.Indexes[i];
        m_indexes[j] = v.Indexes[0];
        m_indexes[j + 1] = v.Indexes[1];
        m_indexes[j + 2] = v.Indexes[2];
    }

    var mdl = new grModel();
    mdl.addMesh(new grMesh(m_vertexes, m_indexes));

    gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function summaryLoadWadScript(data) {
    gr_instance.cleanup();

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
					switch (j) {
						case "Matrix":
						case "DependsEntitiesIds":
							v = JSON.stringify(v);
							break;
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