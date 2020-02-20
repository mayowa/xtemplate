package template

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v2"
)

// XTemplate ...
type XTemplate struct {
	folder string
	shared *template.Template
	cache  map[string]*template.Template
}

var fmRegex *regexp.Regexp

func init() {
	fmRegex = regexp.MustCompile(`(?s){{/\*\*\*(.*)\*\*\*/}}`)
}

// New create new instance of XTemplate
func New(folder string, fnMap template.FuncMap) *XTemplate {

	xt := new(XTemplate)
	xt.cache = make(map[string]*template.Template)
	xt.folder = folder
	xt.shared = template.New("").Funcs(fnMap)
	return xt
}

// Render ...
func (s *XTemplate) Render(wr io.Writer, name string, data interface{}, cache bool) error {

	var (
		tpl   *template.Template
		found bool
		err   error
	)
	// check for template in cache
	tpl, found = s.cache[name]
	if !found {
		tpl, err = s.getTemplate(name)
		if err != nil {
			return err
		}

		// cache template if requested
		if cache {
			s.cache[name] = tpl
		}
	}

	buff := bytes.NewBuffer([]byte{})
	if err = tpl.Execute(buff, data); err != nil {
		return err
	}

	if _, err = buff.WriteTo(wr); err != nil {
		return err
	}

	return nil
}

type frontMatter struct {
	Master  string   `yaml:"master"`
	Include []string `yaml:"include"`
}

func (s *XTemplate) getTemplate(name string) (*template.Template, error) {
	fle := filepath.Join(s.folder, name)

	// read template into a buffer
	fleContent, err := ioutil.ReadFile(fle)
	if err != nil {
		return nil, err
	}

	// check if template contains frontmatter
	match := fmRegex.FindSubmatch(fleContent)
	if len(match) == 0 {
		// this template doesn't contain frontmatter, create template and return
		tpl, err := s.makeTemplate(name, fleContent)
		if err != nil {
			return nil, err
		}

		return tpl, nil
	}

	var tpl *template.Template

	//
	// sample frontmatter (note the '***')
	//
	// {{/***
	// 	master: base.html
	// 	include:
	// 	  - header.html
	// 	  - footer.html
	// ***/}}

	// extract front matter
	fm := new(frontMatter)
	err = yaml.Unmarshal(match[1], fm)
	if err != nil {
		return nil, err
	}

	if len(fm.Master) > 0 {
		// get the master template if requested
		master, err := s.getTemplate(fm.Master)
		if err != nil {
			return nil, err
		}

		// have the master template use this template as an overlay
		_, err = master.Parse(string(fleContent))
		if err != nil {
			return nil, err
		}

		tpl = master
	} else {
		// create template
		tpl, err = s.makeTemplate(name, fleContent)
		if err != nil {
			return nil, err
		}
	}

	// parse includes if requested
	for i := range fm.Include {
		fm.Include[i] = filepath.Join(s.folder, fm.Include[i])
	}
	_, err = tpl.ParseFiles(fm.Include...)
	if err != nil {
		return nil, err
	}

	return tpl, nil
}

func (s *XTemplate) makeTemplate(name string, content []byte) (*template.Template, error) {
	tpl, err := s.shared.Clone()
	if err != nil {
		return nil, err
	}

	return tpl.New(name).Parse(string(content))
}
