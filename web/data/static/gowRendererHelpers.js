function grHelper_PivotMesh(size) {
    if (size == undefined) {
        size = 1000;
    }
    var vertexData = [
        size, 0, 0, -size, 0, 0,
        0, size, 0,
        0, -size, 0,
        0, 0, size,
        0, 0, -size,
    ]
    var colorData = [
        0xff, 0x00, 0x00, 0xff,
        0x00, 0x00, 0x00, 0xff,
        0x00, 0xff, 0x00, 0xff,
        0x00, 0x00, 0x00, 0xff,
        0x00, 0x00, 0xff, 0xff,
        0x00, 0x00, 0x00, 0xff,
    ]
    var indexData = [
        0, 1, 2, 3, 4, 5,
    ]

    var mesh = new grMesh(vertexData, indexData, gl.LINES)
    mesh.setBlendColors(colorData);
    return mesh;
}

function grHelper_Cube(x, y, z, size) {
    if (size == undefined) {
        size = 50;
    }
    var vertexData = [
        x - size, y - size, z - size, x + size, y - size, z - size, x + size, y + size, z - size, x - size, y + size, z - size,
        x - size, y - size, z + size, x + size, y - size, z + size, x + size, y + size, z + size, x - size, y + size, z + size,
        x - size, y - size, z - size, x - size, y + size, z - size, x - size, y + size, z + size, x - size, y - size, z + size,
        x + size, y - size, z - size, x + size, y + size, z - size, x + size, y + size, z + size, x + size, y - size, z + size,
        x - size, y - size, z - size, x - size, y - size, z + size, x + size, y - size, z + size, x + size, y - size, z - size,
        x - size, y + size, z - size, x - size, y + size, z + size, x + size, y + size, z + size, x + size, y + size, z - size,
    ]
    var indexData = [
        0, 1, 2, 0, 2, 3, 4, 5, 6, 4, 6, 7,
        8, 9, 10, 8, 10, 11, 12, 13, 14, 12, 14, 15,
        16, 17, 18, 16, 18, 19, 20, 21, 22, 20, 22, 23
    ]

    return new grMesh(vertexData, indexData, gl.TRIANGLES)
}

function grHelper_CubeLines(x, y, z, size_x, size_y, size_z, diaglines = true) {
    if (size_x == undefined) {
        size_x = 50;
    }
    if (size_y == undefined) {
        size_y = size_x;
    }
    if (size_z == undefined) {
        size_z = size_x;
    }

    var vertexData = [
        x - size_x, y - size_y, z - size_z, x + size_x, y - size_y, z - size_z, x + size_x, y + size_y, z - size_z, x - size_x, y + size_y, z - size_z,
        x - size_x, y - size_y, z + size_z, x + size_x, y - size_y, z + size_z, x + size_x, y + size_y, z + size_z, x - size_x, y + size_y, z + size_z,
        x - size_x, y - size_y, z - size_z, x - size_x, y + size_y, z - size_z, x - size_x, y + size_y, z + size_z, x - size_x, y - size_y, z + size_z,
        x + size_x, y - size_y, z - size_z, x + size_x, y + size_y, z - size_z, x + size_x, y + size_y, z + size_z, x + size_x, y - size_y, z + size_z,
        x - size_x, y - size_y, z - size_z, x - size_x, y - size_y, z + size_z, x + size_x, y - size_y, z + size_z, x + size_x, y - size_y, z - size_z,
        x - size_x, y + size_y, z - size_z, x - size_x, y + size_y, z + size_z, x + size_x, y + size_y, z + size_z, x + size_x, y + size_y, z - size_z,
    ]

    var indexData = diaglines ? [
        0, 1, 1, 2, 0, 2, 2, 3, 4, 5, 5, 6, 4, 6, 6, 7,
        8, 9, 9, 10, 8, 10, 10, 11, 12, 13, 13, 14, 12, 14, 14, 15,
        16, 17, 17, 18, 16, 18, 18, 19, 20, 21, 21, 22, 20, 22, 22, 23
    ] : [
        0, 1, 1, 2, 2, 3, 4, 5, 5, 6,
        8, 9, 9, 10, 10, 11, 12, 13, 13, 14,
        16, 17, 17, 18, 18, 19, 20, 21, 21, 22
    ];

    var mesh = new grMesh(vertexData, indexData, gl.LINES);
    mesh.setJointIds([0], Array(vertexData.length / 3).fill(0));
    return mesh;
}

function grHelper_Pivot(size) {
    var mdl = new grModel();
    mdl.addMesh(grHelper_PivotMesh(size));
    return mdl;
}

function grHelper_SphereLines(x, y, z, radius, latitudeBands, longitudeBands) {
    var vertexData = [];
    for (var latNumber = 0; latNumber <= latitudeBands; latNumber++) {
        var theta = latNumber * Math.PI / latitudeBands;
        var sinTheta = Math.sin(theta);
        var cosTheta = Math.cos(theta);

        for (var longNumber = 0; longNumber <= longitudeBands; longNumber++) {
            var phi = longNumber * 2 * Math.PI / longitudeBands;
            var sinPhi = Math.sin(phi);
            var cosPhi = Math.cos(phi);

            vertexData.push(x + radius * cosPhi * sinTheta);
            vertexData.push(y + radius * cosTheta);
            vertexData.push(z + radius * sinPhi * sinTheta);
        }
    }

    var indexData = [];
    for (var latNumber = 0; latNumber < latitudeBands; latNumber++) {
        for (var longNumber = 0; longNumber < longitudeBands; longNumber++) {
            var first = (latNumber * (longitudeBands + 1)) + longNumber;
            var second = first + longitudeBands + 1;
            indexData.push(first);
            indexData.push(second);
            //indexData.push(first + 1);

            indexData.push(second);
            indexData.push(second + 1);
            //indexData.push(first + 1);
        }
    }

    var mesh = new grMesh(vertexData, indexData, gl.LINES);
    mesh.setJointIds([0], Array(vertexData.length / 3).fill(0));
    return mesh;
}