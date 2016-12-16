attribute highp vec3 aVertexPos;
attribute lowp vec4 aVertexColor;
attribute mediump vec2 aVertexUV;
attribute lowp float aVertexJointID;

uniform highp mat4 umModelTransform;
uniform highp mat4 umProjectionView;
uniform lowp vec4 uMaterialColor;
uniform mediump mat4 umJoints[150];
uniform bool uUseJoints;

varying lowp vec4 vVertexColor;
varying mediump vec2 vVertexUV;
varying lowp vec4 vMaterialColor;

void main(void) {
	vec4 pos = vec4(aVertexPos, 1.0);
	if (uUseJoints) {
		pos = umJoints[int(aVertexJointID)] * pos;
	}
	pos = vec4((umModelTransform * pos).xyz, 1.0);
	//vVertexColor = vec4(vec3(aVertexJointID / 64.0), 1.0);
	vVertexColor = aVertexColor * (256.0 / 128.0);
	gl_Position = umProjectionView * pos;	
	vVertexUV = aVertexUV;
	vMaterialColor = uMaterialColor;
}
