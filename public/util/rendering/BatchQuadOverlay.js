"use strict";

import Overlay from "lumo/src/layer/overlay/Overlay";

// Private Methods

const clipQuads = function (cell, quads) {
  const clipped = [];
  quads.forEach((quad, key) => {
    const result = new Array(quad.length);
    for (let i = 0; i < quad.length; i++) {
      const projected = cell.project(quad[i]);
      result[i] = {
        x: projected.x,
        y: projected.y,
        r: quad[i].r,
        g: quad[i].g,
        b: quad[i].b,
        a: quad[i].a,
        iR: quad[i].iR,
        iG: quad[i].iG,
        iB: quad[i].iB,
        iA: quad[i].iA,
      };
    }
    clipped.push({ points: result, key });
  });
  return clipped;
};

/**
 * Class representing a quad overlay.
 */
export default class BatchQuadOverlay extends Overlay {
  /**
   * Instantiates a new quadOverlay object.
   *
   * @param {object} options - The layer options.
   * @param {Renderer} options.renderer - The layer renderer.
   * @param {number} options.opacity - The layer opacity.
   * @param {number} options.zIndex - The layer z-index.
   */
  constructor(options = {}) {
    super(options);
    this.quads = new Map();
    this.drawModeMap = new Map();
  }

  /**
   * Add a set of points to render as a single quad.
   *
   * @param {string} id - The id to store the quad under.
   * @param {Array} points - The quad points.
   *
   * @returns {quadOverlay} The overlay object, for chaining.
   */
  addQuad(id, points, drawMode) {
    this.quads.set(id, points);
    this.drawModeMap.set(id, drawMode);
    if (this.plot) {
      this.refresh();
    }
    return this;
  }

  /**
   * Remove a quad by id from the overlay.
   *
   * @param {string} id - The id to store the quad under.
   *
   * @returns {quadOverlay} The overlay object, for chaining.
   */
  removeQuad(id) {
    this.quads.delete(id);
    this.drawModeMap.delete(id);
    if (this.plot) {
      this.refresh();
    }
    return this;
  }

  /**
   * Remove all quads from the layer.
   *
   * @returns {quadOverlay} The overlay object, for chaining.
   */
  clearQuads() {
    this.clear();
    this.quads = new Map();
    this.drawModeMap = new Map();
    if (this.plot) {
      this.refresh();
    }
    return this;
  }

  /**
   * Return the clipped geometry based on the current cell.
   *
   * @param {Cell} cell - The rendering cell.
   *
   * @returns {Array} The array of clipped geometry.
   */
  clipGeometry(cell) {
    return clipQuads(cell, this.quads);
  }
}
