export const LAST_STATE = '__LAST_STATE__';

export interface ViewState {
	fetchParamsCache: { [key: string]: string };
}

export const state: ViewState = {
	fetchParamsCache: {}
};
