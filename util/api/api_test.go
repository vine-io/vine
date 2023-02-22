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
)

func TestNewRPCGateway(t *testing.T) {
	addr := "127.0.0.1:35500"
	s := vine.NewService(vine.Address(addr))

	msg := "hello world"
	ns := "go.vine"
	app := NewRPCGateway(s, ns, func(app *gin.Engine) {
		app.GET("/hello", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, msg)
		})
	})

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
