/**
 *
 *    Copyright © 2021 Uncharted Software Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

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
  deleteModel: dispatch(moduleActions.deleteModel),
};

export const mutations = {
  setModels: commit(moduleMutations.setModels),
  setFilteredModels: commit(moduleMutations.setFilteredModels),
};
