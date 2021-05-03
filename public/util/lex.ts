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

import {
  DateTimeEntryState,
  LabelState,
  Lex,
  NumericEntryState,
  RelationState,
  StateTemplate,
  TextEntryState,
  TransitionFactory,
  ValueState,
  ValueStateValue,
} from "@uncharted.software/lex";
import { Highlight, Variable, ClusteredGrouping } from "../store/dataset";
import { Dictionary } from "./dict";
import {
  BIVARIATE_FILTER,
  CATEGORICAL_FILTER,
  DATETIME_FILTER,
  decodeFilters,
  EXCLUDE_FILTER,
  Filter,
  GEOBOUNDS_FILTER,
  GEOCOORDINATE_FILTER,
  INCLUDE_FILTER,
  NUMERICAL_FILTER,
  TEXT_FILTER,
} from "./filters";
import { createFiltersFromHighlights, decodeHighlights } from "./highlights";
import {
  CATEGORICAL_TYPE,
  dateToNum,
  DATE_TIME_LOWER_TYPE,
  GEOBOUNDS_TYPE,
  GEOCOORDINATE_TYPE,
  isCategoricalType,
  isNumericType,
  TIMESERIES_TYPE,
} from "./types";

const HIGHLIGHT = "highlight";

/*
  These are the custom relation options for our distil lex grammar that map our
  filter and highlight actions to lex bar style relation options. Should we
  ever want even more complex filter relations, we can extend these options.
*/
const distilRelationOptions = [
  [HIGHLIGHT, "☀", false],
  [EXCLUDE_FILTER, "≠", true],
].map((o) => new ValueStateValue(o[0], {}, { displayKey: o[1], hidden: o[2] }));

