package main

import (
	"os"
	"flag"
	"fmt"
	"path"
	"strings"
	"runtime"
	"errors"
)

// 用于打印的颜色
var cr0 = "\033[0m"	// 结束
var cr1 = "\033[01;37m"	// 白色
var cr2 = "\033[01;32m"	// 绿色
var cr3 = "\033[01;33m"	// 黄色
var cr4 = "\033[01;31m"	// 红色

func main() {

	// windows下不打印颜色
	if runtime.GOOS == "windows" {
		cr0 = ""
		cr1 = ""
		cr2 = ""
		cr3 = ""
		cr4 = ""
	}

	// 获取appName
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println(cr1, "Usage：gnew [appName]", cr0)
		return
	}
	appName := args[0]
	fmt.Println("Start to create project", cr3, appName, cr0, "...")


	// 创建app路径
	goPath, err := getFirstGoPath()
	if err != nil {
		fmt.Println(cr4, "Get GOPATH failed:", err, cr0)
		return
	}
	appPath := path.Join(goPath, "src", appName)
	if IsExist(appPath) {
		fmt.Println(cr4, "The project already exists!", cr0)
		return
	}
	os.MkdirAll(appPath, 0755)
	os.MkdirAll(path.Join(appPath, "routers"), 0755)
	os.MkdirAll(path.Join(appPath, "controllers"), 0755)

	// 创建各个文件
	Check(Write2File(path.Join(appPath, ".gitignore"), strings.Replace(strGitignore, "{{.AppName}}", appName, -1)))
	Check(Write2File(path.Join(appPath, "main.go"), strings.Replace(strMain, "{{.AppName}}", appName, -1)))
	Check(Write2File(path.Join(appPath, "routers", "router.go"), strings.Replace(strRouter, "{{.AppName}}", appName, -1)))
	Check(Write2File(path.Join(appPath, "controllers", "controller.go"), strings.Replace(strController, "{{.AppName}}", appName, -1)))
	Check(Write2File(path.Join(appPath, "controllers", "example.go"), strings.Replace(strExample, "{{.AppName}}", appName, -1)))

	fmt.Println("Project", cr3, appName, cr0, cr2, "has been created successfully!", cr0)
}

func Check(err error) {
	if err != nil {
		fmt.Println(cr3, "Operation failed:", cr0, cr4, err, cr0)
		panic(err)
	}
}

