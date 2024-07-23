export class PGRestClient {
  #url;
  #clientID;
  #clientSecret;
  #connection;
  #outputFormats;
  #_importKey;

  /**
   * Represents a PGRest client.
   * @constructor
   * @param {string} url - The URL to the PGRest server.
   * @param {string} clientID - The client ID for authentication.
   * @param {string} clientSecret - The client secret for authentication.
   * @param {string} [connection="default"] - The name of the connection configured in PGRest.
   */
  constructor(url, clientID, clientSecret, connection = "default") {
    this.#url = url;
    this.#clientID = clientID;
    this.#clientSecret = clientSecret;
    this.#connection = connection;
    this.#outputFormats = this.#getOutputFormats();
  }

  /**
   * Executes a query against the server.
   *
   * @param {string} query - The query string to execute.
   * @param {object} options - The options for the query.
   * @param {string} [options.connection] - The connection to use for the query. Defaults to the client's connection.
   * @param {string} [options.format="json, jsonDataArray, csv, arrow, parquet"] - The format of the response. Defaults to "default". Options ["json", "jsonDataArray", "csv", "arrow", "parquet"].
   * @param {string} [options.encoding="gzip, br"] - The encoding to use for the response. Defaults to "gzip, br".
   * @param {function} [options.executionTimeFormatter] - A function to format the execution time. Defaults to the client's formatter.
   * @returns {Promise<object>} - A promise that resolves to the response from the server.
   * @throws {object} - If the response status is not 200, an error object is thrown.
   */
  async query(
    query,
    {
      connection = this.#connection,
      format = "json",
      encoding = "gzip, br",
      executionTimeFormatter = undefined,
    } = {}
  ) {
    if (!this.#outputFormats[format]) {
      throw new Error(`Invalid format: ${format}`);
    }

    const body = JSON.stringify({
      query: query,
      format: format,
    });
    const contentType = this.#outputFormats[format].contentType;
    const startTime = performance.now();
    const queryEndpoint = this.#getQueryEndpoint(connection);
    const response = await this.#post(
      queryEndpoint,
      contentType,
      encoding,
      body
    );
    const endTime = performance.now();
    const duration = endTime - startTime;

    if (!response.ok) {
      const error = await response.json();
      throw error;
    }

    return await this.#outputFormats[format].handler(
      response,
      duration,
      executionTimeFormatter
    );
  }

  async #handleJSONResponse(result, duration, executionTimeFormatter) {
    const jsonResponse = await result.json();
    jsonResponse.executionTime = executionTimeFormatter
      ? executionTimeFormatter(duration)
      : this.#formatExecutionTime(duration);

    return jsonResponse;
  }

  async #handleArrowResponse(result, duration, executionTimeFormatter) {
    const arrowData = await result.arrayBuffer();
    return arrowData;
  }

  async #handleCSVResponse(result, duration, executionTimeFormatter) {
    const csvData = await result.text();
    return csvData;
  }

  async #handleParquetResponse(result, duration, executionTimeFormatter) {
    const parquetData = await result.arrayBuffer();
    return parquetData;
  }

  #getOutputFormats() {
    const formats = {
      json: {
        contentType: "application/json",
        handler: async (result, duration, executionTimeFormatter) => { return await this.#handleJSONResponse(result, duration, executionTimeFormatter) },
      },
      jsonDataArray: {
        contentType: "application/json",
        handler: async (result, duration, executionTimeFormatter) => { return await this.#handleJSONResponse(result, duration, executionTimeFormatter) },
      },
      arrow: {
        contentType: "application/vnd.apache.arrow.stream",
        handler: async (result, duration, executionTimeFormatter) => { return await this.#handleArrowResponse(result, duration, executionTimeFormatter) },
      },
      parquet: {
        contentType: "application/octet-stream",
        handler: async (result, duration, executionTimeFormatter) => { return await this.#handleParquetResponse(result, duration, executionTimeFormatter) },
      },
      csv: {
        contentType: "text/csv",
        handler: async (result, duration, executionTimeFormatter) => { return await this.#handleCSVResponse(result, duration, executionTimeFormatter) },
      },
    };

    return formats;
  }

  /**
   * Sends a POST request to the specified query endpoint.
   *
   * @private
   * @param {string} queryEndpoint - The URL of the query endpoint.
   * @param {string} contentType - The content type of the request body.
   * @param {string} encoding - The encoding type of the request.
   * @param {string} body - The request body.
   * @returns {Promise<Response>} - A promise that resolves to the response of the POST request.
   */
  async #post(queryEndpoint, contentType, encoding, body) {
    let currentTime = Math.floor(Date.now() / 1000);
    const authToken = await this.#createAuthToken(body, currentTime);
    const response = await fetch(queryEndpoint, {
      method: "POST",
      headers: {
        "Content-Type": contentType,
        "Accept-Encoding": encoding,
        'X-Request-Time': currentTime,
        Authorization: `Bearer ${authToken}`,
      },
      body: body,
    });

    return response;
  }

  /**
   * Creates an authentication token using the provided body.
   *
   * @private
   * @param {Object} body - The body used to generate the token.
   * @param {Object} time - Unix timestamp in seconds.
   * @returns {string} The generated authentication token.
   */
  async #createAuthToken(body, time) {
    const key = await this.#importKey();
    const content = body ? body + time : time;
    const signature = await this.#signKey(key, content);
    const hmac = btoa(String.fromCharCode(...new Uint8Array(signature)));

    return btoa(`${this.#clientID}.${hmac}`);
  }

  /**
   * Imports the key used for HMAC signing.
   *
   * @private
   * @returns {Promise<CryptoKey>} A promise that resolves to the imported key.
   */
  async #importKey() {
    if (this.#_importKey) {
      return this.#_importKey;
    }

    const key = crypto.subtle.importKey(
      "raw",
      new TextEncoder().encode(this.#clientSecret),
      { name: "HMAC", hash: "SHA-256" },
      false,
      ["sign"]
    );

    this.#_importKey = key;
    return key;
  }

  /**
   * Signs the given body using the provided key.
   *
   * @private
   * @param {CryptoKey} key - The key used for signing.
   * @param {string} body - The body to be signed.
   * @returns {Promise<ArrayBuffer>} - A promise that resolves to the signed data as an ArrayBuffer.
   */
  async #signKey(key, body) {
    return crypto.subtle.sign("HMAC", key, new TextEncoder().encode(body));
  }

  /**
   * Returns the query endpoint URL for the specified connection.
   *
   * @private
   * @param {string} connection - The connection name.
   * @returns {string} The query endpoint URL.
   */
  #getQueryEndpoint(connection) {
    return `${this.#url}${
      this.#url.endsWith("/") ? "" : "/"
    }api/${connection}/query`;
  }

  /**
   * Formats the execution time duration.
   *
   * @private
   * @param {number} duration - The duration of the execution time in milliseconds.
   * @returns {string} The formatted execution time duration.
   */
  #formatExecutionTime(duration) {
    return duration < 1000
      ? `${Math.round(duration)} ms`
      : `${(duration / 1000).toFixed(2)} s`;
  }
}