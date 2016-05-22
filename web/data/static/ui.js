'use strict';

var viewPack, viewTree, viewSummary, view3d;
var dataPack, dataTree, dataSummary, data3d;

function set3dVisible(showOrHide) {
    if (showOrHide) {
        view3d.show();
        viewSummary.attr('style', '')
    } else {
        view3d.hide();
        viewSummary.attr('style', 'width: auto;')
    }
}

function packLoad() {
    $.getJSON('/json/pack', function(data) {
        dataPack.empty();
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
        console.log(filename, data);
        var ext = filename.slice(-3).toLowerCase();
        switch (ext) {
            case 'wad':
                dataTree.append(treeLoadWad(data));
                
                $('#view-tree ol li label').click(function(ev) {
                    var nodeformat = $(this).parent().attr('nodeformat');
                    var nodeid = $(this).parent().attr('nodeid');
                    var wadname = dataTree.children().attr('wadname');
                    treeLoadWadNode(wadname, nodeid, parseInt(nodeformat));
                });
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
        return addNodes(data.Roots);
    else
        return "";
}

function treeLoadWadNode(wad, nodeid, format) {
    dataSummary.empty();
    $.getJSON('/json/pack/' + wad +'/' + nodeid, function(data) {
        switch (format) {
            case 0x00000007: // txr
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
                    
                    dataSummary.append(
                        $('<img>')
                            .attr('src', 'data:image/png;base64,' + img.Image)
                            .attr('alt', 'gfx:' + img.Gfx + '  pal:' + img.Pal));
                }
                break;
            default:
                dataSummary.append(JSON.stringify(data, undefined, 2).replace('\n', '<br>'));
                break;
        }
        console.log('wad ' + wad + ' file ' + nodeid + ' loaded. format:' + format);
    });
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
    
    initGL(data3d);
});