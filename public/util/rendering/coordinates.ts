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

import { BatchQuadOverlayRenderer } from "./BatchQuadOverlayRenderer";
type LatLngBoundsLiteral = import("leaflet").LatLngBoundsLiteral;
import Color from "color";
import { TableRow } from "../../store/dataset";
import { Dictionary } from "../dict";
export enum Coordinate {
  lat,
  lng,
}
export interface VertexPrimitive {
  x: number; // vertex x
  y: number; // vertex y
  r: number; // color r channel
  g: number; // color g channel
  b: number; // color b channel
  a: number; // color alpha channel
  // id's bytes is broken down into 4 channels
  iR: number; // id smallest byte
  iG: number; // id second smallest byte
  iB: number; // id second largest byte
  iA: number; // id largest byte
}

export interface RenderPrimitive {
  vertexPrimitives: VertexPrimitive[];
}

export abstract class CoordinateInfo {
  coordinates: LatLngBoundsLiteral;
  color: string;
  abstract toQuad(
    renderer: BatchQuadOverlayRenderer,
    opacity: number,
    id: number
  ): RenderPrimitive;
  abstract toPoint(
    renderer: BatchQuadOverlayRenderer,
    opacity: number,
    id: number
  ): RenderPrimitive;
  abstract shouldTile(
    renderer: BatchQuadOverlayRenderer,
    extent: number,
    threshold: number
  ): boolean;
}

export class TileInfo extends CoordinateInfo {
  constructor(coordinates: LatLngBoundsLiteral, color: string) {
    super();
    this.coordinates = coordinates;
    this.color = color;
  }
  toQuad(renderer: BatchQuadOverlayRenderer, opacity: number, id: number) {
    const result = [];
    const p1 = renderer.latlngToNormalized(this.coordinates[0]);
    const p2 = renderer.latlngToNormalized(this.coordinates[1]);
    const color = Color(this.color).rgb().object(); // convert hex color to rgb
    const maxVal = 255;
    const v = { x: p1.x - p2.x, y: p1.y - p2.y };
    const magnitude = Math.sqrt(v.x * v.x + v.y * v.y);
    v.x /= magnitude;
    v.y /= magnitude;
    // this add a little distance between tiles to make it easier to see individual tiles in contiguous areas
    const distance = 0.000002;
    p1.x = p1.x - v.x * distance;
    p1.y = p1.y - v.y * distance;
    p2.x = p2.x + v.x * distance;
    p2.y = p2.y + v.y * distance;
    // normalize color values
    color.a = opacity;
    color.r /= maxVal;
    color.g /= maxVal;
    color.b /= maxVal;
    const renderID = renderer.idToRGBA(id); // separate index bytes into 4 channels iR,iG,iB,iA. Used to render the index of the object into webgl FBO
    // need to get rid of spread operators super slow
    result.push({ ...p1, ...color, ...renderID });
    result.push({ x: p2.x, y: p1.y, ...color, ...renderID });
    result.push({ ...p2, ...color, ...renderID });
    result.push({ ...p1, ...color, ...renderID });
    result.push({ x: p1.x, y: p2.y, ...color, ...renderID });
    result.push({ ...p2, ...color, ...renderID });
    return { vertexPrimitives: result };
  }
  toPoint(
    renderer: BatchQuadOverlayRenderer,
    opacity: number,
    id: number
  ): RenderPrimitive {
    const p1 = renderer.latlngToNormalized(this.coordinates[0]);
    const p2 = renderer.latlngToNormalized(this.coordinates[1]);
    const centerPoint = { x: (p1.x + p2.x) / 2, y: (p1.y + p2.y) / 2 };
    const color = Color(this.color).rgb().object(); // convert hex color to rgb
    const maxVal = 255;
    // normalize color values
    color.a = opacity;
    color.r /= maxVal;
    color.g /= maxVal;
    color.b /= maxVal;
    const renderID = renderer.idToRGBA(id); // separate index bytes into 4 channels iR,iG,iB,iA. Used to render the index of the object into webgl FBO
    // need to get rid of spread operators super slow
    return { vertexPrimitives: [{ ...centerPoint, ...color, ...renderID }] };
  }
  shouldTile(
    renderer: BatchQuadOverlayRenderer,
    extent: number,
    threshold: number
  ): boolean {
    const p1 = renderer.latlngToNormalized(this.coordinates[0]);
    const p2 = renderer.latlngToNormalized(this.coordinates[1]);
    const pixelPos1 = { x: p1.x * extent, y: p1.y * extent };
    const pixelPos2 = { x: p2.x * extent, y: p2.y * extent };
    const width = pixelPos2.x - pixelPos1.x;
    const height = pixelPos2.y - pixelPos1.y;
    return width * height > threshold;
  }
}

export class PointInfo extends CoordinateInfo {
  constructor(coordinates: LatLngBoundsLiteral, color: string) {
    super();
    this.coordinates = coordinates;
    this.color = color;
  }
  toPoint(renderer: BatchQuadOverlayRenderer, opacity: number, id: number) {
    const result = [];
    const p1 = renderer.latlngToNormalized(this.coordinates[0]);
    const color = Color(this.color).rgb().object(); // convert hex color to rgb
    const maxVal = 255;
    // normalize color values
    color.a = opacity;
    color.r /= maxVal;
    color.g /= maxVal;
    color.b /= maxVal;
    const renderID = renderer.idToRGBA(id); // separate index bytes into 4 channels iR,iG,iB,iA. Used to render the index of the object into webgl FBO
    // need to get rid of spread operators super slow
    result.push({ ...p1, ...color, ...renderID });
    return { vertexPrimitives: result };
  }
  toQuad(renderer: BatchQuadOverlayRenderer, opacity: number, id: number) {
    return { vertexPrimitives: [] };
  }
  // should never tile because this is a point only primitive
  shouldTile(
    renderer: BatchQuadOverlayRenderer,
    extent: number,
    threshold: number
  ): boolean {
    return false;
  }
}

export function updateVertexPrimitiveColor(
  vertices: VertexPrimitive[],
  data: TableRow[],
  colorFunc: (d, idx) => string,
  originalNumOfItems: number,
  baselineMap: Dictionary<number>
) {
  let idxFetch = (d: TableRow) => {
    return d.d3mIndex - 1;
  }; // d3mIndex starts at 1
  if (baselineMap != null) {
    idxFetch = (d: TableRow) => {
      return baselineMap[d.d3mIndex];
    };
    if (!Object.keys(baselineMap).length) {
      return;
    }
  }
  const step = vertices.length / originalNumOfItems;
  if (step % 1 != 0) {
    console.error("Step is decimal mismatch data length");
    return;
  }
  const maxVal = 255;
  const gray = Color("#999999").rgb().object();
  gray.r /= maxVal;
  gray.g /= maxVal;
  gray.b /= maxVal;
  for (let i = 0; i < vertices.length; ++i) {
    vertices[i].r = gray.r;
    vertices[i].g = gray.g;
    vertices[i].b = gray.b;
  }
  if (!data) {
    return;
  }
  for (let i = 0; i < data.length; ++i) {
    const color = Color(colorFunc(data[i], i)).rgb().object();
    color.r /= maxVal;
    color.g /= maxVal;
    color.b /= maxVal;
    const baseStep = idxFetch(data[i]) * step; //d3m index starts at 1
    for (let j = 0; j < step; ++j) {
      const idx = baseStep + j;
      vertices[idx].r = color.r;
      vertices[idx].g = color.g;
      vertices[idx].b = color.b;
    }
  }
}
