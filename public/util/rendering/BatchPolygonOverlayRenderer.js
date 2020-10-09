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
		uniform vec2 uViewOffset;
		uniform float uScale;
		uniform mat4 uProjectionMatrix;
		void main() {
			vec2 wPosition = (aPosition * uScale) - uViewOffset;
			gl_Position = uProjectionMatrix * vec4(wPosition, 0.0, 1.0);
		}
		`,
  frag: `
		precision highp float;
		uniform vec4 uPolygonColor;
		uniform float uOpacity;
		void main() {
			gl_FragColor = vec4(uPolygonColor.rgb, uPolygonColor.a * uOpacity);
		}
		`,
};

// Private Methods
const getVertexArray = function (points) {
  const vertices = new Float32Array(points.length * 2);
  for (let i = 0; i < points.length; i++) {
    vertices[i * 2] = points[i].x;
    vertices[i * 2 + 1] = points[i].y;
  }
  return vertices;
};

const createBuffers = function (overlay, points) {
  const vertices = getVertexArray(points);
  const vertexBuffer = new VertexBuffer(
    overlay.gl,
    vertices,
    {
      0: {
        size: 2,
        type: "FLOAT",
      },
    },
    {
      mode: "TRIANGLES",
      count: vertices.length / 2, // number of vertices to draw vertices has x,y therefore /2
    }
  );

  return {
    vertex: vertexBuffer,
  };
};

/**
 * Class representing a BatchPolygon Renderer.
 */
export default class BatchPolygonOverlayRenderer extends WebGLOverlayRenderer {
  /**
   * Instantiates a new PolygonOverlayRenderer object.
   *
   * @param {object} options - The overlay options.
   * @param {Array} options.polygonColor - The color of the line.
   */
  constructor(options = {}) {
    super(options);
    this.polygonColor = defaultTo(options.polygonColor, [1.0, 0.4, 0.1, 0.8]);
    this.shader = null;
    this.polygons = null;
  }

  /**
   * Executed when the overlay is attached to a plot.
   *
   * @param {Plot} plot - The plot to attach the overlay to.
   *
   * @returns {BatchPolygonOverlayRenderer} The overlay object, for chaining.
   */
  onAdd(plot) {
    super.onAdd(plot);
    this.shader = this.createShader(SHADER_GLSL);
    return this;
  }

  /**
   * Executed when the overlay is removed from a plot.
   *
   * @param {Plot} plot - The plot to remove the overlay from.
   *
   * @returns {BatchPolygonOverlayRenderer} The overlay object, for chaining.
   */
  onRemove(plot) {
    super.onRemove(plot);
    this.shader = null;
    return this;
  }

  /**
   * Generate any underlying buffers.
   *
   * @returns {BatchPolygonOverlayRenderer} The overlay object, for chaining.
   */
  refreshBuffers() {
    const clipped = this.overlay.getClippedGeometry();
    if (clipped) {
      this.polygons = clipped.map((points) => {
        // generate the buffer
        return createBuffers(this, points);
      });
    } else {
      this.polygons = null;
    }
  }

  /**
   * The draw function that is executed per frame.
   *
   * @returns {BatchPolygonOverlayRenderer} The overlay object, for chaining.
   */
  draw() {
    if (!this.polygons) {
      return this;
    }

    const gl = this.gl;
    const shader = this.shader;
    const polygons = this.polygons;
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
    shader.setUniform("uPolygonColor", this.polygonColor);
    shader.setUniform("uOpacity", opacity);

    // for each polyline buffer
    polygons.forEach((buffer) => {
      // draw the points
      buffer.vertex.bind();
      buffer.vertex.draw();
    });

    return this;
  }
}
