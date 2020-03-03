/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/2/20
   Description :  自动路由
-------------------------------------------------
*/

package auto_route

import (
    "errors"
    "fmt"
    "reflect"
    "strings"

    jsoniter "github.com/json-iterator/go"
    "github.com/kataras/iris/v12"
    "github.com/kataras/iris/v12/context"
    "github.com/kataras/iris/v12/core/router"
)

const (
    // 控制器名后缀
    ControllerSuffix = "Controller"
    // 默认请求方法
    DefaultRequestMethod = "Get"
    // 在上下文中保存结尾路径的字段名
    ParamsFieldName = "params"
)

// 全局自定义上下文生成器
var defaultCustomContextFactory CustomContextFactory = nil

var requestMethods = [...]string{"Get", "Post", "Delete", "Put", "Patch", "Head"}
var typeOfIrisContext = reflect.TypeOf((*iris.Context)(nil)).Elem()

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type CustomContexter interface {
    // 请求处理完毕后会调用这个方法, 如果请求处理函数没有返回值会传入nil
    SetResult(a interface{})
}
type CustomContextFactory func(ctx iris.Context) CustomContexter

type methodType struct {
    method string // 请求方法
    path   string // 请求路径
    fn     reflect.Value
}

func (m *methodType) MakeIrisHandler(service *controller) iris.Handler {
    if service.factory != nil {
        factory := service.factory
        return func(ctx context.Context) {
            a := factory(ctx)
            returnValues := m.fn.Call([]reflect.Value{service.rcvr, reflect.ValueOf(a)})
            if len(returnValues) == 1 {
                a.SetResult(returnValues[0].Interface())
            } else {
                a.SetResult(nil)
            }
        }
    }

    return func(ctx context.Context) {
        returnValues := m.fn.Call([]reflect.Value{service.rcvr, reflect.ValueOf(ctx)})
        if len(returnValues) == 1 {
            v := returnValues[0].Interface()
            if v == nil {
                return
            }

            if err, ok := v.(error); ok {
                _, _ = ctx.WriteString(err.Error())
                return
            }

            bs, err := json.Marshal(v)
            if err != nil {
                ctx.StatusCode(500)
                _, _ = ctx.WriteString(err.Error())
            }
            _, _ = ctx.Write(bs)
        }
    }
}

type controller struct {
    name    string
    rcvr    reflect.Value
    typ     reflect.Type
    methods []*methodType
    factory CustomContextFactory
}

func (m *controller) Parse(a interface{}, name string, factory CustomContextFactory) error {
    m.typ = reflect.TypeOf(a)
    m.rcvr = reflect.ValueOf(a)

    sname := reflect.Indirect(m.rcvr).Type().Name()
    if sname == "" {
        return errors.New("无法获取控制器的名称")
    }
    if !strings.HasSuffix(sname, ControllerSuffix) {
        return fmt.Errorf("控制器 <%s> 后缀不是 %s", sname, ControllerSuffix)
    }
    sname = sname[:len(sname)-len(ControllerSuffix)]

    if name != "" {
        sname = name
    }
    if sname == "" {
        return errors.New("控制器没有名称")
    }

    m.name = snakeString(sname)
    m.factory = factory
    m.methods = m.suitableMethods(m.typ)
    return nil
}

// 匹配方法
func (m *controller) suitableMethods(typ reflect.Type) []*methodType {
    methods := make([]*methodType, 0)
    for i := 0; i < typ.NumMethod(); i++ {
        method := typ.Method(i)
        mtype := method.Type

        // 未导出的方法过滤掉
        if method.PkgPath != "" {
            continue
        }

        // 包括自己本身和接收参数数量
        if mtype.NumIn() != 2 {
            continue
        }

        if m.factory == nil {
            // 第一个参数必须是 iris.Context
            ctxType := mtype.In(1)
            if !ctxType.Implements(typeOfIrisContext) {
                continue
            }
        } else {
            // 第一个参数必须是指针或者接口
            replyType := mtype.In(1)
            kind := replyType.Kind()
            if kind != reflect.Ptr && kind != reflect.Interface {
                continue
            }
        }

        // 方法最多只能有一个输出
        if mtype.NumOut() > 1 {
            continue
        }

        reqMethod, path := m.parserMethod(method.Name)
        methods = append(methods, &methodType{method: reqMethod, path: path, fn: method.Func})
    }
    return methods
}

// 将方法转为为请求方法和路径
func (m *controller) parserMethod(method string) (string, string) {
    for _, s := range requestMethods {
        if strings.HasPrefix(method, s) {
            return s, snakeString(method[len(s):])
        }
    }
    return DefaultRequestMethod, snakeString(method)
}

// 转为蛇形字符串
func snakeString(s string) string {
    data := make([]byte, 0, len(s)*2)
    j := false
    num := len(s)
    for i := 0; i < num; i++ {
        d := s[i]
        if i > 0 && d >= 'A' && d <= 'Z' && j {
            data = append(data, '_')
        }
        if d != '_' {
            j = true
        }
        data = append(data, d)
    }
    return strings.ToLower(string(data[:]))
}

// 注册控制器
// 控制器对象字段名必须以 Controller 结尾
// 如果有一个控制器 TestController 并且它有导出的方法 Fn(iris.Context), 那么会自动注册 Get  /test/fn
// 导出的方法可以控制请求方法, 如 TestController.PostFn 表示 Post /xxx/fn
// 当然, 请求路径可以为空, 如 TestController.Post 表示 Post /xxx
// 请求路径末尾的数据请使用 ctx.Params().Get("params") 来获取值
func RegistryController(party iris.Party, a interface{}) {
    RegistryControllerWithCustom(party, a, "", defaultCustomContextFactory)
}

// 注册控制器并设置控制器名
func RegistryControllerWithName(party iris.Party, a interface{}, name string) {
    RegistryControllerWithCustom(party, a, name, defaultCustomContextFactory)
}

// 注册控制器并设置自定义上下文生成器
func RegistryControllerWithFactory(party iris.Party, a interface{}, factory CustomContextFactory) {
    RegistryControllerWithCustom(party, a, "", factory)
}

// 注册控制器并设置控制器名和自定义上下文生成器
func RegistryControllerWithCustom(party iris.Party, a interface{}, name string, factory CustomContextFactory) {
    service := new(controller)
    if err := service.Parse(a, name, factory); err != nil {
        panic(err)
    }

    for _, method := range service.methods {
        var fn func(string, ...context.Handler) *router.Route
        switch strings.ToLower(method.method) {
        case "get":
            fn = party.Get
        case "post":
            fn = party.Post
        case "delete":
            fn = party.Delete
        case "put":
            fn = party.Put
        case "patch":
            fn = party.Patch
        case "head":
            fn = party.Head
        default:
            panic(fmt.Sprintf("未知的请求方法<%s>", method.method))
        }

        handler := method.MakeIrisHandler(service)
        if method.path != "" {
            fn(fmt.Sprintf("/%s/%s", service.name, method.path), handler)
            fn(fmt.Sprintf("/%s/%s/{%s:path}", service.name, method.path, ParamsFieldName), handler)
        } else {
            fn(fmt.Sprintf("/%s", service.name), handler)
            fn(fmt.Sprintf("/%s/{%s:path}", service.name, ParamsFieldName), handler)
        }
    }
}

// 设置全局自定义上下文生成器
func SetDefaultCustomContextFactory(factory CustomContextFactory) {
    defaultCustomContextFactory = factory
}
