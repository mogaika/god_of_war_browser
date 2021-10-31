// Tracks references count. Used to clean GL resources
class Claimable {
    constructor() {
        this.refs = 0;
    }
    _free() {
        throw new Error(`Free method not implemented for ${this.constructor.name}`);
    }
    claim() { this.refs++; }
    unclaim() {
        if (--this.refs <= 0) {
            this._free();
        }
    }
}

// List of elements of Claimable type
class ClaimedPool {
    _pool = [];
    insert(claimable) {
        claimable.claim();
        if (this._pool.indexOf(claimable) >= 0) {
            throw Error("Element already in pool");
        }
        this._pool.push(claimable);
    }

    get(index) { return this._pool[index]; }
    get list() { return this._pool; }
    get length() { return this._pool.length; }

    remove(claimable) {
        let index = this._pool.indexOf(claimable);
        if (index < 0) {
            throw Error("This claimable not holded by this pool");
        }
        claimable.unclaim();
        this._pool.splice(index, 1);
    }
    removeAll() {
        for (let claimable of this._pool) {
            claimable.unclaim();
        }
        this._pool.length = 0;
    }
}


class ObjectTreeNodeBase extends Claimable {
    constructor(name) {
        super();
        this._localMatrix = mat4.create();
        this._globalMatrix = mat4.create();
        this._name = name;
        this._parent = null;
        this._active = true;
    }

    setLocalMatrix(matrix) {
        this._localMatrix = matrix;
        this.update();
    }

    applyTransform(matrix) {
        this._localMatrix = mat4.mul(this._localMatrix, this._localMatrix, matrix);
        this.update();
    }

    _update() {
        if (this._parent) {
            this._globalMatrix = mat4.mul(this._globalMatrix, this._parent._globalMatrix, this._localMatrix);
        } else {
            this._globalMatrix = mat4.copy(this._globalMatrix, this._localMatrix);
        }
    }

    update() {
        if (!this._active) {
            return;
        }
       this._update();
    }

    get localMatrix() { return this._localMatrix; }
    get globalMatrix() { return this._globalMatrix; }
    get isActive() { return this._active; }
    get parent() { return this._parent; }
    setActive(active = true) { this._active = active; }
}

class ObjectTreeNode extends ObjectTreeNodeBase {
    constructor(name) {
        super(name);
        this._nodes = new ClaimedPool();
    }

    addNode(node) {
        if (node === this) {
            throw Error("Trying to add node to itself");
        }
        if (node._parent) {
            throw Error("Node already has parent");
        }
        node._parent = this;
        this._nodes.insert(node);
        node.update();
    }

    _update() {
        super._update();
        for (const subNode of this._nodes.list) {
            subNode.update();
        }
    }

    get nodes() { return this._nodes.list; }

    _free() {
        this._nodes.removeAll();
    }
}

class ObjectTreeNodeModel extends ObjectTreeNodeBase {
    constructor(name, model) {
        super(name);
        model.claim();
        this._model = model;
    }

    get model() { return this._model; }

    _free() {
        this._model.unclaim();
    }
}


// has inverse matrices for model rendering
// also calculates render matrix based on inverse one
class ObjectTreeNodeSkinJoint extends ObjectTreeNode {
    constructor(name) {
        super(name);
        this._bindToJointMatrix = undefined;
        this._renderMatrix = mat4.create();
    }
    setLocalMatrixWithoutUpdate(matrix) { this._localMatrix = matrix; }
    setBindToJointMatrix(matrix) { this._bindToJointMatrix = matrix; }

    _update() {
        super._update();
        if (this._bindToJointMatrix) {
            this._renderMatrix = mat4.mul(this._renderMatrix, this._globalMatrix, this._bindToJointMatrix);
        } else {
            this._renderMatrix = mat4.copy(this._renderMatrix, this._globalMatrix);
        }
    }

    get bindToJointMatrix() { return this._bindToJointMatrix; }
    get renderMatrix() { return this._renderMatrix; }
}

class ObjectTreeNodeSkinned extends ObjectTreeNode {
    constructor(name) {
        super(name);
        this._jonits = new ClaimedPool();
        this._animation = undefined;
    }

    addJoint(jointNode) { this._jonits.insert(jointNode); }
    get joints() { return this._jonits.list; }
    
    assignAnimation(animation) {
        if (this._animation) {
            this._animation.unclaim();
        }
        animation.claim();
        this._animation = animation;
    }

    _free() {
        if (this._animation) {
            this._animation.unclaim();
        }
        this._jonits.removeAll();
    }
}
