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
  TextEntryState,
  TransitionFactory,
  ValueState,
  ValueStateValue,
} from "@uncharted.software/lex";
import { Highlight, Variable } from "../store/dataset";
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

/*
  This is the core function that actually generates a Lex Bar language. It takes
  a list of distil variables, converts them to an array of Lex Suggestions, then
  combines that with branching logic based on the suggestion's type to provide
  transitions to data entry states that fit that variable's type. As we add
  variable types with distinct entry needs, we can extend this function and the
  functions it depends on to support it in the Lex Bar language.
*/
export function variablesToLexLanguage(variables: Variable[]): Lex {
  const suggestions = variablesToLexSuggestions(variables);
  const catVarLexSuggestions = perCategoricalVariableLexSuggestions(variables);
  return Lex.from("field", ValueState, {
    name: "Choose a variable to search on",
    icon: '<i class="fa fa-filter" />',
    suggestions: suggestions,
  }).branch(
    Lex.from("relation", DistilRelationState, {
      ...TransitionFactory.valueMetaCompare({ type: TEXT_FILTER }),
    }).branch(Lex.from("value", TextEntryState)),
    Lex.from("relation", DistilRelationState, {
      ...TransitionFactory.valueMetaCompare({ type: CATEGORICAL_TYPE }),
    }).branch(
      Lex.from("value", ValueState, {
        allowUnknown: false,
        icon: "",
        name: "Type for suggestions",
        fetchSuggestions: (hint, lexDefintion) => {
          return catVarLexSuggestions[lexDefintion.field.key]
            ? catVarLexSuggestions[lexDefintion.field.key].filter((cat) => {
                return cat.key.toLowerCase().indexOf(hint.toLowerCase()) > -1;
              })
            : [];
        },
      })
    ),
    Lex.from("relation", DistilRelationState, {
      ...TransitionFactory.valueMetaCompare({ type: NUMERICAL_FILTER }),
    }).branch(
      Lex.from(LabelState, { label: "From" })
        .to("min", NumericEntryState, { name: "Enter lower bound" })
        .to(LabelState, { label: "To" })
        .to("max", NumericEntryState, { name: "Enter upper bound" })
    ),
    Lex.from("relation", DistilRelationState, {
      ...TransitionFactory.valueMetaCompare({ type: DATETIME_FILTER }),
    }).branch(
      Lex.from(LabelState, { label: "From" })
        .to("min", DateTimeEntryState, {
          enableTime: true,
          enableCalendar: true,
          timezone: "Greenwich",
        })
        .to(LabelState, { label: "To" })
        .to("max", DateTimeEntryState, {
          enableTime: true,
          enableCalendar: true,
          timezone: "Greenwich",
        })
    ),
    Lex.from("relation", DistilRelationState, {
      ...TransitionFactory.valueMetaCompare({ type: GEOBOUNDS_FILTER }),
    }).branch(
      Lex.from(LabelState, { label: "From Latitude" })
        .to("minX", NumericEntryState, { name: "Enter lower bound" })
        .to(LabelState, { label: "To" })
        .to("maxX", NumericEntryState, { name: "Enter upper bound" })
        .to(LabelState, { label: "From Longitude" })
        .to("minY", NumericEntryState, { name: "Enter lower bound" })
        .to(LabelState, { label: "To" })
        .to("maxY", NumericEntryState, { name: "Enter upper bound" })
    )
  );
}

