attribute highp vec3 aVertexPos;
attribute lowp vec4 aVertexColor;
attribute mediump vec2 aVertexUV;
attribute mediump float aVertexJointID;
attribute mediump float aVertexJointID2; // Probably for normal

uniform highp mat4 umModelTransform;
uniform highp mat4 umProjection;
uniform highp mat4 umView;
uniform lowp vec4 uMaterialColor;
uniform lowp vec4 uLayerColor;
uniform mediump vec2 uLayerOffset;
uniform mediump mat4 umJoints[12];
uniform bool uUseJoints;
uniform bool uUseVertexColor;
uniform bool uUseModelTransform;
uniform bool uUseEnvmapSampler;

varying lowp vec4 vVertexColor;
varying mediump vec2 vVertexUV;
varying mediump vec2 vEnvmapUV;

void main(void) {
	vec4 pos = vec4(aVertexPos, 1.0);
	if (uUseJoints) {
		pos = umJoints[int(aVertexJointID)] * pos +  umJoints[int(aVertexJointID2)] * pos;
		pos *= 0.5;
		//pos = umJoints[int(aVertexJointID2)] * pos;
	}
	if (uUseModelTransform) {
		pos = vec4((umModelTransform * pos).xyz, 1.0);
	} else {
		pos = vec4(pos.xyz, 1.0);
	}

	if (uUseVertexColor) {
		vVertexColor = aVertexColor * (256.0 / 128.0);
	} else {
		vVertexColor = vec4(1.0);
	}
	vVertexColor *= uMaterialColor * uLayerColor;
	gl_Position = umProjection * umView * pos;	
	vVertexUV = aVertexUV + uLayerOffset;
	
	if (uUseEnvmapSampler) {
		//vec3 u = normalize(vec3(pos))
		//vEnvmapUV = 
	}
}
