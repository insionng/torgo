package torgo

/*========================================================================
		Insion / email: insion@lihuashu.com
========================================================================*/
//v0.5

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/astaxie/session"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Controller struct {
	Ctx       *Context
	Tpl       *template.Template
	Statictpl *template.Template
	Data      map[interface{}]interface{}
	ChildName string
	TplNames  string
	Layout    string
	TplExt    string
}

type ControllerInterface interface {
	Init(ct *Context, cn string)
	Prepare()
	Get()
	Post()
	Delete()
	Put()
	Head()
	Patch()
	Options()
	Finish()
	Render() error
	RenderString() (string, []byte, error)
	RenderBytes() ([]byte, error)
}

func (c *Controller) Init(ctx *Context, cn string) {
	c.Data = make(map[interface{}]interface{})
	c.Tpl = template.New(cn + ctx.Request.Method)
	c.Tpl = c.Tpl.Funcs(torgoTplFuncMap)
	c.Layout = ""
	c.TplNames = ""
	c.ChildName = cn
	c.Ctx = ctx
	c.TplExt = "html"

}

func (c *Controller) Prepare() {

}

func (c *Controller) Finish() {

}

func (c *Controller) Get() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Post() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Delete() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Put() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Head() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Patch() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Options() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Render() error {
	rb, err := c.RenderBytes()

	if err != nil {
		return err
	} else {
		c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(rb)), true)
		c.Ctx.ContentType("text/html")
		c.Ctx.ResponseWriter.Write(rb)
		return nil
	}
	return nil
}

func (c *Controller) RenderBytes() ([]byte, error) {

	//if the controller has set layout, then first get the tplname's content set the content to the layout
	if c.Layout != "" {
		if c.TplNames == "" {
			c.TplNames = c.ChildName + "/" + c.Ctx.Request.Method + "." + c.TplExt
		}
		t, err := c.Tpl.ParseFiles(path.Join(ViewsPath, c.TplNames), path.Join(ViewsPath, c.Layout))
		if err != nil {
			Trace("[RenderBytes]template ParseFiles err[a]:", err)
		}
		_, file := path.Split(c.TplNames)
		newbytes := bytes.NewBufferString("")
		t.ExecuteTemplate(newbytes, file, c.Data)
		tplcontent, _ := ioutil.ReadAll(newbytes)
		c.Data["LayoutContent"] = template.HTML(string(tplcontent))
		_, file = path.Split(c.Layout)

		ibytes := bytes.NewBufferString("")
		err = t.ExecuteTemplate(ibytes, file, c.Data)
		icontent, _ := ioutil.ReadAll(ibytes)

		if err != nil {
			Trace("[RenderBytes]template Execute err[b]:", err)
		}
		return icontent, err

	} else {
		if c.TplNames == "" {
			c.TplNames = c.ChildName + "/" + c.Ctx.Request.Method + "." + c.TplExt
		}
		t, err := c.Tpl.ParseFiles(path.Join(ViewsPath, c.TplNames))
		if err != nil {
			Trace("[RenderBytes]template ParseFiles err[c]:", err)
		}
		_, file := path.Split(c.TplNames)

		ibytes := bytes.NewBufferString("")
		err = t.ExecuteTemplate(ibytes, file, c.Data)
		icontent, _ := ioutil.ReadAll(ibytes)
		if err != nil {
			Trace("[RenderBytes]template Execute err[d]:", err)
		}
		return icontent, err
	}

	return []byte{}, nil
}

func (c *Controller) RenderString() (string, []byte, error) {

	//if the controller has set layout, then first get the tplname's content set the content to the layout
	if c.Layout != "" {
		if c.TplNames == "" {
			c.TplNames = c.ChildName + "/" + c.Ctx.Request.Method + "." + c.TplExt
		}
		k, err := c.Statictpl.ParseFiles(path.Join(ViewsPath, c.TplNames), path.Join(ViewsPath, c.Layout))
		if err != nil {
			Trace("[RenderString]template ParseFiles err[a]:", err)
		}
		_, file := path.Split(c.TplNames)
		newbytes := bytes.NewBufferString("")
		k.ExecuteTemplate(newbytes, file, c.Data)
		tplcontent, _ := ioutil.ReadAll(newbytes)
		c.Data["LayoutContent"] = template.HTML(string(tplcontent))
		_, file = path.Split(c.Layout)

		ibytes := bytes.NewBufferString("")
		err = k.ExecuteTemplate(ibytes, file, c.Data)
		icontent, _ := ioutil.ReadAll(ibytes)
		f := string(icontent)

		if err != nil {
			Trace("[RenderString]template Execute err[b]:", err)
		}
		return f, icontent, err

	} else {
		if c.TplNames == "" {
			c.TplNames = c.ChildName + "/" + c.Ctx.Request.Method + "." + c.TplExt
		}
		k, err := c.Statictpl.ParseFiles(path.Join(ViewsPath, c.TplNames))
		if err != nil {
			Trace("[RenderString]template ParseFiles err[c]:", err)
		}
		_, file := path.Split(c.TplNames)

		ibytes := bytes.NewBufferString("")
		err = k.ExecuteTemplate(ibytes, file, c.Data)
		icontent, _ := ioutil.ReadAll(ibytes)
		f := string(icontent)
		if err != nil {
			Trace("[RenderString]template Execute err[d]:", err)
		}
		return f, icontent, err
	}

	return "", []byte{}, nil
}

func (c *Controller) Redirect(url string, code int) {
	c.Ctx.Redirect(code, url)
}

func (c *Controller) ServeJson() {
	content, err := json.MarshalIndent(c.Data["json"], "", "  ")
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(content)), true)
	c.Ctx.ContentType("json")
	c.Ctx.ResponseWriter.Write(content)
}

func (c *Controller) ServeXml() {
	content, err := xml.Marshal(c.Data["xml"])
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(content)), true)
	c.Ctx.ContentType("xml")
	c.Ctx.ResponseWriter.Write(content)
}

func (c *Controller) Input() url.Values {
	c.Ctx.Request.ParseForm()
	return c.Ctx.Request.Form
}

func (c *Controller) StartSession() (sess session.Session) {
	sess = GlobalSessions.SessionStart(c.Ctx.ResponseWriter, c.Ctx.Request)
	return
}
