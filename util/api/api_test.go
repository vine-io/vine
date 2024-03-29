package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vine-io/vine"
	"github.com/vine-io/vine/core/server/grpc"
	ahandler "github.com/vine-io/vine/lib/api/handler"
)

func TestNewPrimpHandler(t *testing.T) {
	addr := "127.0.0.1:35500"
	s := vine.NewService(vine.Address(addr))

	msg := "hello world"
	ns := "go.vine"
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, msg)
	})

	PrimpHandler(ns, s.Name(), app, s.Client(), ahandler.WithMetadata(map[string]string{"request": "api"}))

	s.Server().Init(grpc.HttpHandler(app))

	s.Init()

	go s.Run()

	time.Sleep(time.Second * 1)
	rsp, err := http.Get("http://" + addr + "/hello")
	if err != nil {
		t.Fatal(err)
	}
	defer rsp.Body.Close()

	out, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatal(err)
	}

	outS := strings.Trim(string(out), `"`)
	if outS != msg {
		t.Fatalf("want %s got %s", msg, outS)
	}
}