class DistilRelationState extends RelationState {
  static get HIGHLIGHT() {
    return distilRelationOptions[0];
  }
  static get EXCLUDE() {
    return distilRelationOptions[1];
  }
  constructor(config) {
    config.name = "Highlight";
    config.options = function () {
      return distilRelationOptions;
    };
    config.autoAdvanceDefault = true;
    config.defaultValue = distilRelationOptions[0];
    config.suggestionLimit = 1;
    super(config);
  }
}
export interface VariableInfo {
  // basic Variable
  variable: Variable;
  // number of times this variable exists (used for OR)
  count: number;
  // mode used for background color
  mode: string;
}
export interface TemplateInfo {
  // All of the variables that are present in the filters and highlights
  activeVariables: VariableInfo[];
  // highlightMap based on filter.key (its a collection of all the duplicate filters)
  highlightMap: Map<string, Filter[]>;
  // filterMap based on filter.key (its a collection of all the duplicate filters)
  filterMap: Map<string, Filter[]>;
}
/*
  This is the core function that actually generates a Lex Bar language. It takes
  a list of distil variables, converts them to an array of Lex Suggestions, then
  combines that with branching logic based on the suggestion's type to provide
  transitions to data entry states that fit that variable's type. As we add
  variable types with distinct entry needs, we can extend this function and the
  functions it depends on to support it in the Lex Bar language.
*/
export function variablesToLexLanguage(
  variables: VariableInfo[],
  allVariables: Variable[],
  variableMap: Map<string, Variable>
): Lex {
  // remove timeseries
  const filteredVariables = variables.filter((v) => {
    return v.variable.colType !== TIMESERIES_TYPE;
  });
  const filteredAllVariables = allVariables.filter((v) => {
    return v.colType !== TIMESERIES_TYPE;
  });
  const suggestions = variablesToLexSuggestions(filteredVariables, variableMap);
  // this generates the base templates used for the user typing into the lexbar
  const baseSuggestion = variablesToLexSuggestions(
    filteredAllVariables.map((v) => {
      return { variable: v, count: 1, mode: INCLUDE_FILTER };
    }),
    variableMap
  );

  const catVarLexSuggestions = perCategoricalVariableLexSuggestions(
    allVariables
  );
  const allSuggestions = [...suggestions, ...baseSuggestion];
  return Lex.from("field", ValueState, {
    name: "Choose a variable to search on",
    icon: '<i class="fa fa-filter" />',
    suggestions: allSuggestions,
  }).branch(
    ...distilTextEntryBuilder(allSuggestions),
    ...distilCategoryEntryBuilder(allSuggestions, catVarLexSuggestions),
    ...distilNumericalEntryBuilder(allSuggestions),
    ...distilDateTimeEntryBuilder(allSuggestions),
    ...distilGeoBoundsEntryBuilder(allSuggestions)
  );
}
export function distilTextEntryBuilder(
  suggestions: ValueStateValue[]
): StateTemplate[] {
  const textEntries = [] as ValueStateValue[];
  const textSuggestions = suggestions.filter((suggestion) => {
    return suggestion.meta.type === TEXT_FILTER;
  });
  const uniqueMetaCount = {} as Dictionary<boolean>;
  const uniqueSuggestion = textSuggestions.filter((v) => {
    if (!uniqueMetaCount[v.meta.count + v.meta.mode]) {
      uniqueMetaCount[v.meta.count + v.meta.mode] = true;
      return true;
    }
    return false;
  });
  uniqueSuggestion.forEach((suggestion) => {
    let branch = Lex.from("value_0", TextEntryState, {
      cssClasses: modeToColor(suggestion.meta.mode),
    });
    for (let i = 1; i < suggestion.meta.count; ++i) {
      branch = branch
        .to(LabelState, { label: "OR", vkey: "operator" })
        .to(`value_${i}`, TextEntryState);
    }
    textEntries.push(
      Lex.from("relation", DistilRelationState, {
        ...TransitionFactory.valueMetaCompare({
          type: TEXT_FILTER,
          count: suggestion.meta.count,
          mode: suggestion.meta.mode,
        }),
      }).branch(branch)
    );
  });
  return textEntries;
}
export function distilCategoryEntryBuilder(
  suggestions: ValueStateValue[],
  catVarLexSuggestions: Dictionary<unknown[]>
): StateTemplate[] {
  const categoryEntries = [] as ValueStateValue[];
  const categorySuggestions = suggestions.filter((suggestion) => {
    return suggestion.meta.type === CATEGORICAL_FILTER;
  });
  const uniqueMetaCount = {} as Dictionary<boolean>;
  const uniqueSuggestion = categorySuggestions.filter((v) => {
    if (!uniqueMetaCount[v.meta.count + v.meta.mode + v.meta.variable.key]) {
      uniqueMetaCount[v.meta.count + v.meta.mode + v.meta.variable.key] = true;
      return true;
    }
    return false;
  });
  uniqueSuggestion.forEach((suggestion) => {
    const labelSuggestions =
      catVarLexSuggestions[
        suggestion.meta.variable?.grouping?.clusterCol ??
          suggestion.meta.variable.key
      ] ?? [];
    let branch = Lex.from("value_0", ValueState, {
      allowUnknown: false,
      icon: "",
      name: "Type for suggestions",
      fetchSuggestions: (hint) => {
        return labelSuggestions.filter((cat) => {
          return cat["key"].toLowerCase().indexOf(hint.toLowerCase()) > -1;
        });
      },
      cssClasses: modeToColor(suggestion.meta.mode),
    });
    for (let i = 1; i < suggestion.meta.count; ++i) {
      branch = branch
        .to(LabelState, { label: "OR", vkey: "operator" })
        .to(`value_${i}`, ValueState, {
          allowUnknown: false,
          icon: "",
          name: "Type for suggestions",
          fetchSuggestions: (hint) => {
            return labelSuggestions.filter((cat) => {
              return cat["key"].toLowerCase().indexOf(hint.toLowerCase()) > -1;
            });
          },
        });
    }
    categoryEntries.push(
      Lex.from("relation", DistilRelationState, {
        ...TransitionFactory.valueMetaCompare({
          type: CATEGORICAL_TYPE,
          count: suggestion.meta.count,
          mode: suggestion.meta.mode,
          key: suggestion.meta.variable.key,
        }),
      }).branch(branch)
    );
  });
  return categoryEntries;
}
export function distilNumericalEntryBuilder(
  suggestions: ValueStateValue[]
): StateTemplate[] {
  // returns all the templates for numerical types
  const numericalEntries = [] as ValueStateValue[];
  // we use the supplied suggestions to build our templates therefore we need to find the numerical suggestions
  const numericalSuggestions = suggestions.filter((suggestion) => {
    return suggestion.meta.type === NUMERICAL_FILTER;
  });
  const uniqueMetaCount = {} as Dictionary<boolean>;
  const uniqueSuggestion = numericalSuggestions.filter((v) => {
    if (!uniqueMetaCount[v.meta.count + v.meta.mode]) {
      uniqueMetaCount[v.meta.count + v.meta.mode] = true;
      return true;
    }
    return false;
  });
  // loop through each suggestion
  uniqueSuggestion.forEach((suggestion) => {
    // build the base branch this is what the user will see if typing into the lexbar
    let branch = Lex.from(LabelState, {
      label: "From",
      vkey: "operator",
      cssClasses: modeToColor(suggestion.meta.mode),
    })
      .to("min_0", NumericEntryState, { name: "Enter lower bound" })
      .to(LabelState, { label: "To", vkey: "operator" })
      .to("max_0", NumericEntryState, { name: "Enter upper bound" });
    // adds the OR and the additional filter params if the count is > 0
    for (let i = 1; i < suggestion.meta.count; ++i) {
      branch = branch
        .to(LabelState, { label: "OR", vkey: "operator" })
        .to(LabelState, { label: "From", vkey: "operator" })
        .to(`min_${i}`, NumericEntryState, { name: "Enter lower bound" })
        .to(LabelState, { label: "To", vkey: "operator" })
        .to(`max_${i}`, NumericEntryState, { name: "Enter upper bound" });
    }
    // finished generating template
    numericalEntries.push(
      Lex.from("relation", DistilRelationState, {
        ...TransitionFactory.valueMetaCompare({
          type: NUMERICAL_FILTER,
          count: suggestion.meta.count,
          mode: suggestion.meta.mode,
        }),
      }).branch(branch)
    );
  });
  return numericalEntries;
}
export function distilGeoBoundsEntryBuilder(
  suggestions: ValueStateValue[]
): StateTemplate[] {
  const geoboundEntries = [] as ValueStateValue[];
  const geoboundsSuggestions = suggestions.filter((suggestion) => {
    return suggestion.meta.type === GEOBOUNDS_FILTER;
  });
  const uniqueMetaCount = {} as Dictionary<boolean>;
  const uniqueSuggestion = geoboundsSuggestions.filter((v) => {
    if (!uniqueMetaCount[v.meta.count + v.meta.mode]) {
      uniqueMetaCount[v.meta.count + v.meta.mode] = true;
      return true;
    }
    return false;
  });
  uniqueSuggestion.forEach((suggestion) => {
    let branch = Lex.from(LabelState, {
      label: "From Latitude",
      vkey: "operator",
      cssClasses: modeToColor(suggestion.meta.mode),
    })
      .to("minX_0", NumericEntryState, { name: "Enter lower bound" })
      .to(LabelState, { label: "To", vkey: "operator" })
      .to("maxX_0", NumericEntryState, { name: "Enter upper bound" })
      .to(LabelState, { label: "From Longitude" })
      .to("minY_0", NumericEntryState, { name: "Enter lower bound" })
      .to(LabelState, { label: "To", vkey: "operator" })
      .to("maxY_0", NumericEntryState, { name: "Enter upper bound" });
    for (let i = 1; i < suggestion.meta.count; ++i) {
      branch = branch
        .to(LabelState, { label: "OR", vkey: "operator" })
        .to(LabelState, { label: "From Latitude", vkey: "operator" })
        .to(`minX_${i}`, NumericEntryState, { name: "Enter lower bound" })
        .to(LabelState, { label: "To", vkey: "operator" })
        .to(`maxX_${i}`, NumericEntryState, { name: "Enter upper bound" })
        .to(LabelState, { label: "From Longitude" })
        .to(`minY_${i}`, NumericEntryState, { name: "Enter lower bound" })
        .to(LabelState, { label: "To", vkey: "operator" })
        .to(`maxY_${i}`, NumericEntryState, { name: "Enter upper bound" });
    }
    geoboundEntries.push(
      Lex.from("relation", DistilRelationState, {
        ...TransitionFactory.valueMetaCompare({
          type: GEOBOUNDS_FILTER,
          count: suggestion.meta.count,
          mode: suggestion.meta.mode,
        }),
      }).branch(branch)
    );
  });
  return geoboundEntries;
}
// distilDateTimeEntryBuilder creates an array of DateTimeEntry based on the supplied variables
// this allows us to specify min and max dates
export function distilDateTimeEntryBuilder(
  suggestions: ValueStateValue[]
): StateTemplate[] {
  const dateTimeEntries = [] as ValueStateValue[];
  const dateSuggestions = suggestions.filter((suggestion) => {
    return suggestion.meta.type === DATETIME_FILTER;
  });
  dateSuggestions.forEach((suggestion) => {
    let branch = Lex.from(LabelState, {
      label: "From",
      vkey: "operator",
      cssClasses: modeToColor(suggestion.meta.mode),
    })
      .to("min_0", DateTimeEntryState, {
        enableTime: true,
        enableCalendar: true,
        timezone: "Greenwich",
        hilightedDate: new Date(suggestion.meta.variable.min * 1000),
      })
      .to(LabelState, { label: "To", vkey: "operator" })
      .to("max_0", DateTimeEntryState, {
        enableTime: true,
        enableCalendar: true,
        timezone: "Greenwich",
        hilightedDate: new Date(suggestion.meta.variable.max * 1000),
      });
    for (let i = 1; i < suggestion.meta.count; ++i) {
      branch = branch
        .to(LabelState, { label: "OR", vkey: "operator" })
        .to(LabelState, { label: "From", vkey: "operator" })
        .to(`min_${i}`, DateTimeEntryState, {
          enableTime: true,
          enableCalendar: true,
          timezone: "Greenwich",
          hilightedDate: new Date(suggestion.meta.variable.min * 1000),
        })
        .to(LabelState, { label: "To", vkey: "operator" })
        .to(`max_${i}`, DateTimeEntryState, {
          enableTime: true,
          enableCalendar: true,
          timezone: "Greenwich",
          hilightedDate: new Date(suggestion.meta.variable.max * 1000),
        });
    }
    // default with
    dateTimeEntries.push(
      Lex.from("relation", DistilRelationState, {
        ...TransitionFactory.valueMetaCompare({
          type: DATETIME_FILTER,
          name: suggestion.meta.variable.colName,
          count: suggestion.meta.count,
          mode: suggestion.meta.mode,
        }),
      }).branch(branch)
    );
  });
  return dateTimeEntries;
}
// aggregates all the variables for highlight and filter into VariableInfo in over to generate templates
export function variableAggregation(
  filter: string,
  highlight: string,
  allVariables: Variable[]
): TemplateInfo {
  const decodedFilters = decodeFilters(filter).list.filter(
    (f) => f.type !== "row"
  );
  const decodedHighlights = createFiltersFromHighlights(
    decodeHighlights(highlight),
    HIGHLIGHT
  );

  const variableDict = buildVariableDictionary(allVariables);
  const filterVariables = new Map<string, Variable[]>();
  // check that the filter variables exist
  decodedFilters.forEach((f) => {
    if (variableDict[f.key]) {
      if (filterVariables.has(f.key)) {
        filterVariables.get(f.key).push(variableDict[f.key]);
        return;
      }
      filterVariables.set(f.key, [variableDict[f.key]]);
      if (isCategoricalType(f.type)) {
        filterVariables.get(f.key).length = f.categories.length;
      }
    }
  });
  const highlightVariables = new Map<string, Variable[]>();
  decodedHighlights.forEach((h) => {
    if (variableDict[h.key]) {
      if (highlightVariables.has(h.key)) {
        highlightVariables.get(h.key).push(variableDict[h.key]);
        return;
      }
      highlightVariables.set(h.key, [variableDict[h.key]]);
      if (isCategoricalType(h.type)) {
        highlightVariables.get(h.key).length = h.categories.length;
      }
    }
  });

  let activeVariables = [
    ...Array.from(highlightVariables.values()).map((hv) => {
      return { variable: hv[0], count: hv.length, mode: INCLUDE_FILTER };
    }),
    ...Array.from(filterVariables.values()).map((fv) => {
      return { variable: fv[0], count: fv.length, mode: EXCLUDE_FILTER };
    }),
  ] as VariableInfo[];
  // remove timeseries
  activeVariables = activeVariables.filter((v) => {
    return v.variable.colType !== TIMESERIES_TYPE;
  });
  const activeVariablesMap = new Map(
    activeVariables.map((v) => {
      return [v.variable.key, true];
    })
  );
  const highlightMap = new Map<string, Filter[]>();
  const filterMap = new Map<string, Filter[]>();
  decodedHighlights.forEach((el) => {
    if (activeVariablesMap.has(el.key)) {
      if (highlightMap.has(el.key)) {
        highlightMap.get(el.key).push(el);
        return;
      }
      highlightMap.set(el.key, [el]);
    }
  });
  decodedFilters.forEach((el) => {
    if (activeVariablesMap.has(el.key)) {
      if (filterMap.has(el.key)) {
        filterMap.get(el.key).push(el);
        return;
      }
      filterMap.set(el.key, [el]);
    }
  });
  return { activeVariables, highlightMap, filterMap };
}
export function filterParamsToLexQuery(
  templateInfo: TemplateInfo,
  variableMap: Map<string, Variable>
) {
  // remove highlight if variable does not exist
  const lexableElements = [
    ...templateInfo.highlightMap.values(),
    ...templateInfo.filterMap.values(),
  ];
  const suggestions = variablesToLexSuggestions(
    templateInfo.activeVariables,
    variableMap
  );
  const lexQuery = filtersToValueState(lexableElements, suggestions);
  return lexQuery;
}
export function filtersToValueState(
  filters: Filter[][],
  suggestions: unknown[]
) {
  return filters.map((f, i) => {
    const filterGroupType = f[0].type;
    const result = {
      field: suggestions[i],
      relation: modeToRelation(f[0].mode),
    };
    if (
      filterGroupType === GEOBOUNDS_FILTER ||
      filterGroupType === BIVARIATE_FILTER
    ) {
      for (let i = 0; i < f.length; ++i) {
        result[`minX_${i}`] = new ValueStateValue(f[i].minX);
        result[`maxX_${i}`] = new ValueStateValue(f[i].maxX);
        result[`minY_${i}`] = new ValueStateValue(f[i].minY);
        result[`maxY_${i}`] = new ValueStateValue(f[i].maxY);
      }
      return result;
    } else if (filterGroupType === DATETIME_FILTER) {
      for (let i = 0; i < f.length; ++i) {
        result[`min_${i}`] = new Date(f[i].min * 1000);
        result[`max_${i}`] = new Date(f[i].max * 1000);
      }
      return result;
    } else if (isNumericType(filterGroupType)) {
      for (let i = 0; i < f.length; ++i) {
        result[`min_${i}`] = new ValueStateValue(f[i].min);
        result[`max_${i}`] = new ValueStateValue(f[i].max);
      }
      return result;
    } else {
      for (let i = 0; i < f[0].categories.length; ++i) {
        result[`value_${i}`] = new ValueStateValue(f[0].categories[i], null, {
          displayKey: f[0].categories[i],
        });
      }
      return result;
    }
  });
}
/*
  This translates a lex query's relation and value states to generate a new
  highlight and filter state so that it can be used to update the route and so
  update the filter and highlight state of the application.
*/
export function lexQueryToFiltersAndHighlight(
  lexQuery: any[][],
  dataset: string
): { filters: Filter[]; highlights: Highlight[] } {
  const filters = [] as Filter[];
  const highlights = [] as Highlight[];

  lexQuery[0].forEach((lq) => {
    if (lq.relation.key !== HIGHLIGHT) {
      const key = lq.field.meta.variable.key;
      const displayKey = lq.field.displayKey;
      const type = lq.field.meta.type;
      for (let i = 0; i < lq.field.meta.count; ++i) {
        const filter: Filter = {
          mode: lq.relation.key,
          displayName: displayKey,
          type,
          key,
          set: "",
        };

        if (type === GEOBOUNDS_FILTER || type === GEOCOORDINATE_FILTER) {
          filter.key = filter.key;
          filter.minX = parseFloat(lq[`minX_${i}`].key);
          filter.maxX = parseFloat(lq[`maxX_${i}`].key);
          filter.minY = parseFloat(lq[`minY_${i}`].key);
          filter.maxY = parseFloat(lq[`maxY_${i}`].key);
        } else if (type === DATETIME_FILTER) {
          filter.min = dateToNum(lq[`min_${i}`]);
          filter.max = dateToNum(lq[`max_${i}`]);
        } else if (isNumericType(type)) {
          filter.min = parseFloat(lq[`min_${i}`].key);
          filter.max = parseFloat(lq[`max_${i}`].key);
        } else {
          const values: string[] = [];
          for (i = 0; i < lq.field.meta.count; ++i) {
            values.push(lq[`value_${i}`].key);
          }
          filter.categories = values;
        }

        filters.push(filter);
      }
    } else {
      const key = lq.field.meta.variable.key;
      const type = lq.field.meta.type;
      for (let i = 0; i < lq.field.meta.count; ++i) {
        const highlight = {
          dataset,
          context: "lex-bar",
          key,
          value: {},
        } as Highlight;

        if (
          type === GEOBOUNDS_FILTER ||
          type === GEOCOORDINATE_FILTER ||
          type === BIVARIATE_FILTER
        ) {
          highlight.key = highlight.key;
          highlight.value.minX = parseFloat(lq[`minX_${i}`].key);
          highlight.value.maxX = parseFloat(lq[`maxX_${i}`].key);
          highlight.value.minY = parseFloat(lq[`minY_${i}`].key);
          highlight.value.maxY = parseFloat(lq[`maxY_${i}`].key);
        } else if (type === DATETIME_FILTER) {
          highlight.value.from = dateToNum(lq[`min_${i}`]);
          highlight.value.to = dateToNum(lq[`max_${i}`]);
          highlight.value.type = DATETIME_FILTER;
        } else if (isNumericType(type)) {
          highlight.value.from = parseFloat(lq[`min_${i}`].key);
          highlight.value.to = parseFloat(lq[`max_${i}`].key);
          highlight.value.type = NUMERICAL_FILTER;
        } else {
          const values: string[] = [];
          for (i = 0; i < lq.field.meta.count; ++i) {
            values.push(lq[`value_${i}`].key);
          }
          highlight.value = values;
        }

        highlights.push(highlight);
      }
    }
  });
  return {
    filters: filters,
    highlights: highlights,
  };
}

