# Melo - Web API repository

This repository contains the Web server and clients used within **Melo** to download plugins,
updates and list all devices connected on local network.

## Getting started

The repository is composed of many sub projects:
* server: The **Melo Web API** server written in [Go](https://go.dev/) and using
[Huma](https://huma.rocks/) to generate `OpenAPI` specification from code,
* lib: The client library written in **C++** and using [curl](https://curl.se/) for HTTP requests.

The build is managed by [Bazel](https://bazel.build/).

### Server

To build and run the **server**, the following command can be used:

```sh
bazel run //server
```

The [Gazelle](https://github.com/bazelbuild/bazel-gazelle) is used to auto-generate and update the
**Bazel** rules related to **Go**. To update the rules and dependencies in `MODULE.bazel` and
`server/BUILD.bazel` files, please use:

```sh
bazel run //:gazelle && bazel mod tidy
```

Finally, a **Docker** / **OCI** image can be generated with the following command:

```sh
bazel build //server:image
```

### Client

To build the **C++ client** library, the following command can be used:

```sh
bazel build //lib
```

## Environment variables

Some environment variables can be setup the **MySQL** / **MariaDB** connection and HTTP handler:

| Variable                     | Description |
| :---:                        | ---         |
| `MELO_WEBAPI_MYSQL_HOSTNAME` | Host name of the MySQL / MariaDB server |
| `MELO_WEBAPI_MYSQL_USER`     | Username to use for MySQL / MariaDB server connection |
| `MELO_WEBAPI_MYSQL_PASSWORD` | Password to use for MySQL / MariaDB server connection |
| `MELO_WEBAPI_MYSQL_DATABASE` | Database to use in MySQL / MariaDB server |
| `MELO_WEBAPI_REAL_IP_HEADER` | HTTP header to read from the real IP address of the client |

## Local testing

This server is using a [MariaDB](https://mariadb.org/) database to store all the releases, plugins
and devices.

For local testing, a **Docker compose** file is provided to start a container providing a local
**MariaDB** server with a [Adminer](https://www.adminer.org/) interface on the port `8080`.

To start the server, please run the following command:

```sh
docker compose -f tools/local-db.yml up
```

Then the database can be accessed from [http://localhost:8080/](http://localhost:8080/) with the
following credentials:
* **Server:** `db`
* **Username:** `melo-webapi`
* **Password:** `password`

When the **Melo Web API** server is started with **Bazel**, the local **MariaDB** credentials are
used:

```sh
bazel run //:melo-webapi
```

## Formatting / Linting

Currently, the formatting and linting verification is done by the `//:check` target as a test:

```sh
bazel test //:check
```

To fix the formatting of every languages, please use `//:fix`:

```sh
bazel run //:fix
```

## Copyright / License

**Melo Web API** is licensed under the _GNU Affero General Public License, version 3_ license.
Please read [LICENSE](LICENSE) file or visit
[GNU AGPL 3.0 page](https://www.gnu.org/licenses/agpl-3.0.en.html) for further details.

Copyright @ 2024 Alexandre Dilly - Sparod
