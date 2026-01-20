package lib

import (
	"fmt"
	"github.com/chromedp/cdproto/page"
	"html2pdf/tests"
	"os"
	"sync"
	"testing"
)

func Test_BuildFromLink(t *testing.T) {
	conf, err := loadConfig()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	pdf := NewHTMLPDF(conf)

	worker_count := 10
	wg := new(sync.WaitGroup)
	for i := 0; i < worker_count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			file, err := pdf.BuildFromLink(os.Getenv("TEST_PDF_URL"))
			if err != nil {
				t.Log(err)
				t.Fail()
				return
			}
			t.Log(file)
		}()
	}

	wg.Wait()

	//defer os.Remove(file)

	t.Log("PASS")
}

func TestHTMLPDF_WithParamsRun(t *testing.T) {
	conf, err := loadConfig()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	pdf := NewHTMLPDF(conf)
	file, err := pdf.WithParamsRun("https://v5.geestar.mixmedia.com/api/receipt/proposal?order_id=10", &page.PrintToPDFParams{
		PaperWidth:        8.27,  //A4
		PaperHeight:       11.69, //A4
		Landscape:         false,
		PrintBackground:   true,
		MarginTop:         0,
		MarginBottom:      0,
		MarginLeft:        0,
		MarginRight:       0,
		PreferCSSPageSize: true,
		Scale:             1,
	})
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	//defer os.Remove(file)
	t.Log(file)
	t.Log("PASS")
}

func TestHTMLPDF_BuildFromLink(t *testing.T) {
	conf, err := loadConfig()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	pdf := NewHTMLPDF(conf)

	worker_count := 10
	wg := new(sync.WaitGroup)

	for i := 0; i < worker_count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			file, err := pdf.BuildFromLink(fmt.Sprintf("file://%s",
				tests.GetLocalPath("../tests/index.html")))
			if err != nil {
				t.Log(err)
				t.Fail()
				return
			}
			//defer os.Rename(file, tests.GetLocalPath("../tests/temp.pdf"))
			t.Log(file)
			defer os.Remove(file)
		}()
	}

	for i := 0; i < worker_count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			file, err := pdf.BuildFromLink(fmt.Sprintf("file://%s",
				tests.GetLocalPath("../tests/index.html")))
			if err != nil {
				t.Log(err)
				t.Fail()
				return
			}
			//defer os.Rename(file, tests.GetLocalPath("../tests/temp.pdf"))
			t.Log(file)
			defer os.Remove(file)
		}()
	}

	wg.Wait()

	t.Log("PASS")
}
