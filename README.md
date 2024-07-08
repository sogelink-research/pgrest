# pgrest

A simple RESTful service written in Go to proxy queries to PostgreSQL servers that are not connected to the internet.  

## Features

- Supports multiple postgres sources to query
- Brotli and GZIP compression support
- Streaming response to keep memory footprint low
- DataArray output option to lower transfered bytes and increase speed for lots of rows
- Server binary size < 10MB, Docker image only 16MB

## ToDo

- HMAC
- Ratelimit user

## Security

You're opening up a way to directly query the database so make sure the user set in the connection string has the appropriate operation and access rights.

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

To authorize requests with a Bearer token in the Authorization header, you need to encode the `clientId:apiKey` pair in base64. The resulting string should be included in the Authorization header as follows:

```
Authorization: Bearer <base64(clientId:apiKey)>
```

Replace `<base64(clientId:apiKey)>` with the actual base64-encoded value of `clientId:apiKey`.

For example, if your `clientId` is "pgrest" and your `apiKey` is "myapikey", the Authorization header would look like this:

```
Authorization: Bearer cGdyZXN0Om15YXBraWQiOm15YXBraWQ=
```

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
      "users": [
        {
          "clientId": "pgrest",
          "clientSecret": "mysecret",
          "apiKey": "myapikey",
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

#### Users

Defines the users who can access this connection.

- **Client ID**: Identifier for the client.
- **Client Secret**: A secret key for the client, not send in requests and used for HMAC.
- **API Key**: An API key for additional security.
- **CORS**: Cross-Origin Resource Sharing settings.
  - **Allow Origins**: Specifies the origins that are allowed to access the resource per user.

## Security Notice

Ensure to keep your `clientSecret` and `API Key` confidential to prevent unauthorized access to your database.

For production environments, it is recommended to restrict the `allowOrigins` in the CORS settings to only the domains that need access to the API.