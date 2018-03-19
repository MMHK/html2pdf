package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
	r.HandleFunc("/", s.RedirectSample)
	r.HandleFunc("/htmlpdf", s.HTMLPDF)
	r.HandleFunc("/linkpdf", s.LINKPDF)
	r.HandleFunc("/combine", s.COMBINE)
	r.HandleFunc("/link/combine", s.LinkCombine)
	r.PathPrefix("/sample/").Handler(http.StripPrefix("/sample/",
		http.FileServer(http.Dir(fmt.Sprintf("%s/sample", s.config.WebRoot)))))
	r.NotFoundHandler = http.HandlerFunc(s.NotFoundHandle)

	InfoLogger.Println("http service starting")
	InfoLogger.Printf("Please open http://%s\n", s.config.Listen)
	http.ListenAndServe(s.config.Listen, r)
}

func (s *HTTPService) NotFoundHandle(writer http.ResponseWriter, request *http.Request) {
	http.Error(writer, "handle not found!", 404)
}

func (s *HTTPService) RedirectSample(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/sample/index.html", 301)
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
			ErrLogger.Println(err)
			http.Error(writer, err.Error(), 500)
			return
		}
		defer file.Close()

		bin, err = ioutil.ReadAll(file)
		if err != nil {
			ErrLogger.Println(err)
			http.Error(writer, err.Error(), 500)
			return
		}
	}

	htmlpdf := NewHTMLPDF(s.config)
	file, err := htmlpdf.BuildFromSource(bin)
	if err != nil {
		ErrLogger.Println(err)
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
	_, err = io.Copy(writer, pdf)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	defer pdf.Close()
	defer time.AfterFunc(time.Second*10, func() {
		os.Remove(file)
	})
}

func (s *HTTPService) LINKPDF(writer http.ResponseWriter, request *http.Request) {
	link := request.FormValue("link")

	htmlpdf := NewHTMLPDF(s.config)
	file, err := htmlpdf.BuildFromLink(link)
	if err != nil {
		ErrLogger.Println(err)
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
	_, err = io.Copy(writer, pdf)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	defer pdf.Close()
	defer time.AfterFunc(time.Second*10, func() {
		os.Remove(file)
	})

}

func (s *HTTPService) LinkCombine(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, err.Error(), 500)
	}

	for key, values := range request.PostForm {
		if strings.EqualFold(key, "file") {

			//先转换非pdf文件的url为本地pdf
			input_files := make([]string, len(values))
			task := NewTask(len(values))

			InfoLogger.Println("request params:", values)

			for _, value := range values {
				file_url := value

				task.AddTask(func() (string, error) {
					InfoLogger.Println("handle url:", file_url)
					//分析url路径
					urlInfo, err := url.Parse(file_url)
					if err != nil {
						ErrLogger.Println(err)
						return file_url, nil
					}
					//判定文件后缀是否pdf
					if !strings.EqualFold(strings.ToLower(filepath.Ext(urlInfo.Path)), ".pdf") {
						htmlpdf := NewHTMLPDF(s.config)
						return htmlpdf.BuildFromLink(file_url)
					}
					return file_url, nil
				})
			}

			task.TaskDone(func(list []*TaskResult) {
				InfoLogger.Println("task list:", list)

				for _, item := range list {
					input_files[item.Index] = item.File

					if item.Err != nil {
						http.Error(writer, item.Err.Error(), 500)
						return
					}
				}
			})

			d := NewDownloader(input_files, s.config.TempPath)

			d.Start()
			d.Done(func(list []string) {
				htmlpdf := NewHTMLPDF(s.config)
				savePath, err := htmlpdf.PDFTK_Combine(list)
				if err != nil {
					http.Error(writer, err.Error(), 500)
					return
				}

				download, err := os.Open(savePath)
				if err != nil {
					http.Error(writer, err.Error(), 500)
					return
				}

				writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.pdf", time.Now().UnixNano()))
				writer.Header().Set("Content-Type", "application/pdf")
				_, err = io.Copy(writer, download)
				if err != nil {
					http.Error(writer, err.Error(), 500)
					return
				}

				defer download.Close()

				defer time.AfterFunc(time.Second*10, func() {

					for _, item := range list {
						if !strings.Contains(item, "/cache/") {
							os.Remove(item)
						}
					}
					os.Remove(savePath)
				})
			})

			break
		}
	}
}

func (s *HTTPService) COMBINE(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		http.Error(writer, err.Error(), 500)
	}

	savePath := filepath.Join(s.config.TempPath, fmt.Sprintf("%d", rand.Int())+".pdf")

	for key, values := range request.PostForm {
		if strings.EqualFold(key, "file") {
			d := NewDownloader(values, s.config.TempPath)

			InfoLogger.Println(values)

			d.Start()
			d.Done(func(list []string) {

				CombinePDF(list, savePath)

				download, err := os.Open(savePath)
				if err != nil {
					http.Error(writer, err.Error(), 500)
					return
				}

				writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.pdf", time.Now().UnixNano()))
				writer.Header().Set("Content-Type", "application/pdf")
				_, err = io.Copy(writer, download)
				if err != nil {
					http.Error(writer, err.Error(), 500)
					return
				}

				defer download.Close()

				defer time.AfterFunc(time.Second*10, func() {

					for _, item := range list {
						if !strings.Contains(item, "/cache/") {
							os.Remove(item)
						}
					}
					os.Remove(savePath)
				})
			})

			break
		}
	}
}
