// listsmemo project main.go
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"
	"gopkg.in/xmlpath.v2"
)

func main() {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	response, err := client.Get("http://lists.memo.ru/d36/f85.htm")
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	decoder := charmap.Windows1251.NewDecoder()
	newBody := make([]byte, len(body)*2)
	n, _, err := decoder.Transform(newBody, body, false)
	if err != nil {
		log.Fatalln(err)
	}
	newBody = newBody[:n]
	utf8Body := string(newBody)
	fmt.Println(utf8Body)
	reader := strings.NewReader(utf8Body)
	root, err := html.Parse(reader)
	if err != nil {
		log.Fatal(err)
	}
	var b bytes.Buffer
	html.Render(&b, root)
	fixedHtml := b.String()
	reader = strings.NewReader(fixedHtml)
	xmlroot, xmlerr := xmlpath.ParseHTML(reader)
	if xmlerr != nil {
		log.Fatal(xmlerr)
	}
	var xpath string
	xpath = `//ul[@class='list-right']/li`
	path := xmlpath.MustCompile(xpath)
	iter := path.Iter(xmlroot)
	for iter.Next() {
		log.Println(iter.Node())
		log.Println("---------")
	}
	//if value, ok := path.String(xmlroot); ok {
	//	log.Println("Found:", value)
	//}
	log.Println("the end")
}
