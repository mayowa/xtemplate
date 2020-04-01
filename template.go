package xtemplate

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// XTemplate ...
type XTemplate struct {
	folder string
	shared *template.Template
	cache  map[string]*template.Template
}

var lneRe *regexp.Regexp
var expRe *regexp.Regexp
var actRe *regexp.Regexp
var tplRe *regexp.Regexp

func init() {
	// {{ ...  }}
	lneRe = regexp.MustCompile(`{{.+?}}`)
	// fnMix("index.html", 2)
	expRe = regexp.MustCompile(`([a-zA-Z]+[0-9]*)\((.+)\)`)
	// {{ extend "index.html" }}
	actRe = regexp.MustCompile(`[[:blank:]]*{{ *(.+?) *\"(.+?)\" *}}[[:blank:]]*[\r\n]*`)
	// {{ template button(123) }}
	tplRe = regexp.MustCompile(`{{[ \t]*template ([a-zA-Z0-9\-_]+) ?\(?([^\}]+?)\)?[ \t]*}}`)
}

// New create new instance of XTemplate
func New(folder string) *XTemplate {

	xt := new(XTemplate)
	xt.cache = make(map[string]*template.Template)
	xt.folder = folder
	xt.shared = template.New("")

	funcs := template.FuncMap{
		"title": capitalize,
		"lower": lower,
		"upper": upper,
		"json":  marshalJSON,
	}

	xt.shared.Funcs(funcs)
	return xt
}

// Delims sets the template delimiters to the specified strings,
// must be called before templates are parsed
func (s *XTemplate) Delims(left, right string) *XTemplate {
	s.shared.Delims(left, right)
	return s
}

// Funcs adds the elements of the argument map to the template's function map.
// must be called before templates are parsed
func (s *XTemplate) Funcs(funcMap template.FuncMap) *XTemplate {
	s.shared.Funcs(funcMap)
	return s
}

// Lookup returns the template with the given name in the cache
func (s *XTemplate) Lookup(name string) *template.Template {
	if _, found := s.cache[name]; !found {
		return nil
	}

	return s.cache[name]
}

// ParseFile ...
func (s *XTemplate) ParseFile(name string) error {
	// parse template
	tpl, err := s.getTemplate(name)
	if err != nil {
		return err
	}

	// cache template
	s.cache[name] = tpl

	return nil
}

// ParseDir parse all templates
func (s *XTemplate) ParseDir(root, extension string, onlyPartials bool) error {
	// parse partial templates (i.e files that are named _xxxxx.ext)
	_, err := s.shared.ParseGlob("_*")
	if err != nil {
		return err
	}

	if onlyPartials {
		return err
	}

	// find all template files
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// skip dirs
		if info == nil || info.IsDir() {
			return nil
		}

		// check for extension
		e := filepath.Ext(path)
		if e != extension {
			return nil
		}

		name := strings.TrimPrefix(path, root)

		if strings.HasPrefix(name, "_") {
			return nil
		}

		// parse template
		tpl, err := s.getTemplate(name)
		if err != nil {
			return err
		}

		// cache template
		s.cache[name] = tpl

		return nil
	})

	return err
}

