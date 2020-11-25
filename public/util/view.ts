import { Location } from "vue-router";
import { LAST_STATE } from "../store/view/index";
import localStorage from "store";

import {
  hasGeoordinateFeatures,
  hasImageFeatures,
  hasTimeseriesFeatures,
} from "../util/data";
import { Variable } from "../store/dataset/index";

// Views used to display data
export const GEO_VIEW = "geo" as string;
export const GRAPH_VIEW = "graph" as string;
export const IMAGE_VIEW = "image" as string;
export const TABLE_VIEW = "table" as string;
export const TIMESERIES_VIEW = "timeseries" as string;

// Return a list of views available for the variables
export function filterViews(variables: Variable[]): string[] {
  const views = [TABLE_VIEW];
  if (hasGeoordinateFeatures(variables)) views.push(GEO_VIEW);
  if (hasImageFeatures(variables)) views.push(IMAGE_VIEW);
  // if (hasTimeseriesFeatures(variables)) views.push(TIMESERIES_VIEW); Disabled for now
  return views;
}

export function saveView(args: { view: string; key: string; route: Location }) {
  const value = {
    path: args.route.path,
    query: args.route.query,
  };
  // store under dataset
  if (args.key) {
    localStorage.set(`${args.view}:${args.key}`, value);
  }
  // store last as well in case no dataset available
  localStorage.set(`${args.view}:${LAST_STATE}`, value);
}

export function restoreView(view: string, key: string): Location {
  let res = localStorage.get(`${view}:${key}`);
  if (!res) {
    res = localStorage.get(`${view}:${LAST_STATE}`);
  }
  return res || null;
}
