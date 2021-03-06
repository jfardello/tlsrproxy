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
$ docker run -it --rm -e PROXY_UPSTREAM=https://httpbin.org -p8888:8888 quay.io/jfardello/tlsrproxy:latest
INFO[0000] Running HTTP server on :8888                 
INFO[0000] Forwarding to upstream on https://httpbin.org 
Warning! Serving clear text http!
```
## On another terminal.. 

(it should change "http://" for "https://")

```
$curl -i  -H 'X-foo: bart' http://localhost:8888/anything/barbar/barbar
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: *
Content-Length: 424
Content-Type: application/json
Date: Mon, 28 Dec 2020 20:57:13 GMT
Server: gunicorn/19.9.0

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
    "X-Amzn-Trace-Id": "Root=1-5fea46a9-5bc577145bbcc9195c73730d", 
    "X-Foo": "bert"
  }, 
  "json": null, 
  "method": "GET", 
  "origin": "127.0.0.1, 88.53.64.95", 
  "url": "https://httpbin.org/anything/pepepeo/pepepeo"
}

```
## Customizing the default config file 

It should be overrided in a volume, to get the file from the image:

```
mkdir config
docker run -it --rm --entrypoint cat quai.io/jfardello/tlsrproxy /config/tlsrproxy.yaml > /config/tlsrproxy.yaml
```
After editing `config/tlsrproxy.yaml` you can launch the same docker command with a volume (or attach a permanent one)
```
docker run -it --rm -v ./config:/config -e PROXY_UPSTREAM=https://httpbin.org -p8888:8888 quay.io/jfardello/tlsrproxy:latest
```

## Default config file

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
