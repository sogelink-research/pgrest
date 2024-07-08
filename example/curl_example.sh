#!/usr/bin/env bash

SERVER="http://localhost:8080"
CLIENTID="pgrest"
CLIENTSECRET="mysecret"

# JSON payload
read -r -d '' JSON_PAYLOAD << EOF
{
    "database": "default",
    "format": "default",
    "query": "SELECT * FROM datacore.knmi_historical_weather_measurement WHERE entity_id = 260 limit 100"
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