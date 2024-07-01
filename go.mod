module html2pdf

go 1.12

require (
	github.com/chromedp/cdproto v0.0.0-20240202021202-6d0b6a386732
	github.com/chromedp/chromedp v0.9.5
	github.com/gorilla/mux v1.7.1
	github.com/joho/godotenv v1.5.1
	github.com/jung-kurt/gofpdf v1.1.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pdfcpu/pdfcpu v0.2.4
)

replace html2pdf => ./