// Render parses a template then caches it. Will use cached version unless ignoreCache == true
// if the template isnt found in the cache Render will attempt to locate it and parse
func (s *XTemplate) Render(wr io.Writer, name string, data interface{}, ignoreCache bool) error {

	var (
		tpl   *template.Template
		found bool
		err   error
	)

	if ignoreCache {
		// parse template
		tpl, err = s.getTemplate(name)
		if err != nil {
			return err
		}

	} else {
		tpl, found = s.cache[name]
		if !found {
			// parse template
			tpl, err = s.getTemplate(name)
			if err != nil {
				return err
			}

			// cache template
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

// RenderString renders a template from a string. supports the extend and include actions
func (s *XTemplate) RenderString(tplStr string, data interface{}) (string, error) {

	var (
		tpl *template.Template
		err error
	)

	fleContent := []byte(tplStr)
	var fm *frontMatter
	fleContent, fm, err = preProccess(fleContent)
	if err != nil {
		return "", err
	}

	if fm == nil {
		tpl, err = s.shared.Clone()
		if err != nil {
			return "", err
		}

		_, err = tpl.Parse(string(fleContent))
		if err != nil {
			return "", err
		}
	} else if fm != nil && len(fm.Master) > 0 {
		// get the master template
		master, err := s.getTemplate(fm.Master)
		if err != nil {
			return "", err
		}

		// have the master template use this template as an overlay
		tpl, err = master.Parse(string(fleContent))
		if err != nil {
			return "", err
		}
	}

	if fm != nil && len(fm.Include) > 0 {
		for i := range fm.Include {
			fm.Include[i] = filepath.Join(s.folder, fm.Include[i])
		}
		_, err = tpl.ParseFiles(fm.Include...)
		if err != nil {
			return "", err
		}
	}

	buff := bytes.NewBufferString("")
	if err = tpl.Execute(buff, data); err != nil {
		return "", err
	}

	retv := buff.String()

	return retv, nil
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

	var fm *frontMatter
	fleContent, fm, err = preProccess(fleContent)
	if err != nil {
		return nil, err
	}

	// if template doesn't contains frontmatter
	if fm == nil {
		// this template doesn't contain frontmatter, create template and return
		tpl, err := s.makeTemplate(name, fleContent)
		if err != nil {
			return nil, err
		}

		return tpl, nil
	}

	var tpl *template.Template

	if len(fm.Master) > 0 {
		// get the master template
		master, err := s.getTemplate(fm.Master)
		if err != nil {
			return nil, err
		}

		// have the master template use this template as an overlay
		tpl, err = master.Parse(string(fleContent))
		if err != nil {
			return nil, err
		}
	} else {
		// create template
		tpl, err = s.makeTemplate(name, fleContent)
		if err != nil {
			return nil, err
		}
	}

	// parse includes if requested
	if len(fm.Include) > 0 {
		for i := range fm.Include {
			fm.Include[i] = filepath.Join(s.folder, fm.Include[i])
		}
		_, err = tpl.ParseFiles(fm.Include...)
		if err != nil {
			return nil, err
		}
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

func preProccess(fleContent []byte) ([]byte, *frontMatter, error) {

	// extract front matter
	var fm *frontMatter
	fm, fleContent = extractFrontMatter(actRe, fleContent)

	// handle {{ template }}
	fleContent = convertTemplateSyntax(tplRe, fleContent)

	// translate function syntax sugar
	// fn(arg1, arg2,...) --> fn arg1 arg2 ...
	fleContent = translateFuncSyntax(fleContent)

	return fleContent, fm, nil
}

// convertTemplateSyntax
// {{ template button("args") }} --> {{ template "button" "args" }}
func convertTemplateSyntax(re *regexp.Regexp, src []byte) []byte {
	retv := re.ReplaceAllFunc(src, func(b []byte) []byte {
		part := re.FindSubmatch(b)

		retStr := ""
		if len(part) == 2 {
			retStr = fmt.Sprintf("{{ template \"%s\" }}", string(part[1]))
		} else if len(part) == 3 {
			arg := bytes.TrimSpace(bytes.Replace(part[2], []byte(","), []byte(" "), -1))
			retStr = fmt.Sprintf("{{ template \"%s\" %s }}", string(part[1]), string(arg))
		}
		return []byte(retStr)
	})

	return retv
}

// translateFuncSyntax
// fn(arg1, arg2,...) --> fn arg1 arg2 ...
func translateFuncSyntax(src []byte) []byte {
	retv := lneRe.ReplaceAllFunc([]byte(src), func(b []byte) []byte {
		return expRe.ReplaceAllFunc([]byte(b), func(b2 []byte) []byte {
			rv := bytes.Replace(
				bytes.Replace(b2, []byte(","), []byte(""), -1),
				[]byte("("), []byte(" "), 1,
			)
			return bytes.Replace(rv, []byte(")"), []byte(""), 1)
		})
	})

	return retv
}

func extractFrontMatter(re *regexp.Regexp, src []byte) (*frontMatter, []byte) {
	fm := &frontMatter{}
	retv := re.ReplaceAllFunc(src, func(b []byte) []byte {
		parts := re.FindSubmatch(b)

		if len(parts) < 3 {
			return b
		}
		action := string(bytes.TrimSpace(parts[1]))

		switch action {
		case "extends":
			fm.Master = string(parts[2])
			return []byte("")

		case "include":
			fm.Include = append(fm.Include, string(parts[2]))
			return []byte("")
		}

		return b
	})

	if len(fm.Master) == 0 && len(fm.Include) == 0 {
		fm = nil
	}

	return fm, retv
}
