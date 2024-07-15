#!/usr/bin/env bash

SERVER="http://localhost:8080/api/default/query"
CLIENTID="pgrest"
CLIENTSECRET="98265691-8b9e-44dc-acf9-94610c392c00"

# JSON payload
read -r -d '' JSON_PAYLOAD << EOF
{
    "format": "default",
    "query": "SELECT entity_id, date_timestamp, temperature, humidity, wind_direction, precipitation FROM weather WHERE entity_id = 2 ORDER BY date_timestamp desc limit 10"
}
EOF

# Function to calculate HMAC signature
calculate_hmac_sha256() {
    local message="$1"
    local secret="$2"
    echo -n "$message" | openssl dgst -sha256 -hmac "$secret" -binary | base64
}

# Create HMAC signature
HMAC=$(calculate_hmac_sha256 "$JSON_PAYLOAD" "$CLIENTSECRET")

# Create Authorization header
AUTH_HEADER=$(echo -n "${CLIENTID}.${HMAC}" | base64)

# Send request
time curl -X POST "$SERVER" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $AUTH_HEADER" \
-d "$JSON_PAYLOAD" \
--compressed