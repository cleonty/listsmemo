// listsmemo project main.go
package main

import (
	_ "bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "golang.org/x/net/html"
	_ "golang.org/x/text/encoding/charmap"
	"gopkg.in/xmlpath.v2"
)

func DownloadFile(d string, f string) (htmlContent string, notFound bool, err error) {
	notFound = false
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	link := "http://lists.memo.ru/" + d + "/" + f + ".htm"
	log.Println("Downloading", link)
	response, err := client.Get(link)
	if err != nil {
		return
	}
	if response.StatusCode == 404 {
		notFound = true
		log.Println("File ", link, "not found")
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	sBody := string(body)
	bodyStart := strings.Index(sBody, `<ul class="list-right">`) + len(`<ul class="list-right">`)
	bodyEnd := strings.Index(sBody, "</ul>")
	htmlContent = sBody[bodyStart:bodyEnd]
	return
}

func OpenHtmlDataFile(dtext string) *os.File {
	htmlFile := dtext + ".html"
	file, err := os.Create(htmlFile)
	if err != nil {
		log.Fatalln(err)
	}
	err = WriteHtmlDataFileHeader(file)
	if err != nil {
		log.Fatalln(err)
	}
	return file
}

func OpenTextDataFile(dtext string) *os.File {
	htmlFile := dtext + ".txt"
	file, err := os.Create(htmlFile)
	if err != nil {
		log.Fatalln(err)
	}
	return file
}

func CloseHtmlDataFile(file *os.File) *os.File {
	if file != nil {
		defer file.Close()
		err := WriteHtmlDataFileTrailer(file)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return nil
}

func CloseTextDataFile(file *os.File) *os.File {
	if file != nil {
		file.Close()
	}
	return nil
}

func HtmlToText(html string) string {
	reader := strings.NewReader(html)
	root, err := xmlpath.ParseHTML(reader)
	if err != nil {
		log.Fatalln(err)
	}
	return root.String()
}

func WriteHtmlDataFileHeader(file *os.File) error {
	header := strings.Join([]string{
		`<html>`,
		`<head>`,
		`  <meta http-equiv="Content-Type" content="text/html; charset=windows-1251">`,
		`  <title>lists.memo.ru</title>`,
		`</head>`,
		`<body>`,
		`<ul>`,
	}, "\n")
	_, err := file.WriteString(header)
	return err
}

func WriteHtmlDataFileTrailer(file *os.File) error {
	header := strings.Join([]string{
		`</ul>`,
		`</body>`,
		`</html>`,
	}, "\n")
	_, err := file.WriteString(header)
	return err
}

func DownloadMainDatabase(dStart, dEnd int) {
	d := dStart
	f := 1
	log.Println("begin downloading main database")
	var htmlFile *os.File
	var textFile *os.File
	for d <= dEnd {
		dtext := "d" + strconv.Itoa(d)
		if htmlFile == nil {
			htmlFile = OpenHtmlDataFile(dtext)
		}
		if textFile == nil {
			textFile = OpenTextDataFile(dtext)
		}
		//		if f > 1 {
		//			log.Println("no more than one file")
		//			break
		//		}
		ftext := "f" + strconv.Itoa(f)
		html, notFound, err := DownloadFile(dtext, ftext)
		if err != nil {
			log.Fatalln(err)
		}
		if notFound {
			if f != 1 {
				htmlFile = CloseHtmlDataFile(htmlFile)
				textFile = CloseTextDataFile(textFile)
				d++
				f = 1
				continue
			} else {
				break
			}
		}
		_, err = htmlFile.WriteString(html)
		if err != nil {
			log.Fatalln(err)
		}
		text := HtmlToText(html)
		_, err = textFile.WriteString(text)
		if err != nil {
			log.Fatalln(err)
		}
		f++
	}
	htmlFile = CloseHtmlDataFile(htmlFile)
	log.Println("end downloading main database")
}

func DownloadUpdate() {
	f := 1
	dtext := "dnew"
	log.Println("begin downloading updates")
	htmlFile := OpenHtmlDataFile(dtext)
	textFile := OpenTextDataFile(dtext)
	defer CloseHtmlDataFile(htmlFile)
	defer CloseTextDataFile(htmlFile)
	for {
		//		if f > 1 {
		//			log.Println("no more than one file")
		//			break
		//		}
		ftext := fmt.Sprintf("f%03d", f)
		html, notFound, err := DownloadFile(dtext, ftext)
		if err != nil {
			log.Fatalln(err)
		}
		if notFound {
			break
		}
		_, err = htmlFile.WriteString(html)
		if err != nil {
			log.Fatalln(err)
		}
		text := HtmlToText(html)
		_, err = textFile.WriteString(text)
		if err != nil {
			log.Fatalln(err)
		}
		f++
	}
	log.Println("end downloading updates")
}

func main() {
	var err error
	dMin := 1
	dMax := 1000
	downloadUpdate := true
	args := os.Args
	if len(args) > 1 {
		dMin, err = strconv.Atoi(args[1])
		if err != nil {
			log.Fatalln("First argument must be an integer", err)
		}
	}
	if len(args) > 2 {
		dMax, err = strconv.Atoi(args[2])
		if err != nil {
			log.Fatalln("Second argument must be an integer", err)
		}
	}
	DownloadMainDatabase(dMin, dMax)
	if len(args) > 3 {
		downloadUpdate = 0 == strings.Compare(args[3], "yes")
	}
	if downloadUpdate {
		DownloadUpdate()
	}
}
