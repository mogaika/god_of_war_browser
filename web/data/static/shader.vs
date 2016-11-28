attribute highp vec3 aVertexPosition;
attribute highp vec4 aVertexColor;
attribute mediump vec2 aVertexUV;

uniform highp mat4 uModelMatrix;
uniform highp mat4 uProjectionViewMatrix;

uniform bool bUseVertexUV;
uniform bool bUseVertexColor;

varying lowp vec4 vColor;
varying mediump vec2 vTextureVU;

void main(void) {
    gl_Position = uProjectionViewMatrix * uModelMatrix * vec4(aVertexPosition, 1.0);
    if (bUseVertexColor) {
        vColor = aVertexColor * (256.0 / 128.0);
    } else {
        vColor = vec4(1.0);
    }
    if (bUseVertexUV) {
        vTextureVU = aVertexUV;
    }
}