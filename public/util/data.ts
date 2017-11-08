import _ from 'lodash';
import { Datasets } from '../store/data/index';

// filters datasets by id
export function filterDatasets(ids: string[], datasets: Datasets[]): Datasets[] {
	if (_.isUndefined(ids)) {
		return datasets;
	}
	const idSet = new Set(ids);
	return _.filter(datasets, d => idSet.has(d.name));
}

// fetches datasets from local storage
export function getRecentDatasets(): string[] {
	const datasets = window.localStorage.getItem('recent-datasets');
	return (datasets) ? datasets.split(',') : [];
}

// adds a recent dataset to local storage
export function addRecentDataset(dataset: string) {
	const datasetsStr = window.localStorage.getItem('recent-datasets');
	const datasets = (datasetsStr) ? datasetsStr.split(',') : [];
	datasets.unshift(dataset);
	window.localStorage.setItem('recent-datasets', datasets.join(','));
}
