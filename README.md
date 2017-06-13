# distil

## Dependencies

- [Go](https://golang.org/) programming language binaries with the `GOPATH` environment variable specified and `$GOPATH/bin` in your `PATH`.
- [NodeJS](http://nodejs.org/) JavaScript runtime.

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

Run build the webapp and run the server:
```bash
make dev
```

## Common Issues:

#### "glide: command not found":

- **Cause**: `$GOPATH/bin` has not been added to your `$PATH`.
- **Solution**: Add `export PATH=$PATH:$GOPATH/bin` to your `.bash_profile` or `.bashrc`.

#### "../repo/subpackage/file.go:10:2: cannot find package "github.com/company/package/subpackage" in any of":

- **Cause**: Dependencies are out of date or have not been installed
- **Solution**: Run `make install` to install latest dependencies.
