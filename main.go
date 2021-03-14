package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gmolveau/echotemplate"
	"github.com/gorilla/sessions"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed views/*
var Views embed.FS

//go:embed static/*
var Static embed.FS

func GetStaticFS() http.FileSystem {
	fsys, _ := fs.Sub(Static, "static")
	return http.FS(fsys)
}

var TplConfig = echotemplate.TemplateConfig{
	Root:         "views",
	Extension:    ".html",
	Master:       "layouts/master",
	Partials:     []string{},
	DisableCache: false,
	Funcs:        make(template.FuncMap),
	Delims:       echotemplate.Delims{Left: "{{", Right: "}}"},
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("THE_SECRET_KEY"))))

	e.Renderer = echotemplate.NewWithConfigEmbed(Views, TplConfig)

	e.GET("/", Index)

	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", http.FileServer(GetStaticFS()))))

	e.Logger.Fatal(e.Start(":1323"))
}

func Index(c echo.Context) error {
	SetSuccessFlash(&c, "OK !")
	SetErrorFlash(&c, "There was an error")
	return c.Render(http.StatusOK,
		"index",
		echo.Map{
			"flashes": GetFlashes(&c),
		},
	)
}

//Flash is a flash message with a type
type Flash struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

//ToString returns the flash as a string
func (p Flash) ToString() string {
	out, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(out)
}

//FlashFromString returns a flash from a string
func FlashFromString(s string) (*Flash, error) {
	f := new(Flash)
	err := json.Unmarshal([]byte(s), f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func SetErrorFlash(c *echo.Context, message string) {
	setFlash(c, "error", message)
}

func SetWarningFlash(c *echo.Context, message string) {
	setFlash(c, "warning", message)
}

func SetNoticeFlash(c *echo.Context, message string) {
	setFlash(c, "notice", message)
}

func SetSuccessFlash(c *echo.Context, message string) {
	setFlash(c, "success", message)
}

func setFlash(c *echo.Context, flashType string, message string) {
	sess, _ := session.Get("session", *c)
	flash := Flash{
		Type:    flashType,
		Message: message,
	}
	sess.AddFlash(flash.ToString())
	err := sess.Save((*c).Request(), (*c).Response())
	if err != nil {
		(*c).Logger().Error(err)
	}
}

func GetFlashes(c *echo.Context) (flashes []*Flash) {
	sess, _ := session.Get("session", *c)
	if f := sess.Flashes(); len(f) > 0 {
		for _, ff := range f {
			flash, _ := FlashFromString(ff.(string))
			flashes = append(flashes, flash)
		}
		err := sess.Save((*c).Request(), (*c).Response())
		if err != nil {
			(*c).Logger().Error(err)
		}
	}
	return flashes
}
