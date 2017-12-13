/**
 * Structures and functions to support conditional display of pipeline create parameters.
 * The displayName is intended to be a label that is visible
 */

import _ from 'lodash';
import { PipelineState, PipelineInfo } from '../store/pipelines/index';

const ERROR_VAL = 'ERRORED';

export interface NameInfo {
	displayName: string,
	schemaName: string
}

export interface Task {
	displayName: string,
	schemaName: string,
	metrics: { [name: string]: NameInfo },
	outputs: { [name: string]: NameInfo }
};

// Utility function to determine if a pipeline progress is in an errored state
export function pipelineIsErrored(progress: string): boolean {
	return progress == ERROR_VAL;
}

// Utility function to return all pipeline results that have not ERRORED
// associated with a given request ID
export function getPipelineResultsOkay(state: PipelineState, requestId: string): PipelineInfo[] {
	const pipelineResults = getPipelineResults(state, requestId);
	return _.filter(pipelineResults, p => !pipelineIsErrored(p.progress));
}

// Utility function to return all pipeline results associated with a given request ID
export function getPipelineResults(state: PipelineState, requestId: string): PipelineInfo[] {
	return _.concat(
		_.values(state.runningPipelines[requestId]),
		_.values(state.completedPipelines[requestId]));
}

// Returns a specific pipeline result given request and pipeline IDs.
export function getPipelineResult(state: PipelineState, requestId: string, pipelineId: string): PipelineInfo {
	const pipelineResults = getPipelineResultsOkay(state, requestId);
	return _.find(pipelineResults, p => p.pipelineId === pipelineId);
}

// Returns a specific pipeline result given a request and its ID.
export function getPipelineResultById(state: PipelineState, requestId: string, pipelineId: string): PipelineInfo {
	if (!pipelineId) {
		return null;
	}
	const pipelineResults = getPipelineResultsOkay(state, requestId);
	return _.find(pipelineResults, p => pipelineId === p.pipelineId);
}

// Gets a task object based on a variable type.
export function getTask(varType: string): Task {
	const lowerType = _.toLower(varType);
	return _.get(variableType, lowerType);
}

// Gets the display names for the metrics from a given task.
export function getMetricDisplayNames(task: Task): string[] {
	return _.map(_.get(task, 'metrics', []), (s: NameInfo) => s.displayName);
}

// Gets the schema names for the output types for a given task.
export function getOutputSchemaNames(task): string[] {
	return _.map(_.get(task, 'outputs', []), (s: NameInfo) => s.schemaName);
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

// Checks to see if a supplied schema output type is associated with a classificaiton task
export function isClassificationOutput(schemaOutput: string): boolean {
	return checkOutput(classificationOutputs, schemaOutput);
}

// Checks to see if a supplied schema output type is associated with a regression task
export function isRegressionOutput(schemaOutput: string): boolean {
	return checkOutput(regressionOutputs, schemaOutput);
}

function checkOutput(output: { [name: string]: NameInfo }, schemaOutput: string): boolean {
	const lowerName = _.toLower(schemaOutput);
	return !_.isEmpty(_.find(output, o => o.schemaName === lowerName));
}

// metrics used in classification tasks
const classificationMetrics: { [name: string]: NameInfo } = {
	// Limit the metrics since not all are supported.
	accuracy: {
		displayName: 'Accuracy',
		schemaName: 'accuracy'
	}/*,
	f1: {
		displayName: 'F1',
		schemaName: 'f1',
	},
	f1Micro: {
		displayName: 'F1 Micro',
		schemaName: 'f1_micro'
	},
	f1Macro: {
		displayName: 'F1 Macro',
		schemaName: 'f1_macro'
	},
	rocAuc: {
		displayName: 'ROC-AUC',
		schemaName: 'roc_auc'
	},
	rocAucMicro: {
		displayName: 'ROC-AUC Micro',
		schemaName: 'roc_auc_micro'
	},
	rocAucMacro: {
		displayName: 'ROC-AUC Macro',
		schemaName: 'roc_auc_macro'
	},
	jaccardSimilarityScore: {
		displayName: 'Jaccard Similarity',
		schemaName: 'jaccard_similarity_score'
	},
	normalizedMutualInformation: {
		displayName: 'Normalized Mutual Information',
		schemaName: 'normalized_mutual_information'
	}*/

};

// metrics used in regression tasks
const regressionMetrics: { [name: string]: NameInfo } = {
	rootMeanSquaredError: {
		displayName: 'Root Mean Squared Error',
		schemaName: 'root_mean_squared_error'
	},
	meanSquaredError: {
		displayName: 'Mean Squared Error',
		schemaName: 'mean_squared_error'
	},
	meanAbsoluteErr: {
		displayName: 'Mean Absolute Error',
		schemaName: 'mean_absolute_error'
	},
	rSquared: {
		displayName: 'R Squared',
		schemaName: 'r_squared'
	}
};

const metrics = [classificationMetrics, regressionMetrics];

// output types used in classification tasks
const classificationOutputs: { [name: string]: NameInfo } = {
	classLabel: {
		displayName: 'Label',
		schemaName: 'class_label'
	}
};

// output types used in regression tasks
const regressionOutputs: { [name: string]: NameInfo } = {
	regressionValue: {
		displayName: 'Real',
		schemaName: 'real'
	}
};

// classification task info
const classification: Task = {
	displayName: 'Classification',
	schemaName: 'classification',
	metrics: classificationMetrics,
	outputs: classificationOutputs
};

// regression task info
const regression: Task = {
	displayName: 'Regression',
	schemaName: 'regression',
	metrics: regressionMetrics,
	outputs: regressionOutputs
};

// variable type to task mappings
const variableType: { [varType: string]: Task } = {
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
	text: classification
};
