package xtemplate

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// XTemplate ...
type XTemplate struct {
	rootFolder       string
	partialsFolder   string
	componentsFolder string
	ext              string
	shared           *template.Template
	cache            map[string]*template.Template
	funcs            template.FuncMap
}

// {{ ...  }}
var lneRe = regexp.MustCompile(`{{.+?}}`)

// fnMix("index.html", 2)
var expRe = regexp.MustCompile(`([a-zA-Z]+[0-9]*)\((.+)\)`)

// {{ extend "index.html" }}
var actRe = regexp.MustCompile(`\-*[[:blank:]]*{{ *(.+?) *\"(.+?)\" *}}[[:blank:]]*[\r\n]*\-*`)

// {{ template button(123) }}
var tplRe = regexp.MustCompile(`{{\-*\s*(macro|template)\s*([a-zA-Z0-9\-_]+)\s*\((.*?)\)\s*\-*}}`)

// <tag (attr)>(content)</tag>
var tagRe = regexp.MustCompile(`<tag(\s+[^>]+)?>((.|\n)*?)</tag>([\s]*</tag>)?`)

// {{template "tplName"}}
// {{template "tplName" . }}
// {{template "tplName" $ }}
var tplRe2 = regexp.MustCompile(`{{\-*\s*template\s+"([\w/_.]+)"\s*([.$\w\s_"]*)\s*\-*}}`)

type Config struct {
	RootFolder       string
	PartialsFolder   string
	ComponentsFolder string
	Ext              string
	Funcs            template.FuncMap
}

// New create new instance of XTemplate
func New(cfg Config) *XTemplate {

	xt := new(XTemplate)
	xt.cache = make(map[string]*template.Template)
	xt.rootFolder = cfg.RootFolder
	if xt.rootFolder == "" {
		xt.rootFolder = "./templates"
	}
	xt.partialsFolder = cfg.PartialsFolder
	if xt.partialsFolder == "" {
		xt.partialsFolder = filepath.Join(xt.rootFolder, "_partials")
	}
	xt.componentsFolder = cfg.ComponentsFolder
	if xt.componentsFolder == "" {
		xt.componentsFolder = filepath.Join(xt.rootFolder, "_components")
	}

	xt.shared = template.New("")
	xt.ext = cfg.Ext
	if xt.ext == "" {
		xt.ext = "html"
	}

	funcs := template.FuncMap{
		"args":        args,
		"kwargs":      kwargs,
		"title":       capitalize,
		"lower":       lower,
		"upper":       upper,
		"json":        marshalJSON,
		"tag":         tags,
		"nocache":     NoCache,
		"ifEmpty":     IfEmpty,
		"formatDate":  formatDate,
		"formatCDate": formatCDate,
		"isEmpty":     IsEmpty,
	}

	xt.funcs = funcs
	if len(cfg.Funcs) > 0 {
		for k, v := range cfg.Funcs {
			xt.funcs[k] = v
		}
	}

	xt.shared.Funcs(xt.funcs)
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
	for k, v := range funcMap {
		s.funcs[k] = v
	}
	s.shared.Funcs(s.funcs)
	return s
}

// AddFunc add a function the template's function map.
// must be called before templates are parsed
func (s *XTemplate) AddFunc(name string, fn interface{}) *XTemplate {
	s.funcs[name] = fn
	s.shared.Funcs(s.funcs)
	return s
}

func (s *XTemplate) ListFuncs() {
	for k, v := range s.funcs {
		fmt.Printf("\n %s - %T", k, v)
	}
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
	_, name = getFilename(s.rootFolder, name, s.ext)
	s.cache[name] = tpl

	return nil
}

// ParseDir parse all templates
func (s *XTemplate) parseDir(root, extension string, onlyPartials bool) error {
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

	if err = tpl.Execute(wr, data); err != nil {
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
	fleContent, fm, err = preProcess(s, fleContent)
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
	} else if len(fm.Master) > 0 {
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
			fm.Include[i] = IncludeFile(filepath.Join(s.rootFolder, string(fm.Include[i])))
		}
		_, err = parseFiles(s, tpl, s.rootFolder, s.ext, fm.Include...)
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
	Master  string        `yaml:"master"`
	Include []IncludeFile `yaml:"include"`
}

func getFilename(folder, name, ext string) (fileName string, tplName string) {
	fle := filepath.Join(folder, name)
	// add a file extension if one isn't provided
	if !strings.HasSuffix(name, "."+ext) {
		fle += "." + ext
	}

	// remove file extension if one is provided
	if strings.HasSuffix(name, "."+ext) {
		name = strings.TrimSuffix(name, "."+ext)
	}

	return fle, name
}

