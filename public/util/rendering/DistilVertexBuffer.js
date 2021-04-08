import VertexBuffer from "lumo/src/webgl/vertex/VertexBuffer";

export default class DistilVertexBuffer extends VertexBuffer {
  constructor(gl, arg, pointers = {}, options = {}) {
    super(gl, arg, pointers, options);
    this.options = options;
  }
}
