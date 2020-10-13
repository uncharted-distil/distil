"use strict";

import defaultTo from "lodash/defaultTo";
import VertexBuffer from "lumo/src/webgl/vertex/VertexBuffer";
import WebGLOverlayRenderer from "lumo/src/renderer/overlay/WebGLOverlayRenderer";

// Constants

/**
 * Shader GLSL source.
 *
 * @private
 * @constant {object}
 */
const SHADER_GLSL = {
  vert: `
		precision highp float;
    attribute vec2 aPosition;
    attribute vec4 aColor;
		uniform vec2 uViewOffset;
		uniform float uScale;
    uniform mat4 uProjectionMatrix;
    varying vec4 oColor;
		void main() {
			vec2 wPosition = (aPosition * uScale) - uViewOffset;
      gl_Position = uProjectionMatrix * vec4(wPosition, 0.0, 1.0);
      oColor = aColor;
		}
		`,
  frag: `
    precision highp float;
    varying vec4 oColor;
	  uniform vec4 uquadColor;
		uniform float uOpacity;
		void main() {
			gl_FragColor = vec4(oColor.rgb, oColor.a * uOpacity);
		}
		`,
};

const PICKING_SHADER = {
  vert: `
  precision highp float;
  attribute vec2 aPosition;
  attribute vec4 aColor;
  attribute vec4 id;
  uniform vec2 uViewOffset;
  uniform float uScale;
  uniform mat4 uProjectionMatrix;
  varying vec4 oId;
  void main() {
    vec2 wPosition = (aPosition * uScale) - uViewOffset;
    gl_Position = uProjectionMatrix * vec4(wPosition, 0.0, 1.0);
    oId = id;
  }`,
  frag: `
    precision highp float;
    varying vec4 oId;
		void main() {
			gl_FragColor = oId;
		}
		`,
};

// Private Methods
const getVertexArray = function (points) {
  const numOfAttrs = 10;
  const vertices = new Float32Array(points.length * numOfAttrs);
  for (let i = 0; i < points.length; i++) {
    vertices[i * numOfAttrs] = points[i].x;
    vertices[i * numOfAttrs + 1] = points[i].y;
    vertices[i * numOfAttrs + 2] = points[i].r;
    vertices[i * numOfAttrs + 3] = points[i].g;
    vertices[i * numOfAttrs + 4] = points[i].b;
    vertices[i * numOfAttrs + 5] = points[i].a;
    vertices[i * numOfAttrs + 6] = points[i].iR;
    vertices[i * numOfAttrs + 7] = points[i].iG;
    vertices[i * numOfAttrs + 8] = points[i].iB;
    vertices[i * numOfAttrs + 9] = points[i].iA;
  }
  return vertices;
};

const createBuffers = function (overlay, points) {
  const vertices = getVertexArray(points);
  const floatByteSize = 4;
  const vertSize = 2; // x,y
  const colorSize = 4;
  const vertexBuffer = new VertexBuffer(
    overlay.gl,
    vertices,
    {
      0: {
        size: 2,
        type: "FLOAT",
        byteOffset: 0,
      },
      1: {
        size: 4,
        type: "FLOAT",
        byteOffset: vertSize * floatByteSize,
      },
      2: {
        size: 4,
        type: "FLOAT",
        byteOffset: (colorSize + vertSize) * floatByteSize,
      },
    },
    {
      mode: "TRIANGLES",
      count: vertices.length / 10, // number of vertices to draw vertices has x,y therefore /2
    }
  );

  return {
    vertex: vertexBuffer,
  };
};

/**
 * Class representing a Batchquad Renderer.
 */
