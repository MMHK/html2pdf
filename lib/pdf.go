package lib

import (
	"crypto/md5"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/hhrutter/pdfcpu/pkg/api"
	"github.com/hhrutter/pdfcpu/pkg/pdfcpu"
	"github.com/jung-kurt/gofpdf"
)

type JobItem struct {
	URL       string
	Name      string
	LocalPath string
	Index     int
}

type Downloader struct {
	fileCount   int
	list        []string
	tempPath    string
	cachePath   string
	downloadJob chan JobItem
}

func NewDownloader(UrlList []string, tempPath string) *Downloader {
	return &Downloader{
		fileCount:   len(UrlList),
		list:        UrlList,
		tempPath:    tempPath,
		cachePath:   filepath.Join(tempPath, "cache"),
		downloadJob: make(chan JobItem, 4),
	}
}

//下载远程文件
func (d *Downloader) DownloadRemoteFile(remoteURL string, index int) {
	InfoLogger.Println("begin download file, url:", remoteURL)
	var ext string
	urlInfo, err := url.Parse(remoteURL)
	if err != nil {
		ErrLogger.Println(err)
	}
	ext = filepath.Ext(urlInfo.Path)
	if !strings.Contains(ext, ".") {
		ext = ""
	}
	filename := fmt.Sprintf("%x%s", md5.Sum([]byte(remoteURL)), ext)

	basePath := filepath.Join(d.tempPath, filename)
	cachePath := filepath.Join(d.cachePath, filename)

	//检查是否存在本地缓存
	if _, err := os.Stat(cachePath); err == nil {
		d.downloadJob <- JobItem{
			Name:      filename,
			LocalPath: cachePath,
			URL:       remoteURL,
			Index:     index,
		}
		return
	}
	//检查是否存在本地文件，即remoteURL 为本地文件路劲
	if _, err := os.Stat(remoteURL); err == nil {
		d.downloadJob <- JobItem{
			Name:      filepath.Base(remoteURL),
			LocalPath: remoteURL,
			URL:       remoteURL,
			Index:     index,
		}
		return
	}

	defer (func() {
		if err := recover(); err != nil {
			ErrLogger.Println(err)
		} else {
			d.downloadJob <- JobItem{
				Name:      filename,
				LocalPath: basePath,
				URL:       remoteURL,
				Index:     index,
			}
		}
	})()

	file, err := os.Create(basePath)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	defer file.Close()

	resp, err := http.Get(remoteURL)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	return
}

func (d *Downloader) Start() {
	for i, url := range d.list {
		go d.DownloadRemoteFile(url, i)
	}
}

func (d *Downloader) Done(callback func([]string)) {
	local_list := make([]string, d.fileCount)

	for {
		item := <-d.downloadJob
		local_list[item.Index] = item.LocalPath
		InfoLogger.Println("a download job Done.")
		go d.CacheFile(item)
		d.fileCount--
		if d.fileCount <= 0 {
			break
		}
	}

	defer callback(local_list)

	close(d.downloadJob)
}

func (d *Downloader) CacheFile(item JobItem) {
	destPath := filepath.Join(d.cachePath, item.Name)

	if !strings.EqualFold(item.LocalPath, destPath) && !strings.EqualFold(item.LocalPath, item.URL) {
		err := d.Copy(item.LocalPath, destPath)
		if err != nil {
			ErrLogger.Println(err)
		}
	}
}

func (d *Downloader) Copy(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	dir := filepath.Dir(dst)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func ConvertToPdf(src_image_path string, dest_pdf_path string) error {
	src_image, err := os.Open(src_image_path)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}
	defer src_image.Close()
	img, _, err := image.Decode(src_image)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}
	rect := img.Bounds()

	pdfTpye := "P"
	if rect.Dx() > rect.Dy() {
		pdfTpye = "L"
	}

	pdf := gofpdf.New(pdfTpye, "mm", "A4", ".")
	pdf.AddPage()
	w, _ := pdf.GetPageSize()
	pdf.Image(src_image_path, 0, 0, w, 0, false, "", 0, "")
	err = pdf.OutputFileAndClose(dest_pdf_path)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}

	return nil
}

func CombinePDF(files []string, dest_pdf_path string) error {
	//处理合并过程中可能出现的异常
	defer func() {
		if err := recover(); err != nil {
			ErrLogger.Println(err)
		}
	}()

	config := pdfcpu.NewDefaultConfiguration()
	// config.SetValidationRelaxed()
	cmd := api.MergeCommand(files, dest_pdf_path, config)
	_, err := api.Process(cmd)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}

	return nil
}
