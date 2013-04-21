package torgo

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/insionng/torgo/session"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

type Handler struct {
	Ctx       *Context
	Data      map[interface{}]interface{}
	ChildName string
	TplNames  string
	Layout    string
	TplExt    string
}

type HandlerInterface interface {
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
}

func (c *Handler) Init(ctx *Context, cn string) {
	c.Data = make(map[interface{}]interface{})
	c.Layout = ""
	c.TplNames = ""
	c.ChildName = cn
	c.Ctx = ctx
	c.TplExt = "html"

}

func (c *Handler) Prepare() {

}

func (c *Handler) RenderBefore() {
}

func (c *Handler) Finish() {
}

func (c *Handler) Get() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Handler) Post() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Handler) Delete() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Handler) Put() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Handler) Head() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Handler) Patch() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Handler) Options() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Handler) Render() (err error) {
	return c.RenderCore(nil, err)
}

func (c *Handler) RenderPlus(rb []byte) (err error) {
	return c.RenderCore(rb, err)
}

func (c *Handler) RenderCore(rb []byte, err error) error {
	if rb == nil {
		rb, err = c.RenderBytes()
	}

	if err != nil {
		return err
	} else {
		c.Ctx.ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
		output_writer := c.Ctx.ResponseWriter.(io.Writer)
		if EnableGzip == true && c.Ctx.Request.Header.Get("Accept-Encoding") != "" {
			splitted := strings.SplitN(c.Ctx.Request.Header.Get("Accept-Encoding"), ",", -1)
			encodings := make([]string, len(splitted))

			for i, val := range splitted {
				encodings[i] = strings.TrimSpace(val)
			}
			for _, val := range encodings {
				if val == "gzip" {
					c.Ctx.ResponseWriter.Header().Set("Content-Encoding", "gzip")
					output_writer, _ = gzip.NewWriterLevel(c.Ctx.ResponseWriter, gzip.BestSpeed)

					break
				} else if val == "deflate" {
					c.Ctx.ResponseWriter.Header().Set("Content-Encoding", "deflate")
					output_writer, _ = zlib.NewWriterLevel(c.Ctx.ResponseWriter, zlib.BestSpeed)
					break
				}
			}
		} else {
			c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(rb)), true)
		}
		output_writer.Write(rb)
		switch output_writer.(type) {
		case *gzip.Writer:
			output_writer.(*gzip.Writer).Close()
		case *zlib.Writer:
			output_writer.(*zlib.Writer).Close()
		case io.WriteCloser:
			output_writer.(io.WriteCloser).Close()
		}
		return nil
	}
	return nil
}

func (c *Handler) RenderString() (string, error) {
	b, e := c.RenderBytes()
	return string(b), e
}

func (c *Handler) RenderBytes() ([]byte, error) {
	//if the handler has set layout, then first get the tplname's content set the content to the layout
	if c.Layout != "" {
		if c.TplNames == "" {
			c.TplNames = c.ChildName + "/" + c.Ctx.Request.Method + "." + c.TplExt
		}
		if RunMode == "dev" {
			BuildTemplate(ViewsPath)
		}
		subdir := path.Dir(c.TplNames)
		_, file := path.Split(c.TplNames)
		newbytes := bytes.NewBufferString("")
		if _, ok := BeeTemplates[subdir]; !ok {
			panic("can't find templatefile in the path:" + c.TplNames)
			return []byte{}, errors.New("can't find templatefile in the path:" + c.TplNames)
		}
		BeeTemplates[subdir].ExecuteTemplate(newbytes, file, c.Data)
		tplcontent, _ := ioutil.ReadAll(newbytes)
		c.Data["LayoutContent"] = template.HTML(string(tplcontent))
		_, file = path.Split(c.Layout)
		ibytes := bytes.NewBufferString("")
		err := BeeTemplates[subdir].ExecuteTemplate(ibytes, file, c.Data)
		if err != nil {
			Trace("template Execute err:", err)
		}
		icontent, _ := ioutil.ReadAll(ibytes)
		return icontent, nil
	} else {
		if c.TplNames == "" {
			c.TplNames = c.ChildName + "/" + c.Ctx.Request.Method + "." + c.TplExt
		}
		if RunMode == "dev" {
			BuildTemplate(ViewsPath)
		}
		subdir := path.Dir(c.TplNames)
		_, file := path.Split(c.TplNames)
		ibytes := bytes.NewBufferString("")
		if _, ok := BeeTemplates[subdir]; !ok {
			panic("can't find templatefile in the path:" + c.TplNames)
			return []byte{}, errors.New("can't find templatefile in the path:" + c.TplNames)
		}
		err := BeeTemplates[subdir].ExecuteTemplate(ibytes, file, c.Data)
		if err != nil {
			Trace("template Execute err:", err)
		}
		icontent, _ := ioutil.ReadAll(ibytes)
		return icontent, nil
	}
	return []byte{}, nil
}

func (c *Handler) Redirect(url string, code int) {
	c.Ctx.Redirect(code, url)
}

func (c *Handler) ServeJson() {
	content, err := json.MarshalIndent(c.Data["json"], "", "  ")
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(content)), true)
	c.Ctx.ContentType("json")
	c.Ctx.ResponseWriter.Write(content)
}

func (c *Handler) ServeXml() {
	content, err := xml.Marshal(c.Data["xml"])
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(content)), true)
	c.Ctx.ContentType("xml")
	c.Ctx.ResponseWriter.Write(content)
}

func (c *Handler) Input() url.Values {
	ct := c.Ctx.Request.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		c.Ctx.Request.ParseMultipartForm(MaxMemory) //64MB
	} else {
		c.Ctx.Request.ParseForm()
	}
	return c.Ctx.Request.Form
}

func (c *Handler) GetString(key string) string {
	return c.Input().Get(key)
}

func (c *Handler) GetInt(key string) (int64, error) {
	return strconv.ParseInt(c.Input().Get(key), 10, 64)
}

func (c *Handler) GetBool(key string) (bool, error) {
	return strconv.ParseBool(c.Input().Get(key))
}

func (c *Handler) GetFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.Ctx.Request.FormFile(key)
}

func (c *Handler) SaveToFile(fromfile, tofile string) error {
	file, _, err := c.Ctx.Request.FormFile(fromfile)
	if err != nil {
		return err
	}
	defer file.Close()
	f, err := os.OpenFile(tofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, file)
	return nil
}

func (c *Handler) StartSession() (sess session.SessionStore) {
	sess = GlobalSessions.SessionStart(c.Ctx.ResponseWriter, c.Ctx.Request)
	return
}

func (c *Handler) SetSession(name string, value interface{}) {
	ss := c.StartSession()
	defer ss.SessionRelease()
	ss.Set(name, value)
}

func (c *Handler) GetSession(name string) interface{} {
	ss := c.StartSession()
	defer ss.SessionRelease()
	return ss.Get(name)
}

func (c *Handler) DelSession(name string) {
	ss := c.StartSession()
	defer ss.SessionRelease()
	ss.Delete(name)
}
