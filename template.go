package torgo

//@todo add template funcs

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	torgoTplFuncMap template.FuncMap
	BeeTemplates    map[string]*template.Template
	BeeTemplateExt  []string
)

func init() {
	BeeTemplates = make(map[string]*template.Template)
	torgoTplFuncMap = make(template.FuncMap)
	BeeTemplateExt = make([]string, 0)
	BeeTemplateExt = append(BeeTemplateExt, "tpl", "html")
	torgoTplFuncMap["markdown"] = MarkDown
	torgoTplFuncMap["dateformat"] = DateFormat
	torgoTplFuncMap["date"] = Date
	torgoTplFuncMap["compare"] = Compare
	torgoTplFuncMap["substr"] = Substr
	torgoTplFuncMap["html2str"] = Html2str
	torgoTplFuncMap["str2html"] = Str2html
	torgoTplFuncMap["htmlquote"] = Htmlquote
	torgoTplFuncMap["htmlunquote"] = Htmlunquote
}

// AddFuncMap let user to register a func in the template
func AddFuncMap(key string, funname interface{}) error {
	if _, ok := torgoTplFuncMap[key]; ok {
		return errors.New("funcmap already has the key")
	}
	torgoTplFuncMap[key] = funname
	return nil
}

type templatefile struct {
	root  string
	files map[string][]string
}

func (self *templatefile) visit(paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() {
		return nil
	} else if (f.Mode() & os.ModeSymlink) > 0 {
		return nil
	} else {
		hasExt := false
		for _, v := range BeeTemplateExt {
			if strings.HasSuffix(paths, v) {
				hasExt = true
				break
			}
		}
		if hasExt {
			replace := strings.NewReplacer("\\", "/")
			a := []byte(paths)
			a = a[len([]byte(self.root)):]
			subdir := path.Dir(strings.TrimLeft(replace.Replace(string(a)), "/"))
			if _, ok := self.files[subdir]; ok {
				self.files[subdir] = append(self.files[subdir], paths)
			} else {
				m := make([]string, 1)
				m[0] = paths
				self.files[subdir] = m
			}

		}
	}
	return nil
}

func AddTemplateExt(ext string) {
	for _, v := range BeeTemplateExt {
		if v == ext {
			return
		}
	}
	BeeTemplateExt = append(BeeTemplateExt, ext)
}

func BuildTemplate(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return err
		} else {
			return errors.New("dir open err")
		}
	}
	self := templatefile{
		root:  dir,
		files: make(map[string][]string),
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return self.visit(path, f, err)
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return err
	}
	for k, v := range self.files {
		BeeTemplates[k] = template.Must(template.New("torgoTemplate" + k).Funcs(torgoTplFuncMap).ParseFiles(v...))
	}
	return nil
}
