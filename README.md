# pgrest

A server written in GO that provides secure and efficient querying of PostgreSQL databases via a RESTful API. 

## ToDo

- HMAC

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
|format|The response format: default or dataArray|default|

### Authorization

## Config

pgrest can be configured using a config file.