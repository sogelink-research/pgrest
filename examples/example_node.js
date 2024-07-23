import { PGRestClient } from './pgrest_js_client/pgrest_client.js';

const url = "http://localhost:8080/";
const clientID = "pgrest";
const clientSecret = "98265691-8b9e-44dc-acf9-94610c392c00";
const connection = "default"; // default can be omitted
const defaultQuery = "SELECT entity_id, date_timestamp, temperature, humidity, wind_direction, precipitation FROM weather WHERE entity_id = 2 ORDER BY date_timestamp desc limit 50";
const query = process.argv[2] ? process.argv[2] : defaultQuery;

const client = new PGRestClient(url, clientID, clientSecret, connection);

async function test() {
    try {
        const result = await client.query(query);
        logResult(result);
    } catch (error) {
        logError(error);
    }
}

function logResult(result) {
    console.table(result.data);
    console.log(`Execution time: ${result.executionTime}`);
}

function logError(error) {
    console.log("--------------------------------");
    console.log(`${error.status} - ${error.statusText}`);
    console.log("--------------------------------");
    console.log(`Message: ${error.error}`);
    if(error.details) {
        
        console.log(`Details: ${error.details}`);
    }
}

test();