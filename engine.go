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
	"strings"

	"github.com/jcomo/banana/copy"
)

var (
	postsDir   = "posts"
	pagesDir   = "pages"
	layoutsDir = "layouts"
	staticDir  = "static"

	configName = "banana.yml"

	indexTemplateName = "index.tmpl"
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
	baseTemplate := ""
	templatePath := layout
	templates := make([]string, 0)

	for {
		f, err := os.Open(templatePath)
		if err != nil {
			return nil, err
		}

		defer f.Close()

		// TODO: split out frontmatter parsing
		fm, content, err := parse(f)
		if err != nil {
			return nil, err
		}

		// TODO: check for infinite loop
		templates = append(templates, string(content))
		if fm == nil || fm.Layout == "" {
			// At the base template for this render
			b := path.Base(templatePath)
			baseTemplate = strings.TrimSuffix(b, filepath.Ext(b))
			break
		}

		templatePath = e.layoutPath(fm.Layout)
	}

	t := template.New(baseTemplate).Funcs(funcMap)
	for _, text := range templates {
		_, err := t.Parse(text)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
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

func (e *Engine) writePage(p *Page) error {
	t, err := e.Template(e.layoutPath(p.Layout))
	if err != nil {
		return err
	}

	dir := path.Join(e.outDir, p.Slug())
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// TODO: check for naming collisions
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
		err = e.writePage(p)
		if err != nil {
			return err
		}
	}

	err = filepath.Walk(
		e.path(pagesDir),
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if err != nil {
				return err
			}

			p, err := ParsePage(path)
			if err != nil {
				return err
			}

			return e.writePage(p)
		})

	if err != nil {
		return err
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
