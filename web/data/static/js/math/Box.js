class Box {
    position = vec3.create();
    size = 0.0;

    constructor(position, size) {
        // normalize
        this.position = vec3.copy(this.position, position);
        this.size = size;
    }
}
