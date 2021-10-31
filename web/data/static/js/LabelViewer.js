$(document).ready(function() {
    let data3d = $('.view-item-container');
    gwInitRenderer(data3d);
    gr_instance.setInterfaceCameraMode(true);

    let url = new URL(document.location);
    let packfile = url.searchParams.get('f');
    let flpid = url.searchParams.get('r');
    let commands = JSON.parse(url.searchParams.get('c'));
    let rootMatrix = url.searchParams.get('m');
    if (!rootMatrix) {
        rootMatrix = mat4.create();
    } else {
        rootMatrix = JSON.parse(rootMatrix);
    }

    $.getJSON('/json/pack/' + packfile + '/' + flpid, function(resp) {
        let flp = resp.Data;
        let flpdata = flp.FLP;

        let font = undefined;
        let x = 0;
        let y = 0;
        let fontscale = 1.0;

        let matmap = {};

        for (const cmd of commands) {
            if (cmd.Flags & 8) {
                font = flpdata.Fonts[flpdata.GlobalHandlersIndexes[cmd.FontHandler].IdInThatTypeArray];
                fontscale = cmd.FontScale;
            }
            if (cmd.Flags & 2) {
                x = cmd.OffsetX;
            }
            if (cmd.Flags & 1) {
                y = cmd.OffsetY;
            }

            for (const glyph of cmd.Glyphs) {
                let chrdata = font.MeshesRefs[glyph.GlyphId];

                if (chrdata.MeshPartIndex !== -1) {
                    let mdl = new RenderModel();
                    meshes = loadMeshPartFromAjax(mdl, flp.Model.Meshes[0], chrdata.MeshPartIndex);

                    if (chrdata.Materials && chrdata.Materials.length !== 0 && chrdata.Materials[0].TextureName) {
                        let txr_name = chrdata.Materials[0].TextureName;

                        if (!matmap.hasOwnProperty(txr_name)) {
                            let material = new RenderMaterial();
                            let img = flp.Textures[txr_name].Images[0].Image

                            let texture = new RenderTexture('data:image/png;base64,' + img);
                            texture.markAsFontTexture();

                            let layer = new RenderMaterialLayer();
                            layer.setTextures([texture]);
                            material.addLayer(layer);

                            matmap[txr_name] = material;
                        }
                        mdl.addMaterial(matmap[txr_name]);
                        for (let iMesh in meshes) {
                            meshes[iMesh].setMaterialID(0);
                        }
                    }

                    let matrix = mat4.translate(mat4.create(), rootMatrix, [x, y, 0]);
                    matrix = mat4.scale(matrix, matrix, [fontscale, fontscale, fontscale]);

                    let node = new ObjectTreeNodeModel("glyph", mdl);
                    node.setLocalMatrix(matrix);
                    
                    gr_instance.addNode(node);
                }

                x += glyph.Width;
            }

        }
        gr_instance.flushScene();
        gr_instance.requestRedraw();
    });


});