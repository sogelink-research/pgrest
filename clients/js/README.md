# @sogelink-research/pgrest-client

@sogelink-research/pgrest-client is a JavaScript client library for interacting with [PGRest](https://github.com/sogelink-research/pgrest) servers.

## Installation

### Using npm

To install the package via npm:

```sh
npm install @sogelink-research/pgrest-client
```

### Using jsDelivr

To use the package directly in the browser via jsDelivr, include the following script tag in your HTML file:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PGRestClient Example</title>
</head>
<body>
    <script type="module">
        import { PGRestClient } from 'https://cdn.jsdelivr.net/npm/@sogelink-research/pgrest-client@latest/dist/pgrest-client.esm.js';

        const client = new PGRestClient('https://example.com', 'your-client-id', 'your-client-secret');

        async function fetchData() {
            try {
                const response = await client.query('SELECT * FROM table');
                console.log(response);
            } catch (error) {
                console.error('Error:', error);
            }
        }

        fetchData();
    </script>
</body>
</html>
```

## Usage

### Constructor

Create an instance of PGRestClient with the following constructor:

```js
const client = new PGRestClient(url, clientID, clientSecret, [connection]);
```

- url: The URL to the PGRest server.
- clientID: The client ID for authentication.
- clientSecret: The client secret for authentication.
- connection (optional): The name of the connection to use. Defaults to "default".

### query Method

The query method executes a query against the server:

```js
client.query(query, [options])
```

#### Parameters

- query: A string representing the query to execute.
- options (optional): An object with the following properties:
  - connection: A string specifying the connection to use. Defaults to the client's connection.
  - format: The response format. Defaults to "json". Options include:
    - "json"
    - "jsonDataArray"
    - "csv"
    - "arrow"
    - "parquet"
  - encoding: The response encoding. Defaults to "gzip, br".
  - executionTimeFormatter: A function to format the execution time. Defaults to the client's default formatter.

#### Returns

A promise that resolves to the server's response in the specified format. Throws an error object when result is not `ok`.

#### Example usage

```js
import { PGRestClient } from '@sogelink-research/pgrest-client';

const client = new PGRestClient('https://example.com', 'your-client-id', 'your-client-secret');

function logError(error) {
    console.log("--------------------------------");
    console.log(`${error.status} - ${error.statusText}`);
    console.log("--------------------------------");
    console.log(`Message: ${error.error}`);
    if(error.details) {
        console.log(`Details: ${error.details}`);
    }
}

// Example 1: Execute a query in JSON format
async function fetchData() {
  try {
    const response = await client.query('SELECT * FROM table');
    console.table(result.data);
  } catch (error) {
    logError(error);
  }
}

fetchData();

// Example 2: Execute a query in CSV format
async function fetchCSVData() {
  try {
    const response = await client.query('SELECT * FROM table', { format: 'csv' });
    console.log(response);
  } catch (error) {
    logError(error);
  }
}

fetchCSVData();
```
