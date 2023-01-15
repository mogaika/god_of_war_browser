class Plane {
    normal = vec3.create();
    constant = 0.0;

    constructor(normal, constant) {
        // normalize
        const len = vec3.length(normal);
        this.normal = vec3.scale(this.normal, normal, 1 / vec3.length(normal));
        this.constant = constant / len;
    }

    coplanarPoint(target = vec3.create()) {
        const result = vec3.scale(target, this.normal, -this.constant);
        return result;
    }

    distanceToPoint(point) {
        return vec3.dot(this.normal, point) + this.constant;
    }

    projectPoint(target, point) {
        return vec3.add(target, vec3.scale(target, this.normal, -this.distanceToPoint(point)), point);
    }

    isPointInside(point) {
        const vec = vec3.scale(vec3.create(), this.normal, -this.distanceToPoint(point))
        const dot = vec3.dot(this.normal, vec);
        return dot > -glMatrix.EPSILON;
    }

    static intersect3planes(p1, p2, p3) {
        const n1 = p1.normal,
            n2 = p2.normal,
            n3 = p3.normal;

        const m = mat3.fromValues(n1[0], n1[1], n1[2], n2[0], n2[1], n2[2], n3[0], n3[1], n3[2]);
        const det = mat3.determinant(m);

        if (glMatrix.equals(Math.abs(det), 0)) {
            return;
        }

        const x1 = p1.coplanarPoint();
        const x2 = p2.coplanarPoint();
        const x3 = p3.coplanarPoint();

        const f1 = vec3.scale(vec3.create(), vec3.cross(vec3.create(), n2, n3), vec3.dot(x1, n1));
        const f2 = vec3.scale(vec3.create(), vec3.cross(vec3.create(), n3, n1), vec3.dot(x2, n2));
        const f3 = vec3.scale(vec3.create(), vec3.cross(vec3.create(), n1, n2), vec3.dot(x3, n3));

        const vectorSum = vec3.add(vec3.create(), vec3.add(vec3.create(), f1, f2), f3);
        const intersection = vec3.scale(vec3.create(), vectorSum, (1 / det));

        return intersection;
    }

    // Calculates edjes of list of intersected covex planes
    static planesIntersectionsEdjes(planes) {
        let vectors = [];
        let indices = [];
        let planesIndices = [];
        let planesVectors = [];
        for (let i in planes) {
            planesIndices.push(i);
            planesVectors.push([]);
        }

        for (const combination of k_combinations(planesIndices, 3)) {
            const [i1, i2, i3] = combination;
            const [p1, p2, p3] = [planes[i1], planes[i2], planes[i3]];

            // find intersection of 3 planes
            const intersection = this.intersect3planes(p1, p2, p3);
            if (!intersection) {
                continue;
            }

            // check that this intersection inside hull
            let inside = true;
            for (const iPlane in planes) {
                if (combination.indexOf(iPlane) >= 0) {
                    // skip planes we intersecting with
                    continue;
                }
                if (!planes[iPlane].isPointInside(intersection)) {
                    inside = false;
                    break;
                }
            }
            if (!inside) {
                continue;
            }

            const iVector = vectors.length;
            vectors.push(intersection);
            for (const iPlane of combination) {
                planesVectors[iPlane].push(iVector);
            }
        }

        // for each plane generate indexes for lines of outline
        for (const iPlane in planes) {
            if (iPlane != 13) {
                // continue;
            }
            // sort points on this plane depending on angle to first point
            const vectorIndices = planesVectors[iPlane];
            if (vectorIndices.length == 0) {
                continue;
            }
            if (vectorIndices.length == 1) {
                console.error("Only one point on plane", planes, vectors, planesVectors, iPlane);
            }
            const plane = planes[iPlane];

            let center = vec3.create();
            for (const vi of vectorIndices) {
                center = vec3.add(center, center, vectors[vi]);
            }
            center = vec3.scale(center, center, 1 / vectorIndices.length);

            let rotations = [];

            let compareAxis1 = vec3.cross(vec3.create(), center, vectors[vectorIndices[0]]);
            compareAxis1 = vec3.normalize(compareAxis1, compareAxis1);
            let compareAxis2 = vec3.cross(vec3.create(), plane.normal, compareAxis1);
            compareAxis2 = vec3.normalize(compareAxis2, compareAxis2);
            for (let i = 0; i < vectorIndices.length; i++) {
                const vec = vectors[vectorIndices[i]];
                let diff = vec3.sub(vec3.create(), center, vec);
                diff = vec3.normalize(diff, diff);

                const dot1 = vec3.dot(diff, compareAxis1) / (vec3.length(diff) * vec3.length(compareAxis1));
                const dot2 = vec3.dot(diff, compareAxis2) / (vec3.length(diff) * vec3.length(compareAxis2));
                const angle = Math.atan2(dot1, dot2);
                rotations.push({
                    index: vectorIndices[i],
                    angle: angle,
                    tmp: i,
                });
            }
            rotations = rotations.sort(function(a, b) {
                return a.angle - b.angle;
            })

            for (let i = 0; i < rotations.length; i++) {
                indices.push(rotations[i].index, rotations[(i + 1) % rotations.length].index);
            }
        }

        return [vectors, indices];
    }
}