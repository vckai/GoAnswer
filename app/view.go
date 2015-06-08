package app

import (
	"bytes"
	"html/template"
	"path"
)

type View struct {
	Dir           string
	FuncMap       template.FuncMap
	IsCache       bool
	templateCache map[string]*template.Template
}

func NewView(dir string) *View {
	v := new(View)
	v.Dir = dir
	v.FuncMap = make(template.FuncMap)

	v.IsCache = false
	v.templateCache = make(map[string]*template.Template)
	return v
}

func (this *View) getInstance(tpl string) (*template.Template, error) {
	if this.IsCache && this.templateCache[tpl] != nil {
		return this.templateCache[tpl], nil
	}
	var (
		t *template.Template
		e error
	)
	t = template.New(path.Base(tpl))
	t.Funcs(this.FuncMap)
	t, e = t.ParseFiles(path.Join(this.Dir, tpl))
	if e != nil {
		return nil, e
	}
	if this.IsCache {
		this.templateCache[tpl] = t
	}
	return t, nil
}

func (this *View) Render(tpl string, data map[string]interface{}) ([]byte, error) {
	t, err := this.getInstance(tpl)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
