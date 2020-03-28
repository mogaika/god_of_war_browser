'use strict';

var viewFs, viewPack, viewTree, viewSummary, view3d;
var dataFs, dataPack, dataTree, dataSummary, data3d;
var defferedLoadingWad;
var defferedLoadingWadNode;
var dataSelectors, dataSummarySelectors;
var wad_last_load_view_type = 'nodes';

String.prototype.replaceAll = function(search, replace) {
    if (replace === undefined) {
        return this.toString();
    }
    return this.replace(new RegExp('[' + search + ']', 'g'), replace);
};

function getActionLinkForWadNode(wad, nodeid, action, params = '') {
    return '/action/' + wad + '/' + nodeid + '/' + action + '?' + params;
}

function treeInputFilterHandler($el, localStorageKey) {
    var filterText = $el.val().toLowerCase();
    if (localStorageKey) {
        localStorage.setItem(localStorageKey, filterText);
    }
    $el.parent().find("div li label").each(function(a1, a2, a3) {
        var p = $(this).parent();
        if ($(this).text().toLowerCase().includes(filterText)) {
            while (p.is("li")) {
                p.show();
                p = p.parent().parent();
            }
        } else {
            p.hide();
        }
    });
};

function treePackInputFilterHandler() {
    treeInputFilterHandler($(this), 'tree-filter');
};

function treeItemInputFilterHandler() {
    treeInputFilterHandler($(this), 'item-filter');
};

function set3dVisible(show) {
    if (show) {
        view3d.show();
        viewSummary.attr('style', '');
        gr_instance.setInterfaceCameraMode(false);
        gr_instance.onResize();
    } else {
        view3d.hide();
        viewSummary.attr('style', 'flex-grow:1;');
    }
}

