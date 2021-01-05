'use strict';

let gw_cxt_group_loading = false;
let flp_obj_view_history = [{
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

    let addNodes = function(nodes) {
        let ol = $('<ol>');
        for (let sn in nodes) {
            let node = data.Nodes[nodes[sn]];
            let li = $('<li>')
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
        let node_element = $(this).parent();
        let node_id = parseInt(node_element.attr('nodeid'));
        if (node_id !== 0) {
            gw_cxt_group_loading = false;
            treeLoadWadNode(wadName, node_id);
        } else {
            setLocation(wadName + " => " + node_element.attr('nodename'), '#/' + wadName + '/' + 0);
            dataSummary.empty();
            gr_instance.cleanup();

            gw_cxt_group_loading = true;

            $("#view-tree ol li").each(function(i, node) {
                let $node = $(node);
                if ($node.attr("nodetag") == "30" && $node.attr("nodename").startsWith("PS")) {
                    treeLoadWadNode(wadName, $node.attr("nodeid"), 0x6);
                } else if ($node.attr("nodetag") == "30" && $node.attr("nodename").startsWith("CXT_")) {
                    treeLoadWadNode(wadName, $node.attr("nodeid"), 0x80000001);
                }
            });
        }
    });
}

function treeLoadWadAsTags(wadName, data) {
    wad_last_load_view_type = 'tags';
    dataTree.empty();

    let ol = $('<ol>');
    for (let tagId in data.Tags) {
        let tag = data.Tags[tagId];
        let li = $('<li>')
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
        let node_element = $(this).parent();
        let tagid = parseInt(node_element.attr('tagid'));
        treeLoadWadTag(wadName, tagid);
    });
}

function treeLoadWadNode(wad, tagid, filterServerId = undefined) {
    dataSummary.empty();
    dataSummarySelectors.empty();
    set3dVisible(false);

    $.getJSON('/json/pack/' + wad + '/' + tagid, function(resp) {
        let data = resp.Data;
        let tag = resp.Tag;

        let needHexDump = false;
        let needMarshalDump = false;

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
                if (filterServerId) {
                    if (resp.ServerId != filterServerId) {
                        return;
                    }
                }
                switch (resp.ServerId) {
                    case 0x00000021: // flp
                        summaryLoadWadFlp(data, wad, tagid);
                        break;
                    case 0x00000018: // sbk blk
                    case 0x00040018: // sbk vag
                        summaryLoadWadSbk(data, wad, tagid);
                        needMarshalDump = true;
                        break;
                    case 0x00000006: // light
                        if (gw_cxt_group_loading) {
                            let pos = data.Position;
                            let color = data.Color;

                            let lightName = new grTextMesh("\x0f" + tag.Name, pos[0], pos[1], pos[2], true);
                            lightName.setColor(color[0], color[1], color[2]);
                            lightName.setOffset(-0.5, -0.5);
                            lightName.setMaskBit(5);
                            gr_instance.texts.push(lightName);
                        } else {
                            needMarshalDump = true;
                            needHexDump = true;
                        }
                        break;
                    case 0x00000007: // txr
                        summaryLoadWadTxr(data, wad, tagid);
                        break;
                    case 0x00070007: // ps3 txr
                        summaryLoadWadTxrPs3(data, wad, tagid);
                        break;
                    case 0x00000008: // material
                        summaryLoadWadMat(data);
                        break;
                    case 0x00000011: // collision
                        gr_instance.cleanup();
                        set3dVisible(true);

                        let mdl = new grModel();
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
                    case 0x0003000f: // gmdl
                        summaryLoadWadGmdl(data, wad, tagid);
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

    let form = $('<form class="flexedform" action="' + getActionLinkForWadNode(wad, tagid, 'updatetag') + '" method="post">');
    let tagel = $('li[tagid=' + tagid + ']');
    let tbl = $('<table>');
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
            let fileReader = new FileReader();
            fileReader.onload = function() {
                let arr = new Uint8Array(this.result);
                dataSummary.append($("<h5>").append('Size in bytes:' + arr.length));
                dataSummary.append(hexdump(arr));
            };
            fileReader.readAsArrayBuffer(blob);
        }
    });
}

