import { Location } from "vue-router";
import { LAST_STATE } from "../store/view/index";
import localStorage from "store";
import { local } from "d3";

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
