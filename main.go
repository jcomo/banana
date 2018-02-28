package banana

import (
	"bufio"
	"bytes"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jcomo/banana/copy"
	"gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
)

type Site struct {
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description"`
	Author      string                 `yaml:"author"`
	Vars        map[string]interface{} `yaml:"vars"`
}

type Config struct {
	Site Site `yaml:"site"`
}

type Page struct {
	FrontMatter
	Content []byte
}

func (p *Page) Slug() string {
	re := regexp.MustCompile("[\\W]+")
	title := strings.ToLower(p.Title)
	return re.ReplaceAllLiteralString(title, "-")
}

type FrontMatter struct {
	Layout string     `yaml:"layout"`
	Title  string     `yaml:"title"`
	Author string     `yaml:"author"`
	Meta   string     `yaml:"meta"`
	Date   *time.Time `yaml:"date"`
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

func parse(filename string) (*Page, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	r := bufio.NewReader(f)
	fm, err := parseFrontMatter(r)
	if err != nil {
		return nil, err
	}

	content, err := parseContent(r)
	if err != nil {
		return nil, err
	}

	return &Page{
		FrontMatter: *fm,
		Content:     content,
	}, nil
}

func parseFrontMatter(r *bufio.Reader) (*FrontMatter, error) {
	var buf bytes.Buffer
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	delimiter := strings.Trim(line, "\r\n")
	parser, ok := parsers[delimiter]
	if !ok {
		return nil, errors.New("Invalid front matter delimiter")
	}

	for {
		line, err = r.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if strings.Trim(line, "\r\n") == delimiter {
			break
		}

		buf.WriteString(line)
	}

	fm, err := parser.Parse(buf.Bytes())
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
	Page  *PageContext
	Posts []*PageContext
}

type SiteContext struct {
	Title       string
	Description string
	Author      string
	Time        time.Time
	Vars        map[string]interface{}
}

type PageContext struct {
	URL     string
	Slug    string
	Date    *time.Time
	Title   string
	Author  string
	Meta    string
	Content template.HTML
}

func NewSiteContext(s *Site) *SiteContext {
	return &SiteContext{
		Title:       s.Title,
		Description: s.Description,
		Author:      s.Author,
		Vars:        s.Vars,
		Time:        time.Now(),
	}
}

func NewPageContext(p *Page) *PageContext {
	slug := p.Slug()
	markdown := blackfriday.Run(p.Content)
	return &PageContext{
		URL:     "/" + slug,
		Slug:    slug,
		Title:   p.Title,
		Date:    p.Date,
		Content: template.HTML(markdown),
	}
}

var parsers map[string]FrontMatterParser

func init() {
	parsers = map[string]FrontMatterParser{
		"---": &YAMLFrontMatterParser{},
	}
}

type engine struct {
	BaseDir string
	OutDir  string
	site    *SiteContext
	posts   []*Page
}

func readConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	cfg := new(Config)
	err = yaml.NewDecoder(f).Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func NewEngine() (*engine, error) {
	dir := "."
	postsDir := path.Join(dir, "posts")

	cfg, err := readConfig(path.Join(dir, "banana.yml"))
	if err != nil {
		return nil, err
	}

	fs, err := ioutil.ReadDir(postsDir)
	if err != nil {
		return nil, err
	}

	ps := make([]*Page, len(fs))
	for i, f := range fs {
		p, err := parse(path.Join(postsDir, f.Name()))
		if err != nil {
			return nil, err
		}

		ps[i] = p
	}

	return &engine{
		BaseDir: dir,
		OutDir:  "_build",
		site:    NewSiteContext(&cfg.Site),
		posts:   ps,
	}, nil
}

func (e *engine) context(p *Page) GlobalContext {
	pcs := make([]*PageContext, len(e.posts))
	for i, p := range e.posts {
		pcs[i] = NewPageContext(p)
	}
	sort.Sort(pagesByDate(pcs))

	var pageContext *PageContext
	if p != nil {
		pageContext = NewPageContext(p)
	}

	return GlobalContext{
		Site:  e.site,
		Posts: pcs,
		Page:  pageContext,
	}
}

func (e *engine) path(name string) string {
	return path.Join(e.BaseDir, name)
}

func (e *engine) postPath(name string) string {
	return path.Join(e.BaseDir, "posts", name)
}

func (e *engine) layoutPath(name string) string {
	return path.Join(e.BaseDir, "layouts", name+".tmpl")
}

func (e *engine) Template(layout string) (*template.Template, error) {
	return template.New("main").Funcs(funcMap).ParseFiles(
		e.path("layout.tmpl"),
		layout,
	)
}

func (e *engine) writeIndex() error {
	t, err := e.Template(e.path("index.tmpl"))
	if err != nil {
		return err
	}

	err = os.MkdirAll(e.OutDir, 0755)
	if err != nil {
		return err
	}

	path := path.Join(e.OutDir, "index.html")
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	err = t.Execute(f, e.context(nil))
	if err != nil {
		return err
	}

	return nil
}

func (e *engine) writePost(p *Page) error {
	t, err := e.Template(e.layoutPath(p.Layout))
	if err != nil {
		return err
	}

	dir := path.Join(e.OutDir, p.Slug())
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(dir, "index.html"))
	if err != nil {
		return err
	}

	defer f.Close()

	err = t.Execute(f, e.context(p))
	if err != nil {
		return err
	}

	return nil
}

func (e *engine) writeStaticFiles() error {
	dst := path.Join(e.OutDir, "static")
	err := os.RemoveAll(dst)
	if err != nil {
		return nil
	}

	return copy.Dir(e.path("static"), dst)
}

func (e *engine) Build() error {
	err := e.writeIndex()
	if err != nil {
		return err
	}

	for _, p := range e.posts {
		err = e.writePost(p)
		if err != nil {
			return err
		}
	}

	err = e.writeStaticFiles()
	if err != nil {
		return err
	}

	return nil
}

func (e *engine) Clean() error {
	return os.RemoveAll(e.OutDir)
}

func (e *engine) Watch() (io.Closer, error) {
	dirs := []string{
		e.BaseDir,
		path.Join(e.BaseDir, "layouts"),
		path.Join(e.BaseDir, "posts"),
		path.Join(e.BaseDir, "pages"),
	}

	staticDir := path.Join(e.BaseDir, "static")
	filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		dirs = append(dirs, path)
		return nil
	})

	return StartWatching(dirs, e)
}

func (e *engine) OnChange() error {
	log.Println("Change detected. Rebuilding...")
	err := e.Build()
	if err != nil {
		return err
	}

	log.Println("Rebuild complete")
	return nil
}
