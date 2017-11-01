// listsmemo project main.go
package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"
	"gopkg.in/xmlpath.v2"
)

func DownloadFile(d int, f int) (htmlContent string, notFound bool, err error) {
	notFound = false
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	file := "http://lists.memo.ru/d" + strconv.Itoa(d) + "/f" + strconv.Itoa(f) + ".htm"
	log.Println("Downloading", file)
	response, err := client.Get(file)
	if err != nil {
		return
	}
	if response.StatusCode == 404 {
		notFound = true
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	decoder := charmap.Windows1251.NewDecoder()
	newBody := make([]byte, len(body)*2)
	n, _, err := decoder.Transform(newBody, body, false)
	if err != nil {
		return
	}
	newBody = newBody[:n]
	utf8Body := string(newBody)
	reader := strings.NewReader(utf8Body)
	root, err := html.Parse(reader)
	if err != nil {
		return
	}
	var b bytes.Buffer
	html.Render(&b, root)
	htmlContent = b.String()
	htmlContent = strings.Replace(htmlContent, "windows-1251", "utf-8", 1)
	return
}

func HtmlToText(htmlContent string) (textContent string, err error) {
	reader := strings.NewReader(htmlContent)
	xmlroot, err := xmlpath.ParseHTML(reader)
	if err != nil {
		return
	}
	var xpath string
	xpath = `//ul[@class='list-right']/li`
	path := xmlpath.MustCompile(xpath)
	iter := path.Iter(xmlroot)
	var buffer bytes.Buffer
	for iter.Next() {
		buffer.WriteString(iter.Node().String())
		buffer.WriteString("\n\n")
	}
	textContent = buffer.String()
	return
}

func Download(dStart, dEnd int) {
	d := dStart
	f := 1
	log.Println("begin")

	for d <= dEnd {
		html, notFound, err := DownloadFile(d, f)
		if err != nil {
			log.Fatalln(err)
		}
		if notFound {
			if f != 1 {
				d++
				f = 1
				continue
			} else {
				break
			}
		}
		htmlFile := "d" + strconv.Itoa(d) + "-f" + strconv.Itoa(f) + ".html"
		err = ioutil.WriteFile(htmlFile, []byte(html), 0644)
		if err != nil {
			log.Fatalln(err)
		}
		text, err := HtmlToText(html)
		if err != nil {
			log.Fatalln(err)
		}
		textFile := "d" + strconv.Itoa(d) + "-f" + strconv.Itoa(f) + ".txt"
		err = ioutil.WriteFile(textFile, []byte(text), 0644)
		if err != nil {
			log.Fatalln(err)
		}
		f++
	}
	log.Println("end")

}

func main() {
	Download(1, 1)
}
