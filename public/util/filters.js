import _ from 'lodash';
import * as index from '../store/index';

export function decodeFilter(filter) {
	if (!filter) {
		return null;
	}
	const nameValue = filter.split('=');
	const name = nameValue[0];
	if (_.isEmpty(name)) {
		return null;
	}
	if (nameValue.length > 1) {
		const value = nameValue[1];
		const values = value.split(',');
		if (values.length > 2 ) {
			const enabled = values[0] === '1';
			const type = values[1];
			if (type === index.NUMERICAL_SUMMARY_TYPE) {
				return {
					name: name,
					type: type,
					enabled: enabled,
					min: values[2],
					max: values[3]
				};
			} else if (type === index.CATEGORICAL_SUMMARY_TYPE) {
				return {
					name: name,
					type: type,
					enabled: enabled,
					categories: values.slice(2)
				};
			}
		}
	}
	return {
		name: name,
		enabled: false
	};
}

export function decodeFilters(filters) {
	const results = {};
	filters.forEach(filter => {
		const decoded = decodeFilter(filter);
		if (decoded) {
			results[decoded.name] = decoded;
		}
	});
	return results;
}

export function encodeFilter(filter) {
	if (!filter) {
		return null;
	}
	const enabled = filter.enabled ? '1' : '0';
	// numeric types have type,min,max or no additonal args if the value is unfiltered
	if (filter.type === index.NUMERICAL_SUMMARY_TYPE) {
		if (_.has(filter, 'min') && _.has(filter, 'max')) {
			return `${enabled},${encodeURIComponent(filter.type)},${filter.min},${filter.max}`;
		}
	// categorical type shave type,cat1,cat2...catN or no additional args if the value is unfiltered
	} else if (filter.type === index.CATEGORICAL_SUMMARY_TYPE) {
		if (!_.isEmpty(filter.categories)) {
			return `${enabled},${encodeURIComponent(filter.type)},${filter.categories.join(',')}`;
		}
	}
	return `${enabled}`;
}

export function encodeFilters(filters) {
	const results = {};
	_.forEach(filters, (filter, name) => {
		const encoded = encodeFilter(filter);
		if (encoded !== null) {
			results[name] = encoded;
		}
	});
	return results;
}

export function encodeQueryParam(filter) {
	if (!filter.enabled) {
		return `${filter.name}`;
	}
	if (_.has(filter, 'min') && _.has(filter, 'max')) {
		return `${filter.name}=${encodeURIComponent(filter.type)},${filter.min},${filter.max}`;
	} else if (!_.isEmpty(filter.categories)) {
		return `${filter.name}=${encodeURIComponent(filter.type)},${filter.categories.join(',')}`;
	}
	return null;
}

export function encodeQueryParams(filters) {
	const params = [];
	_.forEach(filters, filter => {
		const param = encodeQueryParam(filter);
		if (param !== null) {
			params.push(param);
		}
	});
	return params.length > 0 ? `?${params.join('&')}` : '';
}

export function isEmpty(filter) {
	return filter.enabled &&
		!_.has(filter, 'min') &&
		!_.has(filter, 'max') &&
		!_.has(filter, 'categories');
}
