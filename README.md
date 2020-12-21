# TLSrproxy
A body-rewriting (hijacking) golang reverse proxy.

## Environment Variables for Configuration

* **HTTP_ADDR:** The host and port. Default: `":8888"`

* **HTTP_CERT_FILE:** Path to cert file. Default: `""`

* **HTTP_KEY_FILE:** Path to key file. Default: `""`

* **HTTP_DRAIN_INTERVAL:** How long application will wait to drain old requests before restarting. Default: `"1s"`

* **UPSTREAM** Forward incoming requests to this host.

## Example:

Setting up a proxy to httpbin.org and post a json.

```
$ mkdir sample && cd sample
$ openssl req -subj '/CN=dsc.127.0.0.1.nip.io/O=dsc/C=ES' -new -newkey rsa:2048 -sha256 -days 365 \
   -nodes -x509 -keyout server.key -out server.crt

$ docker run -e -e UPSTREAM=https://httpbin.org -it --rm -p 8443:8443 quay.io/jfardello/tlsrproxy:latest

```
#On another terminal..
```
#it should change "http://" for "https://"
$ curl -H 'X-foo: http://pepe' http://localhost:8888/anything/foo
```

