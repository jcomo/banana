package banana

import (
	"html/template"
	"time"

	blackfriday "github.com/russross/blackfriday/v2"
)

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