function parseMeshPacket(object, packet, instanceIndex) {
    let m_vertexes = [];
    let m_indexes = [];
    let m_colors;
    let m_textures;
    let m_normals;

    m_vertexes.length = packet.Trias.X.length * 3;

    for (let i in packet.Trias.X) {
        let j = i * 3;
        m_vertexes[j] = packet.Trias.X[i];
        m_vertexes[j + 1] = packet.Trias.Y[i];
        m_vertexes[j + 2] = packet.Trias.Z[i];
        if (!packet.Trias.Skip[i]) {
            m_indexes.push(i - 2);
            m_indexes.push(i - 1);
            m_indexes.push(i - 0);
        }
    }

    let mesh = new grMesh(m_vertexes, m_indexes);
    mesh.setUseBindToJoin(object.UseInvertedMatrix);

    if (packet.Blend.R && packet.Blend.R.length) {
        let m_colors = [];
        m_colors.length = packet.Blend.R.length * 4;
        for (let i in packet.Blend.R) {
            let j = i * 4;
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

        for (let i in packet.Uvs.U) {
            let j = i * 2;
            m_textures[j] = packet.Uvs.U[i];
            m_textures[j + 1] = packet.Uvs.V[i];
        }
        mesh.setUVs(m_textures);
    }

    if (packet.Norms.X && packet.Norms.X.length) {
        m_normals = [];
        m_normals.length = packet.Norms.X.length * 3;

        for (let i in packet.Norms.X) {
            let j = i * 3;
            m_normals[j] = packet.Norms.X[i];
            m_normals[j + 1] = packet.Norms.Y[i];
            m_normals[j + 2] = packet.Norms.Z[i];
        }

        mesh.setNormals(m_normals);
    }

    if (packet.Joints && packet.Joints.length && object.JointMappers && object.JointMappers.length) {
        let jm = object.JointMappers[instanceIndex];
        if (jm && jm.length) {
            //console.log(packet.Joints, packet.Joints2, object.JointMappers);
            let joints1 = packet.Joints;
            let joints2 = (!!packet.Joints2) ? packet.Joints2 : undefined;

            mesh.setJointIds(jm, joints1, joints2, packet.Trias.Weight);
        }
    }

    return mesh;
}

function loadMeshPartFromAjax(model, data, iPart, table = undefined) {
    let part = data.Parts[iPart];
    let totalMeshes = [];

    for (let iGroup in part.Groups) {
        let group = part.Groups[iGroup];
        for (let iObject in group.Objects) {
            let object = group.Objects[iObject];

            for (let iInstance = 0; iInstance < object.InstancesCount; iInstance++) {
                for (let iLayer = 0; iLayer < object.TextureLayersCount; iLayer++) {
                    let iDmaPacket = iInstance * object.TextureLayersCount + iLayer;
                    let dmaPackets = object.Packets[iDmaPacket];

                    let objName = "p" + iPart + "_g" + iGroup + "_o" + iObject + "m" + object.MaterialId;
                    if (object.InstancesCount != 1) {
                        objName += "_i" + iInstance
                    } else if (object.TextureLayersCount != 1) {
                        objName += "_l" + iLayer;
                    }

                    let meshes = [];

                    for (let iPacket in dmaPackets) {
                        let dmaPacket = dmaPackets[iPacket];
                        let mesh = parseMeshPacket(object, dmaPacket, iInstance);

                        if (object.TextureLayersCount != 1) {
                            mesh.setLayer(iLayer);
                        }

                        meshes.push(mesh);
                        mesh.meta['part'] = iPart;
                        mesh.meta['group'] = iGroup;
                        mesh.meta['object'] = iObject;
                        totalMeshes.push(mesh);
                        model.addMesh(mesh);
                    }

                    if (table) {
                        let label = $('<label>');
                        let chbox = $('<input type="checkbox" checked>');
                        let td = $('<td>').append(label);
                        chbox.click(meshes, function(ev) {
                            for (let i in ev.data) {
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
    }
    return totalMeshes;
}

function loadMeshFromAjax(model, data, needTable = false) {
    let table = needTable ? $('<table>') : undefined;
    for (let iPart in data.Parts) {
        loadMeshPartFromAjax(model, data, iPart, table);
    }
    gr_instance.flushScene();
    return table;
}

function summaryLoadWadMesh(data, wad, nodeid) {
    gr_instance.cleanup();
    set3dVisible(true);

    let mdl = new grModel();

    let dumplinkfbx = getActionLinkForWadNode(wad, nodeid, 'fbx');
    dataSummary.append($('<a class="center">').attr('href', dumplinkfbx).append('Download .fbx 2014 bin'));

    let dumplinkobj = getActionLinkForWadNode(wad, nodeid, 'obj');
    dataSummary.append($('<a class="center">').attr('href', dumplinkobj).append('Download .obj'));

    let dumplinkgltf = getActionLinkForWadNode(wad, nodeid, 'gltf');
    dataSummary.append($('<a class="center">').attr('href', dumplinkgltf).append('Download .glb bin glTF 2.0'));

    let table = loadMeshFromAjax(mdl, data, true);
    dataSummary.append(table);

    gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function parseGmdlObjectMesh(part, object, originalMeshObject) {
    let m_indexes = [];

    let streams = part.Streams;

    let streamStart = object.StreamStart;
    let streamCount = object.StreamCount;
    let indexStart = object.IndexStart;
    let indexCount = object.IndexCount;

    let sPos = streams["POS0"].Values.slice(streamStart, streamStart + streamCount);
    m_indexes = part.Indexes.slice(indexStart, indexStart + indexCount);

    let m_vertexes = [];
    m_vertexes.length = sPos.length * 3;
    let m_weights = [];
    m_weights.length = sPos.length;

    for (let i in sPos) {
        let j = i * 3;
        let pos = sPos[i]
        m_vertexes[j + 0] = pos[0];
        m_vertexes[j + 1] = pos[1];
        m_vertexes[j + 2] = pos[2];
        m_weights[i] = pos[3];
    }

    for (let i in m_indexes) {
        m_indexes[i] -= streamStart;
    }

    let mesh = new grMesh(m_vertexes, m_indexes);
    if (originalMeshObject) {
        mesh.setUseBindToJoin(originalMeshObject.UseInvertedMatrix);
    }

    mesh.setMaterialID(object.TextureIndex);

    if ("COL0" in streams) {
        let sCol = streams["COL0"].Values.slice(streamStart, streamStart + streamCount);
        let m_colors = [];
        m_colors.length = sCol.length * 4;

        for (let i in sCol) {
            let j = i * 4;
            let col = sCol[i];
            m_colors[j + 0] = col[0] * 255.0;
            m_colors[j + 1] = col[1] * 255.0;
            m_colors[j + 2] = col[2] * 255.0;
            m_colors[j + 3] = col[3] * 255.0;
        }
        mesh.setBlendColors(m_colors);
    }

    if ("TEX0" in streams) {
        let sTex = streams["TEX0"].Values.slice(streamStart, streamStart + streamCount);
        let m_textures = [];
        m_textures.length = sTex.length * 2;

        for (let i in sTex) {
            let j = i * 2;
            let tex = sTex[i];
            m_textures[j + 0] = tex[0];
            m_textures[j + 1] = tex[1];
        }
        mesh.setUVs(m_textures);
    }

    if ("BONI" in streams) {
        let joints1 = [];
        let joints2 = [];
        let sBoni = streams["BONI"].Values.slice(streamStart, streamStart + streamCount);
        joints1.length = sBoni.length;
        joints2.length = sBoni.length;
        for (let i in sBoni) {
            joints1[i] = sBoni[i][0];
            joints2[i] = sBoni[i][1];
        }
        mesh.setJointIds(object.JointsMap, joints1, joints2, m_weights);
    }

    //console.log(originalMeshObject.Type, originalMeshObject);	
    if (originalMeshObject) {
        if (originalMeshObject.Type == 0x1d) {
            mesh.setps3static(true);
        }
    }

    return mesh;
}

function loadGmdlPartFromAjax(model, data, iPart, originalPart, table = undefined) {
    let part = data.Models[iPart];
    let totalMeshes = [];

    for (let iObject in part.Objects) {
        let object = part.Objects[iObject];

        let objName = "p" + iPart + "_o" + iObject;

        let originalMeshObject;
        if (originalPart) {
            // ignore lods
            originalMeshObject = originalPart.Groups[0].Objects[iObject];
        }
        let mesh = parseGmdlObjectMesh(part, object, originalMeshObject);

        totalMeshes.push(mesh);
        model.addMesh(mesh);

        if (table) {
            let label = $('<label>');
            let chbox = $('<input type="checkbox" checked>');
            let td = $('<td>').append(label);
            chbox.click(mesh, function(ev) {
                ev.data.setVisible(this.checked);
                gr_instance.requestRedraw();
            });
            td.mouseenter([model, mesh], function(ev) {
                ev.data[0].showExclusiveMeshes([ev.data[1]]);
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
    return totalMeshes;
}

function loadGmdlFromAjax(model, data, originalMesh, needTable = false) {
    // console.log(data);

    let table = needTable ? $('<table>') : undefined;
    for (let iPart in data.Models) {
        let originalPart;
        if (originalMesh) {
            originalPart = originalMesh.Parts[iPart];
        }
        loadGmdlPartFromAjax(model, data, iPart, originalPart, table);
    }

    gr_instance.flushScene();
    return table;
}

function summaryLoadWadGmdl(data, wad, nodeid) {
    gr_instance.cleanup();
    set3dVisible(true);

    let mdl = new grModel();

    let table = loadGmdlFromAjax(mdl, data, undefined, true);
    dataSummary.append(table);

    gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function loadMdlFromAjax(mdl, data, parseScripts = false, needTable = false) {
    let tables = [];
    if (data.Meshes && data.Meshes.length) {
        for (let mesh of data.Meshes) {
            if (!!data.GMDL) {
                tables.push(loadGmdlFromAjax(mdl, data.GMDL, mesh, needTable));
            } else {
                tables.push(loadMeshFromAjax(mdl, mesh, needTable));
            }
        }
    }

    for (let iMaterial in data.Materials) {
        let material = new grMaterial();

        let textures = data.Materials[iMaterial].Textures;
        let rawMat = data.Materials[iMaterial].Mat;
        material.setColor(rawMat.Color);

        for (let iLayer in rawMat.Layers) {
            let rawLayer = rawMat.Layers[iLayer];
            let layer = new grMaterialLayer();

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
            // console.log("layer parsing: ", layer, rawLayer);

            if (textures && textures[iLayer] && textures[iLayer].Images) {
                let imgs = textures[iLayer].Images;
                let txrs = [];
                for (let iImg in imgs) {
                    txrs.push(new grTexture('data:image/png;base64,' + imgs[iImg].Image));
                }
                layer.setTextures(txrs);
                layer.setHasAlphaAttribute(textures[iLayer].HaveTransparent);
            }
            material.addLayer(layer);
        }
        mdl.addMaterial(material);

        let anim = data.Materials[iMaterial].Animations;
        if (anim && anim.Groups && anim.Groups.length) {
            let group = anim.Groups[0];
            if (!group.IsExternal && group.Acts && group.Acts.length) {
                for (let iAct in group.Acts) {
                    let act = group.Acts[iAct];
                    for (let dt in anim.DataTypes) {
                        switch (anim.DataTypes[dt].TypeId) {
                            case 8:
                                let animInstance = new gaMatertialLayerAnimation(anim, act, dt, material);
                                animInstance.enable();
                                ga_instance.addAnimation(animInstance);
                                break;
                            case 9:
                                let animSheetInstance = new gaMatertialSheetAnimation(anim, act, dt, material);
                                ga_instance.addAnimation(animSheetInstance);
                                break;
                        }
                    }
                }
            }
        }
    }

    if (parseScripts) {
        for (let i in data.Scripts) {
            let scr = data.Scripts[i];
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
    return tables;
}

function summaryLoadWadMdl(data, wad, nodeid) {
    gr_instance.cleanup();
    set3dVisible(true);

    let mdl = new grModel();

    let table = $('<table>');
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

    let dumplinkfbx = getActionLinkForWadNode(wad, nodeid, 'fbx');
    dataSummary.append($('<a class="center">').attr('href', dumplinkfbx).append('Download .fbx 2014 bin'));

    let dumplinkgltf = getActionLinkForWadNode(wad, nodeid, 'gltf');
    dataSummary.append($('<a class="center">').attr('href', dumplinkgltf).append('Download .glb bin glTF 2.0'));

    let mdlTables = loadMdlFromAjax(mdl, data, false, true);
    for (let mdlTable of mdlTables) {
        dataSummary.append(mdlTable);
    }

    gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function summaryLoadWadTxr(data, wad, nodeid) {
    set3dVisible(false);
    let table = $('<table>');
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
    for (let i in data.Images) {
        let img = data.Images[i];
        dataSummary.append($('<img>')
            .addClass('no-interpolate')
            .attr('src', 'data:image/png;base64,' + img.Image)
            .attr('alt', 'gfx:' + img.Gfx + '  pal:' + img.Pal));
    }

    let form = $('<form action="' + getActionLinkForWadNode(wad, nodeid, 'upload') + '" method="post" enctype="multipart/form-data">');
    form.append($('<input type="file" name="img">'));
    let replaceBtn = $('<input type="button" value="Replace texture">')
    replaceBtn.click(function() {
        let form = $(this).parent();
        $.ajax({
            url: form.attr('action') + "create_new_pal=" + form.find("#create_new_pal")[0].checked,
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
    form.append($('</br></br><b>WARNING: Use checkbox below only if you know what you are doing</b></br><input type="checkbox" id="create_new_pal" name="create_new_pal" value="true">'));
    form.append($('<label>Create new palette for replaced texture. Handy if palette used by multiply textures.</label>'));

    dataSummary.append(form);
}

function summaryLoadWadTxrPs3(data, wad, nodeid) {
    set3dVisible(false);

    let table = $('<table>');
    $.each(data, function(k, val) {
        if (k != 'Images') {
            table.append($('<tr>')
                .append($('<td>').append(k))
                .append($('<td>').append(val)));
        }
    });

    dataSummary.append(table);
    for (let i in data.Images) {
        dataSummary.append($('<img>')
            .addClass('no-interpolate')
            .attr('src', 'data:image/png;base64,' + data.Images[i])
            .attr('alt', 'mipmap: ' + i));
    }
}

function summaryLoadWadMat(data) {
    set3dVisible(false);
    let clr = data.Mat.Color;
    let clrBgAttr = 'background-color: rgb(' + parseInt(clr[0] * 255) + ',' + parseInt(clr[1] * 255) + ',' + parseInt(clr[2] * 255) + ')';

    let table = $('<table>');
    table.append($('<tr>')
        .append($('<td>').append('Color'))
        .append($('<td>').attr('style', clrBgAttr).append(
            JSON.stringify(clr, undefined, 2)
        ))
    );

    for (let l in data.Mat.Layers) {
        let layer = data.Mat.Layers[l];
        let ltable = $('<table>')

        $.each(layer, function(k, v) {
            let td = $('<td>');
            switch (k) {
                case 'Flags':
                    let str = '';
                    for (let i in v) {
                        str = str + '0x' + v[i].toString(0x10) + ', ';
                    }
                    td.append(str);
                    break;
                case 'BlendColor':
                    let r = Array(4);
                    for (let i in data.Mat.Color) {
                        r[i] = v[i] * data.Mat.Color[i];
                    }
                    td.attr('style', 'background-color: rgb(' + parseInt(r[0] * 255) + ',' + parseInt(r[1] * 255) + ',' + parseInt(r[2] * 255) + ')')
                        .append(JSON.stringify(v, undefined, 2) + ';  result:' + JSON.stringify(r, undefined, 2));
                    break;
                case 'Texture':
                    td.append(v);
                    if (v != '') {
                        let txrobj = data.Textures[l];
                        let txrblndobj = data.TexturesBlended[l];
                        td.append(' \\ ' + txrobj.Data.GfxName + ' \\ ' + txrobj.Data.PalName).append('<br>');
                        td.append('Color + Alpha').append('<br>');
                        td.append($('<img>').attr('src', 'data:image/png;base64,' + txrobj.Images[0].Image));
                        td.append('<br>').append(' BLENDED Color + Alpha').append('<br>');
                        td.append($('<img>').attr('src', 'data:image/png;base64,' + txrblndobj.Images[0].Image));
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
        let vec = data.Shape.Vector;
        let mesh = grHelper_SphereLines(vec[0], vec[1], vec[2], vec[3] * 2, 7, 7);
        mdl.addMesh(mesh);
        mesh.setMaskBit(4);
    } else if (data.ShapeName == "") {

    }
}

function loadObjFromAjax(mdl, data, matrix = undefined, parseScripts = false) {
    if (data.Model) {
        let mdlTables = loadMdlFromAjax(mdl, data.Model, parseScripts, true);
        for (let mdlTable of mdlTables) {
            dataSummary.append(mdlTables);
        }
    } else if (data.Collision) {
        loadCollisionFromAjax(mdl, data.Collision);
    }

    if (data.Script) {
        if (data.Script.TargetName == "SCR_Entities") {
            $.each(data.Script.Data.Array, function(entity_id, entity) {
                let objMat = new Float32Array(data.Data.Joints[0].RenderMat);
                let entityMat = new Float32Array(entity.Matrix);

                if (matrix) {
                    // obj = obj * transformMat
                    objMat = mat4.mul(mat4.create(), matrix, objMat);
                }
                // mat = obj * entity
                let mat = mat4.mul(mat4.create(), objMat, entityMat);

                let pos = mat4.getTranslation(vec3.create(), mat);

                let radius = entity.Matrix[0];
                let text3d = new grTextMesh("\x05" + entity.Name, pos[0], pos[1], pos[2], true);
                text3d.setOffset(-0.5, -0.5);
                text3d.setMaskBit(3);

                //let mdl = new grModel();
                //mdl.addMesh(new grHelper_SphereLines(pos[0], pos[1], pos[2], radius, 6, 6));

                let alpha = 1;
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

    let dumplinkfbx = getActionLinkForWadNode(wad, nodeid, 'fbx');
    dataSummary.append($('<a class="center">').attr('href', dumplinkfbx).append('Download .fbx 2014 bin'));

    let dumplinkgltf = getActionLinkForWadNode(wad, nodeid, 'gltf');
    dataSummary.append($('<a class="center">').attr('href', dumplinkgltf).append('Download .glb bin glTF 2.0'));

    let jointsTable = $('<table>');

    if (data.Animations) {
        let $animSelector = $("<select>").attr("size", 6).addClass("animation");

        let anim = data.Animations;
        if (anim && anim.Groups && anim.Groups.length) {
            for (let iGroup in anim.Groups) {
                let group = anim.Groups[iGroup];
                for (let iAct in group.Acts) {
                    let act = group.Acts[iAct];
                    for (let dt in anim.DataTypes) {
                        switch (anim.DataTypes[dt].TypeId) {
                            case 0:
                                let $option = $("<option>").text(group.Name + ": " + act.Name);
                                $option.dblclick([anim, act, dt, data.Data], function(ev) {
                                    let anim = new gaObjSkeletAnimation(ev.data[0], ev.data[1], ev.data[2], ev.data[3], gr_instance.models[0]);
                                    ga_instance.addAnimation(anim);
                                });

                                $animSelector.append($option);
                                break;
                        }
                    }
                }
            }

        }

        let $stopAnim = $("<button>").text("stop anim");
        $stopAnim.click(function() {
            let anims = ga_instance.objSkeletAnimations;
            for (let i in anims) {
                ga_instance.freeAnimation(anims[i]);
            }
        });

        dataSummary.append($animSelector).append($stopAnim);
    }

    $.each(data.Data.Joints, function(joint_id, joint) {
        let row = $('<tr>').append($('<td>').attr('style', 'background-color:rgb(' +
                parseInt((joint.Id % 8) * 15) + ',' +
                parseInt(((joint.Id / 8) % 8) * 15) + ',' +
                parseInt(((joint.Id / 64) % 8) * 15) + ');')
            .append(joint.Id).attr("rowspan", 20));

        let firstRow = true;

        for (let k in joint) {
            if (k === "Name" ||
                k === "IsSkinned" ||
                k === "IsExternal" ||
                k === "OurJointToIdleMat" ||
                k === "ParentToJoint" ||
                k === "BindToJointMat" ||
                k === "RenderMat" ||
                k === "Parent") {

                let d = joint[k];
                if (Array.isArray(d) && d.length == 4 * 4) {
                    row.append($('<td>').text(k));
                    jointsTable.append(row);

                    let t = [d[12].toFixed(1), d[13].toFixed(1), d[14].toFixed(1)];
                    let s = [d[0].toFixed(2), d[5].toFixed(2), d[10].toFixed(2), d[15].toFixed(2)];

                    let ry = Math.asin(d[8]).toFixed(2);
                    let rx = Math.atan2(-d[9] / Math.cos(ry), +d[10] / Math.cos(ry)).toFixed(2);
                    let rz = Math.atan2(-d[4] / Math.cos(ry), +d[0] / Math.cos(ry)).toFixed(2);

                    let appRow = function(name, arr) {
                        jointsTable.append($('<tr>').append($('<td>').text(name + ":" + arr.join(","))));
                    }

                    appRow("t", [d[12].toFixed(2), d[13].toFixed(2), d[14].toFixed(2)]);
                    appRow("r", [rx, ry, rz]);
                    appRow("s", [d[0].toFixed(2), d[5].toFixed(2), d[10].toFixed(2), d[15].toFixed(2)]);
                } else {
                    row.append($('<td>').text(k + ": " + JSON.stringify(d)));
                    jointsTable.append(row);
                }

                row = $('<tr>');
            }
        }
        jointsTable.append(row);
    });
    dataSummary.append(jointsTable);

    if (data.Model || data.Collision) {
        set3dVisible(true);

        let mdl = new grModel();
        loadObjFromAjax(mdl, data);

        gr_instance.models.push(mdl);
        gr_instance.requestRedraw();
    }
}


function summaryLoadWadGameObject(data) {
    gr_instance.cleanup();
    set3dVisible(false);
    let table = $('<table>');
    for (let k in data) {
        table.append($('<tr>').append($('<td>').text(k)).append($('<td>').text(JSON.stringify(data[k]))));
    }
    dataSummary.append(table);
}

function loadCxtFromAjax(data, parseScripts = true) {
    for (let i in data.Instances) {
        let inst = data.Instances[i];
        let obj = data.Objects[inst.Object];

        let rs = 180.0 / Math.PI;
        let rot = quat.fromEuler(quat.create(), inst.Rotation[0] * rs, inst.Rotation[1] * rs, inst.Rotation[2] * rs);

        //let instMat = mat4.fromTranslation(mat4.create(), inst.Position1);
        //instMat = mat4.mul(mat4.create(), instMat, mat4.fromQuat(mat4.create(), rot));
        // same as above
        let instMat = mat4.fromRotationTranslation(mat4.create(), rot, inst.Position1);
        //let instMat = mat4.fromQuat(mat4.create(), rot);
		
		let pos = inst.Position1;
        let text3d = new grTextMesh("\x04" + inst.Name, instMat[12], instMat[13], instMat[14], true);
        text3d.setOffset(-0.5, -0.5);
        text3d.setMaskBit(6);
		gr_instance.texts.push(text3d);

        //console.log(inst.Object, instMat);
        //if (obj && (obj.Model || (obj.Collision && inst.Object.includes("deathzone")))) {
        //if (obj && (obj.Model)) {
        if (obj && (obj.Model || obj.Collision)) {
            let mdl = new grModel();
            loadObjFromAjax(mdl, obj, instMat, parseScripts);

            for (let iScript in inst.Scripts) {
                let scr = inst.Scripts[iScript];
                switch (scr.TargetName) {
                    case "SCR_Sky":
                        // case "SCR_AresSky":
                        mdl.setType("sky");
                        break;
                    default:
                        console.warn("Unknown SCR target: " + scr.TargetName, data, inst, scr);
                        break;
                }
            }

            gr_instance.models.push(mdl);
        }
    }
}

function summaryLoadWadCxt(data, wad, nodeid) {
    if (!gw_cxt_group_loading) {
        gr_instance.cleanup();

        let dumplinkfbx = getActionLinkForWadNode(wad, nodeid, 'fbx');
        dataSummary.append($('<a class="center">').attr('href', dumplinkfbx).append('Download .fbx 2014 bin'));
    } else {
        dataSummary.empty();
        let dumplinkfbx = getActionLinkForWadNode(wad, nodeid, 'fbx_all');
        dataSummary.append($('<a class="center">').attr('href', dumplinkfbx).append('Download .fbx 2014 bin'));
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
    let list = $("<ul>");
    for (let i = 0; i < data.Sounds.length; i++) {
        let snd = data.Sounds[i];
        let link = '/action/' + wad + '/' + nodeid + '/';

        let getSndLink = function(type) {
            return getActionLinkForWadNode(wad, nodeid, type, 'snd=' + snd.Name);
        };

        let vaglink = $("<a>").append(snd.Name).attr('href', getSndLink('vag'));
        let li = $("<li>").append(vaglink).append(" id:" + i + " stream id:" + snd.StreamId);

        if (data.IsVagFiles) {
            let wavplayer = $("<source>").attr("src", getSndLink('wav'));
            let wavlink = $("<audio controls>").attr("preload", "none").append(wavplayer);
            li.append("<br>").append(wavlink);
        } else {
            let commands = $("<table>");
            let banksound = data.Bank.BankSounds[snd.StreamId];

            let banksoundsNoCommand = {};
            for (let key in banksound) {
                if (key != 'Commands') {
                    banksoundsNoCommand[key] = banksound[key];
                }
            }

            commands.append($("<tr>").append(
                $("<td>").text("Parameters")).append(
                $("<td>").attr("colspan", 2).text(JSON.stringify(banksoundsNoCommand))
            ));

            let cmdRow = $("<tr>").append($("<td>").text("Commands").attr("rowspan", banksound.Commands.length));
            for (let command of banksound.Commands) {
                let commandClean = {};
                for (let key in command) {
                    if (key != 'SampleRef' && key != 'Cmd' && key != 'VagRef' && key != 'UnkRef') {
                        commandClean[key] = command[key];
                    }
                }

                let argsCol = $("<td>").text(JSON.stringify(commandClean));
                if (command.SampleRef != null) {
                    let sampleRef = command.SampleRef;

                    let sampleRefClean = {};
                    for (let key in sampleRef) {
                        if (key != 'AdpcmOffset' && key != 'AdpcmSize') {
                            sampleRefClean[key] = sampleRef[key];
                        }
                    }

                    argsCol.append($("<br>")).append("Sample:");
                    argsCol.append($("<br>")).append(JSON.stringify(sampleRefClean));

                    let sndurl = getActionLinkForWadNode(wad, nodeid, 'smpd',
                        'offset=' + sampleRef.AdpcmOffset + '&size=' + sampleRef.AdpcmSize);
                    let wavplayer = $("<source>").attr("src", sndurl);
                    let wavlink = $("<audio controls>").attr("preload", "none").append(wavplayer);

                    argsCol.append($("<br>")).append("Audio offset " + sampleRef.AdpcmOffset + " size " + sampleRef.AdpcmSize);
                    argsCol.append($("<br>")).append(wavlink);
                }
                if (command.VagRef != null) {
                    let vagRef = command.VagRef;
                    argsCol.append($("<br>")).append("ref: " + JSON.stringify(vagRef));
                }
                if (command.UnkRef != null) {
                    let unkRef = command.UnkRef;
                    argsCol.append($("<br>")).append("ref: " + JSON.stringify(unkRef));
                }

                cmdRow.append($("<td>").text(command.Cmd)).append(argsCol);
                commands.append(cmdRow);
                cmdRow = $("<tr>");
            }

            li.append(commands);
        }
        list.append(li);
    }
    dataSummary.append(list);
}

function summaryLoadWadGeomShape(data) {
    gr_instance.cleanup();
    set3dVisible(true);

    let m_vertexes = [];
    m_vertexes.length = data.Vertexes.length * 3;
    for (let i in data.Vertexes) {
        let j = i * 3;
        let v = data.Vertexes[i];
        m_vertexes[j] = v.Pos[0];
        m_vertexes[j + 1] = v.Pos[1];
        m_vertexes[j + 2] = v.Pos[2];
    }

    let m_indexes = [];
    m_indexes.length = data.Indexes.length * 3;
    for (let i in data.Indexes) {
        let j = i * 3;
        let v = data.Indexes[i];
        m_indexes[j] = v.Indexes[0];
        m_indexes[j + 1] = v.Indexes[1];
        m_indexes[j + 2] = v.Indexes[2];
    }

    let mdl = new grModel();
    mdl.addMesh(new grMesh(m_vertexes, m_indexes));

    gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function summaryLoadWadScript(data) {
    gr_instance.cleanup();

    dataSummary.append($("<h3>").append("Scirpt " + data.TargetName));

    if (data.TargetName == 'SCR_Entities') {
        for (let i in data.Data.Array) {
            let e = data.Data.Array[i];

            let ht = $("<table>").append($("<tr>").append($("<td>").attr("colspan", 2).append(e.Name)));
            for (let j in e) {
                let v = e[j];
                if (j == "Handlers") {
                    for (let hi in v) {
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