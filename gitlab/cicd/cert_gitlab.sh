#!/bin/bash
mkdir -p ./config/ssl
openssl req -x509 -newkey rsa:2048 -sha256 -days 365 -nodes \
  -keyout ./config/ssl/gitlab.farengier.local.key \
  -out   ./config/ssl/gitlab.farengier.local.crt \
  -subj "/CN=gitlab.farengier.local" \
  -addext "subjectAltName=DNS:gitlab.farengier.local"
chmod 600 ./config/ssl/gitlab.farengier.local.key
chmod 644 ./config/ssl/gitlab.farengier.local.crt