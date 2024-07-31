package lib

import (
	"net/url"
	//	"os"
	"path/filepath"
	//	"strings"
	"testing"
)

//func Test_BuildFromLink(t *testing.T) {
//	err, conf := NewConfig(getLocalConfigPath("../config.json"))
//	if err != nil {
//		t.Log(err)
//		t.Fail()
//		return
//	}

//	pdf := NewHTMLPDF(conf)

//	for range []int{1, 2} {
//		go func() {
//			file, err := pdf.BuildFromLink("http://www.baidu.com")
//			if err != nil {
//				t.Log(err)
//				t.Fail()
//				return
//			}
//			defer os.Remove(file)
//		}()
//	}

//	file, err := pdf.BuildFromLink("http://www.baidu.com")
//	if err != nil {
//		t.Log(err)
//		t.Fail()
//		return
//	}

//	defer os.Remove(file)

//	t.Log("PASS")
//}

func Test_ext(t *testing.T) {
	urlInfo, err := url.Parse("/usr/local/bin/")
	if err != nil {
		t.Log(err)
	}

	t.Log(filepath.Ext(urlInfo.Path))
}
