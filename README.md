# HTML2PDF

[![Go Report Card](https://goreportcard.com/badge/github.com/MMHK/html2pdf)](https://goreportcard.com/report/github.com/MMHK/html2pdf)
[![Docker Pulls](https://img.shields.io/docker/pulls/mmhk/html2pdf)](https://hub.docker.com/r/mmhk/html2pdf)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

将 `HTML` 渲染成 `PDF` 文件，此项目用 `Golang` 开发，提供一个 HTTP 接口将本地或者远程 HTML 页面输出成为 `PDF` 文件格式。

使用 `HTML` 作为模板的原因是因为 HTML 的表达能力最好，而且 `wkhtmltopdf` 使用 `webkit` 作为渲染内核，可以使用一些现代的 `CSS` 方案，尽量降低模板复杂度。

## 功能

提供 4 个接口，分别对应不同的使用场景：
- `htmlpdf`：将 HTML 源码渲染成 `PDF` 文件格式。
- `linkpdf`：将在线的链接渲染成为 `PDF` 文件格式。
- `combine`：将若干个 PDF URL 合并成一个 PDF 文件。
- `link/combine`：将若干个 PDF/网页 URL 合并成一个 PDF 文件。

## 编译

- 安装 Golang 环境, Go >= 1.16
- 检出源码
- 在源码目录执行 `go get -v` 签出所有的依赖库
- 执行 `go build -o html2pdf .` 编译成二进制可执行文件
- 执行文件 `html2pdf -c ./config.json`

## 依赖

- [Phantomjs](http://phantomjs.org)，基于 `qtwebkit` 的渲染引擎。
- 在 `puppeteer` 分支，还有依赖于 [puppeteer](https://github.com/GoogleChrome/puppeteer)，使用 `chrome-headless` 渲染页面。

## 配置

使用 `Phantomjs` 渲染：

```json
{
    "listen": "127.0.0.1:4444", // HTTP 服务绑定地址
    "tmp_path": "", // 生成 PDF 文件中间的所有过渡临时文件存放路径
    "web_root": "", // HTTP 服务自带了一个示例 sample 存放路径
    "webkit_bin": "", // Phantomjs 执行文件的存放路径
    "webkit_args": ["./render/phantomjs/pdf.js"], // Phantomjs CLI 默认的参数，执行 JS 渲染的具体脚本
    "pdftk_bin": "pdftk.exe", // pdftk 渲染器位置
    "cache_ttl": 3600, // 静态 PDF 缓存时间（秒）
    "worker": 4, // 生成 PDF 的工作进程数
    "timeout": 40 // 生成 PDF 的进程的超时时间
}
```

> 注意: 由于 `Phantomjs` 依赖 `fontconfig`，而不同环境下 `fontconfig` 配置会有不一样的情况，需要从两方面入手解决在不同系统下渲染差别的问题：
> 1. 尽量使用 `embed font` 处理渲染的字体，包括默认的字体。
> 2. 同步 `fontconfig` 的一些公用配置，一般放在 `/etc/fonts/conf.d` 下，修改完后执行 `fc-cache -fv` 重置 `fontconfig`。

## HTML 模板

经过测试发现，如果需要 `A4` 规格的 `PDF` 文件铺满需要使用 `1240px x 1754px` 这个尺寸，但这也只是参考值；因为我们发现不同系统上 `wkhtmltopdf` 渲染网页的页面尺寸是不一样的，这需要控制 `--zoom` 参数进行预匹配。

各种纸张的打印尺寸规格可以参考[这里](http://www.papersizes.org/a-sizes-in-pixels.htm)。

请参考用于生成支持 `HTML2PDF` 的 HTML 模板项目 [html2pdf-template](https://github.com/MMHK/html2pdf-template)。

## Docker

此项目已经打包成 Docker 镜像。

- 签出 Docker 镜像
```bash
docker pull mmhk/html2pdf
```
- 环境变量，具体请参考 `config.json` 的说明。
  - WORKER：同时渲染的进程数，默认为 4
  - HOST：服务绑定的地址及端口，默认为 `127.0.0.1:4444`
  - ROOT：swagger-ui 存放的本地目录，可以设置为空来屏蔽 swagger-ui 的显示，默认为 `/usr/local/html2pdf/web_root`
  - TIMEOUT：每个渲染进程的超时时间（秒），默认为 60
  - TTL：静态 PDF 缓存时间（秒），默认为 3600（1小时）

- 运行
```bash
docker run --name html2pdf -p 4444:4444 mmhk/html2pdf:latest
```

## Docker Compose

你也可以使用 `docker-compose` 来运行此项目。以下是一个示例 `docker-compose.yml` 文件：

```yaml
version: "3.7"

services:
  app:
    image: mmhk/html2pdf
    restart: always
    environment:
      - HOST=0.0.0.0:4444
      - ROOT=/app/web_root
      - TIMEOUT=60
      - WORKER=4
      - TTL=3600
      - LOG_LEVEL=INFO
      - TZ=Asia/Hong_Kong
    ports:
      - 4444:4444
```

使用以下命令启动服务：

```bash
docker-compose up -d
```

## License

此项目使用 [Apache 2.0 许可证](LICENSE)。
```

希望这些信息对你有帮助！如果有任何问题或需要进一步的信息，请随时联系我。