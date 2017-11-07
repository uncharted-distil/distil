// TODO:  Mocha doesn't play nicely with a Typescript using ES6 modules.
// Sounds like Jest can handle it, but figuring that out is low priority given that
// we've found the tests have somewhat limited utility.

// import * as routes from '../../util/routes';
// import {expect} from 'chai';

// describe('route', () => {

// 	describe('#createRouteEntryFromRoute()', () => {
// 		it('should return a task for a valid variable type', () => {

// 			const route = {
// 				path: 'some path',
// 				query: {
// 					a: 'foo',
// 					b: {
// 						c: 'bar'
// 					}
// 				}
// 			};

// 			const expected = {
// 				path: 'some path',
// 				query: {
// 					a: 'fizz',
// 					b: {
// 						c: 'bar'
// 					},
// 					d: 'buzz'
// 				}
// 			};

// 			const args = {
// 				a: 'fizz',
// 				d: 'buzz'
// 			};

// 			const newRoute = routes.createRouteEntryFromRoute(route, args);
// 			expect(newRoute).to.deep.equal(expected);
// 		});
// 	});
// });
