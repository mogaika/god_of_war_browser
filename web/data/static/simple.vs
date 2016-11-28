attribute highp vec3 aVertexPos;

uniform highp mat4 umModelTransform;
uniform highp mat4 umProjectionView;

varying lowp vec4 vPosition;

void main(void) {
    vec4 vPos = umProjectionView * umModelTransform * vec4(aVertexPos, 1.0);
	gl_Position = vPos;
	vPosition = vPos/256.0;
}
