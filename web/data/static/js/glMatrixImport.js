import * as glm from './vendor/gl-matrix/index.js';

for (const mName in glm) {
    window[mName] = glm[mName];
}
