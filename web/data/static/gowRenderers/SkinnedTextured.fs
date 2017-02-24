uniform sampler2D uMaterialDiffuseSampler;
uniform bool uUseMaterialDiffuseSampler;

varying lowp vec4 vVertexColor;
varying mediump vec2 vVertexUV;
varying lowp vec4 vMaterialColor;

void main(void) {
	if (uUseMaterialDiffuseSampler) {
		mediump vec4 clr = vVertexColor * texture2D(uMaterialDiffuseSampler, vVertexUV) * vMaterialColor;
		/*
		if (clr.a < 1.0) {
			clr.rb = vec2(1.0 - clr.a);
			clr *= 0.5;
		}
		*/
		gl_FragColor = clr;
	} else {
		gl_FragColor = vVertexColor;
	}
}
