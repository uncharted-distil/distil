<template>
	<div>
		<div class="image-container">
			<div class="image-elem" ref="imageElem" @click.stop="onClick">
				<div v-if="isErrored">Error</div>
				<div v-if="!isErrored && !isLoaded" v-html="spinnerHTML"></div>
			</div>
		</div>
		<b-modal id="zoom-modal" :title="imageUrl"
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
import { getters as dataGetters, actions as dataActions } from '../store/data/module';

export default Vue.extend({
	name: 'image-preview',

	props: {
		imageUrl: String,
	},

	data() {
		return {
			zoomImage: false
		};
	},

	computed: {
		isLoaded(): boolean {
			const arg = dataGetters.getImages(this.$store)[this.imageUrl];
			return arg && arg.image;
		},
		isErrored(): boolean {
			const arg = dataGetters.getImages(this.$store)[this.imageUrl];
			return arg && arg.err;
		},
		image(): HTMLImageElement {
			const arg = dataGetters.getImages(this.$store)[this.imageUrl];
			return arg ? arg.image : null;
		},
		spinnerHTML(): string {
			return circleSpinnerHTML();
		},
	},

	mounted() {
		dataActions.fetchImage(this.$store, { url: this.imageUrl });
	},

	watch: {
		image(currImage: HTMLImageElement) {
			const $elem = this.$refs.imageElem as any;
			$elem.innerHTML = '';
			if (currImage) {
				$elem.appendChild(currImage.cloneNode());
				const icon = document.createElement('i');
				icon.className += 'fa fa-plus zoom-icon';
				$elem.appendChild(icon);
			}
		}
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
