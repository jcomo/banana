package banana

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/jcomo/banana/copy"
)

type engine struct {
	BaseDir string
	OutDir  string
	site    *SiteContext
	posts   []*Page
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
		p, err := ParsePage(path.Join(postsDir, f.Name()))
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
