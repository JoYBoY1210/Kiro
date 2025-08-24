#!/bin/bash
CA_KEY="ca.key"
CA_CERT="ca.crt"
SERVICES_FILE="services.txt"

while read SERVICE; do
    echo "Generating certs for $SERVICE"
    openssl genrsa -out "${SERVICE}.key" 2048
    openssl req -new -key "${SERVICE}.key" -out "${SERVICE}.csr" -subj "/CN=${SERVICE}" -addext "subjectAltName=DNS:${SERVICE}"
    openssl x509 -req -in "${SERVICE}.csr" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial -out "${SERVICE}.crt" -days 365 -extfile <(printf "subjectAltName=DNS:${SERVICE}")
    echo "Created ${SERVICE}.key, ${SERVICE}.csr, ${SERVICE}.crt"
done < "$SERVICES_FILE"

echo " certificates generated"
