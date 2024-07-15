export class PGRestClient {
  #url;
  #clientID;
  #clientSecret;
  #connection;
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
  }

  /**
   * Executes a query against the server.
   *
   * @param {string} query - The query string to execute.
   * @param {object} options - The options for the query.
   * @param {string} [options.connection] - The connection to use for the query. Defaults to the client's connection.
   * @param {string} [options.format="default"] - The format of the response. Defaults to "default". Options ["default", "dataArray"].
   * @param {string} [options.encoding="gzip, br"] - The encoding to use for the response. Defaults to "gzip, br".
   * @param {function} [options.executionTimeFormatter] - A function to format the execution time. Defaults to the client's formatter.
   * @returns {Promise<object>} - A promise that resolves to the response from the server.
   * @throws {object} - If the response status is not 200, an error object is thrown.
   */
  async query(
    query,
    {
      connection = this.#connection,
      format = "default",
      encoding = "gzip, br",
      executionTimeFormatter = undefined,
    } = {}
  ) {
    const body = JSON.stringify({
      query: query,
      format: format,
    });

    const authToken = await this.#createAuthToken(body);
    const startTime = performance.now();
    const queryEndpoint = this.#getQueryEndpoint(connection);
    const response = await fetch(queryEndpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Accept-Encoding": encoding,
        "Authorization": `Bearer ${authToken}`,
      },
      body: body,
    });
    const endTime = performance.now();
    const duration = endTime - startTime;

    const jsonResponse = await response.json();
    jsonResponse.executionTime = executionTimeFormatter
      ? executionTimeFormatter(duration)
      : this.#formatExecutionTime(duration);

    // if status is not 200, throw an error
    if (!response.ok) {
      throw jsonResponse;
    }

    return jsonResponse;
  }

  /**
   * Creates an authentication token using the provided body.
   *
   * @private
   * @param {Object} body - The body used to generate the token.
   * @returns {string} The generated authentication token.
   */
  async #createAuthToken(body) {
    const key = await this.#importKey();
    const signature = await this.#signKey(key, body);
    const hmac = btoa(String.fromCharCode(...new Uint8Array(signature)));

    return btoa(`${this.#clientID}.${hmac}`);
  }

  /**
   * Imports the key used for HMAC signing.
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

  #getQueryEndpoint(connection) {
    return `${this.#url}${this.#url.endsWith('/') ? '' : '/'}api/${connection}/query`;
  }

  /**
   * Formats the execution time duration.
   *
   * @param {number} duration - The duration of the execution time in milliseconds.
   * @returns {string} The formatted execution time duration.
   */
  #formatExecutionTime(duration) {
    return duration < 1000
      ? `${Math.round(duration)} ms`
      : `${(duration / 1000).toFixed(2)} s`;
  }
}