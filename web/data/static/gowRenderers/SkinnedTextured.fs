uniform sampler2D uLayerDiffuseSampler;
uniform bool uUseLayerDiffuseSampler;
uniform lowp vec4 uMaterialColor;
uniform bool onlyOpaqueRender;

varying lowp vec4 vVertexColor;
varying mediump vec2 vVertexUV;

void main(void) {
	if (uUseLayerDiffuseSampler) {
		mediump vec4 clr = vVertexColor * texture2D(uLayerDiffuseSampler, vVertexUV);
		if (onlyOpaqueRender) {
			if (clr.a < 1.0) {
				clr.rb = vec2(1.0 - clr.a);
				clr *= 0.5;
			}
		}
		gl_FragColor = clr;
	} else {
		gl_FragColor = vVertexColor;
	}
}
