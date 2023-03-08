#version 430

uniform sampler2D uTexture;
uniform vec4 uColor;
uniform bool uUseTexture;

in vec2 vUV;
in vec4 vColor;

out vec4 oColor;

void main() {
    // material color
    vec4 color = uColor;

    // normalize vertex color from ps2 hdr to normal
    color *= vec4(vColor.xyz * (255.0 / 128.0), vColor.a);

    if (uUseTexture) {
        color *= texture(uTexture, vUV);
    }

    oColor = color;
}
