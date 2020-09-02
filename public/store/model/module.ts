import { Module } from "vuex";
import { getters as moduleGetters } from "./getters";
import { actions as moduleActions } from "./actions";
import { mutations as moduleMutations } from "./mutations";
import { state, ModelState } from "./index";

import { DistilState } from "../store";
import { getStoreAccessors } from "vuex-typescript";

export const modelModule: Module<ModelState, DistilState> = {
  getters: moduleGetters,
  actions: moduleActions,
  mutations: moduleMutations,
  state: state,
};

const { commit, read, dispatch } = getStoreAccessors<ModelState, DistilState>(
  null
);

export const getters = {
  getFilteredModels: read(moduleGetters.getFilteredModels),
  getModels: read(moduleGetters.getModels),
  getCountOfModels: read(moduleGetters.getCountOfModels),
};

export const actions = {
  searchModels: dispatch(moduleActions.searchModels),
  fetchModels: dispatch(moduleActions.fetchModels),
};

export const mutations = {
  setModels: commit(moduleMutations.setModels),
  setFilteredModels: commit(moduleMutations.setFilteredModels),
};
