'use strict';

var viewPack, viewTree, viewSummary, view3d;
var dataPack, dataTree, dataSummary, data3d;

function set3dVisible(show) {
    if (show) {
        view3d.show();
        viewSummary.attr('style', '')
    } else {
        view3d.hide();
        viewSummary.attr('style', 'width: auto;')
    }
}

function packLoad() {
    dataPack.empty();
    $.getJSON('/json/pack', function(data) {
        var list = $('<ol>');
        for (var i in data.Files) {
            var fileName = data.Files[i].Name;
            list.append($('<li>')
                    .attr('filename', fileName)
                    .append($('<label>').append(fileName))
                    .append($('<a>')
                            .addClass('button-dump')
                            .attr('href', '/dump/pack/' + fileName)) );
        }
        dataPack.append(list);
        
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
                    .append($('<label>').append(node.Name));
            
            if (node.IsLink) {
                // TODO: link visual
            } else {
                li.append($('<a>')
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
    if (data.Roots)
        dataTree.append(addNodes(data.Roots));
    $('#view-tree ol li label').click(function(ev) {
        var nodeformat = $(this).parent().attr('nodeformat');
        var nodeid = $(this).parent().attr('nodeid');
        var wadname = dataTree.children().attr('wadname');
        treeLoadWadNode(wadname, nodeid, parseInt(nodeformat));
    });
}

function treeLoadWadNode(wad, nodeid, format) {
    dataSummary.empty();
    $.getJSON('/json/pack/' + wad +'/' + nodeid, function(data) {
        switch (format) {
            case 0x00000007: // txr
                summaryLoadWadTxr(data);
                break;
            case 0x00000008: // material
                summaryLoadWadMat(data);
                break;
            case 0x0001000f: // mesh
                summaryLoadWadMesh(data);
                break;
            case 0x0000000c: // gfx pal
            default:
                set3dVisible(false);
                dataSummary.append(JSON.stringify(data, undefined, 2).replace('\n', '<br>'));
                break;
        }
        console.log('wad ' + wad + ' file ' + nodeid + ' loaded. format:' + format);
    });
}

function summaryLoadWadMesh(data) {
    set3dVisible(true);
    reset3d();
    
    console.log(data);
    
    var r_mesh = new Mesh();
    
    for (var iPart in data.Parts) {
        var part = data.Parts[iPart];
        for (var iGroup in part.Groups) {
            var group = part.Groups[iGroup]
            for (var iObject in group.Objects) {
                var object = group.Objects[iObject];
                for (var iPacket in object.Packets) {
                    var packet = object.Packets[iPacket];
                    for (var iBlock in packet.Blocks) {
                        var block = packet.Blocks[iBlock];
                        
                        var m_vertexes = [];
                        var m_indexes = [];
                        
                        m_vertexes.length = block.Trias.length * 3;
                        
                        for (var i in block.Trias) {
                            var tr = block.Trias[i];
                            var j = i * 3;
                            m_vertexes[j] = tr.X;
                            m_vertexes[j+1] = tr.Y;
                            m_vertexes[j+2] = tr.Z;
                            if (!tr.Skip) {
                                m_indexes.push(i-1);
                                m_indexes.push(i-2);
                                m_indexes.push(i-0);
                            }
                        }
                        
                        r_mesh.add(new MeshObject(m_vertexes, m_indexes));
                    }
                }
            }
        }
    }
    
    console.log(new Model(r_mesh));
    
    redraw3d();
}

function summaryLoadWadTxr(data) {
    set3dVisible(false);
    var table = $('<table>');
    $.each(data.Data, function(k, val) {
        table.append($('<tr>')
            .append($('<td>').append(k))
            .append($('<td>').append(val)));
    });
    dataSummary.append(table);
    for (var i in data.Images) {
        var img = data.Images[i];
        dataSummary.append($('<img>')
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
                        var txrobj = JSON.parse(data.Textures[l]);
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

$(document).ready(function(){
    viewPack = $('#view-pack');
    viewTree = $('#view-tree');
    viewSummary = $('#view-summary');
    view3d = $('#view-3d');
    
    dataPack = viewPack.children();
    dataTree = viewTree.children();
    dataSummary = viewSummary.children();
    data3d = view3d.children().children();
    
    packLoad();
    
    init3d(data3d);
});