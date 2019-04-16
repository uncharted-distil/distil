import { Location } from 'vue-router';
import { ViewState } from './index';
import { Variable } from '../dataset/index';
import { Dictionary } from '../../util/dict';
import { sortVariablesByImportance, filterVariablesByPage, NUM_PER_PAGE } from '../../util/data';
import store from '../store';
import { getters as routeGetters } from '../route/module';
import { getters as solutionGetters } from '../solutions/module';

export const getters = {

	getJoinDatasetsPaginatedVariables(state: ViewState): Variable[] {
		const joinDatasetsVarsPage = routeGetters.getRouteJoinDatasetsVarsParge(store);
		const joinDatasetsVariables = routeGetters.getJoinDatasetsVariables(store);
		return filterVariablesByPage(joinDatasetsVarsPage, NUM_PER_PAGE, sortVariablesByImportance(joinDatasetsVariables));
	},

	getSelectTrainingPaginatedVariables(state: ViewState): Variable[] {
		const availableTrainingVarsPage = routeGetters.getRouteAvailableTrainingVarsPage(store);
		const trainingVarsPage = routeGetters.getRouteTrainingVarsPage(store);
		const availableVariables = routeGetters.getAvailableVariables(store);
		const trainingVariables = routeGetters.getTrainingVariables(store);
		const availableTrainingVars = filterVariablesByPage(availableTrainingVarsPage, NUM_PER_PAGE, sortVariablesByImportance(availableVariables));
		const trainingVars = filterVariablesByPage(trainingVarsPage, NUM_PER_PAGE, sortVariablesByImportance(trainingVariables));
		const targetVar = routeGetters.getTargetVariable(store);
		return availableTrainingVars.concat(trainingVars).concat(targetVar ? [ targetVar ] : []);
	},

	getResultsPaginatedVariables(state: ViewState): Variable[] {
		const trainingVars = solutionGetters.getActiveSolutionTrainingVariables(store);
		const resultTrainingVarsPage = routeGetters.getRouteResultTrainingVarsPage(store);
		return filterVariablesByPage(resultTrainingVarsPage, NUM_PER_PAGE, sortVariablesByImportance(trainingVars));
	}
};
