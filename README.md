<img src="https://vikunja.io/images/vikunja-logo.svg" alt="" style="display: block;width: 40%;margin: 0 auto;" width="40%"/>

# vikunja-expanded

A personal fork of [Vikunja](https://vikunja.io) ([go-vikunja/vikunja](https://github.com/go-vikunja/vikunja)), the open-source task manager. This fork adds a few features on top of stock Vikunja for a homelab deployment; everything else is unmodified upstream Vikunja.

All credit for the underlying application goes to the Vikunja maintainers and contributors. This repository exists to track a small, deliberately-scoped set of additions on top of their work, not to replace or compete with it. If you don't need the extras below, use [upstream Vikunja](https://github.com/go-vikunja/vikunja) instead — it's actively maintained and this fork will always lag behind it.

## What this fork adds

- **Project web links** — attach reference links (git repo, prod environment, docs, etc.) to a project, each with a URL, description, and an icon (picked from [simple-icons](https://simple-icons.org/) or uploaded custom), shown in a new resizable sidebar next to the project description. Exposed via the API.
- **Task web links** — the same link concept, but scoped to a task, alongside the existing file-attachments feature.
- **Default task fields per project** — choose which fields (priority, due date, labels, ...) always show on a task in a given project, instead of adding them manually every time.
- **Labels field shows all labels on focus** — no need to type before seeing what labels exist.

Everything else — auth, CalDAV, migrations, mobile apps, etc. — is stock Vikunja. See [upstream's feature list](https://vikunja.io/features/) for the rest.

## Running it

This image is a drop-in replacement for `vikunja/vikunja`. Point your existing compose file's `vikunja` service at this image instead — nothing else about your stack (database, reverse proxy, env vars) needs to change:

```yaml
services:
  vikunja:
    image: warhoeks/vikunja-expanded:latest
    # ...rest of your existing vikunja service config, unchanged
```

All of stock Vikunja's [configuration options](https://vikunja.io/docs/config-options/) and [installation docs](https://vikunja.io/docs/installing/) apply unchanged.

## Staying in sync with upstream

This repo keeps two long-lived branches:

- `upstream-main` — a clean mirror of `go-vikunja/vikunja`'s `main`, never modified directly.
- `main` — `upstream-main` plus this fork's commits. This is what gets built and deployed.

See [MAINTENANCE.md](MAINTENANCE.md) for the sync procedure and the list of files this fork touches (the ones most likely to need conflict resolution on an upstream sync).

## Building

Same as upstream — see [Vikunja's build-from-source docs](https://vikunja.io/docs/build-from-sources/). The `Dockerfile` at the repo root is unmodified from upstream; CI builds it via [`.github/workflows/docker-publish.yml`](.github/workflows/docker-publish.yml) and pushes to [Docker Hub](https://hub.docker.com/r/warhoeks/vikunja-expanded).

## License

Same as upstream: most of this repository is [AGPL-3.0-or-later](LICENSE); [`desktop/`](desktop/) is GPL-3.0-or-later. This fork's source is public here, satisfying AGPL §13 for anyone it's served to over the network.
