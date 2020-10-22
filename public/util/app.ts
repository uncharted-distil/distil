import { getters as appGetters } from "../store/app/module";
import store from "../store/store";

interface TileHandler {
  requestTile(x: number, y: number, z: number): any;
}

export function getTileHandler(): TileHandler {
  const tileRequestURL = appGetters.getTileRequestURL(store);
  const subdomains = appGetters.getSubdomains(store);
  const mapAPIKey = appGetters.getMapAPIKey(store);
  // all the symbols that can potentially reside in the tileRequestURL
  const symbolTokens = {
    token: "{t}",
    subdomains: "{s}",
    x: "{x}",
    y: "{y}",
    z: "{z}",
  };
  const replaceMap = {}; // the map that contains the symbols that need to be replaced
  let subdomainArray = []; // sub domains
  if (subdomains.length) {
    subdomainArray = subdomains.split(",");
  }
  let sessionToken = "";
  if (mapAPIKey.length) {
    sessionToken = appGetters.getSessionToken(store);
  }
  if (tileRequestURL.includes(symbolTokens.token)) {
    replaceMap[symbolTokens.token] = sessionToken;
  }
  if (tileRequestURL.includes(symbolTokens.subdomains)) {
    return {
      requestTile: (x: number, y: number, z: number) => {
        replaceMap[symbolTokens.subdomains] =
          subdomainArray[(x + y + z) % subdomainArray.length];
        replaceMap[symbolTokens.x] = x.toString();
        replaceMap[symbolTokens.y] = y.toString();
        replaceMap[symbolTokens.z] = z.toString();
        return tileRequestURL.replace(
          `/${Object.keys(replaceMap).join("|")}/gi`,
          (matched) => {
            return replaceMap[matched];
          }
        );
      },
    };
  }

  return {
    requestTile: (x: number, y: number, z: number) => {
      replaceMap[symbolTokens.x] = x.toString();
      replaceMap[symbolTokens.y] = y.toString();
      replaceMap[symbolTokens.z] = z.toString();
      return tileRequestURL.replace(
        `/${Object.keys(replaceMap).join("|")}/gi`,
        (matched) => {
          return replaceMap[matched];
        }
      );
    },
  };
}
