// TODO:  Mocha doesn't play nicely with a Typescript using ES6 modules.
// Sounds like Jest can handle it, but figuring that out is low priority given that
// we've found the tests have somewhat limited utility.

// TODO:  Mocha doesn't play nicely with a Typescript + Vue combo.
// Sounds like Jest can handle it, but figuring that out is low priority given that
// we've found the tests have somewhat limited utility.

// import * as mutations from '../../store/mutations';
// import { expect } from 'chai';

// function createTestData(numItems) {
// 	const testData = [];
// 	for (var i = 0; i < numItems; i++) {
// 		const vars = [{ name: `v${i}`, desc: `d${i}` }];
// 		testData.push({ name: `test${i}`, description: `test_description${i}`, variables: vars });
// 	}
// 	return testData;
// }

// describe('mutations', () => {

// 	describe('#setDatasets()', () => {
// 		it('should replace the datasets map with the caller supplied map', () => {
// 			const testData = createTestData(4);
// 			const state = {
// 				datasets: []
// 			};
// 			mutations.setDatasets(state, testData.slice(0, 2));
// 			mutations.setDatasets(state, testData.slice(2, 4));
// 			expect(state.datasets.length).to.equal(2);
// 			expect(state.datasets[0].name).to.equal('test2');
// 		});
// 	});

// 	describe('#setVariableSummaries()', () => {
// 		it('should replace the variable summaries with the caller supplied object', () => {
// 			const testData = { test: 'alpha' };
// 			const state = { variableSummaries: { orig: 'bravo' } };
// 			mutations.setVariableSummaries(state, testData);
// 			expect(state.variableSummaries).to.deep.equal(testData);
// 		});
// 	});

// 	describe('#updateVariableSummaries()', () => {
// 		it('should overwrite the variable summary at the index from the caller supplied object', () => {
// 			const testData = { index: 1, histogram: { name: 'alpha' } };
// 			const state = { variableSummaries: [{ name: 'bravo' }, { name: 'charlie' }] };
// 			mutations.updateVariableSummaries(state, testData);
// 			expect(state.variableSummaries[1]).to.deep.equal(testData.histogram);
// 			expect(state.variableSummaries[0]).to.deep.equal({ name: 'bravo' });
// 			expect(state.variableSummaries.length).equal(2);
// 		});
// 	});

// 	describe('#setFilteredData()', () => {
// 		it('should replace the filtered data with the caller supplied object', () => {
// 			const testData = {
// 				metadata: [
// 					{ name: 'alpha', type: 'integer' },
// 					{ name: 'bravo', type: 'text' }
// 				],
// 				values: [
// 					[0, 'a'],
// 					[1, 'b']
// 				]
// 			};
// 			const state = {
// 				filderedData: {}
// 			};
// 			mutations.setFilteredData(state, testData);
// 			expect(state.filteredData).to.deep.equal(testData);
// 		});
// 	});

// });
