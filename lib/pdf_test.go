package lib

import (
	//	"os"
	"path/filepath"
	//	"strings"
	"testing"
)


//func Test_Covert(t *testing.T) {
//	files, err := filepath.Glob(getLocalConfigPath("../temp/combine/*"))
//	if err != nil {
//		t.Log(err)
//		t.Fail()
//		return
//	}

//	files = Filter(files, func(item string) bool {
//		ext := filepath.Ext(item)

//		if strings.EqualFold(ext, ".jpg") {
//			return true
//		}
//		if strings.EqualFold(ext, ".png") {
//			return true
//		}
//		if strings.EqualFold(ext, ".gif") {
//			return true
//		}
//		if strings.EqualFold(ext, ".jpeg") {
//			return true
//		}

//		return false
//	})

//	for _, file := range files {
//		dest_file := file + ".pdf"
//		err = ConvertToPdf(file, dest_file)
//		if err != nil {
//			t.Log(err)
//			t.Fail()
//			return
//		}
//		err = os.Remove(file)
//		if err != nil {
//			t.Log(err)
//		}
//	}

//	t.Log("PASS")
//}

func Test_Combine(t *testing.T) {
	files, err := filepath.Glob(getLocalConfigPath("../temp/15210063*.pdf"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	dest_file := getLocalConfigPath("../temp/bundle.pdf")

	err = CombinePDF(files, dest_file)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log("PASS")
}

//func Test_Download(t *testing.T) {
//	url_list := []string{
//		"https://s3-ap-southeast-1.amazonaws.com/mm-test-dev/jetso/policy/ATT0099345ZC_policy.pdf",
//		"https://s3-ap-southeast-1.amazonaws.com/mm-test-dev/jetso/policy/ATT0099344ZC_policy.pdf",
//		"D:/_Sam/TestProject/golang/go/src/HTML2PDF/temp/test.pdf",
//		"https://s3-ap-southeast-1.amazonaws.com/mm-test-dev/jetso/policy/ATT0099346ZC_policy.pdf",
//	}

//	d := NewDownloader(url_list, getLocalConfigPath("../temp/combine"))
//	d.Start()
//	d.Done(func(list []string) {
//		t.Log(list)
//	})

//	t.Log("PASS")
//}
