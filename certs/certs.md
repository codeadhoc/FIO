## Generate Certificates 

1. Generate root certificate

```
openssl req -x509 -sha256 -nodes -days 365 \
-newkey rsa:2048 \
-subj '/O=Kubesure Inc./CN=ecomm.kubesure.io' \
-addext "subjectAltName=DNS:ecomm.kubesure.io,DNS:www.ecomm.kubesure.io" \
-keyout ca.key \
-out ca.crt

// verfiy 

openssl x509 --in ca.crt -text --noout | grep DNS

```

2. Generate Server cert and key from root cert

```
openssl req \
-out server.csr \
-newkey rsa:2048 -nodes \
-subj "/CN=ecomm.kubesure.io/O=ecomm kubesure" \
-reqexts SAN \
-config <(cat /etc/ssl/openssl.cnf \
<(printf "\n[SAN]\nsubjectAltName=DNS:ecomm.kubesure.io,DNS:www.ecomm.kubesure.io")) \
-keyout server.key 

//verify
openssl req -in server.csr -text -noout | grep DNS

openssl x509 -req \
-extfile <(printf "subjectAltName=DNS:ecomm.kubesure.io,DNS:www.ecomm.kubesure.io") \
-days 365 \
-in server.csr \
-CA ca.crt \
-CAkey ca.key \
-CAcreateserial \
-out server.crt

//verify

openssl x509 --in server.crt -text --noout | grep DNS
```

## client certs for MTLS

```
openssl req \
-out client.csr \
-newkey rsa:2048 -nodes \
-subj "/CN=ecomm.kubesure.io/O=ecomm kubesure" \
-reqexts SAN \
-config <(cat /etc/ssl/openssl.cnf \
<(printf "\n[SAN]\nsubjectAltName=DNS:ecomm.kubesure.io,DNS:www.ecomm.kubesure.io")) \
-keyout client.key

//verify
openssl req -in client.csr -text -noout | grep DNS

openssl x509 -req \
-extfile <(printf "subjectAltName=DNS:ecomm.kubesure.io,DNS:www.ecomm.kubesure.io") \
-days 365 \
-in client.csr \
-CA ca.crt \
-CAkey ca.key \
-CAcreateserial \
-out client.crt

//verify

openssl x509 --in client.crt -text --noout | grep DNS
```