export function filterParamsToLexQuery(
  filter: string,
  highlight: string,
  allVariables: Variable[]
) {
  const decodedFilters = decodeFilters(filter).list.filter(
    (f) => f.type !== "row"
  );
  const decodedHighlights = createFiltersFromHighlights(
    decodeHighlights(highlight),
    HIGHLIGHT
  );

  const variableDict = buildVariableDictionary(allVariables);
  const filterVariables = [];
  decodedFilters.forEach((f) => {
    if (variableDict[f.key]) {
      filterVariables.push(variableDict[f.key]);
    }
  });
  const highlightVariables = [];
  decodedHighlights.forEach((h) => {
    if (variableDict[h.key]) {
      highlightVariables.push(variableDict[h.key]);
    }
  });

  const activeVariables = [...highlightVariables, ...filterVariables];

  const activeVariablesMap = new Map(
    activeVariables.map((v) => {
      return [v.key, true];
    })
  );
  const lexableElements = [
    ...decodedHighlights.filter((el) => {
      return activeVariablesMap.has(el.key);
    }),
    ...decodedFilters.filter((el) => {
      return activeVariablesMap.has(el.key);
    }),
  ];
  const suggestions = variablesToLexSuggestions(activeVariables);
  const lexQuery = lexableElements.map((f, i) => {
    if (f.type === GEOBOUNDS_FILTER || f.type === BIVARIATE_FILTER) {
      return {
        field: suggestions[i],
        minX: new ValueStateValue(f.minX),
        maxX: new ValueStateValue(f.maxX),
        minY: new ValueStateValue(f.minY),
        maxY: new ValueStateValue(f.maxY),
        relation: modeToRelation(f.mode),
      };
    } else if (f.type === DATETIME_FILTER) {
      return {
        field: suggestions[i],
        min: new Date(f.min * 1000),
        max: new Date(f.max * 1000),
        relation: modeToRelation(f.mode),
      };
    } else if (isNumericType(f.type)) {
      return {
        field: suggestions[i],
        min: new ValueStateValue(f.min),
        max: new ValueStateValue(f.max),
        relation: modeToRelation(f.mode),
      };
    } else {
      return {
        field: suggestions[i],
        value: new ValueStateValue(f.categories[0], null, {
          displayKey: f.categories[0],
        }),
        relation: modeToRelation(f.mode),
      };
    }
  });
  return lexQuery;
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
  const filters = [];
  const highlights = [];

  lexQuery[0].forEach((lq) => {
    if (lq.relation.key !== HIGHLIGHT) {
      const key = lq.field.key;
      const displayKey = lq.field.displayKey;
      const type = lq.field.meta.type;
      const filter: Filter = {
        mode: lq.relation.key,
        displayName: displayKey,
        type,
        key,
      };

      if (type === GEOBOUNDS_FILTER || type === GEOCOORDINATE_FILTER) {
        filter.key = filter.key;
        filter.minX = parseFloat(lq.minX.key);
        filter.maxX = parseFloat(lq.maxX.key);
        filter.minY = parseFloat(lq.minY.key);
        filter.maxY = parseFloat(lq.maxY.key);
      } else if (type === DATETIME_FILTER) {
        filter.min = dateToNum(lq.min);
        filter.max = dateToNum(lq.max);
      } else if (isNumericType(type)) {
        filter.min = parseFloat(lq.min.key);
        filter.max = parseFloat(lq.max.key);
      } else {
        filter.categories = [lq.value.key];
      }

      filters.push(filter);
    } else {
      const key = lq.field.key;
      const type = lq.field.meta.type;
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
        highlight.value.minX = parseFloat(lq.minX.key);
        highlight.value.maxX = parseFloat(lq.maxX.key);
        highlight.value.minY = parseFloat(lq.minY.key);
        highlight.value.maxY = parseFloat(lq.maxY.key);
      } else if (type === DATETIME_FILTER) {
        highlight.value.from = dateToNum(lq.min);
        highlight.value.to = dateToNum(lq.max);
        highlight.value.type = DATETIME_FILTER;
      } else if (isNumericType(type)) {
        highlight.value.from = parseFloat(lq.min.key);
        highlight.value.to = parseFloat(lq.max.key);
        highlight.value.type = NUMERICAL_FILTER;
      } else {
        highlight.value = lq.value.key;
      }

      highlights.push(highlight);
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

/*
  Formats distil variables to Lex Suggestions AKA ValueStateValues so they can
  be used in the Lex Language and in translating filter/highlight state into a
  lex query. Also ungroups some variables such that we can use them in lex
  queries as that reflects the current filter/highlight behavior.
*/
function variablesToLexSuggestions(variables: Variable[]): ValueStateValue[] {
  if (!variables) return;
  return variables.reduce((a, v) => {
    const name = v.key;
    const options = {
      type: colTypeToOptionType(v.colType.toLowerCase()),
    };
    const config = {
      displayKey: v.colDisplayName,
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
  if (colType === GEOBOUNDS_TYPE || colType === GEOCOORDINATE_TYPE) {
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
