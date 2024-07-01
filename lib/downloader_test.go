package lib

import (
	"os"
	"testing"
)

func Test_Download(t *testing.T) {
	url_list := []string{
		"https://s3.ap-southeast-1.amazonaws.com/s3.test.mixmedia.com/sa_whatsapp_llm/upload/2023/09/14/6502d52879a52.pdf",
		"https://s3.ap-southeast-1.amazonaws.com/s3.test.mixmedia.com/sa_whatsapp_llm/upload/2023/08/30/64ef1a9558175.pdf",
		"F:\\TestProject\\golang\\html2pdf\\tests\\test2.pdf",
		"F:\\TestProject\\golang\\html2pdf\\tests\\test1.pdf",
		"https://reads.ie/pages/size-guide",
		"https://reads.ie/cdn/shop/files/reads-ie-logo.png",
		"https://s3.ap-southeast-1.amazonaws.com/s3.test.mixmedia.com/sa_whatsapp_llm/upload/2023/09/12/64ffd0d717e92.pdf",
	}

	conf, err := loadConfig()
	if err != nil {
		t.Fatal(err)
		return
	}
	downloader := NewDownloader(url_list, conf)
	queue := downloader.GetDownloadedFiles()
	for job := range queue {
		if !job.IsLocal {
			defer os.Remove(job.LocalPath)
		}

		t.Logf("%s isPDF:%v, isImage:%v, isHTML:%v",
			job.LocalPath, job.IsPDF(), job.IsImage(), job.IsHTML())
	}

	t.Log("PASS")
}