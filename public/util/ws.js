'use strict';

import _ from 'lodash';

const RETRY_INTERVAL_MS = 5000;

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
		callback(null, conn);
	};
	// on message
	conn.socket.onmessage = function(event) {
		const msg = JSON.parse(event.data);
		if (!conn.streams.has(msg.id)) {
			console.error('Unrecognized response: ', msg,  ', discarding');
			return;
		}
		conn.streams.get(msg.id).dispatch(msg);
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
			establishConnection(conn, () => {
				// once conn is re-established, send pending messages
				conn.streams.forEach(stream => {
					stream.pending.forEach(msg => {
						conn.socket.send(JSON.stringify(msg));
					});
					stream.pending = [];
				});
			});
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
	constructor(id, conn) {
		this.id = id;
		this.conn = conn;
		this.on = new Map();
		this.pending = [];
	}
	on(id, fn) {
		if (!this.on.has(id)) {
			this.on.set(id, []);
		}
		this.on.get(id).push(fn);
	}
	off(id, fn) {
		if (this.on.has(id)) {
			const index = this.on.get(id).indexOf(fn);
			if (index !== -1) {
				this.on.get(id).splice(index, 1);
			}
		}
	}
	send(msg) {
		if (this.conn.isOpen) {
			this.conn.socket.send(JSON.stringify(msg));
		} else {
			this.pending.push(msg);
		}
	}
	dispatch(msg) {
		if (this.on.has(msg.id)) {
			this.on.get(msg.id).forEach(fn => {
				fn(msg);
			});
		}
	}
	close() {
		this.conn.streams.delete(this.id);
	}
}

export default class Connection {
	constructor(url, callback) {
		this.url = stripURL(url);
		this.streams = new Map();
		this.isOpen = false;
		establishConnection(this, callback);
	}
	stream(id) {
		if (this.streams.has(id)) {
			throw `Stream already exists for id \`${id}\``;
		}
		const stream = new Stream(id, this);
		this.streams.set(id, stream);
		return stream;
	}
	close() {
		this.socket.onclose = null;
		this.socket.close();
		this.socket = null;
		console.warn(`WebSocket conn on /${this.url} closed`);
	}
}
