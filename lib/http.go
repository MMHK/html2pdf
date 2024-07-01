package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type HTTPService struct {
	config *Config
}

func NewHTTP(conf *Config) *HTTPService {
	return &HTTPService{
		config: conf,
	}
}

func (s *HTTPService) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/", s.RedirectSwagger)
	r.HandleFunc("/htmlpdf", s.HTMLPDF)
	r.HandleFunc("/linkpdf", s.LINKPDF)
	r.HandleFunc("/combine", s.COMBINE)
	r.HandleFunc("/link/combine", s.LinkCombine)
	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/",
		http.FileServer(http.Dir(fmt.Sprintf("%s/swagger", s.config.WebRoot)))))
	r.NotFoundHandler = http.HandlerFunc(s.NotFoundHandle)

	Log.Info("http service starting")
	Log.Infof("Please open http://%s\n", s.config.Listen)
	http.ListenAndServe(s.config.Listen, r)
}

func (s *HTTPService) NotFoundHandle(writer http.ResponseWriter, request *http.Request) {
	http.Error(writer, "handle not found!", 404)
}

func (s *HTTPService) RedirectSwagger(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/swagger/index.html", 301)
}

func (s *HTTPService) HTMLPDF(writer http.ResponseWriter, request *http.Request) {

	upload_text := request.FormValue("upload")
	var bin []byte
	if len(upload_text) > 0 {
		bin = []byte(upload_text)
	} else {
		request.ParseMultipartForm(32 << 20)
		file, _, err := request.FormFile("upload")
		if err != nil {
			Log.Error(err)
			http.Error(writer, err.Error(), 500)
			return
		}
		defer file.Close()

		bin, err = ioutil.ReadAll(file)
		if err != nil {
			Log.Error(err)
			http.Error(writer, err.Error(), 500)
			return
		}
	}

	htmlpdf := NewHTMLPDF(s.config)
	file, err := htmlpdf.BuildFromSource(bin)
	if err != nil {
		Log.Error(err)
		http.Error(writer, err.Error(), 500)
		return
	}
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.pdf", time.Now().UnixNano()))
	writer.Header().Set("Content-Type", "application/pdf")

	pdf, err := os.Open(file)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	defer pdf.Close()
	defer time.AfterFunc(time.Second*10, func() {
		os.Remove(file)
	})
	_, err = io.Copy(writer, pdf)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
}

func (s *HTTPService) LINKPDF(writer http.ResponseWriter, request *http.Request) {
	link := request.FormValue("link")

	htmlpdf := NewHTMLPDF(s.config)
	file, err := htmlpdf.BuildFromLink(link)
	if err != nil {
		Log.Error(err)
		http.Error(writer, err.Error(), 500)
		return
	}
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.pdf", time.Now().UnixNano()))
	writer.Header().Set("Content-Type", "application/pdf")

	pdf, err := os.Open(file)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	defer pdf.Close()
	defer time.AfterFunc(time.Second*10, func() {
		os.Remove(file)
	})
	_, err = io.Copy(writer, pdf)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
}

func (s *HTTPService) LinkCombine(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, err.Error(), 500)
	}

	input_files := make([]string, 0)
	for key, values := range request.PostForm {
		if strings.EqualFold(key, "file") {
			input_files = append(input_files, values...)
		}
	}

	downloader := NewDownloader(input_files, s.config)
	queue := downloader.GetDownloadedFiles()
	localFiles := make([]string, 0)
	worker := make(chan bool, s.config.Worker)
	defer close(worker)

	var wg sync.WaitGroup
	for item := range queue {

		if item.IsPDF() {
			localFiles = append(localFiles, item.LocalPath)
			continue
		}
		if item.IsImage() {
			defer os.Remove(item.LocalPath)
			wg.Add(1)

			go func(job *JobItem, wg *sync.WaitGroup) {
				defer wg.Done()
				worker <- true
				defer func() {
					<-worker
				}()

				Log.Debug("convert image to pdf")

				pdf_path := fmt.Sprintf("%s.pdf", item.LocalPath)
				err := ConvertToPdf(item.LocalPath, pdf_path)
				if err != nil {
					Log.Error(err)
					return
				}
				localFiles = append(localFiles, pdf_path)
			}(item, &wg)

			continue
		}
		if item.IsHTML() {
			defer os.Remove(item.LocalPath)
			wg.Add(1)

			go func(job *JobItem, wg *sync.WaitGroup) {
				defer wg.Done()
				worker <- true
				defer func() {
					<-worker
				}()

				htmlpdf := NewHTMLPDF(s.config)
				localPath := fmt.Sprintf("file://%s", job.LocalPath)
				Log.Debug("convert html to pdf", localPath)
				pdf_path, err := htmlpdf.BuildFromLink(localPath)
				if err != nil {
					Log.Error(err)
					return
				}
				Log.Debug("convert html to pdf done", pdf_path)
				localFiles = append(localFiles, pdf_path)
			}(item, &wg)

			continue
		}
	}

	Log.Debug("wait for all download job done")
	wg.Wait()

	defer func() {
		for _, file := range localFiles {
			os.Remove(file)
		}
	}()

	Log.Debug("combine pdf")

	var combine_path string

	if len(localFiles) == 1 {
		combine_path = localFiles[0]
	} else {
		combineFile, err := ioutil.TempFile("", "*.pdf")
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}
		combine_path = filepath.ToSlash(combineFile.Name())

		err = combineFile.Close()
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}

		err = CombinePDF(localFiles, combine_path)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}
	}

	download, err := os.Open(combine_path)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	defer download.Close()

	Log.Debug("combine done")

	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.pdf", time.Now().UnixNano()))
	writer.Header().Set("Content-Type", "application/pdf")
	_, err = io.Copy(writer, download)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	defer time.AfterFunc(time.Second * 10, func() {
		os.Remove(combine_path)
	})
}

func (s *HTTPService) COMBINE(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, err.Error(), 500)
	}

	input_files := make([]string, 0)
	for key, values := range request.PostForm {
		if strings.EqualFold(key, "file") {
			input_files = append(input_files, values...)
		}
	}

	downloader := NewDownloader(input_files, s.config)
	queue := downloader.GetDownloadedFiles()

	localFiles := make([]string, 0)
	for item := range queue {
		if item.IsPDF() {
			localFiles = append(localFiles, item.LocalPath)
		}
	}
	defer func() {
		for _, file := range localFiles {
			os.Remove(file)
		}
	}()
	combineFile, err := ioutil.TempFile("", "*.pdf")
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	combineFile.Close()
	combine_path := filepath.ToSlash(combineFile.Name())

	err = CombinePDF(localFiles, combine_path)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	download, err := os.Open(combine_path)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	defer download.Close()

	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.pdf", time.Now().UnixNano()))
	writer.Header().Set("Content-Type", "application/pdf")
	_, err = io.Copy(writer, download)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	defer time.AfterFunc(time.Second * 10, func() {
		os.Remove(combineFile.Name())
	})

}
