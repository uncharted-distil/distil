import { Dictionary } from '../../util/dict';

export interface LoadedImage {
	url: string;
	image: HTMLImageElement;
	err: Error;
	timestamp: Number;
}

export interface ImageState {
	loadedImages: Dictionary<LoadedImage>;
}

export const state: ImageState = {
	loadedImages: {}
}
