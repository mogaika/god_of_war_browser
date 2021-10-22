'use strict';

let viewFs, viewPack, viewTree, viewSummary, view3d;
let dataFs, dataPack, dataTree, dataSummary, data3d;
let defferedLoadingWad;
let defferedLoadingWadNode;
let dataSelectors, dataSummarySelectors;
let wad_last_load_view_type = 'nodes';
let gw_cxt_group_loading = false;
let flp_obj_view_history = [{
    TypeArrayId: 8,
    IdInThatTypeArray: 0
}];

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
    let filterText = $el.val().toLowerCase();
    if (localStorageKey) {
        localStorage.setItem(localStorageKey, filterText);
    }
    $el.parent().find("div li label").each(function(a1, a2, a3) {
        let p = $(this).parent();
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
    let $title = $(view).children(".view-item-title");
    $title.empty();
    $title.append(title);
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
        let list = $('<ol>');
        for (let i in files) {
            const fileName = files[i];
            let li = $(`
                <li filename="${fileName}">
                    <label>${fileName}</label>
                    <a download class="button-dump" title="Download file" href="/dump/pack/${fileName}">
                    <div class="button-upload" title="Upload your version of file" href="/upload/pack/${fileName}">
                </li>
            `);
            li.find(".button-upload").click(uploadAjaxHandler);
            list.append(li);
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
    $.ajax({
        dataType: "json",
        url: '/json/fs',
        success: function(files) {
            let list = $('<ol>');
            for (let i in files) {
                let fileName = files[i];
                list.append($('<li>')
                    .attr('filename', fileName)
                    .append($('<label>').append(fileName))
                    .append($('<a download>')
                        .addClass('button-dump')
                        .attr('title', 'Download file')
                        .attr('href', '/dump/fs/' + fileName)));
            }
            dataFs.append(list);
            dataFs.append($('<p>').text('This window for downloading purposes only'));
            console.log('fs loaded');
        },
        error: function(err) {
            if (err.status == 405) {
                console.warn("Fs is not available for this driver. Hiding window.");
                viewFs.hide();
            }
        }
    });
}

function deleteAjaxHandler() {
    let filename = $(this).attr("filename");
    let ask1 = confirm(
        "You want to delete file\n" + filename +
        "\nDo not forget to backup before deletion!\n" +
        "Are you sure you want to delete file?");
    if (ask1 !== true) {
        return;
    }
    console.warn("Deleting toc file " + filename);

    $.ajax({
        url: $(this).attr("href"),
        processData: false,
        contentType: false,
        success: function(a1) {
            if (a1 !== "") {
                alert('Error deleting: ' + a1);
            } else {
                alert('Success!');
                window.location.reload();
            }
        }
    });
}

function uploadAjaxHandler() {
    let link = $(this).attr("href");
    let form = $('<form action="' + link + '" method="post" enctype="multipart/form-data">');
    let fileInput = $('<input type="file" name="data">');
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


    let $title = $("<span>" + filename + "&#x20;</span>");

    $title.append(
        $('<button>(DEL)</button>')
        .addClass('button-delete')
        .attr('title', 'Delete file')
        .attr('filename', filename)
        .attr("href", '/delete/pack/' + filename)
        .click(deleteAjaxHandler));
    setTitle(viewTree, $title);

    dataTree.append($("<div>loading file..." + filename + "</div>"));

    let onerror = function(error) {
        dataTree.append($("<div>failed to load:<b>" + error + "</b></div>"));
    }

    $.ajax({
        dataType: "json",
        url: '/json/pack/' + filename,
        error: onerror,
        success: function(data, a1, a2, a3) {
            if (data.hasOwnProperty('error')) {
                onerror(data['error']);
                return;
            }
            dataTree.empty();
            let ext = filename.slice(-3).toLowerCase();
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
        }
    });
}

function treeLoadVagVpk(filename, data) {
    set3dVisible(false);
    let list = $("<ul>");
    let wavPath = '/dump/pack/' + filename + '/wav';

    list.append($("<li>").append("SampleRate: " + data.SampleRate));
    list.append($("<li>").append("Channels: " + data.Channels));
    list.append($("<li>").append($("<a>").attr("href", wavPath).append("Download WAV")));
    dataTree.append(list)

    dataTree.append($("<audio controls autoplay>").append($("<source>").attr("src", wavPath)));

    setLocation(filename, '#/' + filename);
}

function treeLoadTxt(filename, data) {
    set3dVisible(false);
    dataSummary.append($("<p>").append(data));
    setLocation(filename, '#/' + filename);
}

function treeLoadPswPss(filename, data) {
    set3dVisible(false);
    let videoPath = '/dump/pack/' + filename;

    let vlc = $('<EMBED pluginspage="http://www.videolan.org"\
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

            if (!$parent.hasClass('collapsed')) {
                $parent.addClass('collapsed');
                $this.text('>> S');
            } else {
                $parent.removeClass('collapsed');
                $this.text('<< HIDE');
            }
            if (view3d.hasClass('collapsed')) {
                viewSummary.addClass('flexgrow');
            } else {
                viewSummary.removeClass('flexgrow');
            }
        });
    });

    let packFilter = localStorage.getItem('tree-filter');
    let itemFilter = localStorage.getItem('item-filter');
    $('#view-pack-filter').on('input', treePackInputFilterHandler).val(packFilter ? packFilter : '.wad');
    $('#view-item-filter').on('input', treeItemInputFilterHandler).val(itemFilter ? itemFilter : '');

    let urlParts = decodeURI(window.location.hash).split("/");
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
    gaInit();

    inputAsRenderMask("#view-3d-config input#show-skeleton-ids", 1, true);
    inputAsRenderMask("#view-3d-config input#show-skeleton", 2, true);
    inputAsRenderMask("#view-3d-config input#show-entity", 3, true);
    inputAsRenderMask("#view-3d-config input#show-collision", 4, true);
    inputAsRenderMask("#view-3d-config input#show-light", 5, true);
    inputAsRenderMask("#view-3d-config input#show-instance", 6, true);
    inputAsRenderMask("#view-3d-config input#show-collision-static", 7, true);
    inputAsRenderMask("#view-3d-config input#show-collision-dbg", 8, true);
   
    inputAsSwitch("#view-3d-config input#enable-backface-culling", function(enable) {
        gr_instance.cull = enable;
    }, false);
    inputAsSwitch("#view-3d-config input#enable-animation", function(enable) {
        ga_instance.active = enable;
    }, true);
    inputAsSwitch(".view-item-container input#enable-3d-helpers", function(enable) {
        let $helpers = $(".view-item-container #hidden-3d-helpers");
        if (enable) {
            $helpers.show();
        } else {
            $helpers.hide();
        }
    }, false);
    $("#hidden-3d-helpers #button-3d-reset-camera").click(function() {
        gr_instance.resetCamera();
    });
});

function hexdump(buffer, blockSize) {
    let table = $('<table>');
    blockSize = blockSize || 16;
    let lines = [];
    let hex = "0123456789ABCDEF";
    let blocks = Math.ceil(buffer.length / blockSize);
    for (let iBlock = 0; iBlock < blocks; iBlock += 1) {
        let blockPos = iBlock * blockSize;

        let line = '';
        let chars = '';
        for (let j = 0; j < Math.min(blockSize, buffer.length - blockPos); j += 1) {
            let code = buffer[blockPos + j];
            line += ' ' + hex[(0xF0 & code) >> 4] + hex[0x0F & code];
            chars += (code > 0x20 && code < 0x80) ? String.fromCharCode(code) : '.';
        }

        let tr = $('<tr>');
        tr.append($('<td>').append(("000000" + blockPos.toString(16)).slice(-6)));
        tr.append($('<td>').append(line));
        tr.append($('<td>').text(chars));
        table.append(tr);
    }
    return table;
}