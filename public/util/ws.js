'use strict';

import _ from 'lodash';

const RETRY_INTERVAL_MS = 5000;

let _trackedID = 1;

function getHost() {
	const loc = window.location;
	const uri = (loc.protocol === 'https:') ? 'wss:' : 'ws:';
	return `${uri}//${loc.host}${loc.pathname}`;
}

function establishConnection(conn, callback) {
	conn.socket = new WebSocket(`${getHost()}${conn.url}`);
	// on open
	conn.socket.onopen = function() {
		conn.isOpen = true;
		console.log(`WebSocket conn established on /${conn.url}`);
		// send pending messages
		conn.pending.forEach(message => {
			conn.socket.send(JSON.stringify(message.payload));
		});
		conn.pending = [];
		// send pending stream messages
		conn.streams.forEach(stream => {
			stream.pending.forEach(msg => {
				conn.socket.send(JSON.stringify(msg));
			});
			stream.pending = [];
		});
		callback(null, conn);
	};
	// on message
	conn.socket.onmessage = function(event) {
		const res = JSON.parse(event.data);
		if (!conn.tracking.has(res.id)) {
			console.error('Unrecognized response: ', res,  ', discarding');
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
				const stream = conn.stream.get(res.id);
				stream.fn(res);
				break;
		}
	};
	// on close
	conn.socket.onclose = function() {
		// log close only if conn was ever open
		if (conn.isOpen) {
			console.warn(`WebSocket connection on /${conn.url} lost, attempting to reconnect in ${RETRY_INTERVAL_MS}ms`);
		} else {
			callback(new Error(`Unable to establish websocket connection on /${conn.url}`), null);
			return;
		}
		// delete socket
		conn.socket = null;
		// flag as closed
		conn.isOpen = false;
		// attempt to re-establish conn
		setTimeout(() => {
			establishConnection(conn, () => {});
		}, RETRY_INTERVAL_MS);
	};
}

function stripURL(url) {
	if (!url || !_.isString(url)) {
		throw `Provided URL \`${url}\` is invalid`;
	}
	// strip leading `/`
	url = (url[0] === '/') ? url.substring(1, url.length) : url;
	// strip trailing `/`
	url = (url[url.length - 1] === '/') ? url.substring(0, url.length - 1) : url;
	return url;
}

class Stream {
	constructor(conn, fn) {
		this.id = `${_trackedID++}`;
		this.conn = conn;
		this.fn = fn;
		this.pending = [];
	}
	send(msg) {
		if (this.conn.isOpen) {
			this.conn.socket.send(JSON.stringify(msg));
		} else {
			this.pending.push(msg);
		}
	}
	close() {
		this.conn.streams.delete(this.id);
		this.conn.tracking.delete(this.id);
	}
}

class Message {
	constructor(payload) {
		this.id = `${_trackedID++}`;
		this.payload = payload;
		this.payload.id = this.id;
		this.promise = new Promise((resolve, reject) => {
			this.resolve = resolve;
			this.reject = reject;
		});
	}
}

const MESSAGE = Symbol();
const STREAM = Symbol();

export default class Connection {
	constructor(url, callback) {
		this.url = stripURL(url);
		this.streams = new Map();
		this.messages = new Map();
		this.pending = [];
		this.tracking = new Map();
		this.isOpen = false;
		establishConnection(this, callback);
	}
	stream(fn) {
		const stream = new Stream(this, fn);
		this.streams.set(stream.id, stream);
		this.tracking.set(stream.id, STREAM);
		return stream;
	}
	send(payload) {
		const message = new Message(payload);
		this.messages.set(message.id, message);
		this.tracking.set(message.id, MESSAGE);
		if (this.isOpen) {
			this.conn.socket.send(JSON.stringify(message.payload));
		} else {
			this.pending.push(message);
		}
		return message.promise;
	}
	close() {
		this.socket.onclose = null;
		this.socket.close();
		this.socket = null;
		console.warn(`WebSocket conn on /${this.url} closed`);
	}
}
