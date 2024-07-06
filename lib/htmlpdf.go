package lib

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"path/filepath"
	"time"
)

//go:embed motorsPDFjsPatch.js
var MOTORS_PDF_JS_PATCH string

type TaskResult struct {
	File  string
	Err   error
	Index int
}

type Task struct {
	taskJob   chan *TaskResult
	taskCount int
}

type PDFOption struct {
	page.PrintToPDFParams

	patchMotors bool
}

type HTMLPDF struct {
	config   *Config
	pdfOption *PDFOption
	jobQueue chan bool
}

func NewHTMLPDF(conf *Config) *HTMLPDF {
	return &HTMLPDF{
		config:   conf,
		jobQueue: make(chan bool, conf.Worker),
		pdfOption: &PDFOption{
			PrintToPDFParams: page.PrintToPDFParams{
				PaperWidth:  8.27, //A4
				PaperHeight: 11.69, //A4
				MarginTop:   0,
				MarginRight: 0,
				MarginBottom: 0,
				MarginLeft: 0,
				Scale: 1,
				Landscape: false,
				PrintBackground: true,
				PreferCSSPageSize: false,
			},
			patchMotors: true,
		},
	}
}

func (pdf *HTMLPDF) WithParams(params *page.PrintToPDFParams) *HTMLPDF {
	pdf.pdfOption = &PDFOption{
		PrintToPDFParams: *params,
		patchMotors: pdf.pdfOption.patchMotors,
	}
	return pdf
}

func (pdf *HTMLPDF) WithParamsRun(url string, params *page.PrintToPDFParams) (string, error) {
	pdf.jobQueue <- true
	defer func() {
		<-pdf.jobQueue
	}()

	return pdf.WithParams(params).run(url)
}

func (pdf *HTMLPDF) run(url string) (string, error) {
	// 將 PrintToPDFParams 轉換為 CSS @page 樣式
	if pdf.pdfOption.Scale == 0  {
		pdf.pdfOption.Scale = 1
	}
	customCSS := ""
	if pdf.pdfOption.Landscape {
		customCSS = fmt.Sprintf(`
			@page {
				size: %.2fin %.2fin;
				margin: %.2fin %.2fin %.2fin %.2fin;
			}
		`, pdf.pdfOption.PaperHeight, pdf.pdfOption.PaperWidth, pdf.pdfOption.MarginTop, pdf.pdfOption.MarginRight, pdf.pdfOption.MarginBottom, pdf.pdfOption.MarginLeft)
	} else {
		customCSS = fmt.Sprintf(`
		@page {
			size: %.2fin %.2fin;
			margin: %.2fin %.2fin %.2fin %.2fin;
		}
	`, pdf.pdfOption.PaperWidth, pdf.pdfOption.PaperHeight, pdf.pdfOption.MarginTop, pdf.pdfOption.MarginRight, pdf.pdfOption.MarginBottom, pdf.pdfOption.MarginLeft)
	}

	PreferCSSPageSize := false;
	if pdf.pdfOption.PreferCSSPageSize || pdf.pdfOption.patchMotors {
		PreferCSSPageSize = true
	}

	dpi := 150.0
	PaperHeight := pdf.pdfOption.PaperHeight
	PaperWidth := pdf.pdfOption.PaperWidth
	if pdf.pdfOption.Landscape {
		PaperHeight = pdf.pdfOption.PaperWidth
		PaperWidth = pdf.pdfOption.PaperHeight
	}
	// 转换为视口尺寸（以像素为单位）
	viewportWidth := int(PaperWidth * dpi)
	viewportHeight := int(PaperHeight * dpi)

	Log.Debugf("PaperHeight: %f, PaperWidth: %f, dpi: %f, viewportWidth: %d, viewportHeight: %d", PaperHeight, PaperWidth, dpi, viewportWidth, viewportHeight)


	// 自定義 Chrome 路徑
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(pdf.config.ChromePath),
		chromedp.Flag("disable-web-security",true),
		chromedp.WindowSize(viewportWidth, viewportHeight + 50),
	)

	//logLevel, ok := os.LookupEnv("LOG_LEVEL")
	//if ok && logLevel == "DEBUG" {
	//	opts = append(opts, chromedp.Flag("headless", false))
	//}

	defaultCtx, cancel := context.WithTimeout(context.Background(), time.Second * time.Duration(pdf.config.Timeout))
	defer cancel()

	// 創建上下文
	ctx, cancel := chromedp.NewExecAllocator(defaultCtx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// 创建一个事件监听器来监听页面加载完成事件
	loadEventFired := make(chan struct{})
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev.(type) {
		case *page.EventLoadEventFired:
			close(loadEventFired)
		}

	})

	waitTime := time.Second * 1
	if pdf.pdfOption.patchMotors {
		waitTime = time.Second * 3
	}

	var buf []byte
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			Log.Debug("chromedp js patch")
			if pdf.pdfOption.patchMotors {
				return chromedp.Evaluate(string(MOTORS_PDF_JS_PATCH), nil).Do(ctx)
			}
			return nil
		}),
		chromedp.WaitReady("span[data-scrip-done=true]"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			Log.Debug("chromedp inject css")
			if PreferCSSPageSize && len(customCSS) > 0 {
				return chromedp.Evaluate(fmt.Sprintf(`(function() {
						var style = document.createElement('style');
						style.type = 'text/css';
						style.innerHTML = %q;
						document.head.appendChild(style);
					})()`, customCSS), nil).Do(ctx)
			}

			return nil
		}),
		chromedp.Sleep(waitTime),
		chromedp.ActionFunc(func(ctx context.Context) error {
			select {
			case <-loadEventFired:
				Log.Debug("chromedp load event fired")
				var err error

				buf, _, err = pdf.pdfOption.Do(ctx)
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}),
	})
	if err != nil {
		Log.Error(err)
		return "", err
	}
	defer page.Close()

	tmpFile, err := ioutil.TempFile("", "*.pdf")
	if err != nil {
		Log.Error(err)
		return "", err
	}
	defer tmpFile.Close()
	// 保存 PDF 文件
	if _, err := tmpFile.Write(buf); err != nil {
		Log.Error(err)
		return "", err
	}

	return filepath.ToSlash(tmpFile.Name()), nil
}

func (pdf *HTMLPDF) BuildFromLink(link string) (local_pdf string, err error) {
	pdf_name, err := pdf.run(link)
	if err != nil {
		return "", err
	}
	return pdf_name, nil
}

func (pdf *HTMLPDF) BuildFromSource(html []byte) (local_pdf string, err error) {

	tmpFile, err := ioutil.TempFile("", "*.html")
	if err != nil {
		Log.Error(err)
		return "", err
	}
	defer tmpFile.Close()
	// 保存 PDF 文件
	if _, err := tmpFile.Write(html); err != nil {
		Log.Error(err)
		return "", err
	}


	pdf_name, err := pdf.run(fmt.Sprintf("file://%s", tmpFile.Name()))
	if err != nil {
		return "", err
	}

	return pdf_name, nil
}

func (pdf *HTMLPDF) Combine(files []string) (dest_pdf_path string, err error) {
	tmpFile, err := ioutil.TempFile("", "*.html")
	if err != nil {
		Log.Error(err)
		return "", err
	}
	tmpFile.Close()

	pdf_name := filepath.ToSlash(tmpFile.Name())
	err = CombinePDF(files, pdf_name)
	if err != nil {
		return pdf_name, err
	}
	return pdf_name, nil
}
