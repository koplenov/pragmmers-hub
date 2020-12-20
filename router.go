package main

import (
	"fmt"
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


	//------
	// week
	//------

	week := echo.New()
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

	// Server
	e := echo.New()
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

	//start
	e.Start(":80")

	//------
	// Reverse Proxy Recipe
	//------
	e = echo.New()

	url1, err := url.Parse("http://git.kinsle.ru:80")
	if err != nil {
		e.Logger.Fatal(err)
	}
	url2, err := url.Parse("http://git.kinsle.ru:3000")
	if err != nil {
		e.Logger.Fatal(err)
	}
	targets := []*middleware.ProxyTarget{
		{
			URL: url1,
		},
		{
			URL: url2,
		},
	}
	e.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(targets)))
	e.Logger.Fatal(e.Start(":1323"))

}