import { Location } from 'vue-router';
import { ViewState } from './index';
import { Variable } from '../dataset/index';
import { Dictionary } from '../../util/dict';
import { sortVariablesByImportance, filterVariablesByPage, NUM_PER_PAGE } from '../../util/data';
import { store } from '../storeProvider';
import { getters as routeGetters } from '../route/module';
import { getters as solutionGetters } from '../solutions/module';

export const getters = {
	getPrevView(state: ViewState): Dictionary<Dictionary<Location>> {
		return state.stack;
	},

	getSelectTrainingPaginatedVariables(state: ViewState): Variable[] {

		const availableVarsPage = routeGetters.getRouteAvailableVarsPage(store())
		const trainingVarsPage = routeGetters.getRouteTrainingVarsPage(store());
		const availableVariables = routeGetters.getAvailableVariables(store());
		const trainingVariables = routeGetters.getTrainingVariables(store());
		const availableVars = filterVariablesByPage(availableVarsPage, NUM_PER_PAGE, sortVariablesByImportance(availableVariables));
		const trainingVars = filterVariablesByPage(trainingVarsPage, NUM_PER_PAGE, sortVariablesByImportance(trainingVariables));
		const targetVar = routeGetters.getTargetVariable(store());
		return availableVars.concat(trainingVars).concat([ targetVar ]);
	},

	getResultsPaginatedVariables(state: ViewState): Variable[] {
		const trainingVars = solutionGetters.getActiveSolutionTrainingVariables(store());
		const resultTrainingVarsPage = routeGetters.getRouteResultTrainingVarsPage(store());
		return filterVariablesByPage(resultTrainingVarsPage, NUM_PER_PAGE, sortVariablesByImportance(trainingVars));
	}
}
