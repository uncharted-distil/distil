import { Store } from "vuex";
import { DistilState } from "./store";

let storeInstance: Store<DistilState> = null;

// Provides global access to Vuex store without requiring the presence of the component
export function store(): Store<DistilState> {
	if (storeInstance === null) {
		console.error('Tried to access uninitialized store instance');
	}
	return storeInstance;
}

// Done once to initialize
export function setStore(store: Store<DistilState>) {
	if (store !== null) {
		storeInstance = store;
	} else {
		console.error('Store can only be set once');
	}
}
