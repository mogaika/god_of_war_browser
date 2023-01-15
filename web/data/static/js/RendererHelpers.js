class RenderHelper {
    static PivotMesh(size = 1000) {
        let vertexData = [
            size, 0, 0, -size, 0, 0,
            0, size, 0,
            0, -size, 0,
            0, 0, size,
            0, 0, -size,
        ]
        let colorData = [
            0xff, 0x00, 0x00, 0xff,
            0x00, 0x00, 0x00, 0xff,
            0x00, 0xff, 0x00, 0xff,
            0x00, 0x00, 0x00, 0xff,
            0x00, 0x00, 0xff, 0xff,
            0x00, 0x00, 0x00, 0xff,
        ]
        let indexData = [
            0, 1, 2, 3, 4, 5,
        ]

        let mesh = new RenderMesh(vertexData, indexData, gl.LINES);
        mesh.setBlendColors(colorData);
        return mesh;
    }

    static CubeMesh(x, y, z, size = 50) {
        let vertexData = [
            x - size, y - size, z - size, x + size, y - size, z - size, x + size, y + size, z - size, x - size, y + size, z - size,
            x - size, y - size, z + size, x + size, y - size, z + size, x + size, y + size, z + size, x - size, y + size, z + size,
            x - size, y - size, z - size, x - size, y + size, z - size, x - size, y + size, z + size, x - size, y - size, z + size,
            x + size, y - size, z - size, x + size, y + size, z - size, x + size, y + size, z + size, x + size, y - size, z + size,
            x - size, y - size, z - size, x - size, y - size, z + size, x + size, y - size, z + size, x + size, y - size, z - size,
            x - size, y + size, z - size, x - size, y + size, z + size, x + size, y + size, z + size, x + size, y + size, z - size,
        ]
        let indexData = [
            0, 1, 2, 0, 2, 3, 4, 5, 6, 4, 6, 7,
            8, 9, 10, 8, 10, 11, 12, 13, 14, 12, 14, 15,
            16, 17, 18, 16, 18, 19, 20, 21, 22, 20, 22, 23
        ]

        return new RenderMesh(vertexData, indexData, gl.TRIANGLES)
    }

    static CubeLinesMesh(x, y, z, size_x, size_y, size_z, diaglines = true) {
        if (size_x == undefined) {
            size_x = 50;
        }
        if (size_y == undefined) {
            size_y = size_x;
        }
        if (size_z == undefined) {
            size_z = size_x;
        }

        let vertexData = [
            x - size_x, y - size_y, z - size_z, x + size_x, y - size_y, z - size_z, x + size_x, y + size_y, z - size_z, x - size_x, y + size_y, z - size_z,
            x - size_x, y - size_y, z + size_z, x + size_x, y - size_y, z + size_z, x + size_x, y + size_y, z + size_z, x - size_x, y + size_y, z + size_z,
            x - size_x, y - size_y, z - size_z, x - size_x, y + size_y, z - size_z, x - size_x, y + size_y, z + size_z, x - size_x, y - size_y, z + size_z,
            x + size_x, y - size_y, z - size_z, x + size_x, y + size_y, z - size_z, x + size_x, y + size_y, z + size_z, x + size_x, y - size_y, z + size_z,
            x - size_x, y - size_y, z - size_z, x - size_x, y - size_y, z + size_z, x + size_x, y - size_y, z + size_z, x + size_x, y - size_y, z - size_z,
            x - size_x, y + size_y, z - size_z, x - size_x, y + size_y, z + size_z, x + size_x, y + size_y, z + size_z, x + size_x, y + size_y, z - size_z,
        ]

        let indexData = diaglines ? [
            0, 1, 1, 2, 0, 2, 2, 3, 4, 5, 5, 6, 4, 6, 6, 7,
            8, 9, 9, 10, 8, 10, 10, 11, 12, 13, 13, 14, 12, 14, 14, 15,
            16, 17, 17, 18, 16, 18, 18, 19, 20, 21, 21, 22, 20, 22, 22, 23
        ] : [
            0, 1, 1, 2, 2, 3, 4, 5, 5, 6,
            8, 9, 9, 10, 10, 11, 12, 13, 13, 14,
            16, 17, 17, 18, 18, 19, 20, 21, 21, 22
        ];

        let mesh = new RenderMesh(vertexData, indexData, gl.LINES);
        return mesh;
    }

    static Pivot(size = 1000) {
        let mdl = new RenderModel();
        mdl.addMesh(this.PivotMesh(size));
        return mdl;
    }

    static SphereLinesMesh(x, y, z, radius, latitudeBands, longitudeBands) {
        let vertexData = [];
        for (let latNumber = 0; latNumber <= latitudeBands; latNumber++) {
            let theta = latNumber * Math.PI / latitudeBands;
            let sinTheta = Math.sin(theta);
            let cosTheta = Math.cos(theta);

            for (let longNumber = 0; longNumber <= longitudeBands; longNumber++) {
                let phi = longNumber * 2 * Math.PI / longitudeBands;
                let sinPhi = Math.sin(phi);
                let cosPhi = Math.cos(phi);

                vertexData.push(x + radius * cosPhi * sinTheta);
                vertexData.push(y + radius * cosTheta);
                vertexData.push(z + radius * sinPhi * sinTheta);
            }
        }

        let indexData = [];
        for (let latNumber = 0; latNumber < latitudeBands; latNumber++) {
            for (let longNumber = 0; longNumber < longitudeBands; longNumber++) {
                let first = (latNumber * (longitudeBands + 1)) + longNumber;
                let second = first + longitudeBands + 1;
                indexData.push(first);
                indexData.push(second);
                //indexData.push(first + 1);

                indexData.push(second);
                indexData.push(second + 1);
                //indexData.push(first + 1);
            }
        }

        let mesh = new RenderMesh(vertexData, indexData, gl.LINES);
        return mesh;
    }

    static SkeletLines(skelet) {
        let model = new RenderModel();
        let vrtxs = [];
        let indxs = [];
        let clrs = [];
        let joints = [];
        let weights = [];

        for (const i in skelet) {
            const currentJoint = skelet[i];
            if (currentJoint.Parent < 0) {
                continue;
            }

            for (const joint of [currentJoint, skelet[currentJoint.Parent]]) {
                indxs.push(indxs.length);
                vrtxs.push(0);
                vrtxs.push(0);
                vrtxs.push(0);
                clrs.push((i % 8) * 15);
                clrs.push(((i / 8) % 8) * 15);
                clrs.push(((i / 64) % 8) * 15);
                clrs.push(127);
                joints.push(joint.Id);
                weights.push(1.0);
            }
        }

        let sklMesh = new RenderMesh(vrtxs, indxs, gl.LINES);
        sklMesh.setDepthTest(false);
        sklMesh.setBlendColors(clrs);
        sklMesh.setJointIds(1, joints, weights);
        sklMesh.setMaskBit(2);
        model.addMesh(sklMesh);

        return model;
    }
}