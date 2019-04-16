export const LAST_STATE = '__LAST_STATE__';

export interface ViewState {
	viewActiveDataset: string;
	viewSelectedTarget: string;
}

export const state: ViewState = {
	viewActiveDataset: '',
	viewSelectedTarget: '',
};
