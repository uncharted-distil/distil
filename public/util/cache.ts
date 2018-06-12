import Vue from 'vue';

export function strongCache(config: any) {
	if (typeof config === 'function') {
		config = { getter: config };
	}

	const vms = {};

	return (state, getters, rootState, rootGetters) => {
		const fn = config.getter(state, getters, rootState, rootGetters)

		if (typeof fn !== 'function') {
			return fn;
		}

		return (...args) => {
			const key = 'c:' + (config.cacheKey ? config.cacheKey(...args) : args.join('~'));

			if (!vms[key]) {
				vms[key] = new Vue({
					computed: {
						value () {
							return fn(...args);
						}
					}
				});
			}

			return vms[key].value;
		}
	}
}

export function strongCaches(getters: any) {
	Object.keys(getters).forEach(getter => {
		getters[getter] = strongCache(getters[getter]);
	})

	return getters;
}
