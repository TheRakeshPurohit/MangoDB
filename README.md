# FerretDB

<!-- textlint-disable one-sentence-per-line -->

> [!TIP]
> Looking for FerretDB v1?
> It is [there](https://github.com/FerretDB/FerretDB/tree/main-v1).

<!-- textlint-enable one-sentence-per-line -->

[![Go Reference](https://pkg.go.dev/badge/github.com/FerretDB/FerretDB/v2/ferretdb.svg)](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/ferretdb)

[![Go](https://github.com/FerretDB/FerretDB/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/FerretDB/FerretDB/branch/main/graph/badge.svg?token=JZ56XFT3DM)](https://codecov.io/gh/FerretDB/FerretDB)

[![Security](https://github.com/FerretDB/FerretDB/actions/workflows/security.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/security.yml)
[![Packages](https://github.com/FerretDB/FerretDB/actions/workflows/packages.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/packages.yml)
[![Docs](https://github.com/FerretDB/FerretDB/actions/workflows/docs.yml/badge.svg?branch=main)](https://github.com/FerretDB/FerretDB/actions/workflows/docs.yml)

FerretDB is an open-source alternative to MongoDB.
It is a proxy that converts MongoDB 5.0+ wire protocol queries to SQL
and uses PostgreSQL with [DocumentDB extension](https://github.com/microsoft/documentdb) as a database engine.

```mermaid
flowchart LR
  A["Any application<br>Any MongoDB driver"]
  F{{FerretDB}}
  P[(PostgreSQL<br>DocumentDB extension)]

  A -- "MongoDB protocol<br>BSON" --> F
  F -- "PostgreSQL protocol<br>SQL" --> P
```

## Why do we need FerretDB?

MongoDB was originally an eye-opening technology for many of us developers,
empowering us to build applications faster than using relational databases.
In its early days, its ease-to-use and well-documented drivers made MongoDB one of the simplest database solutions available.
However, as time passed, MongoDB abandoned its open-source roots;
changing the license to [SSPL](https://www.ferretdb.com/sspl) - making it unusable for many open-source and early-stage commercial projects.

Most MongoDB users do not require any advanced features offered by MongoDB;
however, they need an easy-to-use open-source document database solution.
Recognizing this, FerretDB is here to fill that gap.

## Scope and current state

FerretDB is compatible with MongoDB drivers and popular MongoDB tools.
It functions as a drop-in replacement for MongoDB 5.0+ in many cases.
Features are constantly being added to further increase compatibility and performance.

We welcome all contributors.
See our [public roadmap](https://github.com/orgs/FerretDB/projects/2/views/1),
lists of [known differences and supported commands](https://docs.ferretdb.io/migration/compatibility/),
and [contributing guidelines](CONTRIBUTING.md).

## Quickstart

Run this command to start FerretDB with PostgreSQL, make sure to update `<username>` and `<password>`:

```sh
docker run -d --rm --name ferretdb -p 27017:27017 \
  -e POSTGRES_USER=<username> \
  -e POSTGRES_PASSWORD=<password> \
  ghcr.io/ferretdb/ferretdb-eval:2
```

This command will start a container with FerretDB, pre-packaged PostgreSQL with DocumentDB extension, and MongoDB Shell for quick testing and experiments.
However, it is unsuitable for production use cases because it keeps all data inside and loses it on shutdown.
See our [installation guides](https://docs.ferretdb.io/installation/) for instructions
that don't have those problems.

With that container running, you can:

- Connect to it with any MongoDB client application using the MongoDB URI `mongodb://<username>:<password>@127.0.0.1:27017/`.
- Connect to it using the MongoDB Shell by just running `mongosh`.
  If you don't have it installed locally, you can run `docker exec -it ferretdb mongosh`.
- For PostgreSQL, connect to it by running `docker exec -it ferretdb psql -U <username> postgres`.

You can stop the container with `docker stop ferretdb`.

We also provide binaries and packages for various Linux distributions.
as well as [Go library package](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/ferretdb)
that embeds FerretDB into your application.
See [our documentation](https://docs.ferretdb.io/installation/) for more details.

## Building and packaging

<!-- textlint-disable one-sentence-per-line -->

> [!NOTE]
> We advise users not to build FerretDB themselves.
> Instead, use binaries, Docker images, or packages provided by us.

<!-- textlint-enable one-sentence-per-line -->

FerretDB could be built as any other Go program,
but a few generated files and build tags could affect it.
See [there](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/build/version) for more details.

## Managed FerretDB at cloud providers

- [FerretDB Cloud](https://cloud.ferretdb.com/)
- [Civo](https://www.civo.com/marketplace/FerretDB)
- [Tembo](https://tembo.io/docs/tembo-stacks/mongo-alternative)
- [Elestio](https://elest.io/open-source/ferretdb)
- [Cozystack](https://cozystack.io/docs/components/#managed-ferretdb)

## Community

- Website: https://www.ferretdb.com/.
- Blog: https://blog.ferretdb.io/.
- Documentation: https://docs.ferretdb.io/.
- X (Twitter): [@ferret_db](https://x.com/ferret_db).
- Mastodon: [@ferretdb@techhub.social](https://techhub.social/@ferretdb).
- [Slack chat](https://slack.ferretdb.io/) for quick questions.
- [GitHub Discussions](https://github.com/FerretDB/FerretDB/discussions) for longer topics.
- [GitHub Issues](https://github.com/FerretDB/FerretDB/issues) for bugs and missing features.
- [Open Office Hours meeting](https://calendar.google.com/calendar/event?action=TEMPLATE&tmeid=NGhrZTA5dXZ0MzQzN2gyaGVtZmx2aWxmN2pfMjAyNDA0MDhUMTcwMDAwWiBjX24zN3RxdW9yZWlsOWIwMm0wNzQwMDA3MjQ0QGc&tmsrc=c_n37tquoreil9b02m0740007244%40group.calendar.google.com&scp=ALL)
  every Monday at 17:00 UTC at [Google Meet](https://meet.google.com/mcb-arhw-qbq).

If you want to contact FerretDB Inc., please use [this form](https://www.ferretdb.com/contact/).
