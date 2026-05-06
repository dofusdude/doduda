<p align="center">
  <img src="https://docs.dofusdu.de/dofus2/logo_cropped.png" width="120">
  <h3 align="center">doduda</h3>
  <p align="center">CLI for Dofus asset downloading and unpacking</p>
  <p align="center"><a href="https://goreportcard.com/report/github.com/dofusdude/doduda"><img src="https://goreportcard.com/badge/github.com/dofusdude/doduda" alt=""></a> <a href="https://github.com/dofusdude/doduda/actions/workflows/tests.yml"><img src="https://github.com/dofusdude/doduda/actions/workflows/tests.yml/badge.svg" alt=""></a>
  </p>
</p>

Download the latest Dofus 3 data from Ankama and convert the interesting parts to a developer friendly json format:

```sh
# Install
curl -s https://get.dofusdu.de/doduda | sh

# Run
doduda && doduda map
```

See `doduda --help` for all parameters.

> [!TIP]
> You don't need to run the tool yourself for most cases - just look at the automatically updated outputs of all Dofus 3 versions [here](https://github.com/dofusdude/dofus3-main/releases).

Dofus 3 unpacking just works without installing any additional software.

There may be cases though where you need [Docker](https://docs.docker.com/get-docker/) to be installed and running:

- want to force the legacy Dofus 3 Docker backend (`export DODUDA_UNITY_BACKEND=docker`) because of some missed bugs in the native unpacking backend.
- want to use the `render` command to generate images for Dofus 2.

If you use the Docker backend and have Docker socket problems, the solution is often to find your `docker.sock` path and link it to the missing path or export your path as `DOCKER_HOST` environment variable `export DOCKER_HOST=unix://<your docker.sock path>` before running `doduda`.

### GitHub Releases

Get the latest `doduda` binary from the [release](https://github.com/dofusdude/doduda/releases) page.

### Go install (needs [Go](https://go.dev/doc/install) >= 1.21)

You need to have `$GOPATH/bin` in your `$PATH` for this to work, so `export PATH=$PATH:$(go env GOPATH)/bin` if you haven't already.

```bash
go install github.com/dofusdude/doduda@latest
```

## Known Problems

- Run `doduda` with `--headless` in a server environment or automations to avoid "no tty" errors.

## Credit

To unpack Dofus 2 data, doduda ships a port of the [PyDofus](https://github.com/balciseri/PyDofus) project. Thanks to balciseri for the work on PyDofus!

The terminal visualizations are made with [vhs](https://vhs.charm.sh).

Many thanks to Ankama for developing and updating Dofus! All data belongs to them. I just make it more accessible for the developer community.
