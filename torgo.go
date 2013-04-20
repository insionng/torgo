package torgo

import (
	"fmt"
	"github.com/insionng/torgo/session"
	"html/template"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"runtime"
	"strconv"
)

const VERSION = "0.8.6"

var (
	TorApp        *App
	AppName       string
	AppPath       string
	StaticDir     map[string]string
	TemplateCache map[string]*template.Template
	HttpAddr      string
	HttpPort      int
	RecoverPanic  bool
	AutoRender    bool
	RenderPlus    bool
	PprofOn       bool
	ViewsPath     string
	RunMode       string //"dev" or "prod"
	AppConfig     *Config
	//related to session
	SessionOn            bool   // wheather auto start session,default is false
	SessionProvider      string // default session provider  memory mysql redis
	SessionName          string // sessionName cookie's name
	SessionGCMaxLifetime int64  // session's gc maxlifetime
	SessionSavePath      string // session savepath if use mysql/redis/file this set to the connectinfo
	UseFcgi              bool
	MaxMemory            int64
	EnableGzip           bool // enable gzip

	GlobalSessions *session.Manager //GlobalSessions
)

func init() {
	os.Chdir(path.Dir(os.Args[0]))
	TorApp = NewApp()
	AppPath, _ = os.Getwd()
	StaticDir = make(map[string]string)
	TemplateCache = make(map[string]*template.Template)
	var err error
	AppConfig, err = LoadConfig(path.Join(AppPath, "conf", "app.conf"))
	if err != nil {
		//Trace("open Config err:", err)
		HttpAddr = ""
		HttpPort = 80
		Maxprocs = -1
		AppName = "torgo"
		RunMode = "dev" //default runmod
		AutoRender = false
		RenderPlus = false
		RecoverPanic = true
		PprofOn = false
		ViewsPath = "views"
		SessionOn = false
		SessionProvider = "memory"
		SessionName = "TorgoSessionId"
		SessionGCMaxLifetime = 3600
		SessionSavePath = ""
		UseFcgi = false
		MaxMemory = 1 << 26 //64MB
		EnableGzip = false
	} else {
		HttpAddr = AppConfig.String("httpaddr")
		if v, err := AppConfig.Int("httpport"); err != nil {
			HttpPort = 80
		} else {
			HttpPort = v
		}
		if v, err := AppConfig.Int("maxprocs"); err != nil {
			Maxprocs = -1
		} else {
			Maxprocs = v
		}
		if v, err := AppConfig.Int64("maxmemory"); err != nil {
			MaxMemory = 1 << 26
		} else {
			MaxMemory = v
		}
		AppName = AppConfig.String("appname")
		if runmode := AppConfig.String("runmode"); runmode != "" {
			RunMode = runmode
		} else {
			RunMode = "dev"
		}
		if ar, err := AppConfig.Bool("renderplus"); err != nil {
			RenderPlus = false
		} else {
			RenderPlus = ar
		}
		if ar, err := AppConfig.Bool("autorender"); err != nil {
			AutoRender = false
		} else {
			AutoRender = ar
		}
		if ar, err := AppConfig.Bool("autorecover"); err != nil {
			RecoverPanic = true
		} else {
			RecoverPanic = ar
		}
		if ar, err := AppConfig.Bool("pprofon"); err != nil {
			PprofOn = false
		} else {
			PprofOn = ar
		}
		if views := AppConfig.String("viewspath"); views == "" {
			ViewsPath = "views"
		} else {
			ViewsPath = views
		}
		if ar, err := AppConfig.Bool("sessionon"); err != nil {
			SessionOn = false
		} else {
			SessionOn = ar
		}
		if ar := AppConfig.String("sessionprovider"); ar == "" {
			SessionProvider = "memory"
		} else {
			SessionProvider = ar
		}
		if ar := AppConfig.String("sessionname"); ar == "" {
			SessionName = "TorgoSessionId"
		} else {
			SessionName = ar
		}
		if ar := AppConfig.String("sessionsavepath"); ar == "" {
			SessionSavePath = ""
		} else {
			SessionSavePath = ar
		}
		if ar, err := AppConfig.Int("sessiongcmaxlifetime"); err == nil && ar != 0 {
			int64val, _ := strconv.ParseInt(strconv.Itoa(ar), 10, 64)
			SessionGCMaxLifetime = int64val
		} else {
			SessionGCMaxLifetime = 3600
		}
		if ar, err := AppConfig.Bool("usefcgi"); err != nil {
			UseFcgi = false
		} else {
			UseFcgi = ar
		}
		if ar, err := AppConfig.Bool("enablegzip"); err != nil {
			EnableGzip = false
		} else {
			EnableGzip = ar
		}
	}
	StaticDir["/static"] = "static"

}

