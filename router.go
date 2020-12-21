package main

import (
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type (
	Host struct {
		Echo *echo.Echo
	}
)

var htmlTemplate =
	`
<!DOCTYPE html>
<html>
<head>
	<meta name="description" content="Узнать какая сейчас учебная неделя? Верхняя/нижняя? Четная/нечетная? Вам всегда поможет данный сайт!" />
	<meta name="keywords" content="учебная неделя,верхняя-нижняя неделя,четная-нечетная неделя" />
	<meta property="og:title" content="Какая сейчас учебная неделя?" />
	<meta property="og:description" content="Узнать какая сейчас учебная неделя? Верхняя/нижняя? Четная/нечетная? Вам всегда поможет данный сайт!" />
	<meta property="og:url" content="https://week.kinsle.ru/" />
<title>%s</title>
<style>
body {
  text-align: center;
  font-family: Arial, Helvetica, sans-serif;
}
a {
  color: gray; /* Цвет ссылок */
  text-decoration: none; /* Убираем подчёркивание */
}
</style>
</head>
<body>

<h1>%s</h1>
<p1>powered by <a href="https://vk.com/koplenov">@koplenov</a></p1>

</body>
</html>
`
func main() {
	// Hosts
	hosts := map[string]*Host{}

	//-----
	// API
	//-----

	week := echo.New()

	week.Pre(middleware.HTTPSRedirect())
	week.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")

	week.Use(middleware.Logger())
	week.Use(middleware.Recover())

	hosts["week.kinsle.ru"] = &Host{week}

	week.GET("/", func(c echo.Context) error {
		ts := time.Now().UTC().Unix() //получаем время прошедшее с 1 января 1970 по UTC
		tn := time.Unix(ts, 0) // преобразуем в локальное
		_, week := tn.ISOWeek() //узнаем какая сейчас неделя по счету в этом году

		//четная - верхняя неделя, нечетная - нижняя
		if week % 2 == 0 {
			return c.HTML(http.StatusOK, fmt.Sprintf(htmlTemplate, "Верхняя неделя", "Сейчас верхняя (четная) неделя"))
		} else {
			return c.HTML(http.StatusOK, fmt.Sprintf(htmlTemplate, "Нижняя неделя", "Сейчас нижняя (нечетная) неделя"))
		}
	})
	week.GET("/sitemap.xml", func(c echo.Context) error {
		return c.XML(http.StatusOK,`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd"><url><loc>https://week.kinsle.ru/</loc><lastmod>2020-12-21T00:40:34+00:00</lastmod></url></urlset>`)})
	go week.StartAutoTLS(":443")

	//------
	// Blog
	//------

	blog := echo.New()
	blog.Pre(middleware.NonWWWRedirect())
	blog.Use(middleware.Logger())
	blog.Use(middleware.Recover())

	hosts["blog.kinsle.ru"] = &Host{blog}

	blog.GET("/", func(c echo.Context) error {
		c.HTML(http.StatusOK, "/тут буит блог/")
		return c.String(http.StatusOK, "Blog")
	})


	//------
	// git
	//------

	git := echo.New()

	git.Pre(middleware.HTTPSRedirect())
	git.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")

	git.Use(middleware.Logger())
	git.Use(middleware.Recover())

	hosts["git.kinsle.ru"] = &Host{git}

	// Setup proxy
	//url1, err := url.Parse("http://git.kinsle.ru:80")
	url1, err := url.Parse("https://git.kinsle.ru:80")
	if err != nil {
		git.Logger.Fatal(err)
	}
	url2, err := url.Parse("https://git.kinsle.ru:3000")
	if err != nil {
		git.Logger.Fatal(err)
	}
	targets := []*middleware.ProxyTarget{
		{
			URL: url1,
		},
		{
			URL: url2,
		},
	}
	git.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(targets)))
	go git.StartAutoTLS(":443")

	// Server
	e := echo.New()
	e.Pre(middleware.NonWWWRedirect())
	e.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		host := hosts[req.Host]

		if host == nil {
			err = echo.ErrNotFound
		} else {
			host.Echo.ServeHTTP(res, req)
		}

		return
	})
	e.Logger.Fatal(e.Start(":80"))
}