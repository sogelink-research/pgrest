# pgrest

A server written in GO that provides secure and efficient querying of PostgreSQL databases via a RESTful API.

## Features

- Supports multiple postgres sources to query
- GZIP compression support
- Streaming response to keep memory footprint low
- DataArray output option to lower transfered bytes and increase speed for lots of rows
- Server binary size < 10MB, Docker image only 16MB

## ToDo

- HMAC
- Brotli compression
- Ratelimit user

## Security

You're opening up a way to directly query the database so make sure the user set in the connection string has the appropriate operation and access rights.

## Query

Send a post request to pgrest

### Post

```json
{
    "database": "default",
    "query": "SELECT * from test",
    "format": "default"
}
```

|property|description|default|
|-|-|-|
|database|the name of the database, configured in the config|default|
|query|The query to run|-|
|format|The response format ['default', 'dataArray']|default|

### Authorization

## Config

pgrest can be configured using a config file.