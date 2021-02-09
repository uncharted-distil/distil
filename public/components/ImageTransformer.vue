<template>
  <div class="d-flex justify-content-center flex-column">
    <canvas
      id="canvas-image-transformer"
      ref="canvas"
      class="border"
      :width="size.width"
      :height="size.height"
      onscroll="onScroll"
      @mousedown="onMouseDown"
      @mousemove="onMouseMove"
      @mouseup="setMouseDown(false)"
      @mouseout="setMouseDown(false)"
      @mouseover="setMouseDown(false)"
      @mousewheel="onScroll"
    />
    <div class="d-flex justify-content-center p-1">
      <b-button @click="resetIdentity" title="Reset the Image Position">
        <i class="fa fa-refresh" aria-hidden="true" />
      </b-button>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
export default Vue.extend({
  name: "image-transformer",
  props: {
    width: { type: Number as () => number, default: 300 },
    height: { type: Number as () => number, default: 300 },
    imgSrcs: { type: Array as () => string[], default: [] },
  },
  data() {
    return {
      mouseDown: false,
      translation: { x: 0, y: 0 },
      start: { x: 0, y: 0 },
      scale: 1,
      imgs: [],
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
  },
  mounted() {
    this.initImages();
  },
  methods: {
    resetIdentity() {
      this.ctx.setTransform(1, 0, 0, 1, 0, 0);
      this.translation = { x: 0, y: 0 };
      this.scale = 1;
      this.draw();
    },
    setMouseDown(val: boolean) {
      this.mouseDown = val;
    },
    initImages() {
      if (!this.imgSrcs.length) {
        return;
      }
      this.imgs = [];
      this.imgSrcs.forEach((src) => {
        const image = new Image(this.width, this.height);
        image.src = src;
        this.imgs.push(image);
      });
      const promises = [];
      this.imgs.forEach((img, i) => {
        if (!img.complete) {
          promises.push(
            new Promise((res, rej) => {
              this.imgs[i].onload = () => {
                res(true);
              };
              this.imgs[i].onerror = () => {
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
      this.ctx.save();
      this.ctx.setTransform(1, 0, 0, 1, 0, 0);
      this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
      this.ctx.restore();
      this.ctx.save();
      this.ctx.translate(
        this.translation.x / this.scale,
        this.translation.y / this.scale
      );
      this.imgs.forEach((img) => {
        this.ctx.drawImage(img, 0, 0, this.width, this.height);
      });
      this.ctx.restore();
    },
    onScroll(event: WheelEvent) {
      const x = this.width / 2;
      const y = this.height / 2;
      const scale = event.deltaY > 0 ? 1.1 : 0.9;
      this.scale *= scale;
      this.ctx.translate(x, y);
      this.ctx.scale(scale, scale);
      this.ctx.translate(-x, -y);
      this.draw();
      return;
    },
    onPan(event) {
      console.log(event);
      return;
    },
    onMouseMove(event: MouseEvent) {
      if (this.mouseDown) {
        this.translation.x = event.clientX - this.start.x;
        this.translation.y = event.clientY - this.start.y;
        this.draw();
      }
    },
    onMouseDown(event: MouseEvent) {
      this.setMouseDown(true);
      this.start.x = event.clientX - this.translation.x;
      this.start.y = event.clientY - this.translation.y;
    },
  },
});
</script>
