// TODO:  Mocha doesn't play nicely with a Typescript using ES6 modules.
// Sounds like Jest can handle it, but figuring that out is low priority given that
// we've found the tests have somewhat limited utility.

// import * as pipelines from '../../util/pipelines';
// import {expect} from 'chai';

// describe('pipelines', () => {

// 	describe('#getTask()', () => {
// 		it('should return a task for a valid variable type', () => {
// 			expect(pipelines.getTask('float').displayName).to.equal('Regression');
// 			expect(pipelines.getTask('float').schemaName).to.equal('regression');
// 		});
// 	});

// 	describe('#getMetricDisplayNames()', () => {
// 		it('should return a list of associated metric display names for a task', () => {
// 			const task = pipelines.getTask('float');
// 			expect(pipelines.getMetricDisplayNames(task).length).to.not.equal(0);
// 			expect(pipelines.getMetricDisplayNames(task)).to.include('Mean Absolute Error');
// 		});
// 	});

// 	describe('#getOutputSchemaNames()', () => {
// 		it('should return a list of associated output schema names for a task', () => {
// 			const task = pipelines.getTask('float');
// 			expect(pipelines.getOutputSchemaNames(task).length).to.not.equal(0);
// 			expect(pipelines.getOutputSchemaNames(task)).to.include('real');
// 		});
// 	});

// 	describe('#getMetricSchemaName()', () => {
// 		it('should return a metric\'s schema name given its display name', () => {
// 			expect(pipelines.getMetricSchemaName('Accuracy')).to.equal('accuracy');
// 		});
// 	});

// 	describe('#getMetricDisplayName()', () => {
// 		it('should return a metric\'s display name given its schema name', () => {
// 			expect(pipelines.getMetricDisplayName('accuracy')).to.equal('Accuracy');
// 		});
// 	});
// });
