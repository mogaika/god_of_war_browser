uniform sampler2D uLayerDiffuseSampler;
uniform sampler2D uLayerEnvmapSampler;

uniform bool uUseLayerDiffuseSampler;
uniform bool uUseEnvmapSampler;
uniform lowp vec4 uMaterialColor;

varying lowp vec4 vVertexColor;
varying mediump vec2 vVertexUV;
varying mediump vec2 vEnvmapUV;

void main(void) {
	if (uUseLayerDiffuseSampler) {
		mediump vec4 clr = vVertexColor * texture2D(uLayerDiffuseSampler, vVertexUV);
		if (uUseEnvmapSampler) {
			clr = vec4(clr.rgb, 1.0)*(1.0-clr.a) + texture2D(uLayerEnvmapSampler, vEnvmapUV) * clr.a;
		}
		gl_FragColor = clr;
	} else {
		gl_FragColor = vVertexColor;
	}
}
