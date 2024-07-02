package lib

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"path/filepath"
	"time"
)

type TaskResult struct {
	File  string
	Err   error
	Index int
}

type Task struct {
	taskJob   chan *TaskResult
	taskCount int
}

type HTMLPDF struct {
	config   *Config
	jobQueue chan bool
}

func NewHTMLPDF(conf *Config) *HTMLPDF {
	return &HTMLPDF{
		config:   conf,
		jobQueue: make(chan bool, conf.Worker),
	}
}

func (pdf *HTMLPDF) WithParamsRun(url string, params *page.PrintToPDFParams) (string, error) {
	pdf.jobQueue <- true
	defer func() {
		<-pdf.jobQueue
	}()

	// 將 PrintToPDFParams 轉換為 CSS @page 樣式
	if params.Scale == 0  {
		params.Scale = 1
	}
	customCSS := fmt.Sprintf(`
		@page {
			size: %.2fin %.2fin;
			margin: %.2fin %.2fin %.2fin %.2fin;
		}
		.page { page-break-inside: avoid; }

	`, params.PaperWidth, params.PaperHeight, params.MarginTop, params.MarginRight, params.MarginBottom, params.MarginLeft)
	if params.Landscape {
		customCSS = fmt.Sprintf(`
			@page {
				size: %.2fin %.2fin;
				margin: %.2fin %.2fin %.2fin %.2fin;
			}
			.page { page-break-inside: avoid; }
		`, params.PaperHeight, params.PaperWidth, params.MarginTop, params.MarginRight, params.MarginBottom, params.MarginLeft)
	}

	dpi := 150.0
	PaperHeight := params.PaperHeight
	PaperWidth := params.PaperWidth
	if params.Landscape {
		PaperHeight = params.PaperWidth
		PaperWidth = params.PaperHeight
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

	var buf []byte
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			Log.Debug("chromedp inject css")
			if params.PreferCSSPageSize {
				return chromedp.Evaluate(fmt.Sprintf(`(function() {
						var style = document.createElement('style');
						style.type = 'text/css';
						style.innerHTML = %q;
						document.head.appendChild(style);
					})()`, customCSS), nil).Do(ctx)
			}

			return nil
		}),
		chromedp.Sleep(time.Second * 1),
		chromedp.ActionFunc(func(ctx context.Context) error {
			select {
			case <-loadEventFired:
				Log.Debug("chromedp load event fired")
				var err error

				buf, _, err = params.Do(ctx)
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

func (pdf *HTMLPDF) run(url string) (string, error) {
	params := &page.PrintToPDFParams{
		PaperWidth:  8.27, //A4
		PaperHeight: 11.69, //A4
		Landscape:    false,
		PrintBackground: true,
		MarginTop:    0,
		MarginBottom: 0,
		MarginLeft:   0,
		MarginRight:  0,
		PreferCSSPageSize: true,
		Scale: 0.64,
	}

	return pdf.WithParamsRun(url, params)
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
