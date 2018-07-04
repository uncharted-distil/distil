import { Location } from 'vue-router';
import { ViewState } from './index';
import { Variable } from '../dataset/index';
import { Dictionary } from '../../util/dict';
import { sortVariablesByImportance, filterVariablesByPage, NUM_PER_PAGE } from '../../util/data';

export const getters = {
	getPrevView(state: ViewState): Dictionary<Dictionary<Location>> {
		return state.stack;
	},

	getSelectTrainingPaginatedVariables(state: ViewState, getters: any): Variable[] {
		const availableVarsPage = getters.getRouteAvailableVarsPage;
		const trainingVarsPage = getters.getRouteTrainingVarsPage;
		const availableVariables = getters.getAvailableVariables;
		const trainingVariables = getters.getTrainingVariables;
		const availableVars = filterVariablesByPage(availableVarsPage, NUM_PER_PAGE, sortVariablesByImportance(availableVariables));
		const trainingVars = filterVariablesByPage(trainingVarsPage, NUM_PER_PAGE, sortVariablesByImportance(trainingVariables));
		const targetVar = getters.getTargetVariable;
		return availableVars.concat(trainingVars).concat([ targetVar ]);
	},

	getResultsPaginatedVariables(state: ViewState, getters: any): Variable[] {
		const trainingVars = getters.getActiveSolutionTrainingVariables;
		const resultTrainingVarsPage = getters.getRouteResultTrainingVarsPage;
		return filterVariablesByPage(resultTrainingVarsPage, NUM_PER_PAGE, sortVariablesByImportance(trainingVars));
	}
}