type App struct {
	Handlers *HandlerRegistor
}

// New returns a new PatternServeMux.
func NewApp() *App {
	cr := NewHandlerRegistor()
	app := &App{Handlers: cr}
	return app
}

func (app *App) Run() {
	addr := fmt.Sprintf("%s:%d", HttpAddr, HttpPort)
	var err error
	if UseFcgi {
		l, e := net.Listen("tcp", addr)
		if e != nil {
			BeeLogger.Fatal("Listen: ", e)
		}
		err = fcgi.Serve(l, app.Handlers)
	} else {
		err = http.ListenAndServe(addr, app.Handlers)
	}
	if err != nil {
		BeeLogger.Fatal("ListenAndServe: ", err)
	}
}

func (app *App) Router(path string, c HandlerInterface) *App {
	app.Handlers.Add(path, c)
	return app
}

func (app *App) Filter(filter http.HandlerFunc) *App {
	app.Handlers.Filter(filter)
	return app
}

func (app *App) FilterParam(param string, filter http.HandlerFunc) *App {
	app.Handlers.FilterParam(param, filter)
	return app
}

func (app *App) FilterPrefixPath(path string, filter http.HandlerFunc) *App {
	app.Handlers.FilterPrefixPath(path, filter)
	return app
}

func (app *App) SetViewsPath(path string) *App {
	ViewsPath = path
	return app
}

func (app *App) SetStaticPath(url string, path string) *App {
	StaticDir[url] = path
	return app
}

func (app *App) ErrorLog(ctx *Context) {
	BeeLogger.Printf("[ERR] host: '%s', request: '%s %s', proto: '%s', ua: '%s', remote: '%s'\n", ctx.Request.Host, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), ctx.Request.RemoteAddr)
}

func (app *App) AccessLog(ctx *Context) {
	BeeLogger.Printf("[ACC] host: '%s', request: '%s %s', proto: '%s', ua: '%s', remote: '%s'\n", ctx.Request.Host, ctx.Request.Method, ctx.Request.URL.Path, ctx.Request.Proto, ctx.Request.UserAgent(), ctx.Request.RemoteAddr)
}

func Route(path string, c HandlerInterface) *App {
	TorApp.Router(path, c)
	return TorApp
}

func Router(path string, c HandlerInterface) *App {
	TorApp.Router(path, c)
	return TorApp
}

func RouterHandler(path string, c http.Handler) *App {
	TorApp.Handlers.AddHandler(path, c)
	return TorApp
}

func SetViewsPath(path string) *App {
	TorApp.SetViewsPath(path)
	return TorApp
}

func SetStaticPath(url string, path string) *App {
	StaticDir[url] = path
	return TorApp
}

func Filter(filter http.HandlerFunc) *App {
	TorApp.Filter(filter)
	return TorApp
}

func FilterParam(param string, filter http.HandlerFunc) *App {
	TorApp.FilterParam(param, filter)
	return TorApp
}

func FilterPrefixPath(path string, filter http.HandlerFunc) *App {
	TorApp.FilterPrefixPath(path, filter)
	return TorApp
}

func Run() {
	if PprofOn {
		TorApp.Router(`/debug/pprof`, &ProfHandler{})
		TorApp.Router(`/debug/pprof/:pp([\w]+)`, &ProfHandler{})
	}
	if SessionOn {
		GlobalSessions, _ = session.NewManager(SessionProvider, SessionName, SessionGCMaxLifetime, SessionSavePath)
		go GlobalSessions.GC()
	}
	err := BuildTemplate(ViewsPath)
	if err != nil {
		if RunMode == "dev" {
			Warn(err)
		}
	}

	if Maxprocs == -1 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(Maxprocs)
	}

	TorApp.Run()
}
