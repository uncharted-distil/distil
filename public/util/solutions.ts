/**
 * Structures and functions to support conditional display of solution create parameters.
 * The displayName is intended to be a label that is visible
 */

import _ from 'lodash';
import { Dictionary } from './dict';
import { SolutionState, Solution } from '../store/solutions/index';

export interface NameInfo {
	displayName: string,
	schemaName: string
}

export interface Task {
	displayName: string,
	schemaName: string,
	metrics: Dictionary<NameInfo>
};

// Utility function to return all solution results associated with a given request ID
export function getSolutionsByRequestIds(state: SolutionState, requestIds: string[]): Solution[] {
	const ids = {};
	requestIds.forEach(id => {
		ids[id] = true;
	});

	let solutions = [];
	const filtered = state.requests.filter(request => ids[request.requestId]);
	filtered.forEach(request => {
		solutions = solutions.concat(request.solutions);
	});
	return solutions;
}

// Returns a specific solution result given a request and its solution id.
export function getSolutionById(state: SolutionState, solutionId: string): Solution {
	if (!solutionId) {
		return null;
	}
	let found = null;
	state.requests.forEach(request => {
		request.solutions.forEach(solution => {
			if (solution.solutionId === solutionId) {
				found = solution;
			}
		});
	});
	return found;
}

// Gets a task object based on a variable type.
export function getTask(varType: string): Task {
	const lowerType = _.toLower(varType);
	return _.get(variableType, lowerType);
}

// Gets the display names for the metrics from a given task.
export function getMetricDisplayNames(task: Task): string[] {
	if (!task.metrics) {
		return [];
	}
	return _.map(task.metrics, s => s.displayName);
}

// Gets the schema name for a metric given its display name.
export function getMetricSchemaName(displayName: string): string {
	for(const m of metrics) {
		const result = _.find(m, s => s.displayName === displayName);
		if (!_.isEmpty(result)) {
			return result.schemaName;
		}
	}
	return undefined;
}

// Gets the display name for a metric given its schema name.
export function getMetricDisplayName(schemaName: string): string {
	const lowerName = _.toLower(schemaName);
	for(const m of metrics) {
		const result = _.find(m, s => s.schemaName === lowerName);
		if (!_.isEmpty(result)) {
			return result.displayName;
		}
	}
	return undefined;
}

// metrics used in classification tasks
const classificationMetrics: Dictionary<NameInfo> = {
	// Limit the metrics since not all are supported.
	accuracy: {
		displayName: 'Accuracy',
		schemaName: 'accuracy'
	}
	// Commented out because We are only using accuracy for classification at the moment.
	//
	// ,
	// f1: {
	// 	displayName: 'F1',
	// 	schemaName: 'f1',
	// },
	// f1Micro: {
	// 	displayName: 'F1 Micro',
	// 	schemaName: 'f1_micro'
	// },
	// f1Macro: {
	// 	displayName: 'F1 Macro',
	// 	schemaName: 'f1_macro'
	// },
	// rocAuc: {
	// 	displayName: 'ROC-AUC',
	// 	schemaName: 'roc_auc'
	// },
	// rocAucMicro: {
	// 	displayName: 'ROC-AUC Micro',
	// 	schemaName: 'roc_auc_micro'
	// },
	// rocAucMacro: {
	// 	displayName: 'ROC-AUC Macro',
	// 	schemaName: 'roc_auc_macro'
	// },
	// jaccardSimilarityScore: {
	// 	displayName: 'Jaccard Similarity',
	// 	schemaName: 'jaccard_similarity_score'
	// },
	// normalizedMutualInformation: {
	// 	displayName: 'Normalized Mutual Information',
	// 	schemaName: 'normalized_mutual_information'
	// }
};

// metrics used in regression tasks
const regressionMetrics: Dictionary<NameInfo> = {
	// Commented out because We are only using R2 for regression at the moment.
	//
	// rootMeanSquaredError: {
	// 	displayName: 'Root Mean Squared Error',
	// 	schemaName: 'root_mean_squared_error'
	// },
	meanSquaredError: {
		displayName: 'Mean Squared Error',
		schemaName: 'mean_squared_error'
	}
	// meanAbsoluteErr: {
	// 	displayName: 'Mean Absolute Error',
	// 	schemaName: 'mean_absolute_error'
	// },
	// rSquared: {
	// 	displayName: 'R Squared',
	// 	schemaName: 'r_squared'
	// }
};

const metrics = [classificationMetrics, regressionMetrics];

// classification task info
export const classification: Task = {
	displayName: 'Classification',
	schemaName: 'classification',
	metrics: classificationMetrics,
};

// regression task info
export const regression: Task = {
	displayName: 'Regression',
	schemaName: 'regression',
	metrics: regressionMetrics,
};

// variable type to task mappings
const variableType: Dictionary<Task> = {
	float:  regression,
	latitude:  regression,
	longitude:  regression,
	integer: regression,
	categorical: classification,
	ordinal: classification,
	address: classification,
	city: classification,
	state: classification,
	country: classification,
	email: classification,
	phone: classification,
	postal_code: classification,
	uri: classification,
	datetime: classification,
	text: classification,
	unknown: classification,
	boolean: classification
};
