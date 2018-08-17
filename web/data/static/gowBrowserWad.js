'use strict';

var gw_cxt_group_loading = false;
var flp_obj_view_history = [{
    TypeArrayId: 8,
    IdInThatTypeArray: 0
}];

function treeLoadWad_dumpButtons(li, wadName, tag) {
    li.append($('<div>')
        .addClass('button-upload')
        .attr('title', 'Upload your version of wad tag data')
        .attr('href', '/upload/pack/' + wadName + '/' + tag.Id)
        .click(uploadAjaxHandler));

    li.append($('<a>')
        .addClass('button-dump')
        .attr('title', 'Download wad tag data')
        .attr('href', '/dump/pack/' + wadName + '/' + tag.Id))
}

function treeLoadWadAsNodes(wadName, data) {
    wad_last_load_view_type = 'nodes';
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

            if (node.Tag.Tag == 30 || node.Tag.Tag == 1) {
                if (node.Tag.Size == 0) {
                    li.addClass('wad-node-link');
                } else {
                    li.addClass('wad-node-data');
                }
            } else {
                li.append(' [' + node.Tag.Tag + ']');
            }

            treeLoadWad_dumpButtons(li, wadName, node.Tag);

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
        var node_id = parseInt(node_element.attr('nodeid'));
        if (node_id !== 0) {
            gw_cxt_group_loading = false;
            treeLoadWadNode(wadName, node_id);
        } else {
            setLocation(wadName + " => " + node_element.attr('nodename'), '#/' + wadName + '/' + 0);
            dataSummary.empty();
            gr_instance.cleanup();
            $("#view-tree ol li").each(function(i, node) {
                gw_cxt_group_loading = true;
                var $node = $(node);
                if ($node.attr("nodename").startsWith("CXT_")) {
                    treeLoadWadNode(wadName, $node.attr("nodeid"));
                }
            });
        }
    });
}

