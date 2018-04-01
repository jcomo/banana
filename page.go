package banana

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	parsers map[string]FrontMatterParser
)

type Page struct {
	FrontMatter
	Content []byte
}

func (p *Page) Slug() string {
	if p.Permalink != "" {
		return strings.Trim(p.Permalink, "/")
	}

	re := regexp.MustCompile("[\\W]+")
	title := strings.ToLower(p.Title)
	return re.ReplaceAllLiteralString(title, "-")
}

type FrontMatter struct {
	Layout    string     `yaml:"layout"`
	Permalink string     `yaml:"permalink"`
	Title     string     `yaml:"title"`
	Author    string     `yaml:"author"`
	Meta      string     `yaml:"meta"`
	Date      *time.Time `yaml:"date"`
}

type FrontMatterParser interface {
	Parse(raw []byte) (*FrontMatter, error)
}

type YAMLFrontMatterParser struct{}

func (p *YAMLFrontMatterParser) Parse(raw []byte) (*FrontMatter, error) {
	meta := new(FrontMatter)
	err := yaml.Unmarshal(raw, meta)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func ParsePage(filename string) (*Page, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	fm, content, err := parse(f)
	if err != nil {
		return nil, err
	}

	return &Page{
		FrontMatter: *fm,
		Content:     content,
	}, nil
}

func parse(r io.ReadSeeker) (*FrontMatter, []byte, error) {
	var buf bytes.Buffer
	br := bufio.NewReader(r)

	// Start by parsing the front matter
	line, err := br.ReadString('\n')
	if err != nil {
		return nil, nil, err
	}

	delimiter := strings.Trim(line, "\r\n")
	parser, ok := parsers[delimiter]
	if !ok {
		// No front matter, rewind and read the whole thing
		r.Seek(0, io.SeekStart)
		content, err := ioutil.ReadAll(r)
		return nil, content, err
	}

	for {
		line, err = br.ReadString('\n')
		if err != nil {
			return nil, nil, err
		}

		if strings.Trim(line, "\r\n") == delimiter {
			break
		}

		buf.WriteString(line)
	}

	fm, err := parser.Parse(buf.Bytes())
	if err != nil {
		return nil, nil, err
	}

	// Now parse the content after the front matter
	content, err := ioutil.ReadAll(br)
	if err != nil {
		return nil, nil, err
	}

	return fm, content, nil
}

func init() {
	parsers = map[string]FrontMatterParser{
		"---": &YAMLFrontMatterParser{},
	}
}
