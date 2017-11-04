// listsmemo project main.go
package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"
	"gopkg.in/xmlpath.v2"
)

type Person struct {
	name   string
	d      int
	f      int
	n      string
	author string
	text   string
}

type Downloader struct {
	parser *Parser
}

type Parser struct {
	nPath      *xmlpath.Path
	namePath   *xmlpath.Path
	authorPath *xmlpath.Path
	textPath   *xmlpath.Path
}

func NewParser() *Parser {
	parser := &Parser{}
	parser.nPath = xmlpath.MustCompile(`p/a/@name`)
	parser.namePath = xmlpath.MustCompile(`p[@class="name"]`)
	parser.authorPath = xmlpath.MustCompile(`p[@class="author"]`)
	parser.textPath = xmlpath.MustCompile(`p[@class="cont"]`)
	return parser
}

func (parser *Parser) ParsePerson(node *xmlpath.Node, person *Person) {
	person.n, _ = parser.nPath.String(node)
	person.name, _ = parser.namePath.String(node)
	person.author, _ = parser.authorPath.String(node)
	person.text, _ = parser.textPath.String(node)
	person.text = strings.Replace(person.text, "\n\n\n", "\n", -1)
	person.text = strings.Replace(person.text, "\n\n", "\n", -1)
}

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

func ParsePerson(node *xmlpath.Node) (name, n, author, text string) {
	aPath := xmlpath.MustCompile(`p/a/@name`)
	namePath := xmlpath.MustCompile(`p[@class="name"]`)
	authorPath := xmlpath.MustCompile(`p[@class="author"]`)
	textPath := xmlpath.MustCompile(`p[@class="cont"]`)
	n, _ = aPath.String(node)
	name, _ = namePath.String(node)
	author, _ = authorPath.String(node)
	text, _ = textPath.String(node)
	return
}

func HtmlToText(htmlContent string) (persons []Person, textContent string, err error) {
	parser := NewParser()
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
		node := iter.Node()
		person := Person{}
		parser.ParsePerson(node, &person)
		persons = append(persons, person)
		buffer.WriteString(node.String())
		buffer.WriteString("\n\n")
	}
	textContent = buffer.String()
	//log.Println(persons)
	return
}

func OpenDataFile(d int) *os.File {
	htmlFile := "d" + strconv.Itoa(d) + ".html"
	file, err := os.Create(htmlFile)
	if err != nil {
		log.Fatalln(err)
	}
	return file
}

func Download(dStart, dEnd int) {
	var allPersons []Person
	d := dStart
	f := 1
	log.Println("begin")
	var file *os.File
	for d <= dEnd {
		if file == nil {
			file = OpenDataFile(d)
		}
		html, notFound, err := DownloadFile(d, f)
		if err != nil {
			log.Fatalln(err)
		}
		if notFound {
			if f != 1 {
				if file != nil {
					file.Close()
					file = nil
				}
				d++
				f = 1
				continue
			} else {
				break
			}
		}
		//htmlFile := "d" + strconv.Itoa(d) + "-f" + strconv.Itoa(f) + ".html"
		_, err = file.WriteString(html)
		if err != nil {
			log.Fatalln(err)
		}
		convertToText := false

		//err = ioutil.WriteFile(htmlFile, []byte(html), 0644)
		//if err != nil {
		//	log.Fatalln(err)
		//}
		if convertToText {
			persons, text, err := HtmlToText(html)
			if err != nil {
				log.Fatalln(err)
			}
			allPersons = append(allPersons, persons...)
			textFile := "d" + strconv.Itoa(d) + "-f" + strconv.Itoa(f) + ".txt"
			err = ioutil.WriteFile(textFile, []byte(text), 0644)
			if err != nil {
				log.Fatalln(err)
			}
		}
		f++
	}
	if len(allPersons) > 0 {
		lines := strings.Split(allPersons[0].text, "\n")
		log.Print(strings.Join(lines, "\n"))
		log.Println(len(lines), lines)
	}
	log.Println("end")
}

func main() {
	Download(36, 37)
}
