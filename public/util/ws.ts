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

"use strict";

import _ from "lodash";

const RETRY_INTERVAL_MS = 5000;

const MESSAGE = Symbol();
const STREAM = Symbol();

let _trackedID = 1;

// Creates a new web socket Connection object
export function getWebSocketConnection(): Connection {
  const conn = new Connection("/ws", (err) => {
    if (err) {
      console.warn(err);
      return;
    }
  });
  return conn;
}

function getHost() {
  const loc = window.location;
  const uri = loc.protocol === "https:" ? "wss:" : "ws:";
  return `${uri}//${loc.host}${loc.pathname}`;
}

function establishConnection(
  conn: Connection,
  callback: (err: Error, c: Connection) => void
) {
  conn.socket = new WebSocket(`${getHost()}${conn.url}`);
  // on open
  conn.socket.onopen = () => {
    conn.isOpen = true;
    console.log(`WebSocket conn established on /${conn.url}`);
    // send pending messages
    conn.pending.forEach((message) => {
      conn.socket.send(JSON.stringify(message));
    });
    conn.pending = [];
    // send pending stream messages
    conn.streams.forEach((stream) => {
      stream.pending.forEach((streamMsg) => {
        conn.socket.send(JSON.stringify(streamMsg));
      });
      stream.pending = [];
    });
    callback(null, conn);
  };
  // on message
  conn.socket.onmessage = (event) => {
    const res = JSON.parse(event.data);
    if (!conn.tracking.has(res.id)) {
      console.error("Unrecognized response: ", res, ", discarding");
      return;
    }
    switch (conn.tracking.get(res.id)) {
      case MESSAGE:
        // message
        const message = conn.messages.get(res.id);
        conn.messages.delete(res.id);
        conn.tracking.delete(res.id);
        if (!res.success) {
          message.reject(res.error);
          return;
        }
        message.resolve(res);
        break;

      case STREAM:
        // stream
        const stream = conn.streams.get(res.id);
        stream.fn(res);
        break;
    }
  };
  // on close
  conn.socket.onclose = () => {
    // log close only if conn was ever open
    if (conn.isOpen) {
      console.warn(
        `WebSocket connection on /${conn.url} lost, attempting to reconnect in ${RETRY_INTERVAL_MS}ms`
      );
    } else {
      callback(
        new Error(`Unable to establish websocket connection on /${conn.url}`),
        null
      );
      return;
    }
    // delete socket
    conn.socket = null;
    // flag as closed
    conn.isOpen = false;
    // attempt to re-establish conn
    setTimeout(() => {
      establishConnection(conn, () => {
        return null;
      });
    }, RETRY_INTERVAL_MS);
  };
}

function stripURL(url: string) {
  if (!url || !_.isString(url)) {
    throw new Error(`Provided URL \`${url}\` is invalid`);
  }
  // strip leading `/`
  url = url[0] === "/" ? url.substring(1, url.length) : url;
  // strip trailing `/`
  url = url[url.length - 1] === "/" ? url.substring(0, url.length - 1) : url;
  return url;
}

// message header - contains a  msg type identifier and an id that is generated
// on send
interface Header {
  id: string;
  type: string;
}

class StreamMessage implements Header {
  id: string;
  type: string;
  body: unknown;

  constructor(id: string, type: string, body: unknown) {
    this.body = body;
    this.id = id;
    this.type = type;
  }
}

// Works with an established Connection object to send/receive messages
// over a websocket.  Received messages are handled by using a callback
// function passed into the Stream at the time of construction.
export class Stream {
  id: string;
  conn: Connection;
  fn: (x: unknown) => void;
  pending: StreamMessage[];

  constructor(conn: Connection, fn: (x: unknown) => void) {
    this.id = `${_trackedID++}`;
    this.conn = conn;
    this.pending = [];
    this.fn = fn;
  }

  send(type: string, body: unknown): void {
    const streamMessage = new StreamMessage(this.id, type, body);
    if (this.conn.isOpen) {
      this.conn.socket.send(JSON.stringify(streamMessage));
    } else {
      this.pending.push(streamMessage);
    }
  }

  close(): void {
    this.conn.streams.delete(this.id);
    this.conn.tracking.delete(this.id);
  }
}

type PromiseFunc = (t: unknown) => void;

class Message implements Header {
  id: string;
  type: string;
  resolve: PromiseFunc;
  reject: PromiseFunc;
  promise: Promise<unknown>;
  body: unknown; // amorphous body data

  constructor(type: string, body: unknown) {
    this.id = `${_trackedID++}`;
    this.type = type;
    this.body = body;
    this.promise = new Promise((resolve, reject) => {
      this.resolve = resolve;
      this.reject = reject;
    });
  }
}

// Abstracts a web socket connection.  Once the connection object has been
// created, Stream objects can be created to implement messaging over the
// socket.
export default class Connection {
  url: string;
  streams: Map<string, Stream>;
  messages: Map<string, Message>;
  pending: Message[];
  tracking: Map<string, symbol>;
  isOpen: boolean;
  socket: WebSocket | null;

  constructor(url: string, callback: (err: Error, c: Connection) => void) {
    this.url = stripURL(url);
    this.streams = new Map();
    this.messages = new Map();
    this.pending = [];
    this.tracking = new Map();
    this.isOpen = false;
    establishConnection(this, callback);
  }

  stream(fn: (x: unknown) => void): Stream {
    const stream = new Stream(this, fn);
    this.streams.set(stream.id, stream);
    this.tracking.set(stream.id, STREAM);
    return stream;
  }

  send(type: string, body: unknown): Promise<unknown> {
    const message = new Message(type, body);
    this.messages.set(message.id, message);
    this.tracking.set(message.id, MESSAGE);
    if (this.isOpen) {
      this.socket.send(JSON.stringify(message));
    } else {
      this.pending.push(message);
    }
    return message.promise;
  }

  close(): void {
    this.socket.onclose = null;
    this.socket.close();
    this.socket = null;

    console.info(`WebSocket conn on /${this.url} closed`);
  }
}
