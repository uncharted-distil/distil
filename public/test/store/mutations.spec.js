
import * as mutations from '../../store/mutations';
import { expect } from 'chai';
import * as index from '../../store/index';

function createTestData(numItems) {
	const testData = [];
	for (var i = 0; i < numItems; i++) {
		const vars = [{ name: `v${i}`, desc: `d${i}` }];
		testData.push({ name: `test${i}`, description: `test_description${i}`, variables: vars });
	}
	return testData;
}

describe('mutations', () => {
	describe('#addDataset()', () => {
		it('should add a dataset to the datasets map', () => {
			const testData = createTestData(1);
			const state = {
				datasets: []
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
			mutations.setDatasets(state, testData.slice(2, 4));
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
			const state = { variableSummaries: { orig: 'bravo' } };
			mutations.setVariableSummaries(state, testData);
			expect(state.variableSummaries).to.deep.equal(testData);
		});
	});

	describe('#updateVariableSummaries()', () => {
		it('should overwrite the variable summary at the index from the caller supplied object', () => {
			const testData = { index: 1, histogram: { name: 'alpha' } };
			const state = { variableSummaries: [{ name: 'bravo' }, { name: 'charlie' }] };
			mutations.updateVariableSummaries(state, testData);
			expect(state.variableSummaries[1]).to.deep.equal(testData.histogram);
			expect(state.variableSummaries[0]).to.deep.equal({ name: 'bravo' });
			expect(state.variableSummaries.length).equal(2);
		});
	});

	describe('#setFilteredData()', () => {
		it('should replace the filtered data with the caller supplied object', () => {
			const testData = {
				metadata: [
					{ name: 'alpha', type: 'int' },
					{ name: 'bravo', type: 'text' }
				],
				values: [
					[0, 'a'],
					[1, 'b']
				]
			};
			const state = {
				filderedData: {}
			};
			mutations.setFilteredData(state, testData);
			expect(state.filteredData).to.deep.equal(testData);
		});
	});

	describe('#setVarEnabled()', () => {
		it('should set a variable to enabled in the filter state object', () => {
			const state = {
				filterState: {
					alpha: { enabled: true },
					bravo: { enabled: true }
				}
			};
			mutations.setVarEnabled(state, { name: 'alpha', enabled: false });
			expect(state.filterState.alpha.enabled).to.equal(false);
		});
	});

	describe('#setVarFilterRange()', () => {
		it('should set a variables range in filter state object', () => {
			const state = {
				filterState: {
					alpha: {
						min: 10,
						max: 20,
						type: (index.NUMERICAL_SUMMARY_TYPE)
					}
				}
			};
			mutations.setVarFilterRange(state, { name: 'alpha', min: 20, max: 30 });
			expect(state.filterState.alpha.min).to.equal(20);
			expect(state.filterState.alpha.max).to.equal(30);
		});
	});

	describe('#updateVarFilterState()', () => {
		const state = {
				filterState: {
					alpha: {
						min: 10,
						max: 20,
						enabled: true,
						type: (index.NUMERICAL_SUMMARY_TYPE)
					}
				}
			};

		it('should add/replace the complete filter state for the supplied variable for numerical types', () => {
			const testData = {
				name: 'alpha',
				filterState: {
					min: 11,
					max: 12,
					enabled: false,
					type: (index.NUMERICAL_SUMMARY_TYPE)
				}
			};
			mutations.updateVarFilterState(state, testData);
			expect(state.filterState.alpha).to.deep.equal({ min: 11, max: 12, enabled: false, type: (index.NUMERICAL_SUMMARY_TYPE) });
		});
		it('should add/replace the complete filter state for the supplied variable for categorical types', () => {
			const testData = {
				name: 'alpha',
				filterState: {
					categories: ['a', 'b'],
					enabled: true,
					type: (index.CATEGORICAL_SUMMARY_TYPE)
				}
			};
			mutations.updateVarFilterState(state, testData);
			expect(state.filterState.alpha).to.deep.equal({ categories: ['a', 'b'], enabled: true, type: (index.CATEGORICAL_SUMMARY_TYPE) });
		});
	});
});

describe('#setFilterState()', () => {
		it('should add/replace the complete filter state', () => {
			const state = {
				filterState: {
					alpha: {
						min: 10,
						max: 20,
						enabled: true,
						type: (index.NUMERICAL_SUMMARY_TYPE)
					}
				}
			};

			const testData = {				
				bravo: {
					categories: ['a', 'b'],
					enabled: false,
					type: (index.CATEGORICAL_SUMMARY_TYPE)
				}
			};

			mutations.setFilterState(state, testData);
			expect(state.filterState).to.not.have.property('alpha');
			expect(state.filterState.bravo).to.deep.equal({ categories: ['a', 'b'], enabled: false, type: (index.CATEGORICAL_SUMMARY_TYPE) });
		});
	});
