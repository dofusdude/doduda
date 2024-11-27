<p align="center">
  <img src="https://docs.dofusdu.de/logo_cropped.png" width="120">
  <h3 align="center">doduda</h3>
  <p align="center">The Open Ankama Launcher Terminal Client for Developers</p>
  <p align="center"><a href="https://goreportcard.com/report/github.com/dofusdude/doduda"><img src="https://goreportcard.com/badge/github.com/dofusdude/doduda" alt=""></a> <a href="https://github.com/dofusdude/doduda/actions/workflows/tests.yml"><img src="https://github.com/dofusdude/doduda/actions/workflows/tests.yml/badge.svg" alt=""></a>
  </p>
</p>

Download the latest Dofus 3 version from Ankama and convert the interesting parts to a developer friendly format.

```bash
doduda --release beta && doduda map --release beta
```

## Features

See `doduda --help` for all parameters.

### Load

Download the latest Dofus version.

<img src="https://vhs.charm.sh/vhs-15sHEeT47mgiZ7vrgnenB2.gif" alt="load example" width="600">

The results are saved to `./data`.

### Map

Make the cryptic data easier to use in apps.

<img src="https://vhs.charm.sh/vhs-3YcvO6NALEaRFoNPu9Jhe2.gif" alt="map example" width="600">

The command assumes that all categories (except images) are already downloaded.

### Watchdog

Listen for new Dofus versions and react to their release.

<img src="https://vhs.charm.sh/vhs-g7BGgJ5f4iUhuzRhoYzzR.gif" alt="watchdog example" width="600">

You can use that for getting anything that supports webhooks to react to Dofus version updates. Some ideas are:

- [Discord Channels](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks)
- [ntfy.sh](https://ntfy.sh) (Push notifications to your phone)

### Render

Generate pre-rendered, high resolution `.png` images from the vector files.
This requires [Docker](https://docs.docker.com/engine/install/) to be installed and running.

<img src="https://vhs.charm.sh/vhs-H3yfHga5lcnzj9UhLyP5C.gif" alt="render example" width="600">

## Installation

`doduda` is a single binary that you can download and run without dependencies. There are precompiled versions for Linux, macOS and Windows.

### Precompiled binaries (recommended)

Get the latest `doduda` binary from the [release](https://github.com/dofusdude/doduda/releases) page.

### Go install (needs [Go](https://go.dev/doc/install) >= 1.21)

You need to have `$GOPATH/bin` in your `$PATH` for this to work, so `export PATH=$PATH:$(go env GOPATH)/bin` if you haven't already.

```bash
go install github.com/dofusdude/doduda@latest
```

### Build from source (needs [Go](https://go.dev/doc/install) >= 1.21)

```bash
git clone https://github.com/dofusdude/doduda
cd doduda
go build
```

## The dofusdude auto-update cycle

This tool is the first step in a pipeline that updates the data on [GitHub](https://github.com/dofusdude/dofus2-main) when a new Dofus version is released.

1. Two watchdogs (`doduda listen`) listen for new Dofus versions. One for main and one for beta. When something releases, the watchdog calls the GitHub API to start a workflow that uses `doduda` to download and parse the update to check for new elements and item_types. They hold global IDs for the API, so they must be consistent with each update.
2. At the end of the `doduda` workflow, it triggers the corresponding data repository to do a release, which then downloads the latest `doduda` binary (because it is a different workflow) and runs it to produce the final dataset. The data repository opens a release and uploads the files.
3. After a release, `doduapi` needs to know that a new update is available. The data repository workflow calls the update endpoint. The API then fetches the latest version from GitHub, indexes, starts the image upscaler (if necessary) and does a atomic switch of the database when ready.

## Known Problems

- Run `doduda` with `--headless` in a server environment or automations to avoid "no tty" errors.

- If you get an error regarding a missing Docker socket when running `doduda render`, find out where your `docker.sock` is and link it to the missing path or export your path as `DOCKER_HOST` environment variable `export DOCKER_HOST=unix://<your docker.sock path>`.

## Credit

The code in the `unpack` directory is a port of the [PyDofus](https://github.com/balciseri/PyDofus) project to Go. Thanks to balciseri for the work on PyDofus!

The terminal visualizations are made with [vhs](https://vhs.charm.sh).

Many thanks to Ankama for developing and updating Dofus! All data belongs to them. I just make it more accessible for the developer community.
