# distil

[![CircleCI](https://circleci.com/gh/uncharted-distil/distil/tree/main.svg?style=svg&&circle-token=ff61c235865dd699cc8b923035a80e6e8d39c63a)](https://circleci.com/gh/unchartedsoftware/distil/tree/main)
[![Go Report Card](https://goreportcard.com/badge/github.com/uncharted-distil/distil)](https://goreportcard.com/report/github.com/uncharted-distil/distil)

## Related Projects

- [AutoML Server](https://github.com/uncharted-distil/distil-auto-ml) automated machine learning server component that implements the D3M API.
- [Primitives](https://github.com/uncharted-distil/distil-primitives) set of primitives created for use by Distil as steps in a D3M pipeline and included in the base D3M image.
- [Primitives Addendum](https://github.com/uncharted-distil/distil-primitives-contrib) set of primitives created for use by Distil as steps in a D3M pipeline and not included in the base D3M image.

## Dependencies

- [Git](https://git-scm.com) and [Git LFS](https://git-lfs.github.com) Versioning softwares.
- [Go](https://golang.org/) programming language binaries with the `GOPATH` environment variable specified and `$GOPATH/bin` in your `PATH`.
- [NodeJS](http://nodejs.org/) JavaScript runtime.
- [Docker](https://www.docker.com/) platform.
- [Docker Compose](https://docs.docker.com/compose/) (optional) for managing multi-container dev environments.
- [GDAL](https://gdal.org/) v2.4.2 or better for geospatial data access. Available as a package for most Linux distributions, and OSX through Homebrew.

## Development

#### Clone the repository:

```bash
mkdir -p $GOPATH/src/github.com/uncharted-distil
cd $GOPATH/src/github.com/uncharted-distil
git clone git@github.com:unchartedsoftware/distil.git
cd distil
```

#### Install dependencies:

```bash
make install
```

#### Install datasets:

Datasets are stored using git LFS and can be pulled using the `datasets.sh` script.

```bash
./datasets.sh
```

To add / remove a dataset modify the `$datasets` variable:

```bash
declare -a datasets=("185_baseball" "LL0_acled" "22_handgeometry")
```

#### Generate code (optional):

To regenerate the PANDAS dataframe parser if the `api/compute/result/complex_field.peg` file is changed, run:

```bash
make peg
```

#### Docker images:

The application requires:

- ElasticSearch
- PostgreSQL
- TA2 Pipeline Server Stub

Docker images for each are available at the following registry:

```
docker.uncharted.software
```

##### Login to Docker Registry:

```bash
sudo docker login docker.uncharted.software
```

##### Update `docker-compose.yml`

```yaml
---
distil-auto-ml:
  image: docker.uncharted.software/distil-auto-ml
```

#### Pull Images:

Pull docker images via [Docker Compose](https://docs.docker.com/compose/):

```bash
./update_services.sh
```

#### Running the app:

Using three separate terminals:

##### Terminal 1 - Launch docker containers via [Docker Compose](https://docs.docker.com/compose/):

```bash
./run_services.sh
```

##### Terminal 2 - Build and watch webapp:

```bash
yarn watch
```

The app will be accessible at `localhost:8080`.

##### Terminal 3 - Build, watch, and run server:

```bash
make watch
```

#### Advanced Configuration

The location of the dataset directory can be changed by setting the `D3MINPUTDIR` environment variable, and the location of the temporary data written out during model building can be set using the `D3MOUTPUTDIR` environment variable.
The host IP address of the docker containers if not _localhost_ can be set with `DOCKER_HOST`. (i.e.`export DOCKER_HOST=192.168.0.10 && make watch`.)
These are used by the other Distil services that are launched via the `run_services.sh` script, and are typically set as global environment variables in `.bashrc` or similar.

### Linter Setup

#### VSCODE

For the VsCode editor download and install the eslint extension.
Once installed go to the editor settings (hot key ⌘⇧p -- type settings)
Add the following to your settings file:

```json
  "eslint.lintTask.enable": true, // enable eslint to run
  "eslint.validate": [
    "vue", // tell eslint to read vue files
    "html", // tell eslint to read html files
    "javascript", // tell eslint to read javascript files
    "typescript" // tell eslint to read typescript files
  ],
  "eslint.workingDirectories": [{ "mode": "auto" }], // eslint will try its best to figure out the working directory of the project
```

At this point save your settings file and restart VsCode.
If upon restarting and the linter is not working check the output (^⇧` -- OUTPUT tab -- dropdown -- ESlint)

## Common Issues:

#### "../repo/subpackage/file.go:10:2: cannot find package "github.com/company/package/subpackage" in any of":

- **Cause**: Dependencies are out of date or have not been installed
- **Solution**: Run `make install` to install latest dependencies.

#### "# pkg-config --cflags -- gdal gdal gdal gdal gdal gdal Package gdal was not found in the pkg-config search path."

- **Cause**: GDAL has not been installed
- **Solution**: Install GDAL using a package for your environment or download and build from source.

### Mac

#### runtime error while training "joblib.externals.loky.process_executor.TerminatedWorkerError: A worker process managed by the executor was unexpectedly terminated. This could be caused by a segmentation fault while calling the function or by an excessive memory usage causing the Operating System to kill the worker."

- **Cause**: Not enough Docker resources
- **Solution**: change Docker resources to recommended "CPU:10, RAM:10 gigs, Swap:2.5 gigs, Disk Image Size: 64 gigs"
