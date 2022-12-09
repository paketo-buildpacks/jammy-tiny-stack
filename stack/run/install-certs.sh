#!/bin/bash

set -eu -o pipefail

CERTS_DIR=/tiny/usr/share/ca-certificates
ETC_CERTS_DIR=/tiny/etc/ssl/certs
CERT_FILE=${ETC_CERTS_DIR}/ca-certificates.crt
mkdir -p $(dirname $ETC_CERTS_DIR)

CERTS=$(find "${CERTS_DIR}" -type f | sort)
for cert in ${CERTS}; do
  cat "${cert}" >> "${CERT_FILE}"
  cp "${cert}" "${ETC_CERTS_DIR}"
done

openssl rehash "${ETC_CERTS_DIR}"
rm -rf "${CERTS_DIR}"
