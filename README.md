# HTML2PDF


将 `HTML` 渲染成 `PDF` 文件，此项目用`Golang`开发，提供一个HTTP 接口将本地
或者远程HTML页面输出成为`PDF`文件格式。

使用`HTML`作为模板的原因是因为HTML的表达能力最好，而且`wkhtmltopdf`使用
`webkit`作为渲染内核，可以使用一些现代的`CSS`方案，尽量降低模本复杂度。

## 功能

提供2个接口，分别对应不同的使用场景。
 - `htmlpdf`，将HTML源码渲染成`PDF`文件格式。 
 - `linkpdf`，将online的link渲染成为`PDF`文件格式。 
 - `combine`，将若干个pdf url，合并成一个的pdf文件。
 - `link/combine`，将若干个pdf/网页 url，合并成一个的pdf文件。

## 编译

- 安装Golang环境, Go >= 1.8
- checkout 源码
- 在源码目录 执行 ` go get -v `签出所有的依赖库
- ` go build -o html2pdf .` 编译成二进制可执行文件
- 执行文件 ` html2pdf -c ./config.json`

## 依赖

- [Phantomjs](http://phantomjs.org/) ，基于 `qtwebkit`的渲染引擎。
- 在 `puppeteer` 分支，还有依赖于 [puppeteer](https://github.com/GoogleChrome/puppeteer) , 使用`chrome-headless` 渲染页面。

## 配置

使用 ``Phantomjs`` 渲染：

```
{
    "listen": "127.0.0.1:4444", //http service绑定地址
    "tmp_path": "", //生成pdf文件中间的所有过渡临时文件存放路径
    "web_root": "", //http service自带了一个示例sample存放路径
    "webkit_bin": "", //Phantomjs 执行文件的存放路径
    "webkit_args": [ "./render/phantomjs/pdf.js" ], //Phantomjs cli 默认的参数，执行js渲染的具体脚本
    "pdftk_bin": "pdftk.exe", //pdftk 渲染器位置
    "worker": 4, //生成pdf的工作进程数
    "timeout": 40 //生成PDF的进程的超时时间
}
```

> 注意: 由于`Phantomjs` 依赖 `fontconfig`, 而不同环境下 `fontconfig` 配置会有不一样的情况，
> 需要从两方面入手解决在不同系统下渲染差别的问题。
> 1. 尽量使用`embed font`处理渲染的字体，包括默认的字体。
> 2. 同步 `fontconfig` 的一些公用配置，一般放在 `/etc/fonts/conf.d` 下，修改完后执行 `fc-cache -fv` reset `fontconfig`。


## HTML 模板

经过测试发现如果需要 `A4` 规格的 `PDF` 文件铺满需要使用`1240px x 1754px`
这个尺寸，但是这个也只是参考值；因为我们发现不同系统上 `wkhtmltopdf` 
渲染web page的页面尺寸是不一样的，这个需要控制 `--zoom` 参数进行预匹配。

个种纸张的打印尺寸规格可以参考[这里](http://www.papersizes.org/a-sizes-in-pixels.htm)

## Docker

此项目已经打包成docker 镜像

- 签出docker 镜像
```
docker pull mmhk/html2pdf
```
- 环境变量，具体请参考 `config.json` 的说明。
  - WORKER ，同时渲染的进程数, 默认为 4
  - HOST，service绑定的服务地址及端口，默认为 `127.0.0.1:4444`
  - ROOT, swagger-ui 存放的本地目录，可以设置空来屏蔽 swagger-ui 的显示， 默认为 `/usr/local/html2pdf/web_root`
  - TIMEOUT， 每个渲染进程的超时时间(秒)， 默认为 60
  
- 运行
```
docker run --name html2pdf -p 4444:4444 mmhk/html2pdf:latest
```

