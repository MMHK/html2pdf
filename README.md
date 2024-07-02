# HTML2PDF

[![Go Report Card](https://goreportcard.com/badge/github.com/MMHK/html2pdf)](https://goreportcard.com/report/github.com/MMHK/html2pdf)
[![Docker Pulls](https://img.shields.io/docker/pulls/mmhk/html2pdf)](https://hub.docker.com/r/mmhk/html2pdf)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Convert `HTML` to `PDF` files. This project is developed in `Golang` and provides an HTTP interface to output local or remote HTML pages as `PDF` file format.

The reason for using `HTML` as a template is because of HTML's excellent expressive ability. Additionally, `wkhtmltopdf` uses `webkit` as the rendering engine, allowing the use of modern `CSS` solutions to minimize template complexity.

## Features

Provides multiple interfaces for different use cases:
- `htmlpdf`: Render HTML source code into `PDF` file format.
- `linkpdf`: Render an online link into `PDF` file format.
- `combine`: Combine multiple PDF URLs into a single PDF file.
- `link/combine`: Combine multiple PDF/webpage URLs into a single PDF file.

## Compilation

1. Install Golang environment, Go >= 1.8.
2. Checkout the source code.
3. In the source directory, execute `go get -v` to check out all dependencies.
4. Execute `go build -o html2pdf .` to compile into a binary executable.
5. Run the executable `html2pdf -c ./config.json`.

## Dependencies

- [chromedp](https://github.com/chromedp/chromedp): A faster, simpler way to drive browsers supporting the Chrome DevTools Protocol in Go without external dependencies.

## Configuration

> Note: Since font rendering depends on `fontconfig`, and `fontconfig` configurations can vary across different environments, the following steps are recommended to address rendering differences across systems:
> 1. Use `embed font` to handle rendering fonts, including default fonts as much as possible.
> 2. Synchronize some common `fontconfig` configurations, usually located under `/etc/fonts/conf.d`. After modifications, execute `fc-cache -fv` to reset `fontconfig`.

## HTML Templates

Tests have shown that to fully cover an `A4` size `PDF` file, the dimensions `1240px x 1754px` should be used, but this is just a reference value. Different systems render web pages with `wkhtmltopdf` at different sizes, so the `--zoom` parameter needs to be controlled for pre-matching.

For various paper size specifications, refer to [here](http://www.papersizes.org/a-sizes-in-pixels.htm).

Refer to the project [html2pdf-template](https://github.com/MMHK/html2pdf-template) for generating HTML templates that support `HTML2PDF`.

## Docker

This project is packaged as a Docker image.

- Pull the Docker image:
  ```sh
  docker pull mmhk/html2pdf:chromedp
  ```

- Environment variables, refer to the `config.json` for details:
  - `WORKER`: Number of concurrent rendering processes, default is 4.
  - `LISTEN`: Service binding address and port, default is `0.0.0.0:4444`.
  - `WEB_ROOT`: Local directory for swagger-ui, can be set to empty to disable swagger-ui display, default is `/app/web_root`.
  - `TIMEOUT`: Timeout for each rendering process (seconds), default is 60.
  - `CHROME_PATH`: Path to the chrome-headless executable, default is `/headless-shell/headless-shell`.

- Run:
  ```sh
  docker run --name html2pdf -p 4444:4444 mmhk/html2pdf:chromedp
  ```

### Docker Compose

You can also use Docker Compose to manage the service.

- Create a `docker-compose.yml` file:
  ```yaml
  version: '3'
  services:
    html2pdf:
      image: mmhk/html2pdf:chromedp
      container_name: html2pdf
      ports:
        - "4444:4444"
      environment:
        WORKER: 4
        LISTEN: "0.0.0.0:4444"
        WEB_ROOT: "/app/web_root"
        TIMEOUT: 60
        CHROME_PATH: "/headless-shell/headless-shell"
  ```

- Start the service:
  ```sh
  docker-compose up -d
  ```

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.