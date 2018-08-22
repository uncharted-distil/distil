<template>
	<div>
		<div class="image-container" v-bind:class="{'selected': isSelected&&isLoaded}">
			<div class="image-elem" v-bind:class="{'clickable': hasClick}" ref="imageElem" @click.stop="handleClick" v-bind:style="{'max-width': `${width}px`}">
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

import $ from 'jquery';
import Vue from 'vue';
import { getters as routeGetters } from '../store/route/module';
import { circleSpinnerHTML } from '../util/spinner';
import { D3M_INDEX_FIELD } from '../store/dataset/index';
import { RowSelection } from '../store/highlights/index';
import { isRowSelected } from '../util/row';

export default Vue.extend({
	name: 'image-preview',

	props: {
		row: Object,
		imageUrl: String,
		width: {
			default: 64,
			type: Number
		},
		height: {
			default: 64,
			type: Number
		},
		onClick: Function
	},

	data() {
		return {
			zoomImage: false,
			entry: null,
			zoomedWidth: 400,
			zoomedHeight: 400
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
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		hasClick(): boolean {
			return !!this.onClick;
		},
		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},
		isSelected(): boolean {
			if (this.row) {
				return isRowSelected(this.rowSelection, this.row[D3M_INDEX_FIELD]);
			}
		}
	},

	mounted() {
		this.requestImage(this.imageUrl);
	},

	updated() {
		if (!this.image) {
			return;
		}
		const elem = this.$refs.imageElem as any;
		if (elem) {
			elem.innerHTML = '';
			elem.appendChild(this.clonedImageElement(this.width, this.height));
			const icon = document.createElement('i');
			icon.className += 'fa fa-search-plus zoom-icon';
			$(icon).click(event => {
				this.showZoomedImage();
				event.stopPropagation();
			});
			elem.appendChild(icon);
		}
	},

	methods: {

		handleClick() {
			if (this.onClick) {
				this.onClick({
					row: this.row,
					imageUrl: this.imageUrl,
					image: this.image
				});
			}
		},

		showZoomedImage() {
			if (this.image) {
				const $elem = this.$refs.imageElemZoom as any;
				$elem.innerHTML = '';
				$elem.appendChild(this.clonedImageElement(this.zoomedWidth, this.zoomedHeight));
			}
			this.zoomImage = true;
		},

		hideModal() {
			this.zoomImage = false;
		},

		clonedImageElement(width: number, height: number): HTMLImageElement {
			const img = this.image.cloneNode();
			$(img).css('max-width', `${width}px`);
			$(img).css('max-height', `${height}px`);
			return img as HTMLImageElement;
		},

		requestImage(url: string) {
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
				image.src = `distil/image/${this.dataset}/media/${url}`;
			});
		}
	}
});
</script>

<style>

.image-container {
	border: 2px solid rgba(0,0,0,0);
}
.image-container.selected {
	border: 2px solid #ff0067;
}

.image-elem {
	position: relative;
}
.image-elem:hover {
	background-color: #000;
}
.image-elem img {
	position: relative;
}
.image-elem.clickable {
	cursor: pointer;
}
.image-elem.clickable img:hover {
	opacity: 0.7;
}

.image-elem.clickable zoom-icon:hover {
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
	right: 0;
	top: 0;
	padding: 4px;
	color: #fff;
	visibility: hidden;
}
.image-elem:hover .zoom-icon {
	visibility: visible;
}

.zoom-icon {
	cursor: pointer;
	background-color: #424242;
	/*pointer-events: none;*/
}
</style>
