# distil

[![CircleCI](https://circleci.com/gh/uncharted-distil/distil/tree/master.svg?style=svg&&circle-token=ff61c235865dd699cc8b923035a80e6e8d39c63a)](https://circleci.com/gh/unchartedsoftware/distil/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/uncharted-distil/distil)](https://goreportcard.com/report/github.com/uncharted-distil/distil)
[![GolangCI](https://golangci.com/badges/github.com/uncharted-distil/distil.svg)](https://golangci.com/r/github.com/uncharted-distil/distil)

## Dependencies

- [Git](https://git-scm.com) and [Git LFS](https://git-lfs.github.com) Versioning softwares.
- [Go](https://golang.org/) programming language binaries with the `GOPATH` environment variable specified and `$GOPATH/bin` in your `PATH`.
- [NodeJS](http://nodejs.org/) JavaScript runtime.
- [Docker](https://www.docker.com/) platform.
- [Docker Compose](https://docs.docker.com/compose/) (optional) for managing multi-container dev environments.
- [GDAL](https://gdal.org/) v2.4.2 or better for geospatial data access.  Available as a package for most Linux distributions, and  OSX through Homebrew.

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
docker-compose pull
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

The location of the dataset directory can be changed by setting the `D3MINPUTDIR` environment variable, and the location of the temporary data written out during model building can be set using the `D3MOUTPUTDIR` environment variable. These are used by the other Distil services that are launched via the `run_services.sh` script, and are typically set as global environment variables in `.bashrc` or similar.

## Common Issues:

#### "dep: command not found":

- **Cause**: `$GOPATH/bin` has not been added to your `$PATH`.
- **Solution**: Add `export PATH=$PATH:$GOPATH/bin` to your `.bash_profile` or `.bashrc`.

#### "../repo/subpackage/file.go:10:2: cannot find package "github.com/company/package/subpackage" in any of":

- **Cause**: Dependencies are out of date or have not been installed
- **Solution**: Run `make install` to install latest dependencies.

#### "# pkg-config --cflags  -- gdal gdal gdal gdal gdal gdal Package gdal was not found in the pkg-config search path."

- **Cause**: GDAL has not been installed
- **Solution**: Install GDAL using a package for your environment or download and build from source.
