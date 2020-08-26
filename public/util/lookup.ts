import { Dictionary } from "./dict";

export function buildLookup(strs: any[]): Dictionary<boolean> {
  const lookup = {};
  strs.forEach((str) => {
    if (str) {
      lookup[str] = true;
      lookup[str.toLowerCase()] = true;
    } else {
      console.error(
        "Ignoring NULL string in look-up parameter list.  This should not happen.",
      );
    }
  });
  return lookup;
}
