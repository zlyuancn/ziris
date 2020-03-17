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
    "io/ioutil"
    "net/http"
    "testing"
    "time"

    "github.com/kataras/iris/v12"
)

type testCtxInterface interface {
    Fullpath() string
}

type testCtxStruct struct {
    ctx      iris.Context
    fullpath string
}

func (m *testCtxStruct) SetResult(a interface{}) {
    v := a.(*string)
    _, _ = m.ctx.WriteString(*v)
}

func (m *testCtxStruct) Fullpath() string {
    return m.fullpath
}

type TestController int

func (t *TestController) Fn(ctx iris.Context) {
    fmt.Println("method", ctx.Method())
    fmt.Println("path", ctx.Path())
    fmt.Println("fullpath", ctx.Request().URL.RequestURI())
    fmt.Println("params", ctx.Params().Get("params"))
    _, _ = ctx.WriteString("get")
}
func (t *TestController) PostFn(ctx iris.Context) {
    fmt.Println("method", ctx.Method())
    fmt.Println("path", ctx.Path())
    fmt.Println("fullpath", ctx.Request().URL.RequestURI())
    fmt.Println("params", ctx.Params().Get("params"))
    _, _ = ctx.WriteString("post")
}

type TestCustomContextController int

func (t *TestCustomContextController) Fn(a testCtxInterface) *string {
    fmt.Println(a.Fullpath())
    v := "get"
    return &v
}
func (t *TestCustomContextController) PostFn(a *testCtxStruct) *string {
    fmt.Println(a.Fullpath())
    v := "post"
    return &v
}

func testRegistryController(t *testing.T, controller interface{}, factory CustomContextFactory) {
    app := iris.New()
    NewControllerWithCustom(controller, "test", factory).Registry(app)

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
        bs, _ := ioutil.ReadAll(resp.Body)
        _ = resp.Body.Close()

        if string(bs) == "get" {
            t.Log("成功")
        } else {
            t.Fatal("失败了", string(bs))
        }

        resp, err = http.Post("http://127.0.0.1:8080/test/fn/postparam?a=123", "", nil)
        if err != nil {
            t.Fatal(err)
            return
        }
        bs, _ = ioutil.ReadAll(resp.Body)
        _ = resp.Body.Close()

        if string(bs) == "post" {
            t.Log("成功")
        } else {
            t.Fatal("失败了", string(bs))
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
    testRegistryController(t, new(TestCustomContextController), func(ctx iris.Context) CustomContexter {
        return &testCtxStruct{
            ctx:      ctx,
            fullpath: ctx.Request().URL.RequestURI(),
        }
    })
}