export default class BatchQuadOverlayRenderer extends WebGLOverlayRenderer {
  /**
   * Instantiates a new quadOverlayRenderer object.
   *
   * @param {object} options - The overlay options.
   * @param {Array} options.quadColor - The color of the line.
   */
  constructor(options = {}) {
    super(options);
    this.quadColor = defaultTo(options.quadColor, [1.0, 0.4, 0.1, 0.8]);
    this.shader = null;
    this.quads = null;
    this.pickingShader = null;
    //framebuffer
    this.targetTexture = null;
    this.depthBuffer = null;
    this.fbo = null;
    this.fboDimensions = { width: 0, height: 0 };
    this.hoverCallbacks = [];
    this.leftClickCallbacks = [];
    this.rightClickCallbacks = [];
  }
  /**
   * Executed when the overlay is attached to a plot.
   *
   * @param {Plot} plot - The plot to attach the overlay to.
   *
   * @returns {BatchquadOverlayRenderer} The overlay object, for chaining.
   */
  onAdd(plot) {
    super.onAdd(plot);
    this.shader = this.createShader(SHADER_GLSL);
    this.pickingShader = this.createShader(PICKING_SHADER);
    this.gl.canvas.addEventListener("mouseup", (e) => {
      this.onClick(e);
    });
    this.buildFBO();
    return this;
  }

  /**
   * Executed when the overlay is removed from a plot.
   *
   * @param {Plot} plot - The plot to remove the overlay from.
   *
   * @returns {BatchquadOverlayRenderer} The overlay object, for chaining.
   */
  onRemove(plot) {
    super.onRemove(plot);
    this.shader = null;
    this.pickingShader = null;
    return this;
  }

  /**
   * Generate any underlying buffers.
   *
   * @returns {BatchquadOverlayRenderer} The overlay object, for chaining.
   */
  refreshBuffers() {
    const clipped = this.overlay.getClippedGeometry();
    if (clipped) {
      this.quads = clipped.map((points) => {
        // generate the buffer
        return createBuffers(this, points);
      });
    } else {
      this.quads = null;
    }
  }

