varying lowp vec4 vPosition;

void main(void) {
	gl_FragColor = vec4(vPosition.www, 1.0);
}
