#version 430

uniform mat4 umProjectView;
uniform mat4 umModel;
uniform bool uUseBones;
layout(std140, binding=0) uniform umBonesData {
    mat4 umBones[200];
};

in vec3 aPosition;
in vec3 aNormal;
in vec4 aColor;
in vec2 aUV;
in ivec4 aBoneIndices;
in vec4 aBoneWeights;

out vec2 vUV;
out vec4 vColor;

void main() {
    vec4 vertex = vec4(aPosition, 1.0);
    vec4 normal = vec4(aNormal, 0.0);

    if (uUseBones) {
        mat4 boneTransform = (
            (umBones[aBoneIndices.x] * aBoneWeights.x) +
            (umBones[aBoneIndices.y] * aBoneWeights.y)
        );
        vertex = boneTransform * vertex;
    }

    gl_Position = umProjectView * umModel * vertex;

    vUV = aUV;
    vColor = aColor;
}
