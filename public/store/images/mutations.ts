import _ from 'lodash';
import Vue from 'vue';
import { ImageState } from './index';

export const mutations = {

	setImage(state: ImageState, args: { url: string, image?: HTMLImageElement, err?: Error }) {
		if (args.image) {
			Vue.set(state.loadedImages, args.url, {
				url: args.url,
				image: args.image,
				err: null,
				timestamp: Date.now()
			});
		} else {
			Vue.set(state.loadedImages, args.url, {
				url: args.url,
				image: null,
				err: args.err,
				timestamp: Date.now()
			});
		}

		// LRU
		const MAX_IMAGES = 100;
		let entries = _.values(state.loadedImages);
		if (entries.length > MAX_IMAGES) {
			// take n latest
			entries = entries.sort((a: any, b: any) => {
				return b.timestamp - a.timestamp;
			}).slice(0, MAX_IMAGES);
			// remove all others
			state.loadedImages = {};
			entries.forEach((entry: any) => {
				Vue.set(state.loadedImages, entry.url, entry);
			});
		}

	}
}
