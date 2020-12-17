import {
  GeoCoordinateGrouping,
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
import { layerGroup } from "leaflet";

const distilRelationOptions = [
  [INCLUDE_FILTER, "="],
  [EXCLUDE_FILTER, "≠"],
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
    )
  );
}

export function filterParamsToLexQuery(
  filter: string,
  allVariables: Variable[]
) {
  const filters = decodeFilters(filter);
  const variableDict = buildVariableDictionary(allVariables);
  const filterVariables = filters.map((f) => {
    return variableDict[f.key];
  });
  const suggestions = variablesToLexSuggestions(filterVariables);

  const lexQuery = filters.map((f, i) => {
    if (isNumericType(f.type)) {
      return {
        field: suggestions[i],
        min: new ValueStateValue(f.min),
        max: new ValueStateValue(f.max),
        relation: modeToRelation(f.mode),
      };
    } else if (f.type === DATETIME_FILTER) {
      return {
        field: suggestions[i],
        min: new Date(f.min * 1000),
        max: new Date(f.max * 1000),
        relation: modeToRelation(f.mode),
      };
    } else {
      return {
        field: suggestions[i],
        value: new ValueStateValue(f.categories[0]),
        relation: modeToRelation(f.mode),
      };
    }
  });
  return lexQuery;
}

export function lexQueryToFilters(lexQuery: any[][]): Filter[] {
  const filters = lexQuery[0].map((lq) => {
    const key = lq.field.key;
    const type = lq.field.meta.type;
    const filter: Filter = {
      mode: lq.relation.key,
      displayName: key,
      type,
      key,
    };

    if (isNumericType(type)) {
      filter.min = parseFloat(lq.min.key);
      filter.max = parseFloat(lq.max.key);
    } else if (type === DATETIME_FILTER) {
      filter.min = dateToNum(lq.min);
      filter.max = dateToNum(lq.max);
    } else {
      filter.categories = [lq.value.key];
    }

    return filter;
  });
  return filters;
}

function modeToRelation(mode: string): ValueStateValue {
  if (mode === INCLUDE_FILTER) {
    return distilRelationOptions[0];
  } else {
    return distilRelationOptions[1];
  }
}

function variablesToLexSuggestions(variables: Variable[]): ValueStateValue[] {
  if (!variables) return;

  return variables.reduce((a, v) => {
    const name = v.colDisplayName;
    const options = {
      type: colTypeToOptionType(v.colType.toLowerCase()),
    };
    a.push(new ValueStateValue(name, options));

    if (v.distilRole === "grouping") {
      switch (v.colType) {
        case TIMESERIES_TYPE:
          const grouping = v.grouping as TimeseriesGrouping;
          a.push(new ValueStateValue(grouping.xCol, { type: DATETIME_FILTER }));
          break;
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
  if (isNumericType(colType)) {
    return NUMERICAL_FILTER;
  } else if (colType.toLowerCase() === DATE_TIME_LOWER_TYPE) {
    return DATETIME_FILTER;
  } else if (colType === CATEGORICAL_TYPE) {
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
            colDisplayName: xCol,
            colType: DATE_TIME_LOWER_TYPE,
          } as Variable;
          console.log(grouping, a[xCol]);
          break;
        case GEOCOORDINATE_TYPE:
          const geoGrouping = v.grouping as GeoCoordinateGrouping;
          const lat = geoGrouping.xCol;
          const lon = geoGrouping.yCol;
          a[lat] = {
            colDisplayName: lat,
            colType: NUMERICAL_FILTER,
          } as Variable;
          a[lon] = {
            colDisplayName: lon,
            colType: NUMERICAL_FILTER,
          } as Variable;
          break;
        default:
          console.warn("unknown grouped type");
      }
    }

    return a;
  }, {});
}
