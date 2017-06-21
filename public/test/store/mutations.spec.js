
import * as mutations from '../../store/mutations';
import {expect} from 'chai';

function createTestData(numItems) {
	const testData = [];
	for (var i = 0; i < numItems; i++) {
		const vars = [{name: `v${i}`, desc: `d${i}`}];
		testData.push({name: `test${i}`, description: `test_description${i}`, variables: vars});
	}
	return testData;
}

describe('mutations', () => {
	describe('#addDataset()', () => {
		it('should add a dataset to the datasets map', () => {
			const testData = createTestData(1);
			const state = {
				datasets:[]
			};
			mutations.addDataset(state, testData);
			expect(state.datasets.length).to.equal(1);
			expect(state.datasets[0]).to.deep.equal(testData);
		});
	});

	describe('#setDatasets()', () => {
		it('should replace the datasets map with the caller supplied map', () => {
			const testData = createTestData(4);
			const state = {
				datasets: []
			};
			mutations.setDatasets(state, testData.slice(0, 2));
			mutations.setDatasets(state, testData.slice(2,4));
			expect(state.datasets.length).to.equal(2);
			expect(state.datasets[0].name).to.equal('test2');
		});
	});

	describe('#removeDataset()', () => {
		it('should remove a dataset from the datasets map', () => {
			const testData = createTestData(1);
			const state = {
				datasets: testData
			};
			const result = mutations.removeDataset(state, testData[0].name);
			expect(state.datasets.length).to.equal(0);
			expect(result).to.equal(true);
		});
	});

	describe('#setVariableSummaries()', () => {
		it('should replace the variable summaries with the caller supplied object', () => {
			const testData = { test: 'alpha' };
			const state = { variableSummaries: {orig: 'bravo'} };
			mutations.setVariableSummaries(state, testData);
			expect(state.variableSummaries).to.deep.equal(testData);
		});
	});

	describe('#setFilteredData()', () => {
		it('should replace the filtered data with the caller supplied object', () => {
			const testData = { 
			metadata:[
				{name: 'alpha', type: 'int'},
				{name: 'bravo', type: 'text'}
			],
			values: [
				[0, 'a'],
				[1, 'b']
			]
		};
		const state = {
			data: testData
		};
			mutations.setFilteredData(state, testData);
			expect(state.filteredData).to.deep.equal(testData);
		});
	});
});