func (s *XTemplate) getTemplate(name string) (*template.Template, error) {
	tplName, fm, fleContent, err := s.readTemplate(name)
	if err != nil {
		return nil, err
	}

	/*	// if template doesn't contains front matter
		if fm == nil {
			// this template doesn't contain front matter, create template and return
			tpl, err := s.makeTemplate(tplName, fleContent)
			if err != nil {
				return nil, err
			}

			return tpl, nil
		}*/

	var tpl *template.Template

	if fm != nil && len(fm.Master) > 0 {
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
		tpl, err = s.makeTemplate(tplName, fleContent)
		if err != nil {
			return nil, err
		}

		if err := s.parsePartials(tpl); err != nil {
			return nil, err
		}
	}

	// parse included templates (if any)
	if fm != nil && len(fm.Include) > 0 {
		for i := range fm.Include {
			iNme := string(fm.Include[i])
			_, nme := getFilename(s.rootFolder, iNme, s.ext)
			if tpl.Lookup(IncludeFile(nme).Name()) != nil {
				// template exists as a partial
				continue
			}

			fm.Include[i] = IncludeFile(filepath.Join(s.rootFolder, iNme))
		}
		_, err = parseFiles(s, tpl, s.rootFolder, s.ext, fm.Include...)
		if err != nil {
			return nil, err
		}
	}

	return tpl, nil
}

func (s *XTemplate) readTemplate(name string) (tplName string, fm *frontMatter, content []byte, err error) {
	var fle string
	fle, tplName = getFilename(s.rootFolder, name, s.ext)

	// read template into a buffer
	content, err = ioutil.ReadFile(fle)
	if err != nil {
		return
	}

	// convert extras into standard go template
	content, fm, err = preProcess(s, content)
	if err != nil {
		return
	}

	return
}

func (s *XTemplate) makeTemplate(name string, content []byte) (*template.Template, error) {
	tpl, err := s.shared.Clone()
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return tpl.New(name), nil
	}

	return tpl.New(name).Parse(string(content))
}

