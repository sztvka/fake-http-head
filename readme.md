# Fake HTTP HEAD
This tool is used for scraping response headers on pages that have HTTP HEAD requests disabled. The package tries to establish a connection to the server, reads the first response chunk and drops the connection. This way, the headers are read without downloading the whole page. It does not work well on pages below 50kB, as such a small payload might be received at once.

## Usage
This tool is meant to be piped into an existing scraping software using `os.exec()` or similar methods, hence it takes a single json arg and outputs a single json response to stdout. To use the script provide a following minified JSON: 
```json
{
  "url": "https://google.com",
  "proxy_url": "http://localhost:8888",
  "timeout_ms": 10000,
  "headers": {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"
  }
}
```

```shell
$ ./fake_http_head '{"url":"https://google.com","proxy_url":"http://localhost:8888","timeout_ms":10000,"headers":{"User-Agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"}}'
```

The script will output a JSON response with the following structure:
```json
{
  "http_status_code": 200,
  "status": "ok",
  "headers": {
    "Content-Type": "text/html; charset=UTF-8",
    "Date": "Thu, 01 Jan 1970 00:00:00 GMT",
    "Expires": "-1",
    "Cache-Control": "private, max-age=0",
    "Content-Encoding": "gzip",
    "Server": "gws",
    "X-XSS-Protection": "0",
    "X-Frame-Options": "SAMEORIGIN",
    "Transfer-Encoding": "chunked"
  }
}
```

Or if errored:
```json
{
  "status": "error"
}
```
```json
{
  "status": "timeout"
}
```

## Installation
Download the latest release from the releases page or compile from scratch, then just use in your shell.

## TODO
- Add wrappers for nodejs and python.