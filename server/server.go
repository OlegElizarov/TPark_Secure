package server

import (
	"fmt"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
)

type Server struct {
	port string
	e    *echo.Echo
}

func NewServer(port string, e *echo.Echo) *Server {
	e.GET("/$", Resp)
	e.Any("*", Transfer)
	e.CONNECT("/", TransferSSL)

	return &Server{
		port: port,
		e:    e,
	}
}

func Resp(ctx echo.Context) error {
	fmt.Println(ctx.Request().URL)
	return ctx.String(http.StatusOK, "Hello, World!!!")
}

func Transfer(ctx echo.Context) error {
	var resp *http.Response
	var err error
	switch ctx.Request().Method {
	case "GET":
		resp, err = http.Get(ctx.Request().URL.String())
	case "POST":
		resp, err = http.Post(ctx.Request().URL.String(), ctx.Request().Header.Get("Content-Type"), ctx.Request().Body)
	default:
		resp, err = http.Get(ctx.Request().URL.String())
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	for mime, val := range resp.Header {
		ctx.Response().Header().Set(mime, val[0])
	}
	ctx.Response().Header().Set("Content-Type", string(resp.Header.Get("Content-Type"))+"; charset=utf8")
	return ctx.String(resp.StatusCode, string(body))

}

func TransferSSL(ctx echo.Context) error {
	//resp, err := http.Get(ctx.QueryString())
	//if err != nil {
	//	return err
	//}
	//defer resp.Body.Close()
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}
	//return ctx.String(http.StatusOK, string(body))
	return ctx.String(http.StatusOK, "Transfer")

}

func (s Server) ListenAndServe() error {
	return s.e.Start(s.port)
}
