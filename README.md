# pgrest

A simple RESTful service written in Go to proxy queries to PostgreSQL servers that are not connected to the internet.  

![GitHub License](https://img.shields.io/github/license/sogelink-research/pgrest) 
![GitHub Release](https://img.shields.io/github/v/release/sogelink-research/pgrest)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/sogelink-research/pgrest/docker-publish.yml) 
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/sogelink-research/pgrest?filename=.%2Fsrc%2Fgo.mod)
![GitHub repo size](https://img.shields.io/github/repo-size/sogelink-research/pgrest)

## Features

- Supports multiple postgres sources to query
- Hash-Based Message Authentication
- Brotli and GZIP compression support
- Streaming response to keep memory footprint low
- Server binary size < 10MB, Docker image only 16MB
- Output formats
  - JSON
  - JSONDataArray
  - CSV
  - Apache Arrow (Experimental)
  - Parquet (Experimental)

## Security notice

You're opening up a way to directly query the database so make sure you handle things correctly.

- Make sure the user set in the connection string has the appropriate operation and access rights for configured clients.
- Ensure to keep your `clientSecret` confidential to prevent unauthorized access to a database.

## Usage

### Local development

Download dependencies and start PGRest.

```sh
cd src
go mod download
go run ./cmd/app/main.go
```

### Docker

Example running PGRest using docker and mount/use a config file.

```sh
docker run --network host -v ./config/pgrest.conf:/config/pgrest.conf -e PGREST_CONFIG_PATH="/config/pgrest.conf" ghcr.io/sogelink-research/pgrest:latest
```

### Docker compose

To start PGRest with a database with some mock data and readonly user run the following command.

```sh
docker compose up --build
```

### Examples

Under `./examples` some examples on how to use PGRest can be found for `curl`, `node` and `javascript`.

### PGRest JS client

A basic Javascript Client can be installed from [NPM](https://www.npmjs.com/package/@sogelink-research/pgrest-client) or loaded from jsDelivr, information on usage can be found in the [pgrest-client readme](https://github.com/sogelink-research/pgrest/tree/main/clients/js)

## Endpoints

### Query

Run a query on a connection trough PGRest.

**(POST) /api/{connection}/query**

```json
{
    "query": "SELECT station_id, temperature, humidity, wind_speed FROM weather_station_measurement WHERE station_id = 1",
    "format": "json"
}
```

|property|description|default|
|-|-|-|
|query|The query to run|-|
|format|The response format, one of these options ['json', 'jsonDataArray', 'csv', 'arrow', 'parquet']|json|

#### Authorization

Authorization on the server side utilizes a custom authentication scheme based on the Authorization header with a Bearer token. The token is structured as a base64-encoded string clientId.token, where the token is a SHA-256 HMAC (encoded in base64) generated from the POST body + UNIX timestamp using the clientSecret as the key. When a connection is configured with `"auth": "public"` authorization is skipped, use with cause!.

```
Authorization: Bearer <base64(clientId.token)>
X-Request-Time: <UNIX Timestamp (seconds)>
```

See `examples/curl_example.sh` for an example how to request using curl.

### Status

Check the status of the server, can be used as health check.

**(GET) /api/status**

Returns 200/ok with JSON content

```json
{
  "status": "ok",
  "started": "2024-07-18 14:56:24.474895508 +0000 UTC",
  "uptime": "1d 02h 12m 31s"
}
```

# PGRest Configuration Guide

This document provides an overview of the configuration settings for PGRest as defined in the `./config/pgrest.conf` file. PGRest tries to load the config file from `../config/pgrest.conf` by default and `/root/config/pgrest.conf` for docker. The path to the config file can be set using the environment variable `PGREST_CONFIG_PATH`

## Example

```json
{
  "pgrest": {
    "port": 8080,
    "debug": true,
    "cors": {
      "allowOrigins": ["*"],
      "allowHeaders": ["*"],
      "allowMethods": ["POST", "OPTIONS"]
    },
    "maxConcurrentRequests": 15,
    "timeoudtimeoutSecondsS": 30
  },
  "connections": [
    {
      "name": "default",
      "connectionString": "postgres://readonly_user:readonly_password@pgrest-test-db:5432/postgres",
      "auth": "private",
    },
    ...
  ],
  "users": [
    {
      "clientId": "pgrest",
      "clientSecret": "98265691-8b9e-44dc-acf9-94610c392c00",
      "connections": [
        "default"
      ]
    },
    ...
  ]
}
```

## Configuration Overview

The configuration for PGRest is structured into two main sections: `pgrest` and `connections`.

### PGRest Settings

- **port**: The port on which PGRest will listen for incoming requests. Defaults to `8080`.
- **debug**: This flag controls the log level, if set to false log level defaults to `info`. Default false.
- **cors**: Cross-Origin Resource Sharing settings.
  - **allowOrigins**: Specifies the origins that are allowed to access. Default ["*"]
  - **allowHeaders**: Specifies the allowed headers. Default ["*"]
  - **allowMethods**: Specifies the allowed methods. Default ["OPTIONS", "POST"]
- **maxConcurrentRequests**: Limits number of currently processed requests at a time across all users. Default 15.
- **timeoudtimeoutSecondsS**: The amount of seconds before a request times out.

### Connections

This section defines the database connections that PGRest can use.

#### Connection

- **name**: Identifier for the connection.
- **connectionString**: The connection string used to connect to the PostgreSQL database.
- **auth**: (Do not use!, leave out or set to private) Can be set to public to ignore Authorization: Authorization header/user access will not be checked. Default private.

### Users

Defines the list of users.

- **clientId**: Identifier for the client.
- **clientSecret**: A secret key for the client, will not be send between client/server.
- **connections**: An array of connection names where a user has access to.
