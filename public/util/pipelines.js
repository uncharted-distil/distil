/**
 * Structures and functions to support conditional display of pipeline create parameters.
 * The displayName is intended to be a label that is visible
 */

import _ from 'lodash';

// Gets a task object based on a variable type.
export function getTask(varType) {
	const lowerType = _.toLower(varType);
	return _.get(variableType, [lowerType, 'task'], {});
}

// Gets the display names for the metrics from a given task.
export function getMetricDisplayNames(task) {
	return _.map(_.get(task, 'metrics', []), s => s.displayName);
}

// Gets the schema names for the output types for a given task.
export function getOutputSchemaNames(task) {
	return _.map(_.get(task, 'outputs', []), s => s.schemaName);
}

// Gets the schema name for a metric given its display name.
export function getMetricSchemaName(task, displayName) {
	return _.find(task.metrics, s => s.displayName === displayName).schemaName;
}

// metrics used in classification tasks
const classificationMetrics = {
	precision: {
		displayName: 'Precision',
		schemaName: 'precision'
	},
	recall: {
		displayName: 'Recall',
		schemaName: 'recall'
	},
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
		schemaName: 'roc-auc'
	}
};

// metrics used in regression tasks
const regressionMetrics = {
	logLoss: {
		displayName: 'Log Loss',
		schemaName: 'log_loss'
	},
	meanSquaredErr: {
		displayName: 'Mean Squared Err',
		schemaName: 'mean_squared_err'
	},
	meanAbsoluteErr: {
		displayName: 'Mean Abs Err',
		schemaName: 'mean_abs_err'
	},
	medianAbsoluteErr: {
		displayName: 'MedianAbsErr',
		schemaName: 'median_abs_err'
	},
	r2: {
		displayName: 'R2',
		schemaName: 'r2'
	}
};

// output types used in classification tasks
const classificationOutputs = {
	classLabel: {
		displayName: 'Label',
		schemaName: 'class_label'
	}
	// multilabel: {
	// 	displayName: 'Multi Label',
	// 	schemaName: 'multilabel'
	// }
};

// output types used in regression tasks
const regressionOutputs = {
	regressionValue: {
		displayName: 'Regression Value',
		schemaName: 'regression_value'
	}
	// probability: {
	// 	displayName: 'Probability',
	// 	schemaName: 'probability'
	// },
	// generalScore: {
	// 	displayName: 'General Score',
	// 	schemaName: 'general_score'
	// }
};

// classification task info
const classification = {
	displayName: 'Classification',
	schemaName: 'classification',
	metrics: classificationMetrics,
	outputs: classificationOutputs
};

// regression task info
const regression = {
	displayName: 'Regression',
	schemaName: 'regession',
	metrics: regressionMetrics,
	outputs: regressionOutputs
};

// variable type to task mappings
const variableType = {
	float: {
		task: regression
	},
	integer: {
		task: regression
	},
	categorical: {
		task: classification
	},
	ordinal: {
		task: classification
	},
	boolean: {
		task: classification
	},
	text: {
		task: classification
	}
};
