'use strict';

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
            dataSummary.empty();
            gr_instance.cleanup();

            gw_cxt_group_loading = true;

            setLocation(wadName + " => " + node_element.attr('nodename'), '#/' + wadName + '/' + 0);

            $("#view-tree ol li").each(function(i, node) {
                let $node = $(node);
                if ($node.attr("nodetag") == "30" || $node.attr("nodetag") == "1") {
                    const nn = $node.attr("nodename");
                    if (nn.startsWith("PS")) {
                        treeLoadWadNode(wadName, $node.attr("nodeid"), 0x6);
                    } else if (nn.startsWith("CXT_")) {
                        treeLoadWadNode(wadName, $node.attr("nodeid"), 0x80000001);
                    } else if (nn.startsWith("RIB_sheet")) {
                        treeLoadWadNode(wadName, $node.attr("nodeid"), 0x00000011);
                        //} else if (nn.startsWith("CMZ_") || nn.startsWith("ENZ_") || nn.startsWith("SEZ_")) {
                        //	treeLoadWadNode(wadName, $node.attr("nodeid"), 0x00000011);
                        // // ^^^^^ this should be loaded in CMV_, ENV_, SEV_ 
                    }
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
    if (!gw_cxt_group_loading) {
        dataSummary.empty();
        dataSummarySelectors.empty();
    }
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
                    case 0x0000001B: // flp gow2
                        summaryLoadWadFlp(data, wad, tagid);
                        break;
                    case 0x00000018: // sbk blk
                    case 0x00040018: // sbk vag
                    case 0x00000015: // sbk blk
                        summaryLoadWadSbk(data, wad, tagid);
                        needMarshalDump = true;
                        break;
                    case 0x00000006: // light
                        if (gw_cxt_group_loading) {
                            let pos = data.Position;
                            let color = data.Color;

                            let lightName = new RenderTextMesh("\x0f" + tag.Name, true);
                            lightName.setColor(color[0], color[1], color[2]);
                            lightName.setOffset(0, -0.5);
                            lightName.setMaskBit(5);

                            let node = new ObjectTreeNodeModel("lightName", lightName);
                            node.setLocalMatrix(mat4.fromTranslation(mat4.create(), pos));
                            gr_instance.addNode(node);
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
                        if (gw_cxt_group_loading !== true) {
                            gr_instance.cleanup();
                            set3dVisible(true);
                        }
                        let node = loadCollisionFromAjax(data, wad, tagid);
                        gr_instance.addNode(node);
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
                    case 0x00010001: // obj gow2
                        summaryLoadWadObj(data, wad, tagid);
                        break;
                    case 0x80000001: // cxt
                        summaryLoadWadCxt(data, wad, tagid);
                        break;
                    case 0x00020001: // gameObject
                        summaryLoadWadGameObject(data);
                        break;
                    case 0x00030001: // gameObject gow2
                        summaryLoadWadGameObject(data);
                        break;
                    case 0x00010004: // script
                        summaryLoadWadScript(data, wad, tagid);
                        //needMarshalDump = true;
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
            } else if (tag.Tag == 113 || tag.Tag == 114) {
                summaryLoadWadTWK(data, wad, tagid);
                needMarshalDump = false;
                needHexDump = false;
            } else if (tag.Tag == 500) {
                summaryLoadWadRSRCS(data, wad, tagid);
                needMarshalDump = false;
                needHexDump = false;
            } else if (tag.Tag == 12 || tag.Tag == 13 || tag.Tag == 14 || tag.Tag == 15 || tag.Tag == 16) {
                needMarshalDump = true;
                needHexDump = true;
            } else {
                needHexDump = true;
            }

            // console.log('wad ' + wad + ' file (' + tag.Name + ')' + tag.Id + ' loaded. serverid:' + resp.ServerId);
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

function parseMeshPackets(data, object, packets, instanceIndex) {
    let m_indexes = [];
    let m_vertexes = [];
    let m_colors = [];
    let m_textures = [];
    let m_normals = [];
    let m_jointsIds = [];
    let m_weights = [];
    let m_jointsWidth = 2;
    const jm = object.JointMappers[instanceIndex];

    let offset = 0;
    for (const packet of packets) {
        const vertexCount = packet.Trias.X.length;

        for (const i in packet.Trias.X) {
            m_vertexes.push(packet.Trias.X[i], packet.Trias.Y[i], packet.Trias.Z[i]);
            if (!packet.Trias.Skip[i]) {
                m_indexes.push(offset + parseInt(i) - 2, offset + parseInt(i) - 1, offset + parseInt(i) - 0);
            }
        }

        if (packet.Trias.Weight && packet.Trias.Weight.length && packet.Joints && packet.Joints.length) {
            const joints1 = packet.Joints[0];
            const joints2 = packet.Joints[1];
            const weights = packet.Trias.Weight;

            if (joints1.length !== joints2.length || joints1.length !== weights.length) {
                log.error("inconsistent joints/weights length", packet);
            } else {
                for (const i in weights) {
                    const weight = weights[i];

                    let addedJointAttributes = 0;
                    const insertSingleJointAttribute = function(joint, weight) {
                        // expand array and insert empty spaces if needed
                        if (addedJointAttributes >= m_jointsWidth) {
                            if (m_jointsWidth === 4) {
                                log.error("Over limit for joint width", addedJointAttributes, m_jointsWidth, data.BlendJoints);
                                return
                            }
                            const oldWidth = m_jointsWidth;
                            const oldJointsIds = m_jointsIds;

                            const completedVertexesCount = (m_jointsIds.length - addedJointAttributes) / m_jointsWidth;
                            const oldWeights = m_weights;

                            m_jointsWidth += 1;

                            console.log("expanding additional blends from", oldWidth, "to", m_jointsWidth, "cvc", completedVertexesCount);

                            m_jointsIds = [];
                            m_jointsIds.length = completedVertexesCount * m_jointsWidth;
                            m_weights = [];
                            m_weights.length = m_jointsIds.length;

                            for (let iVert = 0; iVert < completedVertexesCount; iVert++) {
                                const pos = iVert * m_jointsWidth;
                                const oldpos = iVert * oldWidth;

                                for (let j = 0; j < oldWidth; j++) {
                                    m_jointsIds[pos + j] = oldJointsIds[oldpos + j];
                                    m_weights[pos + j] = oldWeights[oldpos + j];
                                }
                                m_jointsIds[pos + oldWidth] = 0;
                                m_weights[pos + oldWidth] = 0;
                            }

                            for (let i = 0; i < addedJointAttributes; i++) {
                                const oldpos = completedVertexesCount * oldWidth + i;
                                m_jointsIds.push(oldJointsIds[oldpos]);
                                m_weights.push(oldWeights[oldpos]);
                            }

                            /*
                            console.log("old j", oldJointsIds, "new j", m_jointsIds);
                            console.log("old w", oldWeights, "new w", m_weights);
                            */

                            if (m_jointsIds[0] === undefined) {
                                console.error("wtf");
                            }
                        }
                        addedJointAttributes += 1;

                        m_jointsIds.push(joint);
                        m_weights.push(weight);
                    }

                    const addJoint = function(joint, weight) {
                        if (joint >= data.SkeletJoints) {
                            const blend = data.BlendJoints[joint - data.SkeletJoints];
                            // console.warn("multiple joints blend", blend);

                            for (const i in blend.Weights) {
                                // console.log("000", m_jointsIds, m_weights, blend.JointIds[i], blend.Weights[i] * weight);
                                insertSingleJointAttribute(blend.JointIds[i], blend.Weights[i] * weight);
                            }
                        } else {
                            insertSingleJointAttribute(joint, weight);
                        }
                    }

                    addJoint(jm[joints1[i]], weight);
                    addJoint(jm[joints2[i]], 1.0 - weight);

                    for (let i = addedJointAttributes; i < m_jointsWidth; i++) {
                        m_jointsIds.push(0);
                        m_weights.push(0);
                    }
                }
            }
        }

        if (packet.Blend.R && packet.Blend.R.length) {
            for (const i in packet.Blend.R) {
                m_colors.push(packet.Blend.R[i], packet.Blend.G[i], packet.Blend.B[i], packet.Blend.A[i]);
            }
        }

        if (packet.Uvs.U && packet.Uvs.U.length) {
            for (const i in packet.Uvs.U) {
                m_textures.push(packet.Uvs.U[i], packet.Uvs.V[i]);
            }
        }

        if (packet.Norms.X && packet.Norms.X.length) {
            for (const i in packet.Norms.X) {
                m_normals.push(packet.Norms.X[i], packet.Norms.Y[i], packet.Norms.Z[i]);
            }
        }

        offset += vertexCount;
    }

    let mesh = new RenderMesh(m_vertexes, m_indexes);
    if (m_colors.length) {
        mesh.setBlendColors(m_colors);
    }
    if (m_textures.length) {
        mesh.setUVs(m_textures);
    }
    if (m_normals.length) {
        mesh.setNormals(m_normals);
    }

    if (object.JointMappers && object.JointMappers.length) {
        if (jm && jm.length && m_jointsWidth && m_jointsIds.length && m_weights.length) {
            if ((m_jointsIds.length / m_jointsWidth) != m_vertexes.length / 3) {
                log.error("joints ids array not matching vertexes length array");
            }
            if ((m_weights.length / m_jointsWidth) != m_vertexes.length / 3) {
                log.error("joints weights array not matching vertexes length array");
            }
            mesh.setJointIds(m_jointsWidth, m_jointsIds, m_weights);
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

                    const mesh = parseMeshPackets(data, object, dmaPackets, iInstance);
                    mesh.meta['part'] = iPart;
                    mesh.meta['group'] = iGroup;
                    mesh.meta['object'] = iObject;
                    mesh.setUseBindToJoin(object.UseInvertedMatrix);
                    mesh.setMaterialID(object.MaterialId);
                    if (object.TextureLayersCount != 1) {
                        mesh.setLayer(iLayer);
                    }
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

    let mdl = new RenderModel();

    let dumplinkobj = getActionLinkForWadNode(wad, nodeid, 'obj');
    dataSummary.append($('<a class="center">').attr('href', dumplinkobj).append('Download .obj'));

    let dumplinkgltf = getActionLinkForWadNode(wad, nodeid, 'gltf');
    dataSummary.append($('<a class="center">').attr('href', dumplinkgltf).append('Download .glb bin glTF 2.0'));

    let table = loadMeshFromAjax(mdl, data, true);
    dataSummary.append(table);

    //console.log(data.Vectors);
    //console.log(data);

    gr_instance.addNode(new ObjectTreeNodeModel("mesh", mdl));
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
    m_weights.length = sPos.length * 2;

    for (let i in sPos) {
        let j = i * 3;
        let pos = sPos[i]
        m_vertexes[j + 0] = pos[0];
        m_vertexes[j + 1] = pos[1];
        m_vertexes[j + 2] = pos[2];
        m_weights[i * 2 + 0] = pos[3];
        m_weights[i * 2 + 1] = 1.0 - pos[3];
    }

    for (let i in m_indexes) {
        m_indexes[i] -= streamStart;
    }

    let mesh = new RenderMesh(m_vertexes, m_indexes);
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
        const jm = object.JointsMap;
        const sBoni = streams["BONI"].Values.slice(streamStart, streamStart + streamCount);
        const width = 2;

        // console.log(sBoni);
        
        let joints = [];
        joints.length = sBoni.length * width;

        for (const i in sBoni) {
            for (let j = 0; j < width; j++) {
                joints[i * width + j] = jm[sBoni[i][j]];
            }
        }
        mesh.setJointIds(width, joints, m_weights);
    }

    //console.log(originalMeshObject.Type, originalMeshObject);	
    if (originalMeshObject) {
        if (originalMeshObject.Type == 0x1d) {
            mesh.setps3static(true);
            if (object.JointsMap.length !== 1 || object.JointsMap[0] !== 0) {
                console.warn("ps3 static jm unexpected", object.JointsMap);
            }
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

    let model = new RenderModel();

    let table = loadGmdlFromAjax(model, data, undefined, true);
    dataSummary.append(table);

    let node = new ObjectTreeNodeModel("gmdl", model);
    gr_instance.addNode(node);

    gr_instance.requestRedraw();
}

function loadMdlFromAjax(data, parseScripts = false, needTable = false) {
    let model = new RenderModel();

    let tables = [];
    if (data.Meshes && data.Meshes.length) {
        for (let mesh of data.Meshes) {
            let mdlTables;
            //if (false) { // (!!data.GMDL) {
            if (!!data.GMDL) {                
                // GMDL static meshes already translated to position?
                mdlTables = loadGmdlFromAjax(model, data.GMDL, mesh, needTable);
            } else {
                mdlTables = loadMeshFromAjax(model, mesh, needTable);
            }
            tables.push(mdlTables);
        }
    }

    for (let iMaterial in data.Materials) {
        let material = new RenderMaterial();

        let textures = data.Materials[iMaterial].Textures;
        let rawMat = data.Materials[iMaterial].Mat;
        material.setColor(rawMat.Color);

        for (let iLayer in rawMat.Layers) {
            let rawLayer = rawMat.Layers[iLayer];
            let layer = new RenderMaterialLayer();

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
                    txrs.push(new RenderTexture('data:image/png;base64,' + imgs[iImg].Image));
                }
                layer.setTextures(txrs);
                layer.setHasAlphaAttribute(textures[iLayer].HaveTransparent);
            }
            material.addLayer(layer);
        }

        model.addMaterial(material);

        const anim = data.Materials[iMaterial].Animations;
        if (anim && anim.Groups && anim.Groups.length) {
            const group = anim.Groups[0];
            if (!group.IsExternal && group.Clips && group.Clips.length) {
                for (const iClip in group.Clips) {
                    const clip = group.Clips[iClip];
                    for (const iDataType in anim.DataTypes) {
                        switch (anim.DataTypes[iDataType].TypeId) {
                            case 8:
                                const animInstance = new AnimationMatertialLayer(anim, clip, iDataType, material);
                                animInstance.enable();
                                ga_instance.addAnimation(animInstance);
                                break;
                            case 9:
                                const animSheetInstance = new AnimationMaterialSheet(anim, clip, iDataType, material);
                                ga_instance.addAnimation(animSheetInstance);
                                break;
                        }
                    }
                }
            }
        }
    }

    if (parseScripts && data.Scripts) {
        for (const script of data.Scripts) {
            switch (script.TargetName) {
                case "SCR_Sky":
                    model.setType("sky");
                    break;
                default:
                    console.warn("Unknown SCR target: " + script.TargetName, data, model, script);
                    break;
            }
        }
    }
    return [model, tables];
}

function summaryLoadWadMdl(data, wad, nodeid) {
    gr_instance.cleanup();
    set3dVisible(true);

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

    let dumplinkgltf = getActionLinkForWadNode(wad, nodeid, 'gltf');
    dataSummary.append($('<a class="center">').attr('href', dumplinkgltf).append('Download .glb bin glTF 2.0'));

    // console.log(data);
    // for (const mesh of data.Meshes) {
    //     for (const i in mesh.Vectors) {
    //         const vec = mesh.Vectors[i];
    //         console.log(i, vec, vec.Value);
    //         //if (vec.Unk00 > 100) {
    //         if (!(i === 0 || i === mesh.Vectors.length - 1)) {
    //         //if (vec.Unk00 === 65494) {
    //             continue;   
    //         }
    //         let pos = vec.Value;
    //         let size = pos[0];
    //         pos = [
    //             pos[1], pos[2], pos[3],
    //             //pos[1], pos[0], pos[3],
    //         ]
    //         let mat = mat4.fromTranslation(mat4.create(), pos);

    //         let model = new RenderModel();
    //         model.addMesh(RenderHelper.SphereLinesMesh(pos[0], pos[1], pos[2], size, 15, 15));
    //         //model.addMesh(RenderHelper.CubeLinesMesh(pos[0], pos[1], pos[2], size, size, size, true));
    //         gr_instance.addNode(new ObjectTreeNodeModel("wtfmesh", model));

    //         const jointText = new RenderTextMesh(`${i}`, true, 10);
    //         jointText.setOffset(-0.5, -0.5);
    //         if (vec.Unk00 === 65535) {
    //             jointText.setColor(1.0, 0.0, 1.0, 0.3);
    //         } else {
    //             jointText.setColor(0.0, 1.0, 1.0, 0.3);
    //         }
    //         let textNode = new ObjectTreeNodeModel("vectorstrange", jointText);
    //         textNode.setLocalMatrix(mat);
    //         gr_instance.addNode(textNode);
    //     }
    // }

    let [model, mdlTables] = loadMdlFromAjax(data, false, true);
    for (let mdlTable of mdlTables) {
        dataSummary.append(mdlTable);
    }

    let node = new ObjectTreeNodeModel("model", model);
    gr_instance.addNode(node);

    // gr_instance.models.push(mdl);
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

function loadCollisionFromAjax(data, wad, nodeid, parentObject = null) {
    let collisionNode = new ObjectTreeNode("collision");

    const adddebugmaterial = function(model, r, g, b, a) {
        let renderMaterial = new RenderMaterial();
        let layer = new RenderMaterialLayer();
        renderMaterial.setColor([r, g, b, a]);
        layer.setColor([1, 1, 1, 1]);
        layer.setMethodAdditive();
        renderMaterial.addLayer(layer);
        model.addMaterial(renderMaterial);
    }

    const loaddebug = function(mdMesh, jointId) {
        let vertices = [];
        for (let vec of mdMesh.Vertices) {
            vertices.push(vec[0], vec[1], vec[2]);
        }
        let mesh = new RenderMesh(vertices, mdMesh.Indices, gl.LINES);
        mesh.setMaskBit(8);
        mesh.setMaterialID(0);
        let model = new RenderModel(collisionNode)
        model.addMesh(mesh);
        adddebugmaterial(model, 0.7, 0, 0.7, 0.3);
        let node = new ObjectTreeNodeModel(`mdbg`, model);
        if (parentObject) {
            parentObject.joints[jointId].addNode(node);
        } else {
            collisionNode.addNode(node);
        }
    };

    if (data.ShapeName == "BallHull") {
        let ball = data.Shape;

        let color = [1.0, 1.0, 1.0];
        switch (ball.Type) {
            case 0:
                color = [1.0, 0.0, 0.0];
                break;
            case 1:
                color = [0.0, 0.0, 1.0];
                break;
            case 2:
                color = [0.0, 1.0, 1.0];
                break;
            case 3:
                color = [0.0, 1.0, 0.0];
                break;
        }

        for (let iMesh in ball.Meshes) {
            let mesh = ball.Meshes[iMesh];

            let calculatedVertices = [];
            let calculatedIndices = [];
            let pointIndices = []; {
                let planes = [];
                for (let vec of mesh.Planes) {
                    planes.push(new Plane([vec[0], vec[1], vec[2]], vec[3]));
                }

                const [intersections, indices] = Plane.planesIntersectionsEdjes(planes);
                for (const v of intersections) {
                    calculatedVertices.push(v[0], v[1], v[2]);
                }
                for (const i in intersections) {
                    pointIndices.push(i);
                }
                calculatedIndices = indices;
            }
            let calculatedMesh = new RenderMesh(calculatedVertices, calculatedIndices, gl.LINES);
            calculatedMesh.setMaskBit(4);
            calculatedMesh.setMaterialID(0);
            let calculatedPointsMesh = new RenderMesh(calculatedVertices, pointIndices, gl.POINTS);
            calculatedPointsMesh.setMaskBit(4);
            calculatedPointsMesh.setMaterialID(0);

            let calculatedModel = new RenderModel(collisionNode)
            calculatedModel.addMesh(calculatedMesh);
            calculatedModel.addMesh(calculatedPointsMesh);
            adddebugmaterial(calculatedModel, color[0], color[1], color[2], 0.3);

            let node = new ObjectTreeNodeModel(`calculated`, calculatedModel);
            if (parentObject) {
                parentObject.joints[mesh.Joint].addNode(node);
            } else {
                collisionNode.addNode(node);
            }

            if (ball.DbgMesh) {
                if (iMesh < ball.DbgMesh.Meshes.length) {
                    loaddebug(ball.DbgMesh.Meshes[iMesh], mesh.Joint);
                } else {
                    console.error("Failed to get dbgmesh for mesh", nodeid, iMesh, ball);
                }
            }
        }

        for (let iBall in ball.Balls) {
            let b = ball.Balls[iBall];

            const vec = [b.Coord[0], b.Coord[1], b.Coord[2]];
            const size = b.Coord[3];

            let mesh = RenderHelper.SphereLinesMesh(vec[0], vec[1], vec[2], size, 7, 7);
            mesh.setMaskBit(4);
            mesh.setMaterialID(0);
            let model = new RenderModel();
            model.addMesh(mesh);
            adddebugmaterial(model, color[0], color[1], color[2], 0.3);

            let node = new ObjectTreeNodeModel(`ball ${iBall}`, model);
            if (parentObject) {
                parentObject.joints[b.Joint].addNode(node);
            } else {
                collisionNode.addNode(node);
            }
        }

        //let vec = ball.BSphere;
        //let mesh = grHelper_SphereLines(vec[0], vec[1], vec[2], vec[3], 7, 7, ball.BSphereJoint);
        //mdl.addMesh(mesh);
        //mesh.setMaskBit(4);
    } else if (data.ShapeName == "mCDbgHdr") {
        for (let mdMesh of data.Shape.Meshes) {
            loaddebug(mdMesh, 0);
        }
    } else if (data.ShapeName == "SheetHdr") {
        let form = $('<form action="' + getActionLinkForWadNode(wad, nodeid, 'frommodel') + '" method="post" enctype="multipart/form-data">');
        form.append($('<input type="file" name="model">'));
        let replaceBtn = $('<input type="button" value="Replace static collision geometry">')
        replaceBtn.click(function() {
            let form = $(this).parent();
            $.ajax({
                url: form.attr('action'),
                type: 'post',
                data: new FormData(form[0]),
                processData: false,
                contentType: false,
                success: function(a1) {
                    if (a1 !== "") {
                        alert('Error: ' + a1);
                    } else {
                        alert('Success!');
                        window.location.reload();
                    }
                }
            });
        });
        form.append(replaceBtn);
        dataSummary.append(form);

        let rib = data.Shape;

        for (let materialId = 0; materialId < rib.Some4Materials.length; materialId++) {
            let material = rib.Some4Materials[materialId];

            let table = $('<table>');
            table.append($('<tr><td>Name</td><td>' + material.Name + '</td></tr>'));
            let colorTd = $('<td>');
            let c = material.EditorColor;
            table.append($('<tr><td>DebugMat</td><td>' + material.EditorMaterial + '</td></tr>'));
            colorTd.attr('style', 'background-color: rgb(' +
                parseInt(c[0] * 255) + ',' + parseInt(c[1] * 255) + ',' + parseInt(c[2] * 255) + ')');
            table.append($('<tr><td>DebugColor</td></tr>').append(colorTd));

            for (let key in material.Values) {
                table.append($('<tr><td>' + key + '</td><td>' + material.Values[key] + '</td></tr>'));
            };
            dataSummary.append(table);

            let indices = [];
            let vertices = [];
            for (let i = 0; i < rib.Some9Points.length; i++) {
                vertices.push(rib.Some9Points[i][0], rib.Some9Points[i][1], rib.Some9Points[i][2]);
            }
            for (let i = 0; i < rib.Some8QuadsIndex.length; i++) {
                let quad = rib.Some8QuadsIndex[i];
                if (quad.MaterialIndex == materialId) {
                    indices.push(quad.Indexes[0], quad.Indexes[1], quad.Indexes[2]);
                    indices.push(quad.Indexes[3], quad.Indexes[0], quad.Indexes[2]);
                }
            }
            for (let i = 0; i < rib.Some7TrianglesIndex.length; i++) {
                let triangle = rib.Some7TrianglesIndex[i];
                if (triangle.MaterialIndex == materialId) {
                    indices.push(triangle.Indexes[0], triangle.Indexes[1], triangle.Indexes[2]);
                }
            }

            let model = new RenderModel();
            let renderMaterial = new RenderMaterial();
            renderMaterial.setColor(material.EditorColor);

            let layer = new RenderMaterialLayer();
            layer.setColor([1, 1, 1, 0.4]);
            layer.setMethodAdditive();
            renderMaterial.addLayer(layer);

            model.addMaterial(renderMaterial);

            if (indices.length) {
                let mesh = new RenderMesh(vertices, indices);
                mesh.setMaterialID(0);
                mesh.setMaskBit(7);
                model.addMesh(mesh);
            }

            collisionNode.addNode(new ObjectTreeNodeModel("sheet", model));
        }
    }
    return collisionNode;
}

function loadObjFromAjax(data, parseScripts = false) {
    const oNode = new ObjectTreeNodeSkinned("object");

    const joints = data.Data.Joints;
    for (const iJoint in joints) {
        const joint = joints[iJoint];

        const jNode = new ObjectTreeNodeSkinJoint(joint.Name, joint.BindToJointMat);
        jNode.setLocalMatrix(joint.ParentToJoint);
        if (joint.IsSkinned) {
            jNode.setBindToJointMatrix(joint.BindToJointMat);
        }

        if (joint.Parent < 0) {
            oNode.addNode(jNode);
        } else {
            oNode.joints[joint.Parent].addNode(jNode);
        }

        oNode.addJoint(jNode);

        const jointText = new RenderTextMesh(iJoint, true, 10);
        jointText.setOffset(-0.5, -0.5);
        jointText.setColor(1.0, 1.0, 1.0, 0.3);
        jointText.setMaskBit(1);
        jNode.addNode(new ObjectTreeNodeModel("label", jointText));
    }

    if (data.Model) {
        const [model, mdlTables] = loadMdlFromAjax(data.Model, true, true);
        /*
        for (const mdlTable of mdlTables) {
            dataSummary.append(mdlTable);
            console.log(mdlTable);
        }
        */
        oNode.addNode(new ObjectTreeNodeModel("model", model));
    }
    if (data.Collisions) {
        for (const collision of data.Collisions) {
            loadCollisionFromAjax(collision, null, null, oNode);
        }
    }

    if (data.Script) {
        if (data.Script.TargetName == "SCR_Entities") {
            $.each(data.Script.Data.Array, function(entity_id, entity) {
                let entityNode = new ObjectTreeNode();
                entityNode.setLocalMatrix(new Float32Array(data.Data.Joints[0].RenderMat));

                let text3d = new RenderTextMesh("\x05" + entity.Name, true);
                text3d.setOffset(-0.5, -0.5);
                text3d.setMaskBit(3);

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

                oNode.addNode(new ObjectTreeNodeModel("entity", text3d));
            });
        }
    }

    oNode.addNode(new ObjectTreeNodeModel("tree", RenderHelper.SkeletLines(joints)));

    return oNode;
}

function summaryLoadWadObj(data, wad, nodeid) {
    gr_instance.cleanup();

    set3dVisible(true);
    const oNode = loadObjFromAjax(data);

    let dumplinkgltf = getActionLinkForWadNode(wad, nodeid, 'gltf');
    dataSummary.append($('<a class="center">').attr('href', dumplinkgltf).append('Download .glb bin glTF 2.0'));

    let jointsTable = $('<table>');

    if (data.Animations) {
        let $animSelector = $("<select>").attr("size", 6).addClass("animation");

        const anim = data.Animations;
        if (anim && anim.Groups && anim.Groups.length) {
            for (const iGroup in anim.Groups) {
                const group = anim.Groups[iGroup];
                for (const iClip in group.Clips) {
                    const clip = group.Clips[iClip];
                    for (const iDataType in anim.DataTypes) {
                        switch (anim.DataTypes[iDataType].TypeId) {
                            case 0:
                                let $option = $("<option>").text(group.Name + ": " + clip.Name);
                                $option.dblclick([anim, clip, iDataType, data.Data], function(ev) {
                                    let anim = new AnimationObjectSkelet(ev.data[0], ev.data[1], ev.data[2], ev.data[3], oNode);
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

    dataSummary.append($("<div>").text("File0x20: " + (data.Data.File0x20 >>> 0).toString(2)));
    dataSummary.append($("<div>").text("File0x24: " + (data.Data.File0x24 >>> 0).toString(2)));

    $.each(data.Data.Joints, function(joint_id, joint) {
        let row = $('<tr>').append($('<td>').attr('style', 'background-color:rgb(' +
                parseInt((joint.Id % 8) * 15) + ',' +
                parseInt(((joint.Id / 8) % 8) * 15) + ',' +
                parseInt(((joint.Id / 64) % 8) * 15) + ');')
            .append(joint.Id).attr("rowspan", 13));

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

    gr_instance.addNode(oNode);
    gr_instance.requestRedraw();
}

function loadGameObjectFromAjax(inst, parseScripts = true) {
    let instNode = new ObjectTreeNode(inst.Name);

    const text3d = new RenderTextMesh("\x04" + inst.Name, true);
    text3d.setOffset(-0.5, -0.5);
    text3d.setColor(1.0, 1.0, 1.0, 0.8);
    text3d.setMaskBit(6);

    if (inst.IsGow2) {
        let instMat = mat4.fromTranslation(mat4.create(), inst.Position);
        // instNode.setLocalMatrix(instMat);

        let text = new ObjectTreeNodeModel("label", text3d);
        text.setLocalMatrix(instMat);
        instNode.addNode(text);

    } else {
        const rs = (180.0 / Math.PI);
        let rot = quat.fromEuler(quat.create(), inst.Rotation[0] * rs, inst.Rotation[1] * rs, inst.Rotation[2] * rs);
        const scale = inst.Rotation[3];

        let instMat = mat4.fromRotationTranslationScale(mat4.create(), rot, inst.Position1, [scale, scale, scale]);

        if (inst.Position1[3] != 1.0) {
            console.warn("posmulincorrect", inst);
        }

        instNode.addNode(new ObjectTreeNodeModel("label", text3d));
        instNode.setLocalMatrix(instMat);
    }

    if (inst.Object) {
        const oNode = loadObjFromAjax(inst.Object, parseScripts);
        instNode.addNode(oNode);
    }

    return instNode;
}


function summaryLoadWadGameObject(data, parseScripts = true) {
    gr_instance.cleanup();
    dataSummary.empty();
    set3dVisible(true);

    const node = loadGameObjectFromAjax(data, parseScripts);

    let table = $('<table>');
    for (let k in data) {
        if (k != "Object") {
            table.append($('<tr>').append($('<td>').text(k)).append($('<td>').text(JSON.stringify(data[k]))));
        }
    }
    dataSummary.append(table);

    gr_instance.addNode(node);
    gr_instance.flushScene();
    gr_instance.requestRedraw();
}

function loadCxtFromAjax(data, parseScripts = true) {
    let cxtNode = new ObjectTreeNode(data.Name);

    for (let i in data.Instances) {
        let inst = data.Instances[i];
        let instNode = loadGameObjectFromAjax(inst, false, parseScripts);
        cxtNode.addNode(instNode);
    }

    return cxtNode;
}

function summaryLoadWadCxt(data, wad, nodeid) {
    if (!gw_cxt_group_loading) {
        gr_instance.cleanup();
        dataSummary.empty();

        let dumplinkgltf = getActionLinkForWadNode(wad, nodeid, 'gltf');
        dataSummary.append($('<a class="center">').attr('href', dumplinkgltf).append('Download .glb bin glTF 2.0'));
    } else {
        let dumplinkgltf = getActionLinkForWadNode(wad, nodeid, 'gltf_all');
        dataSummary.append($('<a class="center">').attr('href', dumplinkgltf).append('Download .glb bin glTF 2.0'));
    }

    if ((data.Instances !== null && data.Instances.length) || gw_cxt_group_loading) {
        set3dVisible(true);
        const node = loadCxtFromAjax(data);
        gr_instance.addNode(node);
        gr_instance.flushScene();
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

    let mdl = new RenderModel();
    mdl.addMesh(new RenderMesh(m_vertexes, m_indexes));

    gr_instance.models.push(mdl);
    gr_instance.requestRedraw();
}

function summaryLoadWadScript(data, wad, tagid) {
    gr_instance.cleanup();

    dataSummary.append($("<h3>").append("Scirpt " + data.TargetName));

    if (data.TargetName == 'SCR_Entities') {
        let asJsonButton = $('<button>').text("Download as json").click(function() {
            window.open(getActionLinkForWadNode(wad, tagid, 'dataasjson'), '_blank');
        });
        dataSummary.append($('<p>').append(asJsonButton));

        let uploadFromJsonButton = $('<button>').text("Upload from json");
        uploadFromJsonButton.attr("href", getActionLinkForWadNode(wad, tagid, 'datafromjson'));
        uploadFromJsonButton.click(function() {
            console.log($(this).attr('href'));
            uploadAjaxHandler.call(this);
        });
        dataSummary.append($('<p>').append(uploadFromJsonButton));

        for (let i in data.Data.Array) {
            let e = data.Data.Array[i];

            let ht = $("<table>").append($("<tr>").append($("<td>").attr("colspan", 2).append(e.Name)));
            for (let j in e) {
                let v = e[j];
                switch (j) {
                    case "Variables":
                        for (let hi in v) {
                            ht.append(
                                $("<tr>").append($("<td>").append(
                                    'Variable name #' + (parseInt(hi) + e.PhysicsObjectId)))
                                .append($("<td>").append(v[hi].Name + " (type " + v[hi].Type + ")")));
                        }
                        break;
                    case "Handlers":
                        for (let ha of v) {
                            ht.append(
                                $("<tr>").append($("<td>").append('Handler #' + ha.Id))
                                .append($("<td style='white-space: pre;'>").append(ha.Decompiled.join('<br>'))));
                        }
                        break;
                    case "DebugTargetEntitiesNames":
                    case "DebugReferencedByNames":
                        ht.append($("<tr>").append($("<td>").append(j)).append($("<td>").append(v.join('<br>'))));
                        break;
                    case "Matrix":
                    case "DebugReferencedBy":
                    case "TargetEntitiesIds":
                        v = JSON.stringify(v);
                        ht.append($("<tr>").append($("<td>").append(j)).append($("<td>").append(v)));
                        break;
                    case "Name":
                        break;
                    default:
                        ht.append($("<tr>").append($("<td>").append(j)).append($("<td>").append(v)));
                        break;
                }
            }
            dataSummary.append(ht);
        }
    }

    set3dVisible(false);
}


function summaryLoadWadRSRCS(data, wad, nodeid) {
    set3dVisible(false);

    let list = $("<ul>");

    let newRSRCSWAd = function(name) {
        let text = $("<span>").text(name);
        let delbtn = $("<button>").text("remove").click(function() {
            $(this).parent().remove();
        });


        list.append($('<li>').append(text).append(delbtn));
    };

    $.each(data.Wads, function(k, val) {
        newRSRCSWAd(val);
    });

    let addWad = $("<input type='text' placeholder='wadname'>");
    let addBtn = $("<button>").text("add").click(function() {
        newRSRCSWAd(addWad.val());
    })

    let saveBtn = $("<button>").text("save").click(function() {
        let params = {};
        $("ul").find("li span").each(function(k, v) {
            params["wad" + k] = $(v).text();
        });

        $.ajax({
            url: getActionLinkForWadNode(wad, nodeid, 'update'),
            data: params,
            success: function(a) {
                if (a != "" && a.error) {
                    alert("Error: " + a.error);
                } else {
                    alert("Success");
                }
            }
        });
    });

    dataSummary.append(addWad).append(addBtn).append(list).append(saveBtn);
}

function summaryLoadWadTWK(data, wad, nodeid) {
    set3dVisible(false);

    let table = $("<table>");
    let twk = data;

    let dumpYamlLink = getActionLinkForWadNode(wad, nodeid, 'asyaml');

    let info = $("<ul>");
    info.append($("<li>").append("Name: " + twk.Name));
    info.append($("<li>").append("MagicHeaderPresened: " + twk.MagicHeaderPresened));
    info.append($("<li>").append("HeaderStrangeMagicUid: " + twk.HeaderStrangeMagicUid));
    info.append($("<li>").append($("<a>").attr('href', dumpYamlLink).append("Download yaml")));
    dataSummary.append(info);

    let form = $('<form action="' + getActionLinkForWadNode(wad, nodeid, 'fromyaml') + '" method="post" enctype="multipart/form-data">');
    form.append($('<input type="file" name="data">'));
    let replaceBtn = $('<input type="button" value="Upload from yaml">')
    replaceBtn.click(function() {
        let form = $(this).parent();
        $.ajax({
            url: form.attr('action'),
            type: 'post',
            data: new FormData(form[0]),
            processData: false,
            contentType: false,
            success: function(a1) {
                if (a1 !== "") {
                    alert('Error: ' + a1);
                } else {
                    alert('Success!');
                    window.location.reload();
                }
            }
        });
    });
    form.append(replaceBtn);
    dataSummary.append(form);

    let valueView = function(value) {
        let bytes = Uint8Array.from(atob(value), c => c.charCodeAt(0));
        let view = new DataView(bytes.buffer, 0);
        let asString = '';
        for (let c of bytes) {
            if (c == 0) {
                break;
            }
            asString += String.fromCharCode(c);
        }

        let s = "int32: " + view.getInt32(0, true) + " uint32: " + view.getUint32(0, true) +
            " float: " + view.getFloat32(0, true) +
            "</br>string: " + asString +
            "</br>hex: " + bytesToHexString(bytes);
        return s;
    }

    let directoryToTable;
    directoryToTable = function(directory) {
        let table = $("<table>");

        // console.log(directory);
        if (directory.Fields) {
            for (let subdir of directory.Fields) {
                if (subdir.Value) {
                    table.append($("<tr>").append(
                        $("<td>").append(subdir.Name),
                        $("<td>").append(valueView(subdir.Value))
                    ));
                } else {
                    table.append($("<tr>").append(
                        $("<td style='vertical-align: top;'>").append(subdir.Name),
                        $("<td>").append(directoryToTable(subdir))
                    ));
                }
            }
        }

        return table;
    };

    dataSummary.append(directoryToTable(twk.Tree));
    console.log(twk);
}

function hexStringToBytes(string) {
    var bytes = new Uint8Array(Math.ceil(string.length / 2));
    for (var i = 0; i < bytes.length; i++) bytes[i] = parseInt(string.substr(i * 2, 2), 16);
    return bytes;
}

function bytesToHexString(bytes) {
    var string = '';
    for (var i = 0; i < bytes.length; i++) {
        if (bytes[i] < 16) string += '0';
        string += bytes[i].toString(16);
    }
    return string;
}