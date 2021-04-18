function gowTransform2d(pos, matrix) {
    this.matrix = mat4.create();
    if (pos !== undefined) {
        this.matrix[12] = pos[0];
        this.matrix[13] = pos[1];
    }
    if (matrix !== undefined) {
        this.matrix[0] = matrix[0];
        this.matrix[1] = matrix[1];
        this.matrix[4] = matrix[2];
        this.matrix[5] = matrix[3];
    }
    //this.pos = (pos === undefined) ? vec2.create() : pos;
    //this.matrix = (matrix === undefined) ? mat2.create() : matrix;
}

gowTransform2d.prototype.applyTransform = function(otherTransform) {
    //let result = new gowTransform2d(
    //	vec2.fromValues(this.pos[0] + otherTransform.pos[0], this.pos[1] + otherTransform.pos[1]),
    //	mat2.mul(mat2.create(), this.matrix, otherTransform.matrix));
    // console.log("AAAAPLLYY", this, otherTransform, result);
    let result = new gowTransform2d();
    mat4.mul(result.matrix, this.matrix, otherTransform.matrix);
    return result;
}

gowTransform2d.prototype.fromTransform = function(source) {
    mat4.copy(this.matrix, source.matrix);
}

gowTransform2d.prototype.scale = function(scale) {
    return mat4.scale(this.matrix, this.matrix, [scale, scale, scale]);
}

gowTransform2d.prototype.toMatrix3d = function() {
    // console.log(this);
    return this.matrix;
    return mat4.fromValues(
        this.matrix[0], this.matrix[1], 0, 0,
        this.matrix[2], this.matrix[3], 0, 0,
        0, 0, 1, 0,
        this.pos[0], this.pos[1], 0, 1);
}

function gowFlp(flp) {
    this.root = flp;
    this.data = flp.FLP;
    this.mdls = [];
    this.texmap = {};
    this.claimed = {};
}

gowFlp.prototype.claimTextures = function() {
    this.claimed = {};
    for (txr in this.texmap) {
        this.texmap[txr].claim();
        this.claimed[txr] = this.texmap[txr];
    }
}

gowFlp.prototype.unclaimTextures = function() {
    for (txr in this.claimed) {
        this.claimed[txr].unclaim();
        if (this.texmap[txr].refs === 0) {
            delete this.texmap[txr];
        }
    }
}

gowFlp.prototype.getObjArrByType = function(_type) {
    switch (_type) {
        case 1:
            return this.data.MeshPartReferences;
        case 3:
            return this.data.Fonts;
        case 4:
            return this.data.StaticLabels;
        case 5:
            return this.data.DynamicLabels;
        case 6:
            return this.data.Datas6;
        case 7:
            return this.data.Datas7;
        case 9:
            return this.data.Transformations;
        case 10:
            return this.data.BlendColors;
    };
};

gowFlp.prototype.getObjByHandler = function(h) {
    if (h.TypeArrayId == 8) {
        return this.data.Data8;
    }
    let arr = this.getObjArrByType(h.TypeArrayId);
    return arr ? arr[h.IdInThatTypeArray] : undefined;
}

gowFlp.prototype.getTransformFromObject = function(transform) {
    return new gowTransform2d([transform.OffsetX, transform.OffsetY], transform.Matrix);
}

gowFlp.prototype.cacheTexture = function(texture_name) {
    // return texture from cache or creates new
    if (!this.texmap.hasOwnProperty(texture_name)) {
        let texture;
        if (this.root.Textures[texture_name].Images.length) {
            let img = this.root.Textures[texture_name].Images[0].Image;
            texture = new grTexture('data:image/png;base64,' + img);
            texture.markAsFontTexture();
        } else {
            texture = gr_instance.emptyTexture;
        }
        this.texmap[texture_name] = texture;
    }
    return this.texmap[texture_name];
}