function setTitle(view, title) {
    $(view).children(".view-item-title").text(title);
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
    dataSelectors.empty();
    $.getJSON('/json/pack', function(files) {
        var list = $('<ol>');
        for (var i in files) {
            var fileName = files[i];
            list.append($('<li>')
                .attr('filename', fileName)
                .append($('<label>').append(fileName))
                .append($('<a download>')
                    .addClass('button-dump')
                    .attr('title', 'Download file')
                    .attr('href', '/dump/pack/' + fileName))
                .append($('<div>')
                    .addClass('button-upload')
                    .attr('title', 'Upload your version of file')
                    .attr("href", '/upload/pack/' + fileName)
                    .click(uploadAjaxHandler)));
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

function driverFsLoad() {
    dataFs.empty();
    $.getJSON('/json/fs', function(files) {
        var list = $('<ol>');
        for (var i in files) {
            var fileName = files[i];
            list.append($('<li>')
                .attr('filename', fileName)
                .append($('<label>').append(fileName))
                .append($('<a download>')
                    .addClass('button-dump')
                    .attr('title', 'Download file')
                    .attr('href', '/dump/fs/' + fileName)));
        }
        dataFs.append(list);
        console.log('fs loaded');
    })
}

function uploadAjaxHandler() {
    var link = $(this).attr("href");
    var form = $('<form action="' + link + '" method="post" enctype="multipart/form-data">');
    var fileInput = $('<input type="file" name="data">');
    form.append(fileInput);

    fileInput.trigger("click");
    fileInput.change(function() {
        if (fileInput[0].files.length == 0) {
            return;
        }

        $.ajax({
            url: form.attr('action'),
            type: 'post',
            data: new FormData(form[0]),
            processData: false,
            contentType: false,
            success: function(a1) {
                if (a1 !== "") {
                    alert('Error uploading: ' + a1);
                } else {
                    alert('Success!');
                    window.location.reload();
                }
            }
        });
    });
}

function packLoadFile(filename) {
    dataTree.empty();
    dataSummary.empty();
    dataSelectors.empty();
    flp_obj_view_history = [{
        TypeArrayId: 8,
        IdInThatTypeArray: 0
    }];
    $.getJSON('/json/pack/' + filename, function(data) {
        var ext = filename.slice(-3).toLowerCase();
        switch (ext) {
            case 'wad':
            case 'ps3':
            case 'sp2':
                treeLoadWad(filename, data);
                break;
            case 'psw':
            case 'pss':
                treeLoadPswPss(filename, data);
                break;
            case 'vag':
            case 'va1':
            case 'va2':
            case 'va3':
            case 'va4':
            case 'va5':
            case 'vpk':
            case 'vp1':
            case 'vp2':
            case 'vp3':
            case 'vp4':
                treeLoadVagVpk(filename, data);
                break;
            case 'txt':
                treeLoadTxt(filename, data);
                break;
            default:
                dataTree.append(JSON.stringify(data, undefined, 2).replaceAll('\n', '<br>'));
                break;
        }
        console.log('pack file ' + filename + ' loaded');
    });
}

function treeLoadVagVpk(filename, data) {
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

function treeLoadTxt(filename, data) {
    set3dVisible(false);
    setTitle(viewTree, filename);
    dataSummary.append($("<p>").append(data));
    setLocation(filename, '#/' + filename);
}

function treeLoadPswPss(filename, data) {
    set3dVisible(false);
    setTitle(viewTree, filename);
    var videoPath = '/dump/pack/' + filename;

    var vlc = $('<EMBED pluginspage="http://www.videolan.org"\
	    type="application/x-vlc-plugin"\
	    version="VideoLAN.VLCPlugin.2"\
	    width="640"\
	    height="360"\
	    toolbar="true"\
	    loop="false"\
	    name="vlc">');
    vlc.attr('target', videoPath);
    dataSummary.append(vlc);

    setLocation(filename, '#/' + filename);
}

function treeLoadWad(wadName, data) {
    setTitle(viewTree, wadName);
    if (!defferedLoadingWadNode) {
        setLocation(wadName, '#/' + wadName);
    }

    dataSelectors.append($('<div class="item-selector">').click(function() {
        treeLoadWadAsNodes(wadName, data);
    }).text("Nodes"));
    dataSelectors.append($('<div class="item-selector">').click(function() {
        treeLoadWadAsTags(wadName, data);
    }).text("Tags"));

    if (wad_last_load_view_type === 'nodes') {
        treeLoadWadAsNodes(wadName, data);
    } else if (wad_last_load_view_type === 'tags') {
        treeLoadWadAsTags(wadName, data);
    }
}

function inputAsRenderMask(selector, bitIndex, init) {
    inputAsSwitch(selector, function(checked) {
        let bit = 1 << bitIndex;
        gr_instance.setFilterMask((gr_instance.filterMask & (~bit)) | (checked ? bit : 0));
    }, init);
}

function inputAsSwitch(selector, updatefunc, init) {
    let $input = $(selector);

    let storageItem = localStorage.getItem(selector);
    $input[0].checked = (storageItem != null) ? (storageItem == "true") : init;
    updatefunc($input[0].checked);
    $input.change(function() {
        updatefunc(this.checked);
        localStorage.setItem(selector, this.checked);
        gr_instance.requestRedraw();
    })
}

function goFullscreen(element) {
    if (element.requestFullscreen) {
        element.requestFullscreen();
    } else if (element.mozRequestFullScreen) {
        element.mozRequestFullScreen();
    } else if (element.webkitRequestFullscreen) {
        element.webkitRequestFullscreen();
    } else if (element.msRequestFullscreen) {
        element.msRequestFullscreen();
    }
}

$(document).ready(function() {
    viewFs = $('#view-fs');
    viewPack = $('#view-pack');
    viewTree = $('#view-tree');
    viewSummary = $('#view-summary');
    view3d = $('#view-3d');

    dataFs = viewFs.children('.view-item-container');
    dataPack = viewPack.children('.view-item-container');
    dataTree = viewTree.children('.view-item-container');
    dataSelectors = viewTree.children('.view-item-selectors');
    dataSummary = viewSummary.children('.view-item-container');
    dataSummarySelectors = viewSummary.children('.view-item-selectors');
    data3d = view3d.children('.view-item-container');

    $('div.collapse-button').each(function(index, el) {
        $(el).click(function(ev) {
            let $this = $(this);
            let $parent = $this.parent();

            console.log($parent, $this.text());
            if (!$parent.hasClass('collapsed')) {
                $parent.addClass('collapsed');
                $this.text('>> S');
            } else {
                $parent.removeClass('collapsed');
                $this.text('<< HIDE');
            }
        });
    });

    var packFilter = localStorage.getItem('tree-filter');
    var itemFilter = localStorage.getItem('item-filter');
    $('#view-pack-filter').on('input', treePackInputFilterHandler).val(packFilter ? packFilter : '.wad');
    $('#view-item-filter').on('input', treeItemInputFilterHandler).val(itemFilter ? itemFilter : '');

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
    driverFsLoad();

    gwInitRenderer(data3d);

    inputAsRenderMask("#view-3d-config input#show-skeleton-ids", 1, true);
    inputAsRenderMask("#view-3d-config input#show-skeleton", 2, true);
    inputAsRenderMask("#view-3d-config input#show-entity", 3, true);
    inputAsRenderMask("#view-3d-config input#show-collision", 4, true);
    inputAsRenderMask("#view-3d-config input#show-light", 5, true);
    inputAsSwitch("#view-3d-config input#enable-backface-culling", function(enable) {
        gr_instance.cull = enable;
    }, false);

    gaInit();
});

function hexdump(buffer, blockSize) {
    var table = $('<table>');
    blockSize = blockSize || 16;
    var lines = [];
    var hex = "0123456789ABCDEF";
    var blocks = Math.ceil(buffer.length / blockSize);
    for (var iBlock = 0; iBlock < blocks; iBlock += 1) {
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
        tr.append($('<td>').text(chars));
        table.append(tr);
    }
    return table;
}