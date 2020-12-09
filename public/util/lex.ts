import { Variable } from "../store/dataset";
import { isNumericType, isTextType, TEXT_TYPE, NUMERIC_TYPE } from "./types";
import {
  LabelState,
  Lex,
  NumericEntryState,
  TextEntryState,
  TransitionFactory,
  ValueState,
  ValueStateValue,
} from "@uncharted.software/lex";

export function variablesToLexLanguage(variables: Variable[]): Lex {
  const suggestions = variablesToLexSuggestions(variables);
  return Lex.from("field", ValueState, {
    name: "Choose a variable to filter",
    icon: '<i class="fa fa-filter" />',
    suggestions: suggestions,
  }).branch(
    Lex.from("value", TextEntryState, {
      // User option meta compare to limit this branch to string fields
      ...TransitionFactory.valueMetaCompare({ type: TEXT_TYPE }),
    }),
    Lex.from(LabelState, {
      label: "From",
      ...TransitionFactory.valueMetaCompare({ type: NUMERIC_TYPE }),
    })
      .to("lower bound", NumericEntryState, { name: "Enter lower bound" })
      .to(LabelState, { label: "To" })
      .to("upper bound", NumericEntryState, { name: "Enter upper bound" })
  );
}

function variablesToLexSuggestions(variables: Variable[]): ValueStateValue[] {
  if (!variables) return;
  return variables.map((variable) => {
    const name = variable.colDisplayName;
    const options = {
      type: colTypeToOptionType(variable.colType.toLowerCase()),
    };
    return new ValueStateValue(name, options);
  });
}

function colTypeToOptionType(colType: string): string {
  if (isNumericType(colType)) {
    return NUMERIC_TYPE;
  } else if (isTextType(colType)) {
    return TEXT_TYPE;
  } else {
    return TEXT_TYPE;
  }
}
