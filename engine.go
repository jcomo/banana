package banana

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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

type Engine struct {
	baseDir string
	outDir  string
	site    *SiteContext
	posts   []*Page
}

func NewEngine() (*Engine, error) {
	dir := "."
	cfg, err := ReadConfig(path.Join(dir, configName))
	if err != nil {
		return nil, err
	}

	return &Engine{
		baseDir: dir,
		outDir:  "_build",
		site:    NewSiteContext(&cfg.Site),
	}, nil
}

func (e *Engine) context(p *Page) GlobalContext {
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

func (e *Engine) path(name string) string {
	return path.Join(e.baseDir, name)
}

func (e *Engine) postPath(name string) string {
	return path.Join(e.baseDir, postsDir, name)
}

func (e *Engine) layoutPath(name string) string {
	return path.Join(e.baseDir, layoutsDir, name+".tmpl")
}

func (e *Engine) Template(layout string) (*template.Template, error) {
	return template.New("main").Funcs(funcMap).ParseFiles(
		e.path(layoutTemplateName),
		layout,
	)
}

func (e *Engine) readPosts() error {
	fs, err := ioutil.ReadDir(e.path(postsDir))
	if err != nil {
		return err
	}

	ps := make([]*Page, len(fs))
	for i, f := range fs {
		p, err := ParsePage(e.postPath(f.Name()))
		if err != nil {
			return err
		}

		ps[i] = p
	}

	e.posts = ps
	return nil
}

func (e *Engine) writeIndex() error {
	t, err := e.Template(e.path(indexTemplateName))
	if err != nil {
		return err
	}

	err = os.MkdirAll(e.outDir, 0755)
	if err != nil {
		return err
	}

	path := path.Join(e.outDir, "index.html")
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

func (e *Engine) writePost(p *Page) error {
	t, err := e.Template(e.layoutPath(p.Layout))
	if err != nil {
		return err
	}

	dir := path.Join(e.outDir, p.Slug())
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

func (e *Engine) writeStaticFiles() error {
	dst := path.Join(e.outDir, staticDir)
	err := os.RemoveAll(dst)
	if err != nil {
		return nil
	}

	return copy.Dir(e.path(staticDir), dst)
}

func (e *Engine) Build() error {
	err := e.readPosts()
	if err != nil {
		return err
	}

	err = e.writeIndex()
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

func (e *Engine) Clean() error {
	return os.RemoveAll(e.outDir)
}

func (e *Engine) Watch() (io.Closer, error) {
	dirs := []string{
		e.baseDir,
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

func (e *Engine) OnChange() error {
	log.Println("Change detected. Rebuilding...")
	err := e.Build()
	if err != nil {
		return err
	}

	log.Println("Rebuild complete")
	return nil
}

func (e *Engine) Serve(addr string) error {
	handler := http.FileServer(http.Dir(e.outDir))
	return http.ListenAndServe(addr, withAccessLog(handler))
}
