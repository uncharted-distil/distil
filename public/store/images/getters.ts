import { ImageState } from './index';
import { Dictionary } from '../../util/dict';

export const getters = {

	getImages(state: ImageState): Dictionary<any> {
		return state.loadedImages;
	}
}
