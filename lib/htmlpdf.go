package lib

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
)

//go:embed motorsPDFjsPatch.js
var MOTORS_PDF_JS_PATCH string
var (
	allocatorCtx    context.Context
	allocatorCancel context.CancelFunc
	allocatorMu     sync.Mutex
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

type PDFOption struct {
	page.PrintToPDFParams

	patchMotors bool
}

type HTMLPDF struct {
	config    *Config
	pdfOption *PDFOption
	jobQueue  chan bool
}

func NewHTMLPDF(conf *Config) *HTMLPDF {
	return &HTMLPDF{
		config:   conf,
		jobQueue: make(chan bool, conf.Worker),
		pdfOption: &PDFOption{
			PrintToPDFParams: page.PrintToPDFParams{
				PaperWidth:        8.27,  //A4
				PaperHeight:       11.69, //A4
				MarginTop:         0,
				MarginRight:       0,
				MarginBottom:      0,
				MarginLeft:        0,
				Scale:             1,
				Landscape:         false,
				PrintBackground:   true,
				PreferCSSPageSize: false,
			},
			patchMotors: true,
		},
	}
}

func (pdf *HTMLPDF) WithParams(params *page.PrintToPDFParams) *HTMLPDF {
	pdf.pdfOption = &PDFOption{
		PrintToPDFParams: *params,
		patchMotors:      pdf.pdfOption.patchMotors,
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

func (pdf *HTMLPDF) getAllocatorOpts() []chromedp.ExecAllocatorOption {

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

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(pdf.config.ChromePath),
		chromedp.DisableGPU,
		chromedp.Flag("font-render-hinting", "none"),             // 禁用字体渲染提示
		chromedp.Flag("disable-font-subpixel-positioning", true), // 禁用字体子像素
		chromedp.Flag("disable-web-security", true),              // 禁用证书验证
		chromedp.Flag("disable-setuid-sandbox", true),            // 禁用沙盒
		chromedp.Flag("disable-dev-shm-usage", true),             // 禁用 /dev/shm
		chromedp.WindowSize(viewportWidth, viewportHeight+50),
	)

	//logLevel, ok := os.LookupEnv("LOG_LEVEL")
	//if ok && logLevel == "DEBUG" {
	//	opts = append(opts, chromedp.Flag("headless", false))
	//}

	return opts
}

// initAllocator 初始化或重啟 Chrome allocator
func (pdf *HTMLPDF) initOrRestartAllocator() error {
	allocatorMu.Lock()
	defer allocatorMu.Unlock()

	// 如果已經有 allocator，先清理
	if allocatorCancel != nil {
		allocatorCancel()
	}

	// 建立新的 ExecAllocator
	allocatorCtx, allocatorCancel = chromedp.NewExecAllocator(context.Background(), pdf.getAllocatorOpts()...)

	// 測試是否成功啟動（避免啟動失敗還繼續用）
	testCtx, testCancel := chromedp.NewContext(allocatorCtx)
	defer testCancel()

	testTimeoutCtx, testCancelTimeout := context.WithTimeout(testCtx, 10*time.Second)
	defer testCancelTimeout()

	err := chromedp.Run(testTimeoutCtx, chromedp.Navigate("about:blank"))
	if err != nil {
		allocatorCancel()
		allocatorCtx = nil
		allocatorCancel = nil
		Log.Infof("Chrome allocator failed to start: %v", err)
		return fmt.Errorf("failed to start Chrome: %w", err)
	}

	Log.Info("Chrome allocator started / restarted successfully")
	return nil
}

// getAllocatorCtx 獲取可用的 allocator context，必要時重啟
func (pdf *HTMLPDF) getAllocatorCtx() (context.Context, context.CancelFunc, error) {
	allocatorMu.Lock()
	if allocatorCtx == nil || allocatorCtx.Err() != nil {
		allocatorMu.Unlock()
		if err := pdf.initOrRestartAllocator(); err != nil {
			return nil, nil, err
		}
		allocatorMu.Lock()
	}
	ctx := allocatorCtx
	cancel := allocatorCancel // 注意：這是 allocator 的 cancel，不要亂呼叫
	allocatorMu.Unlock()

	return ctx, cancel, nil
}

func (pdf *HTMLPDF) PrepareRuntime() error {
	_, _, err := pdf.getAllocatorCtx()
	return err
}

func (pdf *HTMLPDF) run(url string) (string, error) {
	// 將 PrintToPDFParams 轉換為 CSS @page 樣式
	if pdf.pdfOption.Scale == 0 {
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

	PreferCSSPageSize := false
	if pdf.pdfOption.PreferCSSPageSize || pdf.pdfOption.patchMotors {
		PreferCSSPageSize = true
	}

	allocCtx, _, err := pdf.getAllocatorCtx()
	if err != nil {
		Log.Errorf("Failed to create Chrome allocator: %v", err)
		return "", err
	}

	ctx, cancel := context.WithTimeout(allocCtx, time.Second*time.Duration(pdf.config.Timeout))
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
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			Log.Debug("chromedp js patch")
			if pdf.pdfOption.patchMotors {
				err := chromedp.Evaluate(string(MOTORS_PDF_JS_PATCH), nil).Do(ctx)
				if err != nil {
					Log.Error(err)
					return err
				}

				return chromedp.WaitReady("span[data-scrip-done=true]").Do(ctx)
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			Log.Debug("chromedp inject css")
			if !pdf.pdfOption.patchMotors && PreferCSSPageSize && len(customCSS) > 0 {
				return chromedp.Evaluate(fmt.Sprintf(`(function() {
						var style = document.createElement('style');
						style.type = 'text/css';
						style.innerHTML = %q;
						document.head.appendChild(style);
					})()`, customCSS), nil).Do(ctx)
			}

			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			select {
			case <-loadEventFired:
				Log.Debug("chromedp load event fired")
				var err error

				pdf.pdfOption.PreferCSSPageSize = PreferCSSPageSize

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
		Log.Errorf("error with fromlink：%s, error: %s\n", link, err)
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
		Log.Errorf("error with fromSource：%s, error: %s\n", tmpFile.Name(), err)
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
