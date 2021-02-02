import {
  GeoCoordinateGrouping,
  Highlight,
  TimeseriesGrouping,
  Variable,
} from "../store/dataset";
import {
  isNumericType,
  dateToNum,
  DATE_TIME_LOWER_TYPE,
  CATEGORICAL_TYPE,
  TIMESERIES_TYPE,
  GEOCOORDINATE_TYPE,
  GEOBOUNDS_TYPE,
  MULTIBAND_IMAGE_TYPE,
} from "./types";
import {
  decodeFilters,
  Filter,
  EXCLUDE_FILTER,
  CATEGORICAL_FILTER,
  DATETIME_FILTER,
  NUMERICAL_FILTER,
  TEXT_FILTER,
  GEOBOUNDS_FILTER,
  GEOCOORDINATE_FILTER,
} from "./filters";
import {
  LabelState,
  Lex,
  NumericEntryState,
  TextEntryState,
  DateTimeEntryState,
  TransitionFactory,
  ValueState,
  ValueStateValue,
  RelationState,
} from "@uncharted.software/lex";
import { decodeHighlights, createFilterFromHighlight } from "./highlights";
import { Dictionary } from "./dict";

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
  const decodedFilters = decodeFilters(filter).filter((f) => f.type !== "row");
  const decodedHighlight =
    highlight &&
    createFilterFromHighlight(decodeHighlights(highlight), HIGHLIGHT);

  const variableDict = buildVariableDictionary(allVariables);
  const filterVariables = decodedFilters.map((f) => {
    return variableDict[f.key];
  });
  const highlightVariable = variableDict[decodedHighlight?.key];
  const hasHighlight = !!highlight && !!highlightVariable;

  const activeVariables = hasHighlight
    ? [highlightVariable, ...filterVariables]
    : filterVariables;
  const lexableElements = hasHighlight
    ? [decodedHighlight, ...decodedFilters]
    : decodedFilters;

  const suggestions = variablesToLexSuggestions(activeVariables);

  const lexQuery = lexableElements.map((f, i) => {
    if (f.type === GEOBOUNDS_FILTER) {
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
): { filters: Filter[]; highlight: Highlight } {
  const filters = [];
  let highlight = null;

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
        filter.key = filter.key + "_group";
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
      highlight = {
        dataset,
        context: "lex-bar",
        key,
        value: {},
      } as Highlight;

      if (type === GEOBOUNDS_FILTER || type === GEOCOORDINATE_FILTER) {
        highlight.key = highlight.key + "_group";
        highlight.value.minX = parseFloat(lq.minX.key);
        highlight.value.maxX = parseFloat(lq.maxX.key);
        highlight.value.minY = parseFloat(lq.minY.key);
        highlight.value.maxY = parseFloat(lq.maxY.key);
      } else if (type === DATETIME_FILTER) {
        highlight.value.min = dateToNum(lq.min);
        highlight.value.max = dateToNum(lq.max);
        highlight.type = DATETIME_FILTER;
      } else if (isNumericType(type)) {
        highlight.value.min = parseFloat(lq.min.key);
        highlight.value.max = parseFloat(lq.max.key);
        highlight.type = NUMERICAL_FILTER;
      } else {
        highlight.value = lq.value.key;
      }
    }
  });
  return {
    filters: filters,
    highlight: highlight,
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

    if (v.distilRole === "grouping") {
      switch (v.colType) {
        case TIMESERIES_TYPE:
          const grouping = v.grouping as TimeseriesGrouping;
          a.push(
            new ValueStateValue(
              grouping.xCol,
              { type: DATETIME_FILTER },
              { displayKey: grouping.xCol }
            )
          );
          break;
        case MULTIBAND_IMAGE_TYPE:
        case GEOBOUNDS_TYPE:
        case GEOCOORDINATE_TYPE:
          /* not currently ungrouping any information from these types
          for lex suggestions, but not unknown either, so no logging */
          break;
        default:
          console.log("unknown grouped type");
      }
    }

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
  if (colType.toLowerCase() === GEOBOUNDS_TYPE) {
    return GEOBOUNDS_FILTER;
  } else if (colType.toLowerCase() === DATE_TIME_LOWER_TYPE) {
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
  converting a filter/highlight from the distil format to a lex query. Unpacks
  grouped types as filters/highlights can be based on the underlying variable.
*/
function buildVariableDictionary(variables: Variable[]) {
  return variables.reduce((a, v) => {
    a[v.key] = v;
    if (v.distilRole === "grouping") {
      switch (v.colType) {
        case TIMESERIES_TYPE:
          const grouping = v.grouping as TimeseriesGrouping;
          const xCol = grouping.xCol;
          a[xCol] = {
            key: xCol,
            colDisplayName: xCol,
            colType: DATE_TIME_LOWER_TYPE,
          } as Variable;
          break;
        case GEOCOORDINATE_TYPE:
          const geoGrouping = v.grouping as GeoCoordinateGrouping;
          const lat = geoGrouping.xCol;
          const lon = geoGrouping.yCol;
          a[lat] = {
            key: lat,
            colDisplayName: lat,
            colType: NUMERICAL_FILTER,
          } as Variable;
          a[lon] = {
            key: lon,
            colDisplayName: lon,
            colType: NUMERICAL_FILTER,
          } as Variable;
          break;
        case MULTIBAND_IMAGE_TYPE:
        case GEOBOUNDS_TYPE:
          /* to do */
          break;
        default:
          console.warn("unknown grouped type");
      }
    }

    return a;
  }, {});
}