function treeLoadWadAsTags(wadName, data) {
    wad_last_load_view_type = 'tags';
    dataTree.empty();

    var ol = $('<ol>');
    for (var tagId in data.Tags) {
        var tag = data.Tags[tagId];
        var li = $('<li>')
            .attr('tagid', tag.Id)
            .attr('tagname', tag.Name)
            .attr('tagtag', tag.Tag)
            .attr('tagflags', tag.Flags)
            .append($('<label>').append(("0000" + tag.Id).substr(-4, 4) + '.[' + ("000" + tag.Tag).substr(-3, 3) + ']' + tag.Name));

        if (tag.Tag == 30 || tag.Tag == 1) {
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

        treeLoadWad_dumpButtons(li, wadName, tag);

        ol.append(li);
    }

    dataTree.append(ol);

    $('#view-item-filter').trigger('input');
    $('#view-tree ol li label').click(function(ev) {
        var node_element = $(this).parent();
        var tagid = parseInt(node_element.attr('tagid'));
        treeLoadWadTag(wadName, tagid);
    });
}

function treeLoadWadNode(wad, tagid) {
    dataSummary.empty();
    dataSummarySelectors.empty();
    set3dVisible(false);

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
            if (!gw_cxt_group_loading) {
                setTitle(viewSummary, tag.Name);
                setLocation(wad + " => " + tag.Name, '#/' + wad + '/' + tagid);
            }

            if (tag.Tag == 0x1e || tag.Tag == 1) {
                switch (resp.ServerId) {
                    case 0x00000021: // flp
                        summaryLoadWadFlp(data, wad, tagid);
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
                    case 0x00000003: // anim
                        needMarshalDump = true;
                        needHexDump = false;
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

    var form = $('<form class="flexedform" action="' + getActionLinkForWadNode(wad, tagid, 'updatetag') + '" method="post">');
    var tagel = $('li[tagid=' + tagid + ']');
    var tbl = $('<table>');
    tbl.append($('<tr>').append($('<td>').text("tag type")).append($('<td>').append($('<input type="text" id="tagtag" name="tagtag" value="' + tagel.attr("tagtag") + '">'))));
    tbl.append($('<tr>').append($('<td>').text("name")).append($('<td>').append($('<input type="text" id="tagname" name="tagname" value="' + tagel.attr("tagname") + '">'))));
    tbl.append($('<tr>').append($('<td>').text("flags")).append($('<td>').append($('<input type="text" id="tagflags" name="tagflags" value="' + tagel.attr("tagflags") + '">'))));
    tbl.append($('<tr>').append($('<td>')).append($('<td>').append($('<input type="submit" value="Update tag info">'))));

    dataSummary.append(form.append(tbl));
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

function parseMeshPacket(object, packet) {
    var m_vertexes = [];
    var m_indexes = [];
    var m_colors;
    var m_textures;
    var m_normals;

    m_vertexes.length = packet.Trias.X.length * 3;

    for (var i in packet.Trias.X) {
        var j = i * 3;
        m_vertexes[j] = packet.Trias.X[i];
        m_vertexes[j + 1] = packet.Trias.Y[i];
        m_vertexes[j + 2] = packet.Trias.Z[i];
        if (!packet.Trias.Skip[i]) {
            m_indexes.push(i - 1);
            m_indexes.push(i - 2);
            m_indexes.push(i - 0);
        }
    }

    var mesh = new grMesh(m_vertexes, m_indexes);

    if (packet.Blend.R && packet.Blend.R.length) {
        var m_colors = [];
        m_colors.length = packet.Blend.R.length * 4;
        for (var i in packet.Blend.R) {
            var j = i * 4;
            m_colors[j] = packet.Blend.R[i];
            m_colors[j + 1] = packet.Blend.G[i];
            m_colors[j + 2] = packet.Blend.B[i];
            m_colors[j + 3] = packet.Blend.A[i];
        }

        mesh.setBlendColors(m_colors);
    }

    mesh.setMaterialID(object.MaterialId);

    if (packet.Uvs.U && packet.Uvs.U.length) {
        m_textures = [];
        m_textures.length = packet.Uvs.U.length * 2;

        for (var i in packet.Uvs.U) {
            var j = i * 2;
            m_textures[j] = packet.Uvs.U[i];
            m_textures[j + 1] = packet.Uvs.V[i];
        }
        mesh.setUVs(m_textures);
    }

    if (packet.Norms.X && packet.Norms.X.length) {
        m_normals = [];
        m_normals.length = packet.Norms.X.length * 3;

        for (var i in packet.Norms.X) {
            var j = i * 3;
            m_normals[j] = packet.Norms.X[i];
            m_normals[j + 1] = packet.Norms.Y[i];
            m_normals[j + 2] = packet.Norms.Z[i];
        }

        mesh.setNormals(m_normals);
    }

    if (packet.Joints && packet.Joints.length && object.JointMapper && object.JointMapper.length) {
        //console.log(packet.Joints, packet.Joints2, object.JointMapper);
        var joints1 = packet.Joints;
        var joints2 = (!!packet.Joints2) ? packet.Joints2 : undefined;
        mesh.setJointIds(object.JointMapper, joints1, joints2);
    }

    return mesh;
}

function loadMeshPartFromAjax(model, data, iPart, table = undefined) {
    var part = data.Parts[iPart];
    var totalMeshes = [];
    for (var iGroup in part.Groups) {
        var group = part.Groups[iGroup];
        for (var iObject in group.Objects) {
            var object = group.Objects[iObject];

            //var iSkin = 0;
            for (var iSkin in object.Packets) {
                var skin = object.Packets[iSkin];
                var objName = "p" + iPart + "_g" + iGroup + "_o" + iObject + "_s" + iSkin;

                var meshes = [];
                for (var iPacket in skin) {
                    var packet = skin[iPacket];
                    var mesh = parseMeshPacket(object, packet);
                    meshes.push(mesh);
                    totalMeshes.push(mesh);
                    model.addMesh(mesh);
                }

                if (table) {
                    var label = $('<label>');
                    var chbox = $('<input type="checkbox" checked>');
                    var td = $('<td>').append(label);
                    chbox.click(meshes, function(ev) {
                        for (var i in ev.data) {
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
                }
            }
        }
    }
    return totalMeshes;
}

function loadMeshFromAjax(model, data, needTable = false) {
    var table = needTable ? $('<table>') : undefined;
    for (var iPart in data.Parts) {
        loadMeshPartFromAjax(model, data, iPart, table);
    }
    gr_instance.flushScene();
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

    for (var iMaterial in data.Materials) {
        var material = new grMaterial();

        var textures = data.Materials[iMaterial].Textures;
        var rawMat = data.Materials[iMaterial].Mat;
        material.setColor(rawMat.Color);

        for (var iLayer in rawMat.Layers) {
            var rawLayer = rawMat.Layers[iLayer];
            var layer = new grMaterialLayer();

            layer.setColor(rawLayer.BlendColor);
            if (rawLayer.ParsedFlags.RenderingSubstract === true) {
                layer.setMethodSubstract();
            }
            if (rawLayer.ParsedFlags.RenderingUsual === true) {
                layer.setMethodNormal();
            }
            if (rawLayer.ParsedFlags.RenderingAdditive === true) {
                layer.setMethodAdditive();
            }
            if (rawLayer.ParsedFlags.RenderingStrangeBlended === true) {
                layer.setMethodUnknown();
            }

            if (textures && textures[iLayer] && textures[iLayer].Images) {
                var imgs = textures[iLayer].Images;
                var txrs = [];
                for (var iImg in imgs) {
                    txrs.push(new grTexture('data:image/png;base64,' + imgs[iImg].Image));
                }
                layer.setTextures(txrs);
                layer.setHasAlphaAttribute(textures[iLayer].HaveTransparent);
            }
            material.addLayer(layer);
        }
        mdl.addMaterial(material);

        var anim = data.Materials[iMaterial].Animations;
        if (anim && anim.Groups && anim.Groups.length) {
            var group = anim.Groups[0];
            if (!group.IsExternal && group.Acts && group.Acts.length) {
                for (var iAct in group.Acts) {
                    var act = group.Acts[iAct];
                    for (var dt in anim.DataTypes) {
                        switch (anim.DataTypes[dt].TypeId) {
                            case 8:
                                var animInstance = new gaMatertialLayerAnimation(anim, act, dt, material);
                                animInstance.enable();
                                ga_instance.addAnimation(animInstance);
                                break;
                            case 9:
                                var animSheetInstance = new gaMatertialSheetAnimation(anim, act, dt, material);
                                ga_instance.addAnimation(animSheetInstance);
                                break;
                        }
                    }
                }
            }
        }
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
            .append($('<td>').append('Layer ' + l))
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

                var radius = entity.Matrix[0];
                var text3d = new grTextMesh(entity.Name, pos[0], pos[1], pos[2], true);

                //var mdl = new grModel();
                //mdl.addMesh(new grHelper_SphereLines(pos[0], pos[1], pos[2], radius, 6, 6));

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
                //gr_instance.models.push(mdl);
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

    if (data.Animations) {
        var $animSelector = $("<select>").attr("size", 6).addClass("animation");

        var anim = data.Animations;
        if (anim && anim.Groups && anim.Groups.length) {
            for (var iGroup in anim.Groups) {
                var group = anim.Groups[iGroup];
                for (var iAct in group.Acts) {
                    var act = group.Acts[iAct];
                    for (var dt in anim.DataTypes) {
                        switch (anim.DataTypes[dt].TypeId) {
                            case 0:
                                var $option = $("<option>").text(group.Name + ": " + act.Name);
                                $option.dblclick([anim, act, dt, data.Data], function(ev) {
                                    var anim = new gaObjSkeletAnimation(ev.data[0], ev.data[1], ev.data[2], ev.data[3], gr_instance.models[0]);
                                    ga_instance.addAnimation(anim);
                                });

                                $animSelector.append($option);
                                break;
                        }
                    }
                }
            }

        }

        dataSummary.append($animSelector);
    }

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
    if (!gw_cxt_group_loading) {
        gr_instance.cleanup();
    }

    if ((data.Instances !== null && data.Instances.length) || gw_cxt_group_loading) {
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

function summaryLoadWadFlp(flp, wad, tagid) {
    var flpdata = flp.FLP;
    var flp_print_dump = function() {
        set3dVisible(false);
        dataSummary.empty();
        dataSummary.append($("<pre>").append(JSON.stringify(flpdata, null, "  ").replaceAll('\n', '<br>')));
    }

    var flp_scripts_strings = function() {
        set3dVisible(false);
        dataSummary.empty();

        var tbody = $("<tbody>");
        for (var iRef in flp.ScriptPushRefs) {
            var tr = $("<tr>");
            var ref = flp.ScriptPushRefs[iRef];

            var str = atob(ref.String);
            if (flp.FontCharAliases) {
                var originalStr = str;
                str = "";
                for (var i = 0; i < originalStr.length; i++) {
                    var charCode = originalStr.charCodeAt(i);
                    var replaced = false;

                    for (var charToReplace in flp.FontCharAliases) {
                        if (flp.FontCharAliases[charToReplace] === charCode) {
                            str += String.fromCharCode(charToReplace);
                            replaced = true;
                            break;
                        }
                    }

                    if (replaced === false) {
                        str += String.fromCharCode(charCode);
                    }
                }
            }

            tr.append($("<td>").text(iRef));
            tr.append($("<td>").append($("<input type=text>").val(str).css("width", "100%")));
            tr.append($("<td>").append($("<button>").text("Update").click(
                function(ev) {
                    var str = $(this).parent().parent().find('input[type="text"]').val();
                    var id = Number.parseInt($(this).parent().parent().children().first().text());

                    if (flp.FontCharAliases) {
                        var originalStr = str;
                        str = "";
                        for (var char of originalStr) {
                            if (flp.FontCharAliases.hasOwnProperty(char.charCodeAt(0))) {
                                str += String.fromCharCode(flp.FontCharAliases[char.charCodeAt(0)]);
                            } else {
                                str += char;
                            }
                        }
                    }

                    $.ajax({
                        url: getActionLinkForWadNode(wad, tagid, 'scriptstring'),
                        data: {
                            'id': id,
                            'string': btoa(str)
                        },
                        success: function(a) {
                            if (a != "" && a.error) {
                                alert("Error: " + a.error);
                            } else {
                                alert("Success");
                            }
                        }
                    });
                }
            )));
            tbody.append(tr);
        }

        var headtr = $("<tr>");
        headtr.append($("<td>").text("Id"));
        headtr.append($("<td>").text("Text"));
        headtr.append($("<td>"));

        dataSummary.append($("<table>").width("100%").append($("<thead>").append(headtr)).append(tbody));
    }

    let print_static_label_as_tr = function(iSl, needref = true) {
        let sl = flpdata.StaticLabels[iSl];
        let row = $("<tr>");

        if (needref) {
            row.append($("<td>").append($("<a>").addClass('flpobjref').text("id " + iSl).click(function() {
                flp_obj_view_history.unshift({
                    TypeArrayId: 4,
                    IdInThatTypeArray: iSl
                });
                flp_view_object_viewer();
            })));
        }

        var font = undefined;
        var cmdsContainer = $("<td>");
        for (var iCmd in sl.RenderCommandsList) {
            var rcmds = $("<table width='100%'>");
            var cmd = sl.RenderCommandsList[iCmd];

            if (cmd.Flags & 8) {
                var fhi = $("<input type=text id='fonthandler' class=no-width>").val(cmd.FontHandler);
                var fsi = $("<input type=text id='fontscale' class=no-width>").val(cmd.FontScale);
                let $link = $("<a>").addClass('flpobjref').text("handler ").click(function() {
                    flp_obj_view_history.unshift({
                        TypeArrayId: 3,
                        IdInThatTypeArray: cmd.FontHandler
                    });
                    flp_view_object_viewer();
                })
                rcmds.append($("<tr>").append($("<td>").text("Set font")).append($("<td>").append($link).append("#").append(fhi).append(" with scale ").append(fsi)));
                font = flpdata.Fonts[flpdata.GlobalHandlersIndexes[cmd.FontHandler].IdInThatTypeArray];
            }
            if (cmd.Flags & 4) {
                var bclri = $("<input type=text id='blendclr'>").val(JSON.stringify(cmd.BlendColor));
                rcmds.append($("<tr>").append($("<td>").text("Set blend color")).append($("<td>").append(bclri)));
            }
            var xoi = $("<input type=text id='xoffset'>").val(cmd.OffsetX);
            rcmds.append($("<tr>").append($("<td>").text("Set X offset")).append($("<td>").append(xoi)));
            var yoi = $("<input type=text id='yoffset'>").val(cmd.OffsetY);
            rcmds.append($("<tr>").append($("<td>").text("Set Y offset")).append($("<td>").append(yoi)));

            var str = cmd.Glyphs.reduce(function(str, glyph) {
                var char = font.CharNumberToSymbolIdMap.indexOf(glyph.GlyphId);
                if (flp.FontCharAliases) {
                    var map_chars = Object.keys(flp.FontCharAliases).filter(function(charString) {
                        return flp.FontCharAliases[charString] == char;
                    });
                    if (map_chars && map_chars.length !== 0) {
                        char = map_chars[0];
                    }
                }
                return str + (char > 0 ? String.fromCharCode(char) : ("$$" + glyph.GlyphId));
            }, '');

            rcmds.append($("<tr>").append($("<td>").text("Print glyphs")).append($("<td>").append($("<textarea>").val(str))));
            cmdsContainer.append(rcmds);
        }

        var open_preview_for_label = function(sl) {
            var u = new URLSearchParams();
            u.append('c', JSON.stringify(sl.RenderCommandsList));
            u.append('f', wad);
            u.append('r', tagid);

            var t = sl.Transformation;
            var m = t.Matrix;
            u.append('m', JSON.stringify([m[0], m[1], 0, 0, m[2], m[3], 0, 0, 0, 0, 1, 0, t.OffsetX, t.OffsetY, 0, 1]));
            window.open('/label.html?' + u, '_blank');
        }

        var get_label_from_table_tr = function(tr) {
            var sl = {
                'Transformation': JSON.parse(tr.find("td").last().text()),
                'RenderCommandsList': [],
            };

            var fontscale = 1.0;
            var fonthandler = -1;
            tr.find("table").each(function(cmdIndex, tbl) {
                var cmd = {
                    'Flags': 0
                };
                $(tbl).find("tr").each(function(i, row) {
                    var rname = $(row).find("td").first().text();
                    if (rname.includes("font")) {
                        cmd.Flags |= 8;
                        cmd.FontHandler = Number.parseInt($(row).find("#fonthandler").val());
                        cmd.FontScale = Number.parseFloat($(row).find("#fontscale").val());
                        fonthandler = cmd.FontHandler;
                        fontscale = cmd.FontScale;
                    } else if (rname.includes("blend")) {
                        cmd.Flags |= 4;
                        cmd.BlendColor = JSON.parse($(row).find("#blendclr").val());
                    } else if (rname.includes("X offset")) {
                        cmd.OffsetX = Number.parseFloat($(row).find("#xoffset").val());
                        if (Math.abs(cmd.OffsetX) > 0.000001) {
                            cmd.Flags |= 2;
                        }
                    } else if (rname.includes("Y offset")) {
                        cmd.OffsetY = Number.parseFloat($(row).find("#yoffset").val());
                        if (Math.abs(cmd.OffsetY) > 0.000001) {
                            cmd.Flags |= 1;
                        }
                    } else if (rname.includes("glyphs")) {
                        var text = $(row).find("textarea").val();
                        var glyphs = [];

                        var font = flpdata.Fonts[flpdata.GlobalHandlersIndexes[fonthandler].IdInThatTypeArray];
                        for (var char of text) {
                            var charCode = char.charCodeAt(0);
                            if (flp.FontCharAliases) {
                                if (flp.FontCharAliases.hasOwnProperty(charCode)) {
                                    charCode = flp.FontCharAliases[charCode];
                                }
                            }
                            var glyphId = font.CharNumberToSymbolIdMap[charCode];
                            var width = font.SymbolWidths[glyphId] * fontscale;
                            glyphs.push({
                                'GlyphId': glyphId,
                                'Width': width / 16
                            });
                        }
                        cmd.Glyphs = glyphs;
                    }
                });
                sl.RenderCommandsList.push(cmd);
            })
            return sl;
        }

        var btns = $("<div>");
        btns.append($("<button>peview original</button>").click(sl, function(e) {
            open_preview_for_label(e.data);
        }));
        btns.append($("<br>"));
        btns.append($("<button>preview changes</button>").click(function(e) {
            open_preview_for_label(get_label_from_table_tr($(this).parent().parent().parent()));
        }));
        btns.append($("<br>"));
        btns.append($("<button>apply changes</button>").click(iSl, function(e) {
            var sl = get_label_from_table_tr($(this).parent().parent().parent());

            $.post({
                url: getActionLinkForWadNode(wad, tagid, 'staticlabels'),
                data: {
                    'id': e.data,
                    'sl': JSON.stringify(sl)
                },
                success: function(a) {
                    if (a != "" && a.error) {
                        alert('Error uploading: ' + a.error);
                    } else {
                        alert('Success!');
                    }
                }
            });

        }));

        row.append($("<td>").append(cmdsContainer));
        row.append($("<td>").append(btns));
        row.append($("<td>").text(JSON.stringify(sl.Transformation)));
        return row;
    }

    var flp_list_labels = function() {
        set3dVisible(false);
        dataSummary.empty();

        var table = $("<table class='staticlabelrendercommandlist'>");
        table.append($("<tr>").append($("<td>").text("Id")).append($("<td>").text("Render commands")));

        for (let iSl in flpdata.StaticLabels) {
            table.append(print_static_label_as_tr(iSl));
        }

        dataSummary.append(table);
    }

    var flp_view_object_viewer = function() {
        dataSummary.empty();
        gr_instance.cleanup();
        set3dVisible(false);
        let $history_element = $("<div>").css('margin', '7px').css('white-space', 'nowrap').css('overflow', 'hidden');
        let $data_element = $("<div>");
        const objNamesArray = ['Nothing', 'Textured mesh part', 'UNKNOWN', 'Font',
            'Static label', 'Dynamic label', 'Data6', 'Data7',
            'Root', 'Transform', 'Color'
        ];
        let element_view = function(h) {
            $data_element.empty();
            if (h == undefined) {
                h = flp_obj_view_history[0];
            } else {
                flp_obj_view_history.unshift(h);
            } {
                $history_element.empty();
                $history_element.append($("<span>").text("History: ").css('padding', '6px'));
                let new_history = [h];
                for (let i in flp_obj_view_history) {
                    if (i != 0) {
                        if (flp_obj_view_history[i].IdInThatTypeArray != h.IdInThatTypeArray ||
                            flp_obj_view_history[i].TypeArrayId != h.TypeArrayId) {
                            new_history.push(flp_obj_view_history[i]);
                        }
                    }
                }
                flp_obj_view_history = new_history;
                if (flp_obj_view_history.length > 16) {
                    flp_obj_view_history.shift();
                }
                for (let i in flp_obj_view_history) {
                    let h = flp_obj_view_history[i];
                    let $a = $("<a>").text(objNamesArray[h.TypeArrayId] + "[" + h.IdInThatTypeArray + "] ");
                    $a.addClass('flpobjref').click(function() {
                        element_view(flp_obj_view_history[i]);
                    });
                    if (i == 0) {
                        $history_element.append(" > ", $a.css('color', 'white'), " <");
                    } else {
                        $history_element.append(" | ", $a);
                    }
                }
            }

            let get_obj_by_handler = function(h) {
                switch (h.TypeArrayId) {
                    //case 1: return flpdata.GlobalHandlersIndexes[h.IdInThatTypeArray]; break;
                    case 1:
                        return flpdata.MeshPartReferences[h.IdInThatTypeArray];
                        break;
                    case 2:
                        console.error("UNKNOWN element");
                        break;
                    case 3:
                        return flpdata.Fonts[h.IdInThatTypeArray];
                        break;
                    case 4:
                        return flpdata.StaticLabels[h.IdInThatTypeArray];
                        break;
                    case 5:
                        return flpdata.DynamicLabels[h.IdInThatTypeArray];
                        break;
                    case 6:
                        return flpdata.Datas6[h.IdInThatTypeArray];
                        break;
                    case 7:
                        return flpdata.Datas7[h.IdInThatTypeArray];
                        break;
                    case 8:
                        return flpdata.Data8;
                        break;
                    case 9:
                        return flpdata.Transformations[h.IdInThatTypeArray];
                        break;
                    case 10:
                        return flpdata.BlendColors[h.IdInThatTypeArray];
                        break;
                    default:
                        return undefined;
                        break;
                }
            }
            let obj = get_obj_by_handler(h);

            let _row = function() {
                return $("<tr>").append(Array.prototype.slice.call(arguments));
            }

            let _column = function() {
                return $("<td>").append(Array.prototype.slice.call(arguments));
            }

            let print_ref_handler = function(handler) {
                let $a = $("<a>").text('&' + objNamesArray[handler.TypeArrayId] + '[' + handler.IdInThatTypeArray + ']')
                $a.addClass('flpobjref');
                $a.click(function() {
                    element_view(handler);
                });
                switch (handler.TypeArrayId) {
                    case 1:
                        let mats = [];
                        let meshref = get_obj_by_handler(handler);
                        for (let i in meshref.Materials) {
                            let matname = meshref.Materials[i].TextureName;
                            if (matname != "") {
                                mats.push(matname);
                            }
                        }
                        if (mats.length != 0) {
                            return $("<div>").append($a, " (meshpart " + meshref.MeshPartIndex + ", textures: " + mats.join(",") + ")");
                        } else {
                            return $("<div>").append($a, " (meshpart " + meshref.MeshPartIndex + ", no textures used)");
                        }
                        break;
                    case 9:
                        let t = get_obj_by_handler(handler);
                        return $("<div>").append($a, " (x: ", t.OffsetX, " y: ", t.OffsetY, ")");
                    case 10:
                        let clr = get_obj_by_handler(handler).Color;
                        let css_rgb = "rgb(" + (clr[0] / 256.0) * 255 + "," + (clr[1] / 256.0) * 255 + "," + (clr[2] / 256.0) * 255;
                        let $rgb = $("<div>").addClass('flpcolorpreview').css('background-color', css_rgb);
                        let $rgba = $("<div>").addClass('flpcolorpreview').css('background-color', css_rgb).css('opacity', clr[3] / 256.0);
                        return $("<div>").append($a, " (without alpha: ", $rgb, " with alpha: ", $rgba, "  a: ", clr[3], ")");
                        break;
                }
                return $a;
            }

            let print_ref_handler_index = function(handler_index) {
                if (flpdata.GlobalHandlersIndexes[handler_index]) {
                    return print_ref_handler(flpdata.GlobalHandlersIndexes[handler_index]);
                } else {
                    return "%bad handler index " + handler_index + "%";
                }
            }

            let $data_table = $("<table>");

            let print_script = function(script) {
                let code = script.Decompiled;
                let $code_element = $("<div>").text(" > click to show decompiled script < ").css('cursor', 'pointer').click(function() {
                    $(this).empty().css('cursor', '').append(code).off('click');
                })
                return $code_element;
            }

            let print_data6 = function() {
                let $data6_table = print_data6_subtype1(obj.Sub1);
                let $events_table = $("<table>");

                $data_table.append(_row(_column($data6_table)), _row(_column($events_table)));
            }

            let print_data6_subtype1 = function(obj) {
                let $elements_table = $("<table>");
                let $scripts_table = $("<table>");

                for (let i in obj.ElementsAnimation) {
                    let el = obj.ElementsAnimation[i];
                    let $el = $("<div>");

                    let $frames_table = $("<table>");
                    for (let j in el.KeyFrames) {
                        let frame = el.KeyFrames[j];
                        let $frame = $("<table>");
                        $frame.append(_row(_column("name"), _column(frame.Name)));
                        $frame.append(_row(_column("frame end time"), _column(frame.WhenThisFrameEnds)));
                        $frame.append(_row(_column("element"), _column(print_ref_handler(frame.ElementHandler))));
                        $frame.append(_row(_column("color"), _column(print_ref_handler({
                            TypeArrayId: 10,
                            IdInThatTypeArray: frame.ColorId
                        }))));
                        $frame.append(_row(_column("transformation"), _column(print_ref_handler({
                            TypeArrayId: 9,
                            IdInThatTypeArray: frame.TransformationId
                        }))));

                        $frames_table.append(_row(_column("frame " + j), _column($frame)));
                    }
                    $el.append($frames_table);

                    $elements_table.append(_row(_column("element " + i), _column($el)));
                }

                for (let i in obj.FrameScriptLables) {
                    let script = obj.FrameScriptLables[i];
                    let $script = $("<div>");

                    $script.append(_row(_column("triggered after frame"), _column(script.TriggerFrameNumber)));
                    $script.append(_row(_column("name"), _column(script.LabelName)));
                    let $streams_table = $("<table>");
                    for (let iStream in script.Subs) {
                        $streams_table.append(_row(_column(print_script(script.Subs[iStream].Script))));
                    }
                    $script.append(_row(_column("threads"), _column($streams_table)));

                    $scripts_table.append(_row(_column("script " + i), _column($script)));
                }

                $data_table.append(_row(_column("elements"), _column($elements_table)), _row(_column("methods"), _column($scripts_table)));
            }

            let print_mesh = function(obj) {
                $data_table.append(_row(_column("Mesh part index "),
                    _column("<b>" + obj.MeshPartIndex + "</b><br><sub>You can open related MDL_%flpname% resource and check this object part (mesh that index starts with o_" + obj.MeshPartIndex + "_g0_...) </sub>")));
                let $materials = [];
                for (let i in obj.Materials) {
                    console.log(obj.Materials, obj, flp);
                    let mat = obj.Materials[i];
                    let $mat = $("<div>");
                    $mat.append("Color: <b>0x" + mat.Color.toString(16) + "</b><br>");
                    $mat.append("Texture name: <b>" + mat.TextureName + "</b><br>");
                    if (mat.TextureName != "") {
                        $mat.append($('<img>').addClass('no-interpolate').attr('src', 'data:image/png;base64,' + flp.Textures[mat.TextureName].Images[0].Image));
                    }
                    $materials.push(_row(_column("material " + i), _column($mat)));
                }
                $data_table.append($materials);
            }

            let print_transform = function(obj) {
                let $form = $("<div>");
                $data_table.append(_row(_column("Offset X"), _column($("<input id='x' type='text'>").val(obj.OffsetX))));
                $data_table.append(_row(_column("Offset Y"), _column($("<input id='y' type='text'>").val(obj.OffsetY))));
                let $matrix = $("<textarea id='matrix'>").css('height', '8em').val(JSON.stringify(obj.Matrix, null, ' '));
                $matrix.append("<sub>You can read about 2d matrices <a href='https://en.wikipedia.org/wiki/Transformation_matrix#Examples_in_2D_computer_graphics'>there</a></sub>")
                $data_table.append(_row(_column("Matrix"), _column($matrix)));
                let $submit = $("<button>").text("Update resource").click(function() {
                    $table = $(this).parent().parent().parent();
                    let newTransform = {
                        OffsetX: Number.parseFloat($table.find("#x").val()),
                        OffsetY: Number.parseFloat($table.find("#y").val()),
                        Matrix: JSON.parse($table.find("#matrix").val()),
                    };
                    $.post({
                        url: getActionLinkForWadNode(wad, tagid, 'transofrm'),
                        data: {
                            'id': h.IdInThatTypeArray,
                            'data': JSON.stringify(newTransform),
                        },
                        success: function(a) {
                            if (a != "" && a.error) {
                                alert('Error uploading: ' + a.error);
                            } else {
                                flpdata.Transformations[h.IdInThatTypeArray] = newTransform;
                                alert('Success!');
                            }
                        }
                    });
                })
                let warning = ("<sub>You can miss changes in web interface, but they must appear on disk</sub>")
                $data_table.append(_row(_column(), _column($submit, warning)));
            }

            switch (h.TypeArrayId) {
                default: $data_table.append(JSON.stringify(obj));
                break;
                case 1:
                        print_mesh(obj);
                    break;
                case 4:
                        $data_table.append(print_static_label_as_tr(h.IdInThatTypeArray), false);
                    break;
                case 6:
                        print_data6(obj);
                    break;
                case 7:
                        print_data6_subtype1(obj);
                    break;
                case 8:
                        print_data6_subtype1(obj);
                    break;
                case 9:
                        print_transform(obj);
                    break;
            }

            let get_parents = function(child_h) {
                let parents = [];
                let check_parenting = function(parent, h) {
                    if (h.IdInThatTypeArray == child_h.IdInThatTypeArray && h.TypeArrayId == child_h.TypeArrayId) {
                        let already = false;
                        for (let i in parents) {
                            if (parent.IdInThatTypeArray == parents[i].IdInThatTypeArray && parent.TypeArrayId == parents[i].TypeArrayId) {
                                already = true;
                            }
                        }
                        if (!already) {
                            parents.push(parent);
                        }
                    }
                }
                for (var i in flpdata.GlobalHandlersIndexes) {
                    let h = flpdata.GlobalHandlersIndexes[i];
                    let o = get_obj_by_handler(h);

                    let parse_parenting_data6_sub1 = function(o) {
                        for (let i in o.ElementsAnimation) {
                            let anim = o.ElementsAnimation[i];
                            for (let j in anim.KeyFrames) {
                                let frame = anim.KeyFrames[j];
                                check_parenting(h, frame.ElementHandler);
                                check_parenting(h, {
                                    TypeArrayId: 9,
                                    IdInThatTypeArray: frame.TransformationId
                                });
                                check_parenting(h, {
                                    TypeArrayId: 10,
                                    IdInThatTypeArray: frame.ColorId
                                });
                            }
                        }
                    }
                    switch (h.TypeArrayId) {
                        case 4:
                            for (let j in o.RenderCommandsList) {
                                if (o.RenderCommandsList[j].Flags & 8 != 0) {
                                    check_parenting(h, {
                                        TypeArrayId: 3,
                                        IdInThatTypeArray: o.RenderCommandsList[j].FontHandler
                                    });
                                }
                            }
                            break;
                        case 5:
                            check_parenting(h, o.FontHandler);
                            break;
                        case 6:
                            check_parenting(h, {
                                TypeArrayId: 3,
                                IdInThatTypeArray: o.FontHandler
                            });
                            break;
                        case 7:
                            parse_parenting_data6_sub1(o);
                            break;
                        case 8:
                            parse_parenting_data6_sub1(o);
                            break;
                    }
                }
                return parents;
            }

            console.log(obj);
            let $table = $("<table>");


            let $header = $("<span>").text(" Viewing object " + objNamesArray[h.TypeArrayId] + "[" + h.IdInThatTypeArray + "]");

            let parents_list = [];
            let parents = get_parents(h);
            let curParentRow = _row();
            let colums_cnt = 6;
            for (var i in parents) {
                if (i != 0 && (i % colums_cnt == 0)) {
                    parents_list.push(curParentRow);
                    curParentRow = _row().attr('colspan', colums_cnt);
                }
                curParentRow.append(_column(print_ref_handler(parents[i])));
            }
            if (parents.length < colums_cnt || (parents.length % colums_cnt != 0)) {
                parents_list.push(curParentRow);
            }

            console.log("parents", parents_list, parents);
            $table.append(_row(_column($header).attr('colspan', colums_cnt + 1)));
            if (parents.length != 0) {
                $table.append(_row(_column("parents").attr('rowspan', parents_list.length + 1)), parents_list);
            } else {
                $table.append(_row(_column("parents"), _column("no parents found")));
            }
            $table.append(_row(_column($data_table).attr('colspan', colums_cnt + 1)));
            $data_element.append($table);
            console.log($('#view-item-container'));
            $('#view-item-container').animate({
                scrollTop: 0
            }, 200);
        }
        dataSummary.append($history_element, $data_element);
        element_view();
    }

    var flp_view_font = function() {
        gr_instance.cleanup();
        set3dVisible(true);
        gr_instance.setInterfaceCameraMode(true);
        dataSummary.empty();

        var importBMFontScale = $('<input id="importbmfontscale" type="number" min="0" max="20" value="1" step="0.1">');
        var importBMFontInput = $('<button>');
        importBMFontInput.text('Import glyphs from BMFont file');
        importBMFontInput.attr("href", getActionLinkForWadNode(wad, tagid, 'importbmfont')).click(function() {
            $(this).attr('href', getActionLinkForWadNode(wad, tagid, 'importbmfont', 'scale=' + $("#importbmfontscale").val()));
            console.log($(this).attr('href'));
            uploadAjaxHandler.call(this);
        });
        var importDiv = $('<div id="flpimportfont">');
        importDiv.append($('<label>').text('font scale').append(importBMFontScale));
        importDiv.append(importBMFontInput);
        importDiv.append($('<a>').text('Link to usage instruction').attr('target', '_blank')
            .attr('href', 'https://github.com/mogaika/god_of_war_browser/blob/master/LOCALIZATION.md'));
        dataSummary.append(importDiv);

        var charstable = $("<table>");

        var mdl = new grModel();
        var matmap = {};

        for (var iFont in flpdata.Fonts) {
            var font = flpdata.Fonts[iFont];
            for (var iChar in font.CharNumberToSymbolIdMap) {
                if (font.CharNumberToSymbolIdMap[iChar] == -1) {
                    continue;
                }

                var glyphId = font.CharNumberToSymbolIdMap[iChar];
                if (glyphId >= font.CharsCount) {
                    continue;
                }

                var chrdata = font.MeshesRefs[glyphId];

                var meshes = [];
                if (chrdata.MeshPartIndex !== -1) {
                    meshes = loadMeshPartFromAjax(mdl, flp.Model.Meshes[0], chrdata.MeshPartIndex);
                    var txrid = undefined;
                    if (chrdata.Materials && chrdata.Materials.length !== 0 && chrdata.Materials[0].TextureName) {
                        var txr_name = chrdata.Materials[0].TextureName;

                        if (!matmap.hasOwnProperty(txr_name)) {
                            var material = new grMaterial();
                            console.log(txr_name);

                            var img = flp.Textures[txr_name].Images[0].Image

                            var texture = new grTexture('data:image/png;base64,' + img);
                            texture.markAsFontTexture();

                            var layer = new grMaterialLayer();
                            layer.setTextures([texture]);
                            material.addLayer(layer);

                            matmap[txr_name] = mdl.materials.length;
                            mdl.addMaterial(material);
                        }
                        txrid = matmap[txr_name];
                    }
                    for (var iMesh in meshes) {
                        meshes[iMesh].setMaterialID(txrid);
                    }
                }

                var symbolWidth = font.SymbolWidths[glyphId];
                var cubemesh = grHelper_CubeLines(symbolWidth / 32, 0, 0, symbolWidth / 32, 500, 5, false);
                mdl.addMesh(cubemesh);
                meshes.push(cubemesh);

                var char = String.fromCharCode(iChar);
                if (flp.FontCharAliases) {
                    var map_chars = Object.keys(flp.FontCharAliases).filter(function(charUnicode) {
                        return flp.FontCharAliases[charUnicode] == iChar
                    });
                    if (map_chars && map_chars.length !== 0) {
                        char = String.fromCharCode(map_chars[0]);
                    }
                }

                var table = $("<table>");

                var tr1 = $("<tr>");
                var tr2 = $("<tr>");
                tr1.append($("<td>").text('#' + glyphId));
                tr1.append($("<td>").text('width ' + symbolWidth));
                tr1.append($("<td>").text('ansii ' + iChar));
                tr2.append($("<td>").append($("<h2>").text(char)));
                tr2.append($("<td>").text('mesh pt ' + chrdata.MeshPartIndex));

                table.mouseenter([mdl, meshes], function(ev) {
                    ev.data[0].showExclusiveMeshes(ev.data[1]);
                    gr_instance.flushScene();
                    gr_instance.requestRedraw();
                });

                charstable.append($("<tr>").append(table.append(tr1).append(tr2)));
            }
        }

        dataSummary.append(charstable);
        gr_instance.models.push(mdl);
        gr_instance.requestRedraw();
    }

    dataSummarySelectors.append($('<div class="item-selector">').click(flp_list_labels).text("Labels editor"));
    dataSummarySelectors.append($('<div class="item-selector">').click(flp_print_dump).text("Dump"));
    dataSummarySelectors.append($('<div class="item-selector">').click(flp_scripts_strings).text("Scripts strings"));
    dataSummarySelectors.append($('<div class="item-selector">').click(flp_view_font).text("Font viewer"));
    dataSummarySelectors.append($('<div class="item-selector">').click(flp_view_object_viewer).text("Obj viewer"));

    // flp_list_labels();
    flp_view_object_viewer();
}