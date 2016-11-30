attribute highp vec3 aVertexPos;
attribute lowp vec4 aVertexColor;

uniform highp mat4 umModelTransform;
uniform highp mat4 umProjectionView;

varying lowp vec4 vColor;

void main(void) {
	gl_Position = umProjectionView * umModelTransform * vec4(aVertexPos, 1.0);
	vColor = aVertexColor * (256.0 / 128.0);
}