function modeToRelation(mode: string): ValueStateValue {
  switch (mode) {
    case HIGHLIGHT:
      return distilRelationOptions[0];
    case EXCLUDE_FILTER:
      return distilRelationOptions[1];
    default:
      return distilRelationOptions[0];
  }
}
function modeToColor(mode: string): string[] {
  switch (mode) {
    case HIGHLIGHT:
      return ["include-filter"];
    case EXCLUDE_FILTER:
      return ["exclude-filter"];
    default:
      return ["include-filter"];
  }
}

/*
  Formats distil variables to Lex Suggestions AKA ValueStateValues so they can
  be used in the Lex Language and in translating filter/highlight state into a
  lex query. Also ungroups some variables such that we can use them in lex
  queries as that reflects the current filter/highlight behavior.
*/
function variablesToLexSuggestions(
  variables: VariableInfo[],
  variableMap: Map<string, Variable>
): ValueStateValue[] {
  if (!variables) return;
  return variables.reduce((a, v) => {
    const grouping = v.variable.grouping as ClusteredGrouping;
    const clusterVar = variableMap.get(grouping?.clusterCol);
    const name = v.variable.colDisplayName;
    const options = {
      type: colTypeToOptionType(
        clusterVar?.colType?.toLowerCase() ?? v.variable?.colType?.toLowerCase()
      ),
      variable: v.variable,
      name,
      count: v.count,
      mode: v.mode,
      key: v.variable.key,
    };
    const config = {
      displayKey: v.variable.colDisplayName,
    };
    a.push(new ValueStateValue(name, options, config));
    return a;
  }, []);
}

