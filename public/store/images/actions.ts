import { mutations } from './module'
import { ImageState } from './index';
import { ActionContext } from 'vuex';
import { DistilState } from '../store';

export type ImagesContext = ActionContext<ImageState, DistilState>;

export const actions = {

	fetchImage(context: ImagesContext, args: { url: string }) {
		const IMAGES = [
			'a.jpeg',
			'b.jpeg',
			'c.jpeg'
		];
		return new Promise((resolve, reject) => {
			const image = new Image();
			image.onload = () => {
				mutations.setImage(context, { url: args.url, image: image });
				resolve(image);
			};
			image.onerror = (event: any) => {
				const err = new Error(`Unable to load image from URL: \`${event.path[0].currentSrc}\``);
				mutations.setImage(context, { url: args.url, err: err });
				reject(err);
			};
			image.crossOrigin = 'anonymous';
			image.src = `images/${IMAGES[Math.floor(Math.random() * IMAGES.length)]}`;
			//image.src = `images/${args.url}`;
		});

	}
}
