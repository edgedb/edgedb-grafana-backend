# Grafana Data Source Backend Plugin for EdgeDB

This plugin is currently in Alpha and is not ready for production use.

## What is Grafana Data Source Backend Plugin for EdgeDB?

This plugin allows connecting any EdgeDB server instance to Grafana and
visualize arbitrary EdgeQL queries.

For more information about backend plugins, refer to the documentation on [Grafana backend plugins](https://grafana.com/docs/grafana/latest/developers/plugins/backend/).

## Installation

Use Grafana 8.0 or higher is required for this plugin.  To install,
follow instruction below to build the plugin and then copy the
resulting `dist/` directory as `$GF_PATHS_PLUGINS/edgedb`.

This plugin is not signed yet, and Grafana will refuse to load it,
unless explicitly configured to do so via the
`GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=edgedb` environment
variable, or `allow_loading_unsigned_plugins = edgedb` in the
`[plugins]` section in `grafana.ini`.

## Building from source

A data source backend plugin consists of both frontend and backend components.

### Frontend

1. Install dependencies

   ```bash
   yarn install
   ```

2. Build plugin in development mode or run in watch mode

   ```bash
   yarn dev
   ```

   or

   ```bash
   yarn watch
   ```

3. Build plugin in production mode

   ```bash
   yarn build
   ```

### Backend

1. Update [Grafana plugin SDK for
Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/)
dependency to the latest minor version:

   ```bash
   go get -u github.com/grafana/grafana-plugin-sdk-go
   go mod tidy
   ```

2. Build backend plugin binaries for Linux, Windows and Darwin:

   ```bash
   mage -v
   ```

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```
## Authentication

### Local or self-hosted

For a local or self-hosted EdgeDB instance your grafana datasource
file will look like:

```
---
apiVersion: 1

datasources:
  - name: Metrics (edgedb.com)
    type: edgedb
    uid: edgedb_site_metrics
    access: proxy
    isDefault: true
    editable: false
    jsonData:
      host: <edgedb-hostname>
      port: <edgedb-port-number>
      user: <edgedb-metrics-user>
      database: <edgedb-metrics-database>
      tlsSecurity: <"strict"|"insecure"|"no_host_verification">
      tlsCA: <optional root cert if self-signed>
    secureJsonData:
      password: <edgedb-metrics-users-password>
```

### EdgeDB Cloud

The [edgedb-go](https://github.com/edgedb/edgedb-go) client library
authenticates [EdgeDB SAAS hosted Cloud](https://www.edgedb.com/cloud)
differntly than local or self-hosted.  The EdgeDB Cloud instances
authenticate with a combination of an instance name and a cloud secret
key.  These are obtained from the EdgeDB Cloud UI when you [create an
instance](https://cloud.edgedb.com/org/edgedb/create-instance).

The grafana datasource must indicate that the cloud authentication
settings are in use.  A sample [grafana datasource
file](https://grafana.com/docs/grafana/latest/datasources) is shown
below.

The EdgeDB connection uses TLS.  The server presents a certificate
signed by a widely trusted root so the `tlsCA` parameter should not be
needed.

```
---
apiVersion: 1

datasources:
  - name: <My Instance> (edgedb cloud)
    type: edgedb
    uid: edgedb_site_metrics
    access: proxy
    isDefault: true
    editable: false
    jsonData:
      deployment: edgedb-cloud
      cloudInstance: <my-edge-db-instance-name>
      database: <my-edgedb-schema-name>
      tlsSecurity: strict
    secureJsonData:
      cloudSecret: <my-instance-cloud-secret>
```


