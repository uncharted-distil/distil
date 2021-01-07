import {
  GeoCoordinateGrouping,
  Highlight,
  TimeseriesGrouping,
  Variable,
} from "../store/dataset";
import {
  isNumericType,
  isTextType,
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
  INCLUDE_FILTER,
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

const HIGHLIGHT = "highlight";

const distilRelationOptions = [
  [INCLUDE_FILTER, "="],
  [EXCLUDE_FILTER, "≠"],
  [HIGHLIGHT, "☀"],
].map((o) => new ValueStateValue(o[0], {}, { displayKey: o[1] }));

class DistilRelationState extends RelationState {
  static get INCLUDE() {
    return distilRelationOptions[0];
  }
  static get EXCLUDE() {
    return distilRelationOptions[1];
  }
  constructor(config) {
    if (config.name === undefined) config.name = "Include or exclude";
    config.options = function () {
      return distilRelationOptions;
    };
    super(config);
  }
}

export function variablesToLexLanguage(variables: Variable[]): Lex {
  const suggestions = variablesToLexSuggestions(variables);
  return Lex.from("field", ValueState, {
    name: "Choose a variable to filter",
    icon: '<i class="fa fa-filter" />',
    suggestions: suggestions,
  }).branch(
    Lex.from("relation", DistilRelationState, {
      ...TransitionFactory.valueMetaCompare({ type: TEXT_FILTER }),
    }).branch(Lex.from("value", TextEntryState)),
    Lex.from("relation", DistilRelationState, {
      ...TransitionFactory.valueMetaCompare({ type: CATEGORICAL_FILTER }),
    }).branch(Lex.from("value", TextEntryState)),
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
        .to(LabelState, { label: "To Latitude" })
        .to("maxX", NumericEntryState, { name: "Enter upper bound" })
        .to(LabelState, { label: "From Longitude" })
        .to("minY", NumericEntryState, { name: "Enter lower bound" })
        .to(LabelState, { label: "To Longitude" })
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
  const highlightVariable = highlight && variableDict[decodedHighlight.key];

  const activeVariables = highlight
    ? [highlightVariable, ...filterVariables]
    : filterVariables;
  const lexableElements = highlight
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

export function lexQueryToFiltersAndHighlight(
  lexQuery: any[][]
): { filters: Filter[]; highlight: Highlight } {
  const filters = [];
  let highlight = null; // to do: populate highlight

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
    }
  });
  return {
    filters: filters,
    highlight: highlight,
  };
}

function modeToRelation(mode: string): ValueStateValue {
  switch (mode) {
    case INCLUDE_FILTER:
      return distilRelationOptions[0];
    case EXCLUDE_FILTER:
      return distilRelationOptions[1];
    case HIGHLIGHT:
      return distilRelationOptions[2];
    default:
      return distilRelationOptions[1];
  }
}

function variablesToLexSuggestions(variables: Variable[]): ValueStateValue[] {
  if (!variables) return;

  return variables.reduce((a, v) => {
    const name = v.colName;
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
          break;
        default:
          console.log("unknown grouped type");
      }
    }

    return a;
  }, []);
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
  } else if (isTextType(colType)) {
    return TEXT_FILTER;
  } else {
    return TEXT_FILTER;
  }
}

function buildVariableDictionary(variables: Variable[]) {
  return variables.reduce((a, v) => {
    a[v.colName] = v;
    if (v.distilRole === "grouping") {
      switch (v.colType) {
        case TIMESERIES_TYPE:
          const grouping = v.grouping as TimeseriesGrouping;
          const xCol = grouping.xCol;
          a[xCol] = {
            colName: xCol,
            colDisplayName: xCol,
            colType: DATE_TIME_LOWER_TYPE,
          } as Variable;
          break;
        case GEOCOORDINATE_TYPE:
          const geoGrouping = v.grouping as GeoCoordinateGrouping;
          const lat = geoGrouping.xCol;
          const lon = geoGrouping.yCol;
          a[lat] = {
            colName: lat,
            colDisplayName: lat,
            colType: NUMERICAL_FILTER,
          } as Variable;
          a[lon] = {
            colName: lon,
            colDisplayName: lon,
            colType: NUMERICAL_FILTER,
          } as Variable;
          break;
        case MULTIBAND_IMAGE_TYPE:
        case GEOBOUNDS_TYPE:
          break;
        default:
          console.warn("unknown grouped type");
      }
    }

    return a;
  }, {});
}
