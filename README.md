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
$ docker run -it --rm -e PROXY_UPSTREAM=https://httpbin.org quai.io/jfardello/tlsrproxy
```
#On another terminal..
```
#it should change "http://" for "https://"
$ $curl -H 'X-foo: http://pepe' http://localhost:8888/anything/foo
{
  "args": {}, 
  "data": "", 
  "files": {}, 
  "form": {}, 
  "headers": {
    "Accept": "*/*", 
    "Accept-Encoding": "gzip", 
    "Host": "httpbin.org", 
    "User-Agent": "curl/7.69.1", 
    "X-Amzn-Trace-Id": "Root=1-5fe07d1d-2fb61c05572f65ce45bdf472", 
    "X-Foo": "https://pepe"
  }, 
  "json": null, 
  "method": "GET", 
  "origin": "127.0.0.1, 88.24.169.196", 
  "url": "https://httpbin.org/anything/foo"
}

```

The default config should be overrided in a volume:

```
mkdir config
docker run -it --rm --entrypoint cat quai.io/jfardello/tlsrproxy /config/tlsrproxy.yaml > /conf/tlsrproxy.yaml
docker run -it --rm -v ./config:/config quai.io/jfardello/tlsrproxy
```