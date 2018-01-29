package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"

	"gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
)

type Site struct {
	Name    string
	Content template.HTML
}

func upper(value string) string {
	return value + "SDIANSDKANS"
}

type Post struct {
	FrontMatter
	Content []byte
}

type FrontMatter struct {
	Layout string `yaml:"layout"`
	Title  string `yaml:"title"`
}

type FrontMatterParser interface {
	IsStart(string) bool
	IsEnd(string) bool
	Parse([]byte) (*FrontMatter, error)
}

type YAMLFrontMatterParser struct{}

func (p *YAMLFrontMatterParser) IsStart(line string) bool {
	return line == "---"
}

func (p *YAMLFrontMatterParser) IsEnd(line string) bool {
	return line == "---"
}

func (p *YAMLFrontMatterParser) Parse(raw []byte) (*FrontMatter, error) {
	meta := new(FrontMatter)
	err := yaml.Unmarshal(raw, meta)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func parse(filename string) (*Post, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(f)
	fm, err := parseFrontMatter(r)
	if err != nil {
		return nil, err
	}

	content, err := parseContent(r)
	if err != nil {
		return nil, err
	}

	return &Post{
		FrontMatter: *fm,
		Content:     content,
	}, nil
}

func parseFrontMatter(r *bufio.Reader) (*FrontMatter, error) {
	p := &YAMLFrontMatterParser{}

	var buf bytes.Buffer
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if !p.IsStart(strings.Trim(line, "\r\n")) {
		return nil, errors.New("Expected start of front matter")
	}

	for {
		line, err = r.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if p.IsEnd(strings.Trim(line, "\r\n")) {
			break
		}

		buf.WriteString(line)
	}

	fm, err := p.Parse(buf.Bytes())
	if err != nil {
		return nil, err
	}

	return fm, nil
}

func parseContent(r *bufio.Reader) ([]byte, error) {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type GlobalContext struct {
	Site  *SiteContext
	Post  *PostContext
	Posts []*PostContext
}

type SiteContext struct {
	Name string
}

type PostContext struct {
	Title   string
	Content template.HTML
}

func (p *Post) AsContext() *PostContext {
	markdown := blackfriday.Run(p.Content)
	return &PostContext{
		Title:   p.Title,
		Content: template.HTML(markdown),
	}
}

func main() {
	p, err := parse("example/site/posts/example.md")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(p)

	funcs := map[string]interface{}{
		"upper": upper,
	}

	ctx := GlobalContext{
		Site: &SiteContext{Name: "Test Site"},
		Post: p.AsContext(),
		Posts: []*PostContext{
			p.AsContext(),
		},
	}
	t, _ := template.New("page").Funcs(funcs).ParseFiles(
		"example/site/index.tmpl",
		"example/site/content.tmpl",
	)
	err = t.Execute(os.Stdout, ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
}
