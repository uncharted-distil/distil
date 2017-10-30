export interface Variable {
	name: string;
	type: string;
	suggestedTypes: string;
}

export interface Datasets {
	name: string;
	description: string;
	variables: Variable[];
}

export interface Extrema {
	min: number;
	max: number;
}

export interface Bucket {
	key: string;
	count: number;
}

export interface VariableSummary {
	name: string;
	feature: string;
	buckets: Bucket[];
	extrema: Extrema;
	type?: string;
	err?: string;
	pending?: string;
}

export interface Data {
	name: string;
	columns: string[];
	types: string[];
	values: any[][];
}
