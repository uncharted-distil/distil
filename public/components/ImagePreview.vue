<template>
	<div>
		<div class="image-container">
			<div class="image-elem" ref="imageElem" @click.stop="onClick">
				<div v-if="isErrored">Error</div>
				<div v-if="!isErrored && !isLoaded" v-html="spinnerHTML"></div>
			</div>
		</div>
		<b-modal id="image-zoom-modal" :title="imageUrl"
			@hide="hideModal"
			:visible="!!zoomImage"
			hide-footer>
			<div class="image-elem-zoom" ref="imageElemZoom"></div>
		</b-modal>
	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import { circleSpinnerHTML } from '../util/spinner';

export default Vue.extend({
	name: 'image-preview',

	props: {
		imageUrl: String,
	},

	data() {
		return {
			zoomImage: false,
			entry: null
		};
	},

	computed: {
		isLoaded(): boolean {
			return this.entry && this.entry.image;
		},
		isErrored(): boolean {
			return this.entry && this.entry.err;
		},
		image(): HTMLImageElement {
			return this.entry ? this.entry.image : null;
		},
		spinnerHTML(): string {
			return circleSpinnerHTML();
		},
	},

	mounted() {
		this.requestImage(this.imageUrl);
	},

	updated() {
		if (!this.image) {
			return;
		}
		const $elem = this.$refs.imageElem as any;
		$elem.innerHTML = '';
		$elem.appendChild(this.image.cloneNode());
		const icon = document.createElement('i');
		icon.className += 'fa fa-plus zoom-icon';
		$elem.appendChild(icon);
	},

	methods: {
		onClick() {
			if (this.image) {
				const $elem = this.$refs.imageElemZoom as any;
				$elem.innerHTML = '';
				$elem.appendChild(this.image.cloneNode());
			}
			this.zoomImage = true;
		},

		hideModal() {
			this.zoomImage = false;
		},

		requestImage(url: string) {
			const IMAGES = [
				'a.jpeg',
				'b.jpeg',
				'c.jpeg'
			];
			return new Promise((resolve, reject) => {
				const image = new Image();
				image.onload = () => {
					this.entry = { url: url, image: image };
				};
				image.onerror = (event: any) => {
					const err = new Error(`Unable to load image from URL: \`${event.path[0].currentSrc}\``);
					this.entry = { url: url, err: err };
					reject(err);
				};
				image.crossOrigin = 'anonymous';
				image.src = `images/${IMAGES[Math.floor(Math.random() * IMAGES.length)]}`;
				//image.src = `images/${url}`;
			});
		}
	}
});
</script>

<style>

.image-elem {
	position: relative;
	max-width: 64px;
	border-radius: 4px;
}
.image-elem:hover {
	background-color: #000;
}
.image-elem img {
	position: relative;
	max-height: 64px;
	max-width: 64px;
	border-radius: 4px;
}
.image-elem img:hover {
	opacity: 0.7;
}
.image-elem-zoom {
	position: relative;
	text-align: center;
}
.image-elem-zoom img {
	position: relative;
	padding: 8px 16px;
	max-width: 100%;
	border-radius: 4px;
}
.image-elem .zoom-icon {
	position: absolute;
	right: 4px;
	top: 4px;
	color: #fff;
	visibility: hidden;
}
.image-elem:hover .zoom-icon {
	visibility: visible;
}

.zoom-icon {
	pointer-events: none;
}
</style>