  /**
   * The draw function that is executed per frame.
   *
   * @returns {BatchquadOverlayRenderer} The overlay object, for chaining.
   */
  draw() {
    if (!this.quads) {
      return this;
    }
    const gl = this.gl;
    const shader = this.shader;
    const quads = this.quads;
    const plot = this.overlay.plot;
    const cell = plot.cell;
    const proj = this.getOrthoMatrix();
    const scale = Math.pow(2, plot.zoom - cell.zoom);
    const opacity = this.overlay.opacity;

    // get view offset in cell space
    const offset = cell.project(plot.viewport, plot.zoom);

    // set blending func
    gl.enable(gl.BLEND);
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

    // bind shader
    shader.use();

    // set global uniforms
    shader.setUniform("uProjectionMatrix", proj);
    shader.setUniform("uViewOffset", [offset.x, offset.y]);
    shader.setUniform("uScale", scale);
    // shader.setUniform("uquadColor", this.quadColor);
    shader.setUniform("uOpacity", opacity);

    // for each quad buffer
    quads.forEach((buffer) => {
      // draw the points
      buffer.vertex.bind();
      buffer.vertex.draw();
    });
    if (this.canvasResize(gl.canvas)) {
      // the canvas was resized, make the framebuffer attachments match
      this.setFramebufferAttachmentSizes(gl.canvas.width, gl.canvas.height);
    }
    this.pickingShader.use();
    gl.bindFramebuffer(gl.FRAMEBUFFER, this.fbo);
    gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
    //gl.enable(gl.CULL_FACE);
    gl.disable(gl.BLEND);
    gl.enable(gl.DEPTH_TEST);

    // Clear the canvas AND the depth buffer.
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    //uniforms
    this.pickingShader.setUniform("uProjectionMatrix", proj);
    this.pickingShader.setUniform("uViewOffset", [offset.x, offset.y]);
    this.pickingShader.setUniform("uScale", scale);
    quads.forEach((buffer) => {
      buffer.vertex.bind();
      buffer.vertex.draw();
    });
    if (this.clicked) {
      const pixelX = (this.x * gl.canvas.width) / gl.canvas.clientWidth;
      const pixelY =
        gl.canvas.height -
        (this.y * gl.canvas.height) / gl.canvas.clientHeight -
        1;

      this.readPixels(pixelX, pixelY);
      this.clicked = false;
    }
    gl.bindFramebuffer(gl.FRAMEBUFFER, null);
    gl.disable(gl.DEPTH_TEST);
    return this;
  }
  canvasResize(canvas) {
    return (
      canvas.width !== this.fboDimensions.width ||
      canvas.height !== this.fboDimensions.height
    );
  }
  buildFBO() {
    const gl = this.gl;
    this.targetTexture = gl.createTexture();
    gl.bindTexture(gl.TEXTURE_2D, this.targetTexture);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    // create a depth renderbuffer
    this.depthBuffer = gl.createRenderbuffer();
    gl.bindRenderbuffer(gl.RENDERBUFFER, this.depthBuffer);

    // Create and bind the framebuffer
    this.fbo = gl.createFramebuffer();
    gl.bindFramebuffer(gl.FRAMEBUFFER, this.fbo);

    // attach the texture as the first color attachment
    const attachmentPoint = gl.COLOR_ATTACHMENT0;
    const level = 0;
    gl.framebufferTexture2D(
      gl.FRAMEBUFFER,
      attachmentPoint,
      gl.TEXTURE_2D,
      this.targetTexture,
      level
    );

    // make a depth buffer and the same size as the targetTexture
    gl.framebufferRenderbuffer(
      gl.FRAMEBUFFER,
      gl.DEPTH_ATTACHMENT,
      gl.RENDERBUFFER,
      this.depthBuffer
    );
    gl.bindFramebuffer(gl.FRAMEBUFFER, null);
  }
  readPixels(x, y) {
    const gl = this.gl;
    gl.flush();

    //gl.bindFramebuffer(gl.FRAMEBUFFER, this.fbo);
    const data = new Uint8Array(4);
    gl.readPixels(
      x, // x
      y, // y
      1, // width
      1, // height
      gl.RGBA, // format
      gl.UNSIGNED_BYTE, // type
      data
    );
    const id = this.RGBAToId(data);
    console.log(
      id,
      `background-color: rgb(${data[0]}, ${data[1]}, ${data[2]})`
    );
  }
  onClick(event) {
    this.clicked = true;
    this.x = event.layerX;
    this.y = event.layerY;
  }
  setFramebufferAttachmentSizes(width, height) {
    const gl = this.gl;
    gl.bindTexture(gl.TEXTURE_2D, this.targetTexture);
    // define size and format of level 0
    const level = 0;
    const internalFormat = gl.RGBA;
    const border = 0;
    const format = gl.RGBA;
    const type = gl.UNSIGNED_BYTE;
    const data = null;
    gl.texImage2D(
      gl.TEXTURE_2D,
      level,
      internalFormat,
      width,
      height,
      border,
      format,
      type,
      data
    );

    gl.bindRenderbuffer(gl.RENDERBUFFER, this.depthBuffer);
    gl.renderbufferStorage(
      gl.RENDERBUFFER,
      gl.DEPTH_COMPONENT16,
      width,
      height
    );
    this.fboDimensions.height = height;
    this.fboDimensions.width = width;
  }
  idToRGBA(id) {
    // 0 is reserved for background
    const ID = id + 1;
    return {
      iR: ((ID >> 0) & 0xff) / 0xff,
      iG: ((ID >> 8) & 0xff) / 0xff,
      iB: ((ID >> 16) & 0xff) / 0xff,
      iA: ((ID >> 24) & 0xff) / 0xff,
    };
  }
  RGBAToId(pixels) {
    return pixels[0] + (pixels[1] << 8) + (pixels[2] << 16) + (pixels[3] << 24);
  }
}
