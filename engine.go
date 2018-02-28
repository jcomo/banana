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

var (
	postsDir   = "posts"
	pagesDir   = "pages"
	layoutsDir = "layouts"
	staticDir  = "static"

	configName = "banana.yml"

	indexTemplateName  = "index.tmpl"
	layoutTemplateName = "layout.tmpl"
)

type engine struct {
	BaseDir string
	OutDir  string
	site    *SiteContext
	posts   []*Page
}

func NewEngine() (*engine, error) {
	dir := "."
	postsPath := path.Join(dir, postsDir)

	cfg, err := ReadConfig(path.Join(dir, configName))
	if err != nil {
		return nil, err
	}

	fs, err := ioutil.ReadDir(postsPath)
	if err != nil {
		return nil, err
	}

	ps := make([]*Page, len(fs))
	for i, f := range fs {
		p, err := ParsePage(path.Join(postsPath, f.Name()))
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
	return path.Join(e.BaseDir, postsDir, name)
}

func (e *engine) layoutPath(name string) string {
	return path.Join(e.BaseDir, layoutsDir, name+".tmpl")
}

func (e *engine) Template(layout string) (*template.Template, error) {
	return template.New("main").Funcs(funcMap).ParseFiles(
		e.path(layoutTemplateName),
		layout,
	)
}

func (e *engine) writeIndex() error {
	t, err := e.Template(e.path(indexTemplateName))
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
	dst := path.Join(e.OutDir, staticDir)
	err := os.RemoveAll(dst)
	if err != nil {
		return nil
	}

	return copy.Dir(e.path(staticDir), dst)
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
		e.path(layoutsDir),
		e.path(postsDir),
		e.path(pagesDir),
	}

	filepath.Walk(
		e.path(staticDir),
		func(path string, info os.FileInfo, err error) error {
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
