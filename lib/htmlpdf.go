package lib

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

type TaskResult struct {
	File  string
	Err   error
	Index int
}

type Task struct {
	taskJob   chan *TaskResult
	taskCount int
}

type HTMLPDF struct {
	config   *Config
	buildJob chan bool
}

var HTMLPDF_INSTANCE *HTMLPDF

func NewHTMLPDF(conf *Config) *HTMLPDF {
	if HTMLPDF_INSTANCE != nil {
		return HTMLPDF_INSTANCE
	}

	HTMLPDF_INSTANCE = &HTMLPDF{
		config:   conf,
		buildJob: make(chan bool, conf.Worker),
	}

	return HTMLPDF_INSTANCE
}

func NewTask(worker int) *Task {
	return &Task{
		taskJob:   make(chan *TaskResult, worker),
		taskCount: 0,
	}
}

func (t *Task) AddTask(task func() (string, error)) {
	go func(index int) {
		file, err := task()
		t.taskJob <- &TaskResult{
			File:  file,
			Err:   err,
			Index: index,
		}
	}(t.taskCount)
	t.taskCount++
}

func (t *Task) TaskDone(callback func([]*TaskResult)) {
	count := 0
	list := make([]*TaskResult, t.taskCount)
	for {
		result := <-t.taskJob
		list[count] = result
		count++
		if count >= t.taskCount {
			break
		}
	}

	close(t.taskJob)

	callback(list)
}

func (pdf *HTMLPDF) run(source_path string, pdf_path string) error {
	pdf.buildJob <- true
	Logger.Infof("current html2pdf job count:%d\n", len(pdf.buildJob))

	source_path = filepath.ToSlash(source_path)
	bin_args := append(pdf.config.WebKitArgs, source_path, pdf_path)
	cmd := exec.Command(pdf.config.WebKitBin, bin_args...)
	var outbuffer bytes.Buffer
	var errbuffer bytes.Buffer
	cmd.Stderr = &outbuffer
	cmd.Stderr = &errbuffer
	Logger.Debugf("raw command line: %s\n", cmd)

	done := make(chan error, 1)

	var err error
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err = <-done:
		Logger.Info("Done")
	case <-time.After(time.Second * time.Duration(pdf.config.Timeout)):
		cmd.Process.Kill()
		err = errors.New("timeout!")
	}

	<-pdf.buildJob
	if err != nil {
		Logger.Error(err)
		Logger.Error(errbuffer.String())
		return err
	}
	return nil
}

func (pdf *HTMLPDF) BuildFromLink(link string) (local_pdf string, err error) {
	pdf_name := fmt.Sprintf("%s.pdf", MakeUUID())
	pdf_name = path.Join(pdf.config.TempPath, pdf_name)

	err = pdf.run(link, pdf_name)
	if err != nil {
		return "", err
	}
	return pdf_name, nil
}

func (pdf *HTMLPDF) BuildFromSource(html []byte) (local_pdf string, err error) {
	tmp_name := fmt.Sprintf("%s.html", MakeUUID())
	tmp_name = path.Join(pdf.config.TempPath, tmp_name)

	err = os.WriteFile(tmp_name, html, os.ModePerm)
	if err != nil {
		return
	}
	//defer os.Remove(tmp_name)

	pdf_name := fmt.Sprintf("%s.pdf", MakeUUID())
	pdf_name = path.Join(pdf.config.TempPath, pdf_name)

	tmp_name = fmt.Sprintf("file:///%s", tmp_name)

	err = pdf.run(tmp_name, pdf_name)
	if err != nil {
		return "", err
	}

	return pdf_name, nil
}

func (pdf *HTMLPDF) PDFTK_Combine(files []string) (dest_pdf_path string, err error) {
	pdf_name := fmt.Sprintf("%s.pdf", MakeUUID())
	pdf_name = path.Join(pdf.config.TempPath, pdf_name)

	err = CombinePDF(files, pdf_name)
	if err != nil {
		return pdf_name, err
	}
	return pdf_name, nil
}
