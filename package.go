package articles

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/golang-commonmark/markdown"
)

const homeTempl = `
<html>
<head>
	<title>Sheki articles of interest</title>
</head>
<body>
	<h1>Articles of interest</h1>
	{{range .}}
		<p>
			<a href="{{.Link}}">{{.Header}}</a>
		</p>
	{{end}}
</body>
</html>`

const pageTempl = `
<html>
  <head>
	  <title>Sheki articles of interest</title>
	</head>
	<body>
    <p>
	  {{.Content}}
		</p>
		<p>Date: {{.StringDate}}</p>
		<p>
		{{range .Tags}}
		  <a href="tag/{{.}}.html">{{.}}</a>
		{{end}}
		</p>
	</body>
</html>
`

const tagTempl = `
<html>
<head>
	<title>Sheki articles of interest</title>
</head>
<body>
<h1>Tag: {{.Tag}}</h1>
	{{range .Articles}}
		<p>
			<a href="/{{.Index}}.html">{{.Header}}</a>
		</p>
	{{end}}
</body>
</html>`

// Represents an article
type Article struct {
	Header  string
	Content string
	Tags    []string
	Date    time.Time
	Index   int
}

type pageArticle struct {
	Content    template.HTML
	StringDate string
	Tags       []string
}

type renderArticle struct {
	Header string
	Link   string
}

func prepareForRender(articles []Article) []renderArticle {
	var res []renderArticle
	for _, v := range articles {
		r := renderArticle{
			Header: v.Header,
			Link:   fmt.Sprintf("%d.html", v.Index),
		}
		res = append(res, r)
	}
	return res
}

func indexFile(base string) string {
	return path.Join(base, "index.html")
}

func articleFile(base string, index int) string {
	return path.Join(base, fmt.Sprintf("%d.html", index))
}

func generateArticlePage(base string, article Article) error {
	t, err := template.New("article").Parse(pageTempl)

	if err != nil {
		return err
	}
	f, err := os.OpenFile(
		articleFile(base, article.Index),
		os.O_RDWR|os.O_CREATE,
		0644,
	)

	if err != nil {
		return err
	}
	defer f.Close()

	p := pageArticle{
		Content:    template.HTML(article.Content),
		StringDate: article.Date.Format("2 Jan 2006"),
		Tags:       article.Tags,
	}

	return t.Execute(f, p)

}

func generateArticlePages(base string, articles []Article) error {
	for _, a := range articles {
		if err := generateArticlePage(base, a); err != nil {
			return err
		}
	}
	return nil
}

func tagFile(base, tag string) string {
	return path.Join(base, "tag", fmt.Sprintf("%s.html", tag))
}

func generateTagFile(base string, tag string, articles []Article) error {
	t, err := template.New("tags").Parse(tagTempl)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(tagFile(base, tag), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	render := struct {
		Articles []Article
		Tag      string
	}{Articles: articles, Tag: tag}
	return t.Execute(f, render)
}

func generateTags(base string, articles []Article) error {
	tagDir := path.Join(base, "tag")
	if _, err := os.Stat(tagDir); os.IsNotExist(err) {
		mkdirErr := os.Mkdir(tagDir, 0700)
		if mkdirErr != nil {
			return mkdirErr
		}
	}

	tagMap := make(map[string][]Article)
	for _, v := range articles {
		for _, tag := range v.Tags {
			tagMap[tag] = append(tagMap[tag], v)
		}
	}

	f, err := os.OpenFile(indexFile(base), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for k, v := range tagMap {
		if err := generateTagFile(base, k, v); err != nil {
			return err
		}
	}
	return nil

}

func generateIndex(base string, articles []Article) error {
	t, err := template.New("webpage").Parse(homeTempl)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(indexFile(base), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, prepareForRender(articles))
}

func Generate(notePath string, baseDir string) error {
	arr, err := parseFile(notePath)
	if err != nil {
		return err
	}

	if err := generateIndex(baseDir, arr); err != nil {
		return err
	}

	if err := generateArticlePages(baseDir, arr); err != nil {
		return err
	}

	return generateTags(baseDir, arr)

}

// Parses a File into various articles
func parseFile(path string) ([]Article, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var res []Article

	last := ""
	var lines []string
	index := 1
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "# ") {
			if last != "" {
				a, err := parseArticle(last, lines)
				a.Index = index
				index++
				res = append(res, a)
				if err != nil {
					return nil, err
				}
				lines = nil
			}
			last = text
			continue

		}

		lines = append(lines, text)

	}
	if last != "" {
		a, err := parseArticle(last, lines)
		if err != nil {
			return nil, err
		}

		a.Index = index
		index++
		res = append(res, a)

	}
	return reverse(res), nil
}

func reverse(res []Article) []Article {
	n := make([]Article, len(res))
	for i := len(res) - 1; i >= 0; i-- {
		n = append(n, res[i])
	}
	return n
}

func parseArticle(header string, lines []string) (Article, error) {
	h := strings.TrimLeft(header, "# ")
	h = strings.TrimSpace(h)
	var markDown = []string{
		header,
	}
	a := Article{
		Header: h,
	}

	dateLine := ""
	tagsLine := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "Date:") {
			dateLine = line
			continue
		}
		if strings.HasPrefix(line, "Tags:") {
			tagsLine = line
			continue
		}
		markDown = append(markDown, line)
	}

	dateLine = strings.TrimLeft(dateLine, "Date:")
	dateLine = strings.TrimSpace(dateLine)
	var err error
	a.Date, err = time.Parse("2006/01/02", dateLine)
	if err != nil {
		return Article{}, err
	}

	tagsLine = strings.TrimLeft(tagsLine, "Tags:")
	for _, tag := range strings.Split(tagsLine, ",") {
		a.Tags = append(a.Tags, strings.TrimSpace(tag))
	}

	md := markdown.New(markdown.HTML(true), markdown.Nofollow(true))
	a.Content = md.RenderToString([]byte(strings.Join(markDown, "\n")))
	return a, nil
}
