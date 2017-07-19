attribute highp vec3 aVertexPos;
attribute lowp vec4 aVertexColor;
attribute mediump vec2 aVertexUV;

uniform highp mat4 umModelTransform;
uniform highp mat4 umProjectionView;
uniform lowp vec4 uMaterialColor;

varying lowp vec4 vVertexColor;
varying mediump vec2 vVertexUV;
varying lowp vec4 vMaterialColor;

void main(void) {
	gl_Position = umProjectionView * umModelTransform * vec4(aVertexPos, 1.0);
	vVertexColor = aVertexColor * (256.0 / 128.0);
	vVertexUV = aVertexUV;
	vMaterialColor = uMaterialColor;
}
