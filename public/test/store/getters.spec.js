
import * as getters from '../../store/getters';
import {expect} from 'chai';

function createTestData(numItems) {
	const testData = [];
	for (var i = 0; i < numItems; i++) {
		const vars = [{name: `v${i}`, desc: `d${i}`}];
		testData.push({name: `test${i}`, description: `test_description${i}`, variables: vars});
	}
	return testData;
}

describe('getters', () => {

	describe('#getVariables()', () => {
		it('should retrieve a dataset\'s variables', () => {
			const testData = createTestData(1);
			const state = {
				variables: testData[0].variables
			};
			expect(getters.getVariables(state)(testData[0].name)).to.deep.equal(testData[0].variables);
		});
	});

	describe('#getDatasets()', () => {
		const testData = createTestData(2);
		const state = {
			datasets: testData
		};

		it('should retrieve a list of datasets from datasets map', () => {
			expect(getters.getDatasets(state)([testData[0].name, testData[1].name])).to.deep.equals(testData);
		});

		it('should retrieve all datasets from datasets map if id list is not supplied', () => {
			expect(getters.getDatasets(state)()).to.deep.equals(testData);
		});
	});

	describe('#getVariableSummaries()', () => {
		const testData = { test: 'alpha' };
		const state = {
			variableSummaries: testData
		};
		it('should retrieve the variable summaries object', () => {
			expect(getters.getVariableSummaries(state)()).to.deep.equals(testData);
		});

	});

	describe('#getFilteredData()', () => {
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
			filteredData: testData
		};
		it('should retrieve the data object', () => {
			expect(getters.getFilteredData(state)()).to.deep.equals(testData);
		});
	});
});
