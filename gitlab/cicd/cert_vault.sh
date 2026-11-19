#!/bin/bash
mkdir -p ./certs
openssl req -x509 -newkey rsa:2048 -sha256 -days 365 -nodes \
  -keyout ./certs/vault.farengier.local.key \
  -out   ./certs/vault.farengier.local.crt \
  -subj "/CN=vault.farengier.local" \
  -addext "subjectAltName=DNS:vault.farengier.local"
chmod 600 ./certs/vault.farengier.local.key
chmod 644 ./certs/vault.farengier.local.crt
sudo chown 100:100 certs/vault.farengier.local.*
sudo cp ./certs/vault.farengier.local.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Копируем сертификат гитлаба
mkdir gitlabcerts
cp ../gitlab/config/ssl/gitlab.farengier.local.crt gitlabcerts/