gowFlp.prototype.renderData2 = function(o, handler, frameIndex, transform, color) {
    if (o.MeshPartIndex < 0) {
        return [];
    }

    let model = new grModel();

    // console.log("MESH PART INDEX", o.MeshPartIndex);
    meshes = loadMeshPartFromAjax(model, this.root.Model.Meshes[0], o.MeshPartIndex);

    for (let iMesh in meshes) {
        let mesh = meshes[iMesh];
        if (mesh.meta.hasOwnProperty('object')) {
            meshes[iMesh].setMaterialID(mesh.meta.object);
        } else {
            meshes[iMesh].setMaterialID(0);
        }
    }

    if (o.Materials && o.Materials.length !== 0) {
        for (let iMaterial in o.Materials) {
            let flpMaterial = o.Materials[iMaterial];
            let material = new grMaterial();
            let layer = new grMaterialLayer();
            if (flpMaterial.TextureName) {
                layer.setTextures([this.cacheTexture(flpMaterial.TextureName)]);
            }
            layer.setHasAlphaAttribute();
            let newColor = [];
            for (let i = 0; i < 4; i++) {
                newColor[i] = color[i] * (((flpMaterial.Color >> (8 * i)) & 0xff) / 257);
            }
            layer.setColor(newColor);
            material.addLayer(layer);
            model.addMaterial(material);
        }
    }

    // console.log("rendered data2 ", o, handler, transform.pos, transform.Matrix, color);

    model.matrix = transform.toMatrix3d();

    // console.log("MODELS FROM DATA2", [model]);
    return [model];
}

gowFlp.prototype.renderData4 = function(o, handler, frameIndex, transform, color) {
    let elementTransform = this.getTransformFromObject(o.Transformation);
    let font, fontscale;
    let x = 0;
    let y = 0;
    let models = [];
    let baseColor = color;
    let commands = o.RenderCommandsList;

    transform = transform.applyTransform(elementTransform);

    for (let iCmd = 0; iCmd < commands.length; iCmd++) {
        let cmd = commands[iCmd];

        if (cmd.Flags & 8) {
            font = this.data.Fonts[this.data.GlobalHandlersIndexes[cmd.FontHandler].IdInThatTypeArray];
            fontscale = cmd.FontScale;
        }
        if (cmd.Flags & 4) {
            color = [];
            for (let i = 0; i < baseColor.length; i++) {
                color.push(baseColor[i] * (cmd.BlendColor[i] / 255.0));
            }
        }
        if (cmd.Flags & 2) {
            x = cmd.OffsetX;
        }
        if (cmd.Flags & 1) {
            y = cmd.OffsetY;
        }

        for (let iGlyph = 0; iGlyph < cmd.Glyphs.length; iGlyph++) {
            let glyph = cmd.Glyphs[iGlyph];
            let data2 = font.MeshesRefs[glyph.GlyphId];

            let charTransform = transform.applyTransform(new gowTransform2d([x, y]));
            charTransform.scale(cmd.FontScale);

            models = models.concat(this.renderData2(data2, handler, frameIndex, charTransform, color));
            x += glyph.Width;
        }
    }
    //console.log("MODELS FROM DATA4", models);
    return models;
}

gowFlp.prototype.renderData6sub1 = function(o, handler, frameIndex, transform, color) {
    let models = [];
    for (let iElement = 0; iElement < o.ElementsAnimation.length; iElement++) {
        let flpElement = o.ElementsAnimation[iElement];
        // search current frame
        let currentFrame;

        for (let iFrame = 0; iFrame < flpElement.KeyFrames.length; iFrame++) {
            currentFrame = flpElement.KeyFrames[iFrame];
            // console.info(frameIndex, "+ 1 >=", currentFrame.WhenThisFrameEnds);
            if (frameIndex >= currentFrame.WhenThisFrameEnds - 1) {
                continue;
            } else {
                break;
            }
        }
        if (currentFrame === undefined) {
            console.warn("DIDNT FOUND FRAME TO RENDER");
            continue;
        }

        let elementTransform = this.data.Transformations[currentFrame.TransformationId];
        let elementColor = this.data.BlendColors[currentFrame.ColorId];

        let newTransform = elementTransform ?
            transform.applyTransform(this.getTransformFromObject(elementTransform)) : transform;

        let newColor = color;
        if (elementColor) {
            newColor = [];
            for (let i in elementColor.Color) {
                newColor.push((elementColor.Color[i] / 256.0) * color[i]);
            }
        }

        // console.info("idx", handler.IdInThatTypeArray, "el", iElement, "frame",
        // 	currentFrame.WhenThisFrameEnds, newTransform.matrix, color, currentFrame.ElementHandler);	
        let elementModels = this.renderElementByHandler(currentFrame.ElementHandler, 0, newTransform, newColor);
        models = models.concat(elementModels);
    }
    //console.log("MODELS FROM DATA6sub1", models);
    return models;
}

