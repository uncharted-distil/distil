
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
	describe('#getDataset()', () => {
		it('should retrieve a dataset from datasets map by id', () => {
			const testData = createTestData(1);
			const state = {
				datasets: testData
			};
			expect(getters.getDataset(state)(testData[0].name)).to.deep.equal(testData[0]);
		});
	});

	describe('#getVariables()', () => {
		it('should retrieve a dataset\'s variables', () => {
			const testData = createTestData(1);
			const state = {
				datasets: testData
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
});