func Write2File(filename, content string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.WriteString(content)
	fmt.Println(cr3, "-- Created: ", cr0, cr2, filename, cr0)
	return
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func getFirstGoPath() (goPath string, err error) {
	gopath := os.Getenv("GOPATH")
	var paths []string
	if runtime.GOOS == "windows" {
		gopath = strings.Replace(gopath, "\\", "/", -1)
		paths = strings.Split(gopath, ";")
	} else {
		paths = strings.Split(gopath, ":")
	}
	if len(paths) <= 0 {
		err = errors.New("GOPATH environment variable is not valid!")
		return
	}
	goPath = paths[0]
	return
}

var strMain = `package main

import (
    "github.com/gorilla/mux"
    "net/http"
    "{{.AppName}}/routers"
    "fmt"
    "github.com/urfave/negroni"
    "time"
    "github.com/15125505/zlog/log"
    "sort"
    "strings"
)

func init() {
    log.Log.SetLogFile("logs/{{.AppName}}")
    log.Log.SetFileColor(true)
    log.Log.SetAdditionalErrorFile(true)
    log.Log.SetLogLevel(log.LevelDebug)
}


func main() {
    r := mux.NewRouter()
    rr := routers.NewRouter(r)
    rr.CreateHandle()
    n := negroni.New(negroni.NewRecovery(), negroni.NewStatic(http.Dir("public")), negroni.HandlerFunc(PreProcess))
    n.UseHandler(r)
    n.Run(":5000")
}

func PreProcess(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

    // 在页面被处理之前，你可以做一些工作
    start := time.Now()

    // 这句代码放在next之前，可以避免每次获取参数之前都进行ParseForm操作
    r.ParseForm()

    // 继续后续的处理
    next(rw, r)

    // 获取http状态码
    res := rw.(negroni.ResponseWriter)
    code := res.Status()
    var color string
    switch {
    case code >= 200 && code < 300:
        color = "\033[01;42;34m" // 绿色
    case code >= 300 && code < 400:
        color = "\033[01;47;34m" // 白色
    case code >= 400 && code < 500:
        color = "\033[01;43;34m" // 黄色
    default:
        color = "\033[01;41;33m" // 红色
    }

    // 获取参数信息
    var keys []string
    for k := range r.Form {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    var param string
    for _, k := range keys {
        param += " " + k + ":"
        param += fmt.Sprint(r.Form[k])
    }
    param = strings.TrimLeft(param, " ")

    // 获取请求者IP地址
    ip := "127.0.0.1"
    if forwards := r.Header.Get("X-Forwarded-For"); forwards != "" {
        ips := strings.Split(forwards, ",")
        if len(ips) > 0 {
            ip = ips[0]
        }
    } else {
        ip = r.RemoteAddr
    }
    ips := strings.Split(ip, ":")
    if len(ips) > 0 {
        ip = ips[0]
    }

    // 显示请求详情
    tmpUA := []byte(r.UserAgent())
    if len(tmpUA) > 40 {
        tmpUA = tmpUA[:40]
    }
    if r.Method == "HEAD" {
        // 为了避免阿里云的SLB日志过多，不打印HEAD请求，
        return
    }
    log.Info(fmt.Sprintf("%v %v \033[0m\033[37m|%12v\033[32m|%15s\033[01;37m|%5v\033[0m\033[33m|%40v\033[32m|%v\033[33m|%v",
        color,
        res.Status(),
        time.Since(start),
        ip,
        r.Method,
        string(tmpUA),
        r.URL.Path,
        param,
    ))
}`

var strRouter = `package routers

import (
    "github.com/gorilla/mux"
    "{{.AppName}}/controllers"
)

// 根路由
type Router struct {
    r *mux.Router
}

// 创建一个根路由实例
func NewRouter(r *mux.Router) (*Router) {
    return &Router{r: r}
}

// 创建路由
func (r *Router) CreateHandle() {

    // todo: 添加自定义的Controller（Controller名字和前缀名字需要设置）
    (&controllers.ExampleController{*controllers.NewController(r.r, "/example")}).Handle()
}`

var strController = `package controllers

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/unrolled/render"
)


// Controller基类定义
type Controller struct {
    muxRounter *mux.Router
    subRounter *mux.Router
}

var Render *render.Render = render.New()


// 创建一个Controller实例
func NewController(r *mux.Router, tpl string) (*Controller) {
    return &Controller{
        muxRounter: r,
        subRounter: r.PathPrefix(tpl).Subrouter(),
    }
}


// 添加路由
func (c *Controller) HandleFunc(path string, f func(http.ResponseWriter,
    *http.Request)) (*mux.Route) {
    return c.muxRounter.HandleFunc(path, f)
}

// 添加子路由
func (c *Controller) HandleSubFunc(path string, f func(http.ResponseWriter,
    *http.Request)) (*mux.Route) {
    return c.subRounter.HandleFunc(path, f)
}`

var strExample = `package controllers

import (
    "net/http"
)

// todo: 从Controller类派生一个自己的类
type ExampleController struct {
    Controller
}

// todo: 为自己的类添加一个这样的Handle函数，注意名称不能随意改动
func (c *ExampleController) Handle() {

    // todo: 本函数演示了使用绝对路由的方式
    c.HandleFunc("/path1", absolutePath)

    // todo: 本函数演示了使用相对路由的方式
    c.HandleSubFunc("/path2", relativePath).Methods("POST")
}

// todo: 访问 http://localhost:5000/path1 将触发本函数的反馈
func absolutePath(w http.ResponseWriter, r *http.Request) {

    // todo:本函数演示了json的输出方法
    type o struct {
        A int ` + "`" + `json:"IntValue"` + "`" + `
        B string ` + "`" + `json:"StringValue"` + "`" + `
    }
    Render.JSON(w, http.StatusOK, o{1, "This is a string value."})
    return
}

// todo: 使用POST方式访问 http://localhost:5000/example/path2 将触发本函数的反馈
func relativePath(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is a relative path."))
    return
}`

var strGitignore = `.idea
*.exe
*.log`
