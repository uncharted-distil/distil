# distil

## Dependencies

- [Go](https://golang.org/) programming language binaries with the `GOPATH` environment variable specified and `$GOPATH/bin` in your `PATH`.
- [NodeJS](http://nodejs.org/) JavaScript runtime.
- [Docker](https://www.docker.com/) platform.

## Development

Clone the repository:

```bash
mkdir -p $GOPATH/src/github.com/unchartedsoftware
cd $GOPATH/src/github.com/unchartedsoftware
git clone git@github.com:unchartedsoftware/distil.git
```

Install dependencies:

```bash
cd distil
make install
```
Install protocol buffer compiler:

Linux

```bash
curl -OL https://github.com/google/protobuf/releases/download/v3.3.0/protoc-3.3.0-linux-x86_64.zip
unzip protoc-3.3.0-linux-x86_64.zip -d protoc3
sudo mv protoc3/bin/protoc /usr/bin/protoc
```

OSX

```bash
curl -OL https://github.com/google/protobuf/releases/download/v3.3.0/protoc-3.3.0-osx-x86_64.zip
unzip protoc-3.3.0-osx-x86_64.zip -d protoc3
sudo mv protoc3/bin/protoc /usr/bin/protoc
```

Compile protobuffers:

```bash
make protoc
```

The application depends on ElasticSearch for data, and a stub TA2 system for back end integration.  Docker images (with data) for both are available:

```bash
docker pull docker.uncharted.software/distil_dev_es
docker pull docker.uncharted.software/distil_dev_postgres
docker pull docker.uncharted.software/distil-pipeline-server
```

Launch docker containers:

```bash
./es_run.sh
./pg_run.sh
./pipeline_server_run.sh
```

Build and watch webapp:
```bash
yarn watch
```

Build, watch, and run server:
```bash
make watch
```

## Vue + Vuex + Vue-Router Flow

We use [vue](https://github.com/vuejs/vue), [vuex](https://github.com/vuejs/vuex), [vue-router](https://github.com/vuejs/vue-router) and [vuex-router-sync](https://github.com/vuejs/vuex-router-sync) in the frontend app.

### Components / Views (vue)

The application is split into views, each comprised of one or more components.

### Routes (vue-router)

Everything is based off the route. The route contains **_entire_** reproducible state of the application. Therefore copy and pasting the current route into a new tab **_should_** result in the exact same view for a user.

The route is the ground truth and **_everything_** must be derivable from it. That is not to say that everything should go in the route. It should only contain the minimal information that is required to regenerate the state of the application. Any other data, typically pulled from the server via asynchronous requests, will be the result of actions dispatched to the store.

### Store (vuex + vuex-router-sync)

The store contains the route (via [vuex-router-sync](https://github.com/vuejs/vuex-router-sync)) and any auxiliary state that can be derived from the route.

### Application Architecture / Flow

Views are routed based off the URL, which is registered in `public/main.js`:

```javascript
const router = new VueRouter({
	routes: [
		{ path: '/route0', component: View0 },
		{ path: '/route1', component: View1 },
	]
});
```

Any change of state through user interaction is pushed to the router via the respective component:

```javaScript
methods: {
	clickOnButton() {
		this.$router.push({
			path: '/path',
			query: {
				someValue: this.computedValue,
			}
		});
	}
}
```

Components retrieve their values / data from the store via computed values:

```javascript
computed: {
	someRouteValue() {
		return this.$store.getters.getRouteValue();
	}
	someOtherData() {
		return this.$store.getters.otherData();
	}
}
```

Components watch the route for any change that may affect them. When a change occurs, and required action is then dispatched to the store. The store will update (via a commit), and a new value will be computed, thus updating the component.

```javaScript
watch: {
	'$route.query.someValue'() {
		this.$store.dispatch('someAction', this.someValue);
	}
}
```

Therefore the overall flow is:

- User interaction
- Component pushes to route
- Affected components dispatch actions from route watch
- Affected components computed new values from store changes
- View updates

NOTE: Any state that is shared between components should be managed by a higher level component rather than redundantly watching the route in multiple components. Ex. If components A and B both need state C which is dependent upon route query param D, a third component E should be created to watch the state and dispatch a single action upon change. Components A and B will then read from the store via computed values.

## Common Issues:

#### "glide: command not found":

- **Cause**: `$GOPATH/bin` has not been added to your `$PATH`.
- **Solution**: Add `export PATH=$PATH:$GOPATH/bin` to your `.bash_profile` or `.bashrc`.

#### "../repo/subpackage/file.go:10:2: cannot find package "github.com/company/package/subpackage" in any of":

- **Cause**: Dependencies are out of date or have not been installed
- **Solution**: Run `make install` to install latest dependencies.
