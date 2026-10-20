# Генерация самоподписанного сертификата
```bash
mkdir -p ./certs
openssl req -x509 -newkey rsa:2048 -sha256 -days 365 -nodes \
  -keyout ./certs/<domain>.key \
  -out   ./certs/<domain>.crt \
  -subj "/CN=<domain>" \
  -addext "subjectAltName=DNS:<domain>"
chmod 600 ./certs/<domain>.key
chmod 644 ./certs/<domain>.crt
sudo chown 100:100 certs/<domain>.*
```

# Добавление сертификата
```bash
sudo cp ./certs/<domain>.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```