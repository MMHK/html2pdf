package lib

import (
	"fmt"
	"github.com/chromedp/cdproto/page"
	"html2pdf/tests"
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

	file, err := pdf.BuildFromLink("http://www.baidu.com")
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(file)

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
	file, err := pdf.WithParamsRun("https://v5.geestar.mixmedia.com/api/receipt/proposal?order_id=883", &page.PrintToPDFParams{
		PaperWidth:  8.27, //A4
		PaperHeight: 11.69, //A4
		Landscape:    false,
		PrintBackground: true,
		MarginTop:    0,
		MarginBottom: 0,
		MarginLeft:   0,
		MarginRight:  0,
		PreferCSSPageSize: true,
		Scale: 0.84,
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
	file, err := pdf.WithParamsRun(fmt.Sprintf("file://%s",
		tests.GetLocalPath("../tests/index.html")),
		&page.PrintToPDFParams{
			PrintBackground: true,
			MarginTop:       0,
			MarginBottom:    0,
			MarginLeft:      0,
			MarginRight:     0,
			Landscape:       false,
			Scale:           0.84,
			PreferCSSPageSize: true,
		})
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	//defer os.Rename(file, tests.GetLocalPath("../tests/temp.pdf"))
	t.Log(file)
	t.Log("PASS")
}
