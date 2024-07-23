# @sogelink-research/pgrest-client

A JavaScript client for interacting with the [PGRest](https://github.com/sogelink-research/pgrest) service, which proxies queries to PostgreSQL servers. This client simplifies the process of creating and setting the Authorization token and sending queries to PGRest.


## Installation

To install the PGRest JS client, use npm:

```sh
npm install @sogelink-research/pgrest-client
```

## Usage

### Initialization

First, import the PGRestClient and create an instance of it:

```js
import { PGRestClient } from '@sogelink-research/pgrest-client';

const client = new PGRestClient(
  "http://localhost:8080", // Host
  "pgrest", // Client ID
  "98265691-8b9e-44dc-acf9-94610c392c00", // Client Secret
  "default" // Connection name, can be left out if name is default
);
```

### Sending queries

To send a query to the PGRest service, use the query method:

```js
const result = await client.query(
  "SELECT entity_id, date_timestamp, temperature, humidity, wind_direction, precipitation FROM weather WHERE entity_id = 2 ORDER BY date_timestamp desc limit 10"
);
console.log(result);
```

### Optional parameters

The query method accepts an optional second parameter for additional configurations:

```js
const result = await client.query("my awesome query", {
  connection: "default",
  format: "parquet",
  encoding: "gzip, br",
  executionTimeFormatter: (duration) => { return `${duration} ms` }
});
console.log(result);
```