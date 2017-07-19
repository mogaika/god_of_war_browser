uniform sampler2D uMaterialDiffuseSampler;
uniform bool uUseMaterialDiffuseSampler;

varying lowp vec4 vVertexColor;
varying mediump vec2 vVertexUV;
varying lowp vec4 vMaterialColor;

void main(void) {
	if (uUseMaterialDiffuseSampler) {
		gl_FragColor = vVertexColor * texture2D(uMaterialDiffuseSampler, vVertexUV) * vMaterialColor;
	} else {
		gl_FragColor = vVertexColor * vMaterialColor;	
	}
}
