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

import { Highlight, VariableSummary } from "../store/dataset";
import { DATE_TIME_TYPE, isCategoricalType } from "./types";
import { TableRow } from "../store/dataset/index";
import {
  interpolateInferno,
  interpolateMagma,
  interpolatePlasma,
  interpolateTurbo,
  interpolateViridis,
} from "d3-scale-chromatic";
import { Filter } from "./filters";

// ColorScaleNames is an enum that contains all the supported color scale names. Can be used to access COLOR_SCALES functions
export enum ColorScaleNames {
  viridis = "viridis",
  magma = "magma",
  inferno = "inferno",
  plasma = "plasma",
  turbo = "turbo",
  plotly = "plotly",
  d3 = "d3",
  g10 = "g10",
  t10 = "t10",
  alphabet = "alphabet",
  dark24 = "dark24",
  light24 = "light24",
}

export function getGradientScales(): ColorScaleNames[] {
  return [
    ColorScaleNames.viridis,
    ColorScaleNames.magma,
    ColorScaleNames.inferno,
    ColorScaleNames.plasma,
    ColorScaleNames.turbo,
  ];
}
export function getDiscreteScales(): ColorScaleNames[] {
  return [
    ColorScaleNames.plotly,
    ColorScaleNames.d3,
    ColorScaleNames.g10,
    ColorScaleNames.t10,
    ColorScaleNames.alphabet,
    ColorScaleNames.dark24,
    ColorScaleNames.light24,
  ];
}
function getColor(discreteScale: string[], idx: number): string {
  if (idx < 0) {
    console.error("Index out of range for discrete color scale");
  }
  // clamp discrete
  return discreteScale[Math.min(discreteScale.length - 1, idx)];
}
// colorByFacet returns a function that is used to calculate the normalized color scale value for a given facet type
export function colorByFacet(
  variable: VariableSummary,
  highlight?: Highlight,
  clusterKey = ""
): (item: TableRow, idx: number) => number {
  let key = clusterKey ? clusterKey : variable.key;
  // this is for result/prediction variables
  if (variable.key.includes(":")) {
    key = variable.key.split(":")[1];
  }
  if (isCategoricalType(variable.type)) {
    const keyMap = new Map(
      variable.baseline.buckets.map((b, i) => {
        return [b.key === "<none>" ? "" : b.key, i];
      })
    );
    return (item: TableRow, idx: number) => {
      return keyMap.get(item[key]?.value ?? item[variable.key]?.value);
    };
  } else if (variable.varType === DATE_TIME_TYPE) {
    const min = highlight
      ? highlight.value.from
      : variable.baseline.extrema.min;
    // the way the buckets are created makes the max value in the ds appear to be at the 90% mark instead of 100
    const max = Math.min(
      highlight ? highlight.value.to : variable.baseline.extrema.max,
      variable.baseline.extrema.max
    );
    const diff = max - min;
    return (item: TableRow, idx: number) => {
      if (diff === 0) {
        return 0;
      }
      return (new Date(item[key].value).getTime() / 1000 - min) / diff;
    };
  }
  // assume range
  else {
    const min = highlight
      ? highlight.value.from
      : variable.baseline.extrema.min;
    // the way the buckets are created makes the max value in the ds appear to be at the 90% mark instead of 100
    const max = Math.min(
      highlight ? highlight.value.to : variable.baseline.extrema.max,
      variable.baseline.extrema.max
    );
    const diff = max - min;
    return (item: TableRow) => {
      const itemValue = item[key]?.value ?? 0;
      return (itemValue - min) / diff;
    };
  }
}
/**
 * ****************** COLOR DEFINES *********************
 */
export const BLUE_PALETTE = [
  "#D8EAFA",
  "#CCE1F8",
  "#C0D9F6",
  "#B4D0F4",
  "#A8C8F2",
  "#9CBEEF",
  "#90B5EB",
  "#84ABE8",
  "#78A1E4",
  "#6C97E1",
  "#618EDD",
  "#5584DA",
  "#497AD6",
  "#3D70D3",
  "#3167CF",
  "#255DCC",
];

export const BLACK_PALETTE = [
  "#7F7F7F",
  "#777777",
  "#707070",
  "#696969",
  "#626262",
  "#5B5B5B",
  "#545454",
  "#4D4D4D",
  "#464646",
  "#3F3F3F",
  "#383838",
  "#313131",
  "#2A2A2A",
  "#232323",
  "#1C1C1C",
  "#151515",
  "#0E0E0E",
  "#070707",
  "#000000",
];

export const GRAY = "#999999";

export const BLUE = "#255DCC";

export const SELECTION_RED = "#ff0067";

export const BLACK = "#000000";

export const RESULT_RED = "#be0000";
export const RESULT_GREEN = "#03c003";

