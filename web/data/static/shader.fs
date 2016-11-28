uniform sampler2D uSampler;
uniform bool bUseVertexUV;

varying lowp vec4 vColor;
varying mediump vec2 vTextureVU;

void main(void) {
    if (bUseVertexUV) {
        gl_FragColor = vColor * texture2D(uSampler, vTextureVU);
    } else {
        gl_FragColor = vColor;
    }
}