gowFlp.prototype.renderElementByHandler = function(handler, frameIndex, transform, color) {
    // color in range [0, 1]
    if (transform === undefined) {
        transform = new gowTransform2d();
    }
    if (frameIndex === undefined) {
        frameIndex = 0;
    }
    if (color === undefined) {
        color = [1, 1, 1, 1];
    }

    let o = this.getObjByHandler(handler);

    // console.info("rendering item", handler, o);

    switch (handler.TypeArrayId) {
        case 0:
            return [];
        case 1:
            return this.renderData2(o, handler, frameIndex, transform, color);
        case 4:
            return this.renderData4(o, handler, frameIndex, transform, color);
        case 6:
            return this.renderData6sub1(o.Sub1, handler, frameIndex, transform, color);
        case 7:
        case 8:
            return this.renderData6sub1(o, handler, frameIndex, transform, color);
    }
    console.warn("unknown render for handler", handler, "object", o);
    return [];
}

function summaryLoadWadFlp(flp, wad, tagid) {
    let flpdata = flp.FLP;

    let object_renderer_handler = {
        TypeArrayId: 1,
        IdInThatTypeArray: 2
    };
    let object_renderer_frame = 0;

    let flp_print_dump = function() {
        set3dVisible(false);
        dataSummary.empty();
        dataSummary.append($("<pre>").append(JSON.stringify(flpdata, null, "  ").replaceAll('\n', '<br>')));
    }

    const objNamesArray = ['Nothing', 'Textured mesh part', 'UNKNOWN', 'Font',
        'Static label', 'Dynamic label', 'Data6', 'Data7',
        'Root', 'Transform', 'Color'
    ];

    let get_obj_arr_by_id = function(t) {
        switch (t) {
            case 1:
                return flpdata.MeshPartReferences;
                break;
            case 3:
                return flpdata.Fonts;
                break;
            case 4:
                return flpdata.StaticLabels;
                break;
            case 5:
                return flpdata.DynamicLabels;
                break;
            case 6:
                return flpdata.Datas6;
                break;
            case 7:
                return flpdata.Datas7;
                break;
            case 9:
                return flpdata.Transformations;
                break;
            case 10:
                return flpdata.BlendColors;
                break;
        };
        return undefined;
    };

    let get_obj_by_handler = function(h) {
        if (h.TypeArrayId == 8) {
            return flpdata.Data8;
        } else {
            let arr = get_obj_arr_by_id(h.TypeArrayId);
            if (arr) {
                return arr[h.IdInThatTypeArray];
            } else {
                return undefined;
            }
        }
    }

    let flp_scripts_strings = function() {
        set3dVisible(false);
        dataSummary.empty();

        let tbody = $("<tbody>");
        for (let iRef in flp.ScriptPushRefs) {
            let tr = $("<tr>");
            let ref = flp.ScriptPushRefs[iRef];

            let str = atob(ref.String);
            if (flp.FontCharAliases) {
                let originalStr = str;
                str = "";
                for (let i = 0; i < originalStr.length; i++) {
                    let charCode = originalStr.charCodeAt(i);
                    let replaced = false;

                    for (let charToReplace in flp.FontCharAliases) {
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
                    let str = $(this).parent().parent().find('input[type="text"]').val();
                    let id = Number.parseInt($(this).parent().parent().children().first().text());

                    if (flp.FontCharAliases) {
                        let originalStr = str;
                        str = "";
                        for (let char of originalStr) {
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

        let headtr = $("<tr>");
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

        let font = undefined;
        let cmdsContainer = $("<td>");
        for (let iCmd in sl.RenderCommandsList) {
            let rcmds = $("<table width='100%'>");
            let cmd = sl.RenderCommandsList[iCmd];

            if (cmd.Flags & 8) {
                let fhi = $("<input type=text id='fonthandler' class=no-width>").val(cmd.FontHandler);
                let fsi = $("<input type=text id='fontscale' class=no-width>").val(cmd.FontScale);
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
                let bclri = $("<input type=text id='blendclr'>").val(JSON.stringify(cmd.BlendColor));
                rcmds.append($("<tr>").append($("<td>").text("Set blend color")).append($("<td>").append(bclri)));
            }
            let xoi = $("<input type=text id='xoffset'>").val(cmd.OffsetX);
            rcmds.append($("<tr>").append($("<td>").text("Set X offset")).append($("<td>").append(xoi)));
            let yoi = $("<input type=text id='yoffset'>").val(cmd.OffsetY);
            rcmds.append($("<tr>").append($("<td>").text("Set Y offset")).append($("<td>").append(yoi)));

            let str = cmd.Glyphs.reduce(function(str, glyph) {
                let char = font.CharNumberToSymbolIdMap.indexOf(glyph.GlyphId);
                if (flp.FontCharAliases) {
                    let map_chars = Object.keys(flp.FontCharAliases).filter(function(charString) {
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

        let open_preview_for_label = function(sl) {
            let u = new URLSearchParams();
            u.append('c', JSON.stringify(sl.RenderCommandsList));
            u.append('f', wad);
            u.append('r', tagid);

            let t = sl.Transformation;
            let m = t.Matrix;

            u.append('m', JSON.stringify([m[0], m[1], 0, 0, m[2], m[3], 0, 0, 0, 0, 1, 0, t.OffsetX, t.OffsetY, 0, 1]));
            window.open('/label.html?' + u, '_blank');
        }

        let get_label_from_table_tr = function(tr) {
            let sl = {
                'Transformation': JSON.parse(tr.find("td").last().find("textarea").last().val()),
                'RenderCommandsList': [],
            };

            let fontscale = 1.0;
            let fonthandler = -1;
            tr.find("table").each(function(cmdIndex, tbl) {
                let cmd = {
                    'Flags': 0
                };
                $(tbl).find("tr").each(function(i, row) {
                    let rname = $(row).find("td").first().text();
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
                        let text = $(row).find("textarea").val();
                        let glyphs = [];

                        let font = flpdata.Fonts[flpdata.GlobalHandlersIndexes[fonthandler].IdInThatTypeArray];
                        for (let char of text) {
                            let charCode = char.charCodeAt(0);
                            if (flp.FontCharAliases) {
                                if (flp.FontCharAliases.hasOwnProperty(charCode)) {
                                    charCode = flp.FontCharAliases[charCode];
                                }
                            }
                            let glyphId = font.CharNumberToSymbolIdMap[charCode];
                            let width = font.SymbolWidths[glyphId] * fontscale;
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

        let btns = $("<div>");
        btns.append($("<button>peview original</button>").click(sl, function(e) {
            open_preview_for_label(e.data);
        }));
        btns.append($("<br>"));
        btns.append($("<button>preview changes</button>").click(function(e) {
            open_preview_for_label(get_label_from_table_tr($(this).parent().parent().parent()));
        }));
        btns.append($("<br>"));
        btns.append($("<button>apply changes</button>").click(iSl, function(e) {
            let sl = get_label_from_table_tr($(this).parent().parent().parent());

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

        let $transform = $("<textarea id='matrix'>").css('height', '12em').val(JSON.stringify(sl.Transformation, null, ' '));
        row.append($("<td>").append($transform));
        return row;
    }

    let flp_list_labels = function() {
        set3dVisible(false);
        dataSummary.empty();

        let table = $("<table class='staticlabelrendercommandlist'>");
        table.append($("<tr>").append($("<td>").text("Id")).append($("<td>").text("Render commands")));

        for (let iSl in flpdata.StaticLabels) {
            table.append(print_static_label_as_tr(iSl));
        }

        dataSummary.append(table);
    }

    let flp_view_object_viewer = function() {
        dataSummary.empty();
        gr_instance.cleanup();
        set3dVisible(false);
        let $history_element = $("<div>").css('margin', '7px').css('white-space', 'nowrap').css('overflow', 'hidden');
        let $data_element = $("<div>");

        let element_view = function(h) {
            $data_element.empty();
            if (h == undefined) {
                h = flp_obj_view_history[0];
            } else {
                flp_obj_view_history.unshift(h);
            }

            {
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
                if (script) {
                    let code = script.Decompiled;
                    let $code_element = $("<div>").text(" > click to show decompiled script < ").css('cursor', 'pointer').click(function() {
                        $(this).empty().css('cursor', '').append(code).off('click');
                    })
                    return $code_element;
                } else {
                    return $("<div>").text("not implemented");
                }
            }

            let print_data6 = function() {
                print_data6_subtype1(obj.Sub1);
                let $events = $("<div>");

                let $events_table = $("<table>");
                for (let i in obj.Sub2s) {
                    let ev = obj.Sub2s[i]
                    let $event_table = $("<table>");
                    $event_table.append(
                        _row(_column("Mask"), _column(ev.EventKeysMask)),
                        _row(_column("Mask2"), _column(ev.EventUnkMask)),
                        _row(_column("Script"), _column(print_script(ev.Script))),
                    );
                    $events_table.append(_row(_column("event" + i), _column($event_table)));
                }
                $data_table.append(_row(_column("events"), _column($events_table)));
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
                        url: getActionLinkForWadNode(wad, tagid, 'transform'),
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
                if (child_h.TypeArrayId == 8) {
                    return parents;
                }
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
                let parse_parenting_data6_sub1 = function(h, o) {
                    for (let anim of o.ElementsAnimation) {
                        for (let frame of anim.KeyFrames) {
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
                for (let h of flpdata.GlobalHandlersIndexes) {
                    let o = get_obj_by_handler(h);

                    switch (h.TypeArrayId) {
                        case 4:
                            for (let rc of o.RenderCommandsList) {
                                if (rc.Flags & 8 != 0) {
                                    check_parenting(h, {
                                        TypeArrayId: 3,
                                        IdInThatTypeArray: rc.FontHandler
                                    });
                                }
                            }
                            break;
                        case 5:
                            check_parenting(h, o.FontHandler);
                            break;
                        case 6:
                            parse_parenting_data6_sub1(h, o.Sub1);
                            break;
                        case 7:
                            parse_parenting_data6_sub1(h, o);
                            break;
                        case 8:
                            parse_parenting_data6_sub1(h, o);
                            break;
                    }
                }
                parse_parenting_data6_sub1({
                    TypeArrayId: 8,
                    IdInThatTypeArray: 0
                }, flpdata.Data8);
                return parents;
            }

            let $table = $("<table>");

            let $header = $("<span>").text(" Viewing object " + objNamesArray[h.TypeArrayId] + "[" + h.IdInThatTypeArray + "]");

            let parents_list = [];
            let parents = get_parents(h);
            let curParentRow = _row();
            let colums_cnt = 6;
            for (let i in parents) {
                if (i != 0 && (i % colums_cnt == 0)) {
                    parents_list.push(curParentRow);
                    curParentRow = _row().attr('colspan', colums_cnt);
                }
                curParentRow.append(_column(print_ref_handler(parents[i])));
            }
            if (parents.length < colums_cnt || (parents.length % colums_cnt != 0)) {
                parents_list.push(curParentRow);
            }

            $table.append(_row(_column($header).attr('colspan', colums_cnt + 1)));
            if (parents.length != 0) {
                $table.append(_row(_column("parents").attr('rowspan', parents_list.length + 1)), parents_list);
            } else {
                $table.append(_row(_column("parents"), _column("no parents found")));
            }

            let $renderButton = $("<a class='flpobjref'>").click(function() {
                object_renderer_handler = h;
                flp_view_object_renderer();
            }).text('render');
            $table.append(_row(_column($renderButton).attr('colspan', colums_cnt + 1)));

            if (h.TypeArrayId != 8) {
                let $nav_row = _row();
                let arr = get_obj_arr_by_id(h.TypeArrayId);
                if (h.IdInThatTypeArray > 0 || h.IdInThatTypeArray + 1 < arr.length) {
                    if (h.IdInThatTypeArray > 0) {
                        $nav_row.append(_column("Prev:"));
                        $nav_row.append(_column(print_ref_handler({
                            TypeArrayId: h.TypeArrayId,
                            IdInThatTypeArray: h.IdInThatTypeArray - 1,
                        })));
                    }
                    if (h.IdInThatTypeArray + 1 < arr.length) {
                        $nav_row.append(_column("Next:"));
                        $nav_row.append(_column(print_ref_handler({
                            TypeArrayId: h.TypeArrayId,
                            IdInThatTypeArray: h.IdInThatTypeArray + 1,
                        })));
                    }
                    $table.append(_row(_column("nav"), _column($("<table>").append($nav_row))));
                }
            }

            $table.append(_row(_column($data_table).attr('colspan', colums_cnt + 1)));
            $data_element.append($table);
            $('#view-summary .view-item-container').animate({
                scrollTop: 0
            }, 200);
        }
        dataSummary.append($history_element, $data_element);
        element_view();
    }

    let flp_view_font = function() {
        gr_instance.cleanup();
        set3dVisible(true);
        gr_instance.setInterfaceCameraMode(true);
        dataSummary.empty();

        let importBMFontScale = $('<input id="importbmfontscale" type="number" min="0" max="20" value="1" step="0.1">');
        let importBMFontInput = $('<button>');
        importBMFontInput.text('Import glyphs from BMFont file');
        importBMFontInput.attr("href", getActionLinkForWadNode(wad, tagid, 'importbmfont')).click(function() {
            $(this).attr('href', getActionLinkForWadNode(wad, tagid, 'importbmfont', 'scale=' + $("#importbmfontscale").val()));
            console.log($(this).attr('href'));
            uploadAjaxHandler.call(this);
        });
        let importDiv = $('<div id="flpimportfont">');
        importDiv.append($('<label>').text('font scale').append(importBMFontScale));
        importDiv.append(importBMFontInput);
        importDiv.append($('<a>').text('Link to usage instruction').attr('target', '_blank')
            .attr('href', 'https://github.com/mogaika/god_of_war_browser/blob/master/LOCALIZATION.md'));
        dataSummary.append(importDiv);

        let charstable = $("<table>");

        let mdl = new grModel();
        let matmap = {};

        for (let iFont in flpdata.Fonts) {
            let font = flpdata.Fonts[iFont];
            charstable.append($("<tr>").append("font: " + iFont));
            charstable.append($("<tr>").append("reversed map (utf-16): " + !(font.Flags & 1)));

            for (let iChar in font.CharNumberToSymbolIdMap) {
                if (font.CharNumberToSymbolIdMap[iChar] == -1) {
                    continue;
                }

                let glyphId = font.CharNumberToSymbolIdMap[iChar];
                let char = iChar;

                if (!(font.Flags & 1)) {
                    char = glyphId;
                    glyphId = iChar;
                }

                if (glyphId >= font.CharsCount) {
                    continue;
                }

                let chrdata = font.MeshesRefs[glyphId];

                let meshes = [];
                if (chrdata.MeshPartIndex !== -1) {
                    meshes = loadMeshPartFromAjax(mdl, flp.Model.Meshes[0], chrdata.MeshPartIndex);
                    let txrid = undefined;
                    if (chrdata.Materials && chrdata.Materials.length !== 0 && chrdata.Materials[0].TextureName) {
                        let txr_name = chrdata.Materials[0].TextureName;

                        if (!matmap.hasOwnProperty(txr_name) &&
                            flp.Textures.hasOwnProperty(txr_name) &&
                            flp.Textures[txr_name].Images.length !== 0 &&
                            flp.Textures[txr_name].Images[0].hasOwnProperty('Image')) {
                            let img = flp.Textures[txr_name].Images[0].Image;

                            let material = new grMaterial();

                            let texture = new grTexture('data:image/png;base64,' + img);
                            texture.markAsFontTexture();

                            let layer = new grMaterialLayer();
                            layer.setTextures([texture]);
                            material.addLayer(layer);

                            matmap[txr_name] = mdl.materials.length;
                            mdl.addMaterial(material);
                        }
                        txrid = matmap[txr_name];
                    }
                    for (let iMesh in meshes) {
                        meshes[iMesh].setMaterialID(txrid);
                    }
                }

                let symbolWidth = font.SymbolWidths[glyphId];
                let cubemesh = grHelper_CubeLines(symbolWidth / 32, 0, 0, symbolWidth / 32, 500, 5, false);
                mdl.addMesh(cubemesh);
                meshes.push(cubemesh);

                let charS = String.fromCharCode(char);

                if (flp.FontCharAliases) {
                    let map_chars = Object.keys(flp.FontCharAliases).filter(function(charUnicode) {
                        return flp.FontCharAliases[charUnicode] == char
                    });
                    if (map_chars && map_chars.length !== 0) {
                        charS = String.fromCharCode(map_chars[0]);
                    }
                }

                let table = $("<table>");

                let tr1 = $("<tr>");
                let tr2 = $("<tr>");
                tr1.append($("<td>").text('#' + glyphId));
                tr1.append($("<td>").text('width ' + symbolWidth));
                tr1.append($("<td>").text('ansii ' + char));
                tr2.append($("<td>").append($("<h2>").text(charS)));
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

    let flp_view_object_renderer = function() {
        let h = object_renderer_handler;
        object_renderer_frame = 0;

        dataSummary.empty();

        let o = get_obj_by_handler(h);

        switch (h.TypeArrayId) {
            case 6:
                o = o.Sub1;
                break;
            case 7:
            case 8:
                break;
            default:
                dataSummary.append("Invalid object type for rendering");
                break;
        }

        set3dVisible(true);
        gr_instance.setInterfaceCameraMode(true);

        let frames_amount = o.hasOwnProperty('TotalFramesCount') ? o.TotalFramesCount : 1;

        dataSummary.append("Rendering object ");
        let $a = $("<a>").text(objNamesArray[h.TypeArrayId] + "[" + h.IdInThatTypeArray + "] ");
        $a.addClass('flpobjref').click(function() {
            flp_obj_view_history.unshift(h);
            flp_view_object_viewer();
        });
        dataSummary.append($("Rendering object "));

        let f = new gowFlp(flp);

        let $currentFrame = $("<p>");

        let renderFrame = function(frame) {
            if (frame === undefined) {
                frame = object_renderer_frame;
            } else if (frame == object_renderer_frame) {
                return;
            }
            $currentFrame.text("Current frame " + frame + " / " + frames_amount);
            object_renderer_frame = frame;

            f.claimTextures();

            gr_instance.cleanup();

            let elementsRenderModels = f.renderElementByHandler(object_renderer_handler, object_renderer_frame);
            gr_instance.models = gr_instance.models.concat(elementsRenderModels);
            // console.log("Rendered frame", frame);

            gr_instance.flushScene();
            gr_instance.requestRedraw();

            f.unclaimTextures();
        }


        dataSummary.append($a);
        $rangeInput = $('<input type="range" min="0" value="0">').attr("max", frames_amount - 1);

        $rangeInput.on('input', function(ev) {
            let newFrame = parseInt(this.value);
            this.value = newFrame;
            if (gr_instance.frameChecker == 0) {
                if (newFrame != object_renderer_frame) {
                    renderFrame(newFrame);
                    gr_instance.frameChecker = 1;
                }
            } else {
                console.warn("skipping frames");
            }
        });

        dataSummary.append($("<div>").append($currentFrame));
        dataSummary.append($("<div>").append($rangeInput));

        renderFrame();
    }

    dataSummarySelectors.append($('<div class="item-selector">').click(flp_list_labels).text("Labels editor"));
    dataSummarySelectors.append($('<div class="item-selector">').click(flp_print_dump).text("Dump"));
    dataSummarySelectors.append($('<div class="item-selector">').click(flp_scripts_strings).text("Scripts strings"));
    dataSummarySelectors.append($('<div class="item-selector">').click(flp_view_font).text("Font viewer"));
    dataSummarySelectors.append($('<div class="item-selector">').click(flp_view_object_viewer).text("Obj explorer"));
    dataSummarySelectors.append($('<div class="item-selector">').click(flp_view_object_renderer).text("Obj renderer"));

    // flp_list_labels();
    flp_view_object_viewer();
    flp_view_object_renderer();
}