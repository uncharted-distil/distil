# distil

[![CircleCI](https://circleci.com/gh/unchartedsoftware/distil/tree/master.svg?style=svg&&circle-token=ff61c235865dd699cc8b923035a80e6e8d39c63a)](https://circleci.com/gh/unchartedsoftware/distil/tree/master)

## Dependencies

- [Go](https://golang.org/) programming language binaries with the `GOPATH` environment variable specified and `$GOPATH/bin` in your `PATH`.
- [NodeJS](http://nodejs.org/) JavaScript runtime.
- [Docker](https://www.docker.com/) platform.
- [Docker Compose](https://docs.docker.com/compose/) (optional) for managing multi-container dev environments.

## Development

#### Clone the repository:

```bash
mkdir -p $GOPATH/src/github.com/unchartedsoftware
cd $GOPATH/src/github.com/unchartedsoftware
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
- TA2 Pipeline Runner
- D3M Resource Server

Docker images for each are available at the following registry:

```
docker.uncharted.software
```

##### Login to Docker Registry:

```bash
sudo docker login docker.uncharted.software
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
docker-compose up
```

##### Terminal 2 - Build and watch webapp:

```bash
yarn watch
```

##### Terminal 3 - Build, watch, and run server:
```bash
make watch
```

## Common Issues:

#### "dep: command not found":

- **Cause**: `$GOPATH/bin` has not been added to your `$PATH`.
- **Solution**: Add `export PATH=$PATH:$GOPATH/bin` to your `.bash_profile` or `.bashrc`.

#### "../repo/subpackage/file.go:10:2: cannot find package "github.com/company/package/subpackage" in any of":

- **Cause**: Dependencies are out of date or have not been installed
- **Solution**: Run `make install` to install latest dependencies.
