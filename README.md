# TLSrproxy
A body-rewriting (hijacking) golang reverse proxy.


TLSrproxy is a proxy server that replaces strings in request and responses, it serves as a sandbox 
for debugging mixed-content errors, it uses a yaml config file which can be overrided by environment
variaboles.




## Environment Variables for Configuration

* **SERVER_HTTP_ADDR:** The host and port. Default: `":8888"`

* **SERVER_CERT:** Path to cert file. Default: `""`

* **CERT_KEY:** Path to key file. Default: `""`

* **SERVER_DRAIN:** How long application will wait to drain old requests before restarting. Default: `"1s"`

* **PROXY_UPSTREAM** Forward incoming requests to this host.

## Example:

Setting up a proxy to httpbin.org and post a json.

```
$ docker run -it --rm -e PROXY_UPSTREAM=https://httpbin.org quai.io/jfardello/tlsrproxy
```
## On another terminal..
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

## Config file

```yaml
server:
  cert: ""
  key: ""
  drain: 1s
proxy:
  upstream: http://localhost:9090
  #Replaces body strings we get from upstream.
  replaces:
    request: #Replaces the headers we send.
      headers: 
        - - bart
          - bert
    response: #Replaces the content in the response.
      mimes: #Only deal with this types (Dont break downloads or othe content.)
      - text/html
      - text/css
      - application/javascript
      - application/json
      - application/xhtml+xml
      body:
        - - http://
          - https://
        - - barbar
          - pepepeo
      headers:
        - - header1
          - header2
        - - headerA
          - headerB

```

## Monitoring endpoint

## /_health/status