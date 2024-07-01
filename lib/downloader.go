package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type JobItem struct {
	URL       string
	Name      string
	LocalPath string
	IsLocal   bool
	Index     int
}

func (this *JobItem) IsPDF() bool {
	return strings.HasSuffix(this.Name, ".pdf")
}

func (this *JobItem) IsImage() bool {
	return strings.HasSuffix(this.Name, ".jpg") || strings.HasSuffix(this.Name, ".png") || strings.HasSuffix(this.Name, ".gif") || strings.HasSuffix(this.Name, ".jpeg")
}

func (this *JobItem) IsHTML() bool {
	return strings.HasSuffix(this.Name, ".html") || strings.HasSuffix(this.Name, ".htm")
}

type JobList []*JobItem

func (j *JobList) Urls() []string {
	list := make([]string, len(*j))
	for i, item := range *j {
		if item.IsLocal {
			list[i] = fmt.Sprintf("file://%s", item.LocalPath)
			continue
		}
		list[i] = item.URL
	}
	return list
}

type Downloader struct {
	list   []string
	config *Config
}

func NewDownloader(UrlList []string, conf *Config) *Downloader {
	return &Downloader{
		list:   UrlList,
		config: conf,
	}
}

//下载远程文件
func (d *Downloader) DownloadRemoteFile(remoteURL string, index int) (error, *JobItem) {
	Log.Debug("begin download file, url:", remoteURL)
	var ext string
	urlInfo, err := url.Parse(remoteURL)
	if err != nil {
		Log.Error(err)
		return err, nil
	}
	ext = filepath.Ext(urlInfo.Path)

	//检查是否存在本地文件，即remoteURL 为本地文件路劲
	if _, err := os.Stat(remoteURL); err == nil {
		Log.Debug("local file hint, path:", remoteURL)
		return nil, &JobItem{
			Name:      filepath.Base(remoteURL),
			LocalPath: filepath.ToSlash(remoteURL),
			URL:       filepath.ToSlash(remoteURL),
			IsLocal:   true,
			Index:     index,
		}
	}

	client := &http.Client{
		Timeout: time.Second * time.Duration(d.config.Timeout), // 设置整个请求的超时时间
	}
	resp, err := client.Get(remoteURL)
	if err != nil {
		Log.Error(err)
		return err, nil
	}
	defer resp.Body.Close()

	if !strings.Contains(ext, ".") {
		ext = ".tmp"
		// 获取 Content-Type 并解析文件扩展名
		contentType := resp.Header.Get("Content-Type")
		extList, err := mime.ExtensionsByType(contentType)
		if err != nil {
			Log.Error(err)
		} else {
			if len(extList) > 0 {
				ext = extList[0]
			}
		}
	}

	tmpFile, err := ioutil.TempFile("", fmt.Sprintf("*%s", ext))
	if err != nil {
		Log.Error(err)
		return err, nil
	}
	defer tmpFile.Close()

	Log.Debugf("local Path:%s", tmpFile.Name())

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		Log.Error(err)
		return err, nil
	}

	return nil, &JobItem{
		Name:      filepath.Base(tmpFile.Name()),
		LocalPath: filepath.ToSlash(tmpFile.Name()),
		URL:       filepath.ToSlash(remoteURL),
		Index:     index,
		IsLocal:   false,
	}
}

func (d *Downloader) GetDownloadedFiles() <-chan *JobItem {
	total := len(d.list)
	queue := make(chan *JobItem, total)
	var wg sync.WaitGroup
	maxConcurrentWorkers := make(chan bool, d.config.Worker)
	wg.Add(total)
	defer close(queue)
	defer close(maxConcurrentWorkers)

	for i, fileUrl := range d.list {
		go func(queue chan *JobItem, url string, i int, wg *sync.WaitGroup, limit chan bool) error {
			defer wg.Done()
			limit <- true
			defer func() { <-limit }()

			err, jobItem := d.DownloadRemoteFile(url, i)
			if err != nil {
				Log.Error(err)
				return err
			}
			queue <- jobItem
			Log.Debugf("%+v", jobItem)
			return nil
		}(queue, fileUrl, i, &wg, maxConcurrentWorkers)
	}

	wg.Wait()

	return queue
}
