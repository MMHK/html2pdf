package lib

import (
	"crypto/md5"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/jung-kurt/gofpdf"

	"github.com/pdfcpu/pdfcpu/pkg/cli"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
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
	config      *Config
	downloadJob chan JobItem
}

func NewDownloader(UrlList []string, tempPath string, conf *Config) *Downloader {
	return &Downloader{
		fileCount:   len(UrlList),
		list:        UrlList,
		tempPath:    tempPath,
		config:      conf,
		cachePath:   filepath.Join(tempPath, "cache"),
		downloadJob: make(chan JobItem, 4),
	}
}

// 下载远程文件
func (d *Downloader) DownloadRemoteFile(remoteURL string, index int) {
	Logger.Infof("begin download file, url:\n", remoteURL)
	var ext string
	urlInfo, err := url.Parse(remoteURL)
	if err != nil {
		Logger.Error(err)
	}
	ext = filepath.Ext(urlInfo.Path)
	if !strings.Contains(ext, ".") {
		ext = ""
	}
	filename := fmt.Sprintf("%x%s", md5.Sum([]byte(remoteURL)), ext)

	basePath := filepath.Join(d.tempPath, filename)
	cachePath := filepath.Join(d.cachePath, filename)

	//检查是否存在本地缓存
	if info, err := os.Stat(cachePath); err == nil &&
		info.Size() > 0 {
		//检查cache file 是否已经失效
		if info.ModTime().Add(time.Duration(d.config.CacheTTL) * time.Second).After(time.Now()) {
			Logger.Infof("cache file hint, path:%s\n", cachePath)
			d.downloadJob <- JobItem{
				Name:      filename,
				LocalPath: cachePath,
				URL:       remoteURL,
				Index:     index,
			}
			return
		}
	}
	//检查是否存在本地文件，即remoteURL 为本地文件路劲
	if _, err := os.Stat(remoteURL); err == nil {
		Logger.Infof("local file hint, path:%s\n", remoteURL)
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
			Logger.Error(err)
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
		Logger.Error(err)
		return
	}
	defer file.Close()

	resp, err := http.Get(remoteURL)
	if err != nil {
		Logger.Error(err)
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
		Logger.Info("a download job Done.")
		d.CacheFile(item)
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
	lockPath := destPath + ".lock"

	if _, err := os.Stat(lockPath); err == nil {
		time.AfterFunc(time.Second*60, func() {
			d.CacheFile(item)
		})
		return
	} else {
		ioutil.WriteFile(lockPath, []byte{}, 0777)
	}

	if !strings.EqualFold(item.LocalPath, destPath) && !strings.EqualFold(item.LocalPath, item.URL) {
		tempPath := destPath + ".tmp"
		err := d.Copy(item.LocalPath, tempPath)
		if err != nil {
			Logger.Error(err)
		}
		err = os.Rename(tempPath, destPath)
		if err != nil {
			Logger.Error(err)
		}
	}

	os.Remove(lockPath)
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
		Logger.Error(err)
		return err
	}
	defer src_image.Close()
	img, _, err := image.Decode(src_image)
	if err != nil {
		Logger.Error(err)
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
		Logger.Error(err)
		return err
	}

	return nil
}

func CombinePDF(files []string, dest_pdf_path string) error {
	//处理合并过程中可能出现的异常
	defer func() {
		if err := recover(); err != nil {
			Logger.Error(err)
		}
	}()

	config := pdfcpu.NewDefaultConfiguration()
	config.ValidationMode = pdfcpu.ValidationRelaxed
	cmd := cli.MergeCommand(files, dest_pdf_path, config)
	_, err := cli.Process(cmd)
	if err != nil {
		Logger.Error(err)
		return err
	}

	return nil
}
