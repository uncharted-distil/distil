/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

"use strict";

import Overlay from "lumo/src/layer/overlay/Overlay";

// Private Methods

const clipQuads = function (cell, quads) {
  const clipped = [];
  quads.forEach((quad, key) => {
    clipped.push({ points: quad, key });
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
   * @param {Renderer=} options.renderer - The layer renderer.
   * @param {number=} options.opacity - The layer opacity.
   * @param {number=} options.zIndex - The layer z-index.
   * @param {boolean=} options.hidden - Whether or not the overlay is visible.
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
   * @param {Boolean} refresh - to refresh buffers or not
   * @returns {quadOverlay} The overlay object, for chaining.
   */
  addQuad(id, points, drawMode, refresh = true) {
    this.quads.set(id, points);
    this.drawModeMap.set(id, drawMode);
    this.renderer.addDrawLayer(id);
    if (this.plot && refresh) {
      this.refresh();
    }
    return this;
  }
  /**
   * Get a set of points
   *
   * @param {string} id - The id the quad is under.
   *
   * @returns {Array} points - The quad points.
   */
  getQuad(id) {
    return this.quads.get(id);
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
