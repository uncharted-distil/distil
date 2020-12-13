// the geo coordinate type
type LatLngBounds = import("leaflet").LatLngBounds;
type LatLng = import("leaflet").LatLng;
import Color from "color";
import { TableRow } from "../../store/dataset/index";

// contains rgba normalized values for webgl
interface ColorBuffer {
  r: number;
  g: number;
  b: number;
  a: number;
}
// contains x,y vertex data for webgl
interface Point {
  x: number;
  y: number;
}
// contains the an int id where each byte is a property
interface IdBuffer {
  iR: number;
  iG: number;
  iB: number;
  iA: number;
}
// contains required data for a primitive
interface PrimitiveData {
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
interface Area {
  coordinates: LatLngBounds;
  color: string;
  imageUrl: string;
  item: TableRow;
}
// PrimitiveManager is responsible for all primitives render in an overlay fashion for the Lumo maps
export class PrimitiveManager {
  pointBuffer: PrimitiveData[];
  quadBuffer: PrimitiveData[];
  primitiveOpacity: number;
  getCirclePrimitiveBuffer(): PrimitiveData[] {
    return this.pointBuffer;
  }
  getQuadPrimitiveBuffer(): PrimitiveData[] {
    return this.quadBuffer;
  }
  // add area data to primitive manager
  addLatLngBounds(areas: Area[]): void {
    areas.forEach((area, idx) => {
      // create circle buffer
      const point = this.normalizeLatLng(area.coordinates.getCenter());
      const colorBuffer = this.colorToRgb(area.color);
      const idBuffer = this.idToRGBA(idx);
      this.pointBuffer.push({ ...point, ...colorBuffer, ...idBuffer });
      // create quad
      const p1 = this.normalizeLatLng(area.coordinates.getNorthWest());
      const p2 = this.normalizeLatLng(area.coordinates.getSouthEast());
      const colorId = { ...colorBuffer, ...idBuffer };
      this.quadBuffer.push({ ...p1, ...colorId });
      this.quadBuffer.push({ x: p2.x, y: p1.y, ...colorId });
      this.quadBuffer.push({ ...p2, ...colorId });
      this.quadBuffer.push({ ...p1, ...colorId });
      this.quadBuffer.push({ x: p1.x, y: p2.y, ...colorId });
      this.quadBuffer.push({ ...p2, ...colorId });
    });
  }
  clearPrimitives() {
    this.pointBuffer = [];
    this.quadBuffer = [];
  }
  idToRGBA(id: number): IdBuffer {
    // 0 is reserved for background
    const ID = id + 1;
    return {
      iR: ((ID >> 0) & 0xff) / 0xff,
      iG: ((ID >> 8) & 0xff) / 0xff,
      iB: ((ID >> 16) & 0xff) / 0xff,
      iA: ((ID >> 24) & 0xff) / 0xff,
    };
  }
  colorToRgb(hex: string): ColorBuffer {
    const maxVal = 255;
    const color = Color(hex).rgb.object() as ColorBuffer;
    color.a = this.primitiveOpacity;
    color.r /= maxVal;
    color.g /= maxVal;
    color.b /= maxVal;
    return color;
  }
  normalizeLatLng(latLngBounds: LatLng): Point {
    const maxLon = 180.0;
    const degreesToRadians = Math.PI / 180.0; // Factor for changing degrees to radians
    const latRadians = latLngBounds.lat * degreesToRadians;
    const x = (latLngBounds.lng + maxLon) / (maxLon * 2);
    const y =
      (1 -
        Math.log(Math.tan(latRadians) + 1 / Math.cos(latRadians)) / Math.PI) /
      2;

    return { x, y: 1 - y } as Point; // have to invert y
  }
}