func (s *XTemplate) parsePartials(tpl *template.Template) error {
	path := filepath.Join(s.partialsFolder)
	fi, err := os.Lstat(path)
	if err != nil {
		return nil
	}
	if !fi.IsDir() {
		return nil
	}

	root := strings.TrimPrefix(s.rootFolder, "/")
	root = strings.TrimPrefix(root, "./")

	// _, err = tpl.ParseGlob(path + "/*." + s.ext)
	err = filepath.Walk(path, func(name string, info fs.FileInfo, err error) error {
		// skip folders
		if info.IsDir() {
			return nil
		}

		// skip files with different extensions
		if !strings.HasSuffix(name, s.ext) {
			return nil
		}

		// log.Println(s.folder, name, info.Name(), strings.TrimPrefix(name, root))

		_, _, content, err := s.readTemplate(strings.TrimPrefix(name, root))
		if err != nil {
			return err
		}

		tplName := strings.TrimSuffix(strings.TrimPrefix(name, s.partialsFolder+"/"), "."+s.ext)
		_, err = tpl.New(tplName).Parse(string(content))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return err
}

type IncludeFile string

func (i IncludeFile) Name() string {
	p := strings.Split(string(i), "|")
	return p[0]
}
func (i IncludeFile) String() string {
	return i.Name()
}
func (i IncludeFile) FromTemplateAction() bool {
	p := strings.Split(string(i), "|")
	return len(p) > 1 && p[1] == "ta"
}
func (i *IncludeFile) SetTemplateActionFlag() {
	if strings.HasSuffix((*i).String(), "|ta") {
		return
	}

	*i += "|ta"
}

// parseFiles expects filenames to have extensions
func parseFiles(xt *XTemplate, t *template.Template, baseFolder, ext string, filenames ...IncludeFile) (*template.Template, error) {
	baseFolder = filepath.Join(baseFolder) + "/"

	for _, filename := range filenames {

		fName, name := getFilename("", filename.Name(), ext)
		b, err := ioutil.ReadFile(fName)
		if err != nil {
			if filename.FromTemplateAction() {
				// this template was referenced in a {{template "name"}} action
				// and included by extractTemplates so it's possible that it's
				// not an actual file but a template defined elsewhere
				continue
			}

			return nil, err
		}
		prd, _, err := preProcess(xt, b)
		if err != nil {
			return nil, err
		}

		s := string(prd)
		name = strings.Replace(name, baseFolder, "", 1)
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func preProcess(tpl *XTemplate, fleContent []byte) ([]byte, *frontMatter, error) {

	// extract front matter
	var (
		fm  *frontMatter
		err error
	)

	fm, fleContent = extractFrontMatter(actRe, fleContent)

	// add template "name" to includes
	fm = extractTemplates(tplRe2, fm, fleContent)

	fleContent, err = translateComponents(tpl, fleContent)
	if err != nil {
		return nil, nil, err
	}

	// <tag> --> tag .type .attr . content
	fleContent = translateTags(tpl, fleContent)

	// handle {{ template }}
	fleContent = convertTemplateSyntax(tplRe, fleContent)

	// translate function syntax sugar
	// fn(arg1, arg2,...) --> fn arg1 arg2 ...
	fleContent = translateFuncSyntax(fleContent)

	// fmt.Println("\n>>>>\n", string(fleContent), "\n>>>>")
	return fleContent, fm, nil
}

func extractTemplates(re *regexp.Regexp, fm *frontMatter, fleContent []byte) *frontMatter {

	// get all actions matching {{template "tplName" .}}
	matches := re.FindAll(fleContent, -1)
	if matches == nil {
		return fm
	}

	if fm == nil {
		fm = &frontMatter{}
	}

	for _, i := range matches {
		parts := re.FindStringSubmatch(string(i))
		if len(parts) >= 2 {
			fm.Include = append(fm.Include, IncludeFile(parts[1]+"|ta"))
		}
	}

	return fm
}

// convertTemplateSyntax
// {{ template button("args") }} --> {{ template "button" "args" }}
// {{ macro button("args") }} --> {{ template "button" "args" }}
// {{ macro button("a",1,2) }} --> {{ template "button" args "a" 1 2 }}
// {{ macro button("a"::1,"b"::22) }} --> {{ template "button" kwargs "a" 1 "b" 22 }}
func convertTemplateSyntax(re *regexp.Regexp, src []byte) []byte {
	retv := re.ReplaceAllFunc(src, func(b []byte) []byte {
		part := re.FindSubmatch(b)

		retStr := ""
		if len(part) == 3 {
			retStr = fmt.Sprintf("{{ template \"%s\" -}}", string(part[2]))
		} else if len(part) == 4 {
			if bytes.Count(part[3], []byte("::")) > 0 {
				arg := bytes.TrimSpace(bytes.Replace(part[3], []byte("::"), []byte(" "), -1))
				arg = bytes.TrimSpace(bytes.Replace(arg, []byte(","), []byte(" "), -1))
				retStr = fmt.Sprintf("{{ template \"%s\" kwargs %s -}}", string(part[2]), string(arg))
			} else if bytes.Count(part[3], []byte(",")) > 0 {
				arg := bytes.TrimSpace(bytes.Replace(part[3], []byte(","), []byte(" "), -1))
				retStr = fmt.Sprintf("{{ template \"%s\" args %s -}}", string(part[2]), string(arg))
			} else {
				arg := part[3]
				retStr = fmt.Sprintf("{{ template \"%s\" %s -}}", string(part[2]), string(arg))
			}
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
				bytes.Replace(b2, []byte(","), []byte(" "), -1),
				[]byte("("), []byte(" "), 1,
			)
			return bytes.Replace(rv, []byte(")"), []byte(""), 1)
		})
	})

	return retv
}

/*
 translateTags
 Examples.
 <tag type="input" class="abc" disabled><p>hello</p></tag>

 <tag type="field">
 	<tag type="select" value=1>
		{{range .options}}
		<option value="{{.id}}">{{.label}}</option>
		{{end}}
  </tag>
 </tag>
*/
func translateTags(xt *XTemplate, src []byte, mode ...int) []byte {
	// match <tag></tag>
	proc := func(b []byte) []byte {
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
		if err != nil {
			return b
		}

		tag := doc.Find("tag")
		tagType := tag.AttrOr("type", "div")
		tagAttr := tag.Nodes[0].Attr
		tagHTML, err := tag.Html()
		if err != nil {
			return b
		}
		tagHTML = html.UnescapeString(tagHTML)

		// varName := randString(10)
		attrMap := map[string]interface{}{}
		for _, i := range tagAttr {
			if (i.Key) == "type" {
				continue
			}

			attrMap[i.Key] = i.Val
		}

		tagFunc, valid := xt.funcs["tag"].(func(typ string, attr map[string]interface{}, content string) template.HTML)
		if !valid {
			return b
		}

		retv := string(tagFunc(tagType, attrMap, tagHTML))

		// check if tag output includes tag construct
		if tagRe.Match([]byte(retv)) {
			retv = string(translateTags(xt, []byte(retv)))
		}

		// fmt.Println("\n\nsrc: ", string(b), "\nretv: ", retv)
		return []byte(retv)
	}

	retv := tagRe.ReplaceAllFunc(src, proc)

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
			fm.Include = append(fm.Include, IncludeFile(parts[2]))
			return []byte("")
		}

		return b
	})

	if len(fm.Master) == 0 && len(fm.Include) == 0 {
		fm = nil
	}

	return fm, retv
}
