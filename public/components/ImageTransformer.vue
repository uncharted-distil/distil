<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->

<template>
  <div
    class="d-flex justify-content-center flex-column position-relative"
    @mouseenter="setIsMouseOnCanvas(true)"
    @mouseleave="setIsMouseOnCanvas(false)"
  >
    <canvas
      id="canvas-image-transformer"
      ref="canvas"
      :class="{ selected: selected, border: !selected }"
      :width="size.width"
      :height="size.height"
      onscroll="onScroll"
      @mousedown="onMouseDown"
      @mousemove="onMouseMove"
      @mouseup="onMouseUp"
      @mouseout="setMouseDown(false)"
      @mouseover="setMouseDown(false)"
      @mousewheel="onScroll"
    />
    <transition name="fade">
      <div
        v-if="mouseOnCanvas"
        class="d-flex justify-content-center p-1 position-absolute"
        :style="refreshStyle"
      >
        <b-button @click="resetIdentity" title="Reset the Image Position">
          <i class="fa fa-refresh" aria-hidden="true" />
        </b-button>
      </div>
    </transition>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { EventList } from "../util/events";
export default Vue.extend({
  name: "image-transformer",
  props: {
    width: { type: Number as () => number, default: 300 },
    height: { type: Number as () => number, default: 300 },
    imgSrcs: { type: Array as () => HTMLImageElement[], default: [] },
    selected: { type: Boolean as () => boolean, default: false },
  },
  data() {
    return {
      mouseDown: false,
      start: { x: 0, y: 0 },
      isMouseOnCanvas: false,
      startMouseEvent: null as MouseEvent,
    };
  },
  watch: {
    imgSrcs() {
      this.initImages();
    },
  },
  computed: {
    ctx(): CanvasRenderingContext2D {
      return this.canvas.getContext("2d");
    },
    size(): { width: number; height: number } {
      return { width: this.width ?? 300, height: this.height ?? 300 };
    },
    canvas(): HTMLCanvasElement {
      return this.$refs.canvas as HTMLCanvasElement;
    },
    refreshStyle(): string {
      const bottomPadding = 45;
      return `top: ${this.size.height - bottomPadding}px; left: ${
        this.size.width / 2 - 16
      }px;z-index:1;`;
    },
    mouseOnCanvas(): boolean {
      return this.isMouseOnCanvas;
    },
  },
  mounted() {
    this.initImages();
  },
  methods: {
    resetIdentity() {
      this.ctx.setTransform(1, 0, 0, 1, 0, 0);
      this.draw();
    },
    setMouseDown(val: boolean) {
      this.mouseDown = val;
    },
    setIsMouseOnCanvas(val: boolean) {
      this.isMouseOnCanvas = val;
    },
    initImages() {
      if (!this.imgSrcs.length) {
        return;
      }
      const promises = [];
      this.imgSrcs.forEach((img, i) => {
        if (!img.complete) {
          promises.push(
            new Promise((res, rej) => {
              this.imgSrcs[i].onload = () => {
                res(true);
              };
              this.imgSrcs[i].onerror = () => {
                rej();
              };
            })
          );
          return;
        }
        this.draw();
      });
      // await until all images are loaded
      Promise.all(promises).then(() => {
        this.$nextTick(() => {
          this.draw();
        });
      });
    },
    draw() {
      // clears canvas
      this.ctx.save();
      this.ctx.setTransform(1, 0, 0, 1, 0, 0);
      this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
      this.ctx.restore();
      // draws any images
      this.imgSrcs.forEach((img) => {
        this.ctx.drawImage(img, 0, 0, this.width, this.height);
      });
    },
    onScroll(event: WheelEvent) {
      // on scroll transform mouse coordinates to canvas coordinates
      const p = this.getTransformedPoint(event.offsetX, event.offsetY);
      // scale for the view matrix
      const scale = event.deltaY > 0 ? 0.9 : 1.1;
      // translate to mouse canvas coordinates
      this.ctx.translate(p.x, p.y);
      // scale view matrix
      this.ctx.scale(scale, scale);
      // translate back
      this.ctx.translate(-p.x, -p.y);
      this.draw();
      return;
    },
    // converts screen coordinates (browser coordinates) to canvas coordinates
    getTransformedPoint(x: number, y: number) {
      // invert view matrix (opengl type stuff)
      const inverseTransform = this.ctx.getTransform().invertSelf();
      const transformedX =
        inverseTransform.a * x + inverseTransform.c * y + inverseTransform.e;
      const transformedY =
        inverseTransform.b * x + inverseTransform.d * y + inverseTransform.f;

      return { x: transformedX, y: transformedY };
    },
    onMouseMove(event: MouseEvent) {
      if (this.mouseDown) {
        const curPos = this.getTransformedPoint(event.offsetX, event.offsetY);
        this.ctx.translate(curPos.x - this.start.x, curPos.y - this.start.y);
        this.draw();
      }
    },
    onMouseDown(event: MouseEvent) {
      this.setMouseDown(true);
      this.start = this.getTransformedPoint(event.offsetX, event.offsetY);
      this.startMouseEvent = event;
    },
    onMouseUp(event: MouseEvent) {
      // row selection
      if (
        this.startMouseEvent.offsetX === event.offsetX &&
        this.startMouseEvent.offsetY === event.offsetY
      ) {
        this.$emit(EventList.TABLE.ROW_SELECTION_EVENT);
      }
      this.setMouseDown(false);
    },
  },
});
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s;
}
.fade-enter, .fade-leave-to /* .fade-leave-active below version 2.1.8 */ {
  opacity: 0;
}
.selected {
  border: 2px solid #ff0067;
}
</style>
