# pgrest

A simple RESTful service written in Go to proxy queries to PostgreSQL servers that are not connected to the internet.  

## Features

- Supports multiple postgres sources to query
- Hash-Based Message Authentication
- Brotli and GZIP compression support
- Streaming response to keep memory footprint low
- DataArray output option to lower transfered bytes and increase speed for lots of rows
- Server binary size < 10MB, Docker image only 16MB

## Future?

- Ratelimit user
- IP-Whitelisting

## Security notice

You're opening up a way to directly query the database so make sure you handle things correctly.

- Make sure the user set in the connection string has the appropriate operation and access rights for configured clients.
- Ensure to keep your `clientSecret` confidential to prevent unauthorized access to a database.


## Docker

Example running PGRest using docker and mount/use a config file.

```sh
docker run --network host -v ./my.conf:/config/my.conf -e PGREST_CONFIG_PATH="/config/my.conf" ghcr.io/sogelink-research/pgrest:latest
```

## Query

Send a post request to pgrest

### Post

```json
{
    "database": "default",
    "query": "SELECT station_id, temperature, humidity, wind_speed FROM weather_station_measurement WHERE station_id = 1",
    "format": "default"
}
```

|property|description|default|
|-|-|-|
|database|the name of the database, configured in the config|default|
|query|The query to run|-|
|format|The response format ['default', 'dataArray']|default|

### Authorization

Authorization on the server side utilizes a custom authentication scheme based on the Authorization header with a Bearer token. The token is structured as a base64-encoded string clientId.token, where the token is a SHA-256 HMAC (encoded in base64) generated from the POST body using the clientSecret as the key. When a connection is configured with `"auth": "public"` authorization is skipped, use with cause!.

```
Authorization: Bearer <base64(clientId.token)>
```

See `examples/curl_example.sh` for an example how to request using curl.

# PGRest Configuration Guide

This document provides an overview of the configuration settings for PGRest as defined in the `./config/pgrest.conf` file. PGRest tries to load the config file from `./config/pgrest.conf` by default. The path to the config file can be set using the environment variable `PGREST_CONFIG_PATH`

## Example

```json
{
  "pgrest": {
    "port": 8080,
    "debug": true
  },
  "connections": [
    {
      "name": "default",
      "connectionString": "postgres://user:password@localhost:5432/database",
      "auth": "private",
      "users": [
        {
          "clientId": "pgrest",
          "clientSecret": "mysecret",
          "cors": {
            "allowOrigins": ["*"]
          }
        },
        ...
      ]
    },
    ...
  ]
}
```

## Configuration Overview

The configuration for PGRest is structured into two main sections: `pgrest` and `connections`.

### PGRest Settings

- **Port**: The port on which PGRest will listen for incoming requests. Defaults to `8080`.
- **Debug**: This flag controls the log level, if set to false log level defaults to `info`. Defaults to false.

### Connections

This section defines the database connections that PGRest can use.

#### Connection

- **Name**: Identifier for the connection.
- **Connection String**: The connection string used to connect to the PostgreSQL database.
- **Auth**: (Do not use!, leave out or set to private) Can be set to public to ignore Authorization: Authorization header/user access will not be checked. Default private.

#### Users

Defines the users who can access this connection.

- **Client ID**: Identifier for the client.
- **Client Secret**: A secret key for the client, will not be send between client/server.
- **CORS**: Cross-Origin Resource Sharing settings.
  - **Allow Origins**: Specifies the origins that are allowed to access the resource per user.
