/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/2/20
   Description :
-------------------------------------------------
*/

package auto_route

import (
    "context"
    "fmt"
    "net/http"
    "testing"
    "time"

    "github.com/kataras/iris/v12"
)

var testValue string

type testCtxInterface interface {
    Fullpath() string
}

type testCtxStruct struct {
    fullpath string
}

func (m *testCtxStruct) Fullpath() string {
    return m.fullpath
}

type TestController int

func (t *TestController) Fn(ctx iris.Context) {
    testValue = "get"
    fmt.Println("method", ctx.Method())
    fmt.Println("path", ctx.Path())
    fmt.Println("fullpath", ctx.Request().URL.RequestURI())
    fmt.Println("params", ctx.Params().Get("params"))
}
func (t *TestController) PostFn(ctx iris.Context) {
    testValue = "post"
    fmt.Println("method", ctx.Method())
    fmt.Println("path", ctx.Path())
    fmt.Println("fullpath", ctx.Request().URL.RequestURI())
    fmt.Println("params", ctx.Params().Get("params"))
}

type TestCustomContextController int

func (t *TestCustomContextController) Fn(a testCtxInterface) {
    testValue = "get"
    fmt.Println(a.Fullpath())
}
func (t *TestCustomContextController) PostFn(a *testCtxStruct) {
    testValue = "post"
    fmt.Println(a.Fullpath())
}

func testRegistryController(t *testing.T, controller interface{}, factory CustomContextFactory) {
    app := iris.New()
    RegistryControllerWithCustom(app, controller, "test", factory)

    go func() {
        time.Sleep(1e9)
        defer func() {
            _ = app.Shutdown(context.Background())
        }()

        resp, err := http.Get("http://127.0.0.1:8080/test/fn/getparam?a=123")
        if err != nil {
            t.Fatal(err)
            return
        }
        _ = resp.Body.Close()

        if testValue == "get" {
            t.Log("成功")
        } else {
            t.Fatal("失败了")
        }

        resp, err = http.Post("http://127.0.0.1:8080/test/fn/postparam?a=123", "", nil)
        if err != nil {
            t.Fatal(err)
            return
        }
        _ = resp.Body.Close()

        if testValue == "post" {
            t.Log("成功")
        } else {
            t.Fatal("失败了")
        }
    }()

    err := app.Run(iris.Addr(":8080"), iris.WithoutPathCorrection)
    if err != nil && err != iris.ErrServerClosed {
        t.Fatal(err)
    }
}

func TestRegistryController(t *testing.T) {
    testRegistryController(t, new(TestController), nil)
}
func TestRegistryWithCustomController(t *testing.T) {
    testRegistryController(t, new(TestCustomContextController), func(ctx iris.Context) interface{} {
        return &testCtxStruct{
            fullpath: ctx.Request().URL.RequestURI(),
        }
    })
}