/*
  uses the value data in categorical variables to build a per variable dictionary
  of suggestion lists whose values are LexBar ValueStateValues
*/
function perCategoricalVariableLexSuggestions(
  variables: Variable[]
): Dictionary<ValueStateValue[]> {
  const categoryDict = new Object() as Dictionary<ValueStateValue[]>;

  variables.forEach((v) => {
    if (v.colType === CATEGORICAL_TYPE && v.values !== null) {
      categoryDict[v.key] = v.values.map((c) => new ValueStateValue(c));
    }
  });

  return categoryDict;
}

function colTypeToOptionType(colType: string): string {
  if (
    colType === GEOBOUNDS_TYPE ||
    colType === GEOCOORDINATE_TYPE ||
    colType === BIVARIATE_FILTER
  ) {
    return GEOBOUNDS_FILTER;
  } else if (colType === DATE_TIME_LOWER_TYPE) {
    return DATETIME_FILTER;
  } else if (isNumericType(colType)) {
    return NUMERICAL_FILTER;
  } else if (colType === CATEGORICAL_TYPE || colType === TIMESERIES_TYPE) {
    return CATEGORICAL_FILTER;
  } else {
    return TEXT_FILTER;
  }
}

/*
  Convert Distil Variable Array To a Dictionary For O(1) look up. Used when
  converting a filter/highlight from the distil format to a lex query.
*/
function buildVariableDictionary(variables: Variable[]) {
  return variables.reduce((a, v) => {
    a[v.key] = v;
    return a;
  }, {});
}
