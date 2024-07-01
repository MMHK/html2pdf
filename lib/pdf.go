package lib

import (
	"image"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/jung-kurt/gofpdf"

	"github.com/pdfcpu/pdfcpu/pkg/cli"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

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
	Log.Debugf("ConvertToPdf: %s => %s\n", src_image_path, dest_pdf_path)
	src_image, err := os.Open(src_image_path)
	if err != nil {
		Log.Error(err)
		return err
	}
	defer src_image.Close()
	img, _, err := image.Decode(src_image)
	if err != nil {
		Log.Error(err)
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
		Log.Error(err)
		return err
	}

	return nil
}

func CombinePDF(files []string, dest_pdf_path string) error {
	Log.Debugf("CombinePDF: %v => %s\n", files, dest_pdf_path)
	//处理合并过程中可能出现的异常
	defer func() {
		if err := recover(); err != nil {
			Log.Error(err)
		}
	}()

	config := pdfcpu.NewDefaultConfiguration()
	config.ValidationMode = pdfcpu.ValidationNone
	cmd := cli.MergeCommand(files, dest_pdf_path, config)
	_, err := cli.Process(cmd)
	if err != nil {
		Log.Error(err)
		return err
	}

	return nil
}
