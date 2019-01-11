// TODO:  Mocha doesn't play nicely with a Typescript using ES6 modules.
// Sounds like Jest can handle it, but figuring that out is low priority given that
// we've found the tests have somewhat limited utility.

// import * as solutions from '../../util/solutions';
// import {expect} from 'chai';

// describe('solutions', () => {

// 	describe('#getTask()', () => {
// 		it('should return a task for a valid variable type', () => {
// 			expect(solutions.getTask(FLOAT_TYPE).displayName).to.equal('Regression');
// 			expect(solutions.getTask(FLOAT_TYPE).schemaName).to.equal('regression');
// 		});
// 	});

// 	describe('#getMetricDisplayNames()', () => {
// 		it('should return a list of associated metric display names for a task', () => {
// 			const task = solutions.getTask(FLOAT_TYPE);
// 			expect(solutions.getMetricDisplayNames(task).length).to.not.equal(0);
// 			expect(solutions.getMetricDisplayNames(task)).to.include('Mean Absolute Error');
// 		});
// 	});

// 	describe('#getOutputSchemaNames()', () => {
// 		it('should return a list of associated output schema names for a task', () => {
// 			const task = solutions.getTask(FLOAT_TYPE);
// 			expect(solutions.getOutputSchemaNames(task).length).to.not.equal(0);
// 			expect(solutions.getOutputSchemaNames(task)).to.include(REAL_TYPE);
// 		});
// 	});

// 	describe('#getMetricSchemaName()', () => {
// 		it('should return a metric\'s schema name given its display name', () => {
// 			expect(solutions.getMetricSchemaName('Accuracy')).to.equal('accuracy');
// 		});
// 	});

// 	describe('#getMetricDisplayName()', () => {
// 		it('should return a metric\'s display name given its schema name', () => {
// 			expect(solutions.getMetricDisplayName('accuracy')).to.equal('Accuracy');
// 		});
// 	});
// });
