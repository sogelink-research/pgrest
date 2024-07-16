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
- DataArray output option to lower transfered bytes and increase speed for lots of rows
- Server binary size < 10MB, Docker image only 16MB

## Security notice

You're opening up a way to directly query the database so make sure you handle things correctly.

- Make sure the user set in the connection string has the appropriate operation and access rights for configured clients.
- Ensure to keep your `clientSecret` confidential to prevent unauthorized access to a database.

## Usage

### Local development

Download dependencies and start PGRest

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

### JS client

To make it easier to use PGRest a JS Client can be found under `./examples/pgrest_js_client/pgrest_client.mjs` This is a simple client that helps creating and setting the Authorization token.

```js
import { PGRestClient } from './pgrest_js_client/pgrest_client.mjs';

const client = new PGRestClient(
  "http://localhost:8080", 
  "pgrest",
  "98265691-8b9e-44dc-acf9-94610c392c00", 
  "default"
);

const result = await client.query("SELECT entity_id, date_timestamp, temperature, humidity, wind_direction, precipitation FROM weather WHERE entity_id = 2 ORDER BY date_timestamp desc limit 10");
```

## Query

Send a post request to pgrest

### (POST) /api/{connection}/query

```json
{
    "query": "SELECT station_id, temperature, humidity, wind_speed FROM weather_station_measurement WHERE station_id = 1",
    "format": "default"
}
```

|property|description|default|
|-|-|-|
|query|The query to run|-|
|format|The response format ['default', 'dataArray']|default|

### Authorization

Authorization on the server side utilizes a custom authentication scheme based on the Authorization header with a Bearer token. The token is structured as a base64-encoded string clientId.token, where the token is a SHA-256 HMAC (encoded in base64) generated from the POST body using the clientSecret as the key. When a connection is configured with `"auth": "public"` authorization is skipped, use with cause!.

```
Authorization: Bearer <base64(clientId.token)>
```

See `examples/curl_example.sh` for an example how to request using curl.

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
    "maxConcurrentRequests": 15
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
- **debug**: This flag controls the log level, if set to false log level defaults to `info`. Defaults to false.
- **cors**: Cross-Origin Resource Sharing settings.
  - **allowOrigins**: Specifies the origins that are allowed to access. Default ["*"]
  - **allowHeaders**: Specifies the allowed headers. Default ["*"]
  - **allowMethods**: Specifies the allowed methods. Default ["OPTIONS", "POST"]
- **maxConcurrentRequests**: Limits number of currently processed requests at a time across all users. Defatul 15.

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