export const PLOTLY_PALETTE = [
  "#636EFA",
  "#EF553B",
  "#00CC96",
  "#AB63FA",
  "#FFA15A",
  "#19D3F3",
  "#FF6692",
  "#B6E880",
  "#FF97FF",
  "#FECB52",
];
export const D3_PALETTE = [
  "#1F77B4",
  "#FF7F0E",
  "#2CA02C",
  "#D62728",
  "#9467BD",
  "#8C564B",
  "#E377C2",
  "#7F7F7F",
  "#BCBD22",
  "#17BECF",
];
export const G10_PALETTE = [
  "#3366CC",
  "#DC3912",
  "#FF9900",
  "#109618",
  "#990099",
  "#0099C6",
  "#DD4477",
  "#66AA00",
  "#B82E2E",
  "#316395",
];
export const T10_PALETTE = [
  "#4C78A8",
  "#F58518",
  "#E45756",
  "#72B7B2",
  "#54A24B",
  "#EECA3B",
  "#B279A2",
  "#FF9DA6",
  "#9D755D",
  "#BAB0AC",
];
export const ALPHABET_PALETTE = [
  "#AA0DFE",
  "#3283FE",
  "#85660D",
  "#782AB6",
  "#565656",
  "#1C8356",
  "#16FF32",
  "#F7E1A0",
  "#E2E2E2",
  "#1CBE4F",
  "#C4451C",
  "#DEA0FD",
  "#FE00FA",
  "#325A9B",
  "#FEAF16",
  "#F8A19F",
  "#90AD1C",
  "#F6222E",
  "#1CFFCE",
  "#2ED9FF",
  "#B10DA1",
  "#C075A6",
  "#FC1CBF",
  "#B00068",
  "#FBE426",
  "#FA0087",
];
export const DARK24_PALETTE = [
  "#2E91E5",
  "#E15F99",
  "#1CA71C",
  "#FB0D0D",
  "#DA16FF",
  "#222A2A",
  "#B68100",
  "#750D86",
  "#EB663B",
  "#511CFB",
  "#00A08B",
  "#FB00D1",
  "#FC0080",
  "#B2828D",
  "#6C7C32",
  "#778AAE",
  "#862A16",
  "#A777F1",
  "#620042",
  "#1616A7",
  "#DA60CA",
  "#6C4516",
  "#0D2A63",
  "#AF0038",
];
export const LIGHT24_PALETTE = [
  "#FD3216",
  "#00FE35",
  "#6A76FC",
  "#FED4C4",
  "#FE00CE",
  "#0DF9FF",
  "#F6F926",
  "#FF9616",
  "#479B55",
  "#EEA6FB",
  "#DC587D",
  "#D626FF",
  "#6E899C",
  "#00B5F7",
  "#B68E00",
  "#C9FBE5",
  "#FF0092",
  "#22FFA7",
  "#E3EE9E",
  "#86CE00",
  "#BC7196",
  "#7E7DCD",
  "#FC6955",
  "#E48F72",
];
/**
 * ************ MAP DEFINES *********************
 */
export const DISCRETE_COLOR_MAPS: Map<ColorScaleNames, string[]> = new Map([
  [ColorScaleNames.plotly, PLOTLY_PALETTE],
  [ColorScaleNames.d3, D3_PALETTE],
  [ColorScaleNames.g10, G10_PALETTE],
  [ColorScaleNames.t10, T10_PALETTE],
  [ColorScaleNames.alphabet, ALPHABET_PALETTE],
  [ColorScaleNames.dark24, DARK24_PALETTE],
  [ColorScaleNames.light24, LIGHT24_PALETTE],
]);

// COLOR_SCALES contains the color scalefunctions that are js. This is for wrapping it in typescript.
export const COLOR_SCALES: Map<
  ColorScaleNames,
  (t: number) => string
> = new Map([
  [ColorScaleNames.viridis, interpolateViridis],
  [ColorScaleNames.magma, interpolateMagma],
  [ColorScaleNames.inferno, interpolateInferno],
  [ColorScaleNames.plasma, interpolatePlasma],
  [ColorScaleNames.turbo, interpolateTurbo],
  [
    ColorScaleNames.plotly,
    (t: number) => {
      return getColor(PLOTLY_PALETTE, t);
    },
  ],
  [
    ColorScaleNames.d3,
    (t: number) => {
      return getColor(D3_PALETTE, t);
    },
  ],
  [
    ColorScaleNames.g10,
    (t: number) => {
      return getColor(G10_PALETTE, t);
    },
  ],
  [
    ColorScaleNames.t10,
    (t: number) => {
      return getColor(T10_PALETTE, t);
    },
  ],
  [
    ColorScaleNames.alphabet,
    (t: number) => {
      return getColor(ALPHABET_PALETTE, t);
    },
  ],
  [
    ColorScaleNames.dark24,
    (t: number) => {
      return getColor(DARK24_PALETTE, t);
    },
  ],
  [
    ColorScaleNames.light24,
    (t: number) => {
      return getColor(LIGHT24_PALETTE, t);
    },
  ],
]);
