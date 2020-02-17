uniform sampler2D uLayerDiffuseSampler;
uniform sampler2D uLayerEnvmapSampler;

uniform bool uUseLayerDiffuseSampler;
uniform bool uUseEnvmapSampler;
uniform lowp vec4 uMaterialColor;
uniform lowp vec4 uLayerColor;

varying lowp vec4 vVertexColor;
varying mediump vec2 vVertexUV;
varying mediump vec2 vEnvmapUV;

void main(void) {
	mediump vec4 clr = vec4(1.0, 1.0, 1.0, 1.0);
	if (uUseLayerDiffuseSampler) {
		clr = texture2D(uLayerDiffuseSampler, vVertexUV);
	}
	if (uUseEnvmapSampler) {
		//clr = vec4(clr.rgb, 1.0);
		//clr = vec4(clr.rgb*(1.0-clr.a) + texture2D(uLayerEnvmapSampler, vEnvmapUV).xyz*(clr.a), 1.0);
		//clr = vec4(1.0, 0.0, 1.0, 1.0);
	}
	// gl_FragColor = clr * vVertexColor * uMaterialColor * uLayerColor;
	gl_FragColor = vVertexColor * 0.8 + 0.2 * clr * uMaterialColor * uLayerColor;
}
