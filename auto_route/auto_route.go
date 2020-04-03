/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/2/20
   Description :  自动路由
-------------------------------------------------
*/

package auto_route

import (
    "fmt"
    "reflect"
    "strings"

    jsoniter "github.com/json-iterator/go"
    "github.com/kataras/iris/v12"
)

const (
    // 控制器名后缀
    ControllerSuffix = "Controller"
    // 默认请求方法
    DefaultRequestMethod = "Get"
    // 在上下文中保存结尾路径的字段名
    ParamsFieldName = "params"
)

var requestMethods = [...]string{"Get", "Post", "Delete", "Put", "Patch", "Head"}
var typeOfIrisContext = reflect.TypeOf((*iris.Context)(nil)).Elem()

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type methodType struct {
    reqMethod     string // 请求方法
    controlMethod string // 控制器方法
    fn            reflect.Value
}

func (m *methodType) Handler(service *controller, ctx iris.Context) {
    if service.factory != nil {
        a := ctx.(CustomContexter)
        returnValues := m.fn.Call([]reflect.Value{service.rcvr, reflect.ValueOf(a)})
        if len(returnValues) == 1 {
            a.SetResult(returnValues[0].Interface())
        } else {
            a.SetResult(nil)
        }
        return
    }

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

        switch data := v.(type) {
        case []byte:
            _, _ = ctx.Write(data)
        case *[]byte:
            _, _ = ctx.Write(*data)
        default:
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
    name        string
    parentPath  string
    rcvr        reflect.Value
    typ         reflect.Type
    methods     map[string]*methodType
    factory     CustomContextFactory
    reqHandlers []ReqMiddleware
}

// 创建控制器
func NewController(a interface{}) *controller {
    return NewControllerWithCustom(a, "", defaultCustomContextFactory)
}

// 创建控制器并设置控制器名和自定义上下文生成器
func NewControllerWithCustom(a interface{}, name string, factory CustomContextFactory) *controller {
    m := new(controller)
    m.typ = reflect.TypeOf(a)
    m.rcvr = reflect.ValueOf(a)

    if name == "" {
        sname := reflect.Indirect(m.rcvr).Type().Name()
        if sname == "" {
            panic("无法获取控制器的名称")
        }

        if strings.HasSuffix(sname, ControllerSuffix) {
            sname = sname[:len(sname)-len(ControllerSuffix)]
        }

        name = sname
    }

    if name == "" {
        panic("控制器没有名称")
    }

    m.name = snakeString(name)
    m.factory = factory
    m.methods = m.suitableMethods(m.typ)
    return m
}

// 注册控制器
func (m *controller) Registry(party iris.Party, handler ...ReqMiddleware) {
    path := party.GetRelPath()
    if strings.HasSuffix(path, "/") {
        path = path[:len(path)-1]
    }
    m.parentPath = path
    party.CreateRoutes(nil, fmt.Sprintf("/%s", m.name), m.handler)
    party.CreateRoutes(nil, fmt.Sprintf("/%s/{%s:path}", m.name, ParamsFieldName), m.handler)

    m.reqHandlers = append(([]ReqMiddleware)(nil), handler...)
}

// 匹配方法
func (m *controller) suitableMethods(typ reflect.Type) map[string]*methodType {
    methods := make(map[string]*methodType, 0)
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

        reqMethod, controlMethod := m.parserMethod(method.Name)
        key := m.makeMethodKey(reqMethod, controlMethod)
        methods[key] = &methodType{reqMethod: reqMethod, controlMethod: controlMethod, fn: method.Func}
    }
    return methods
}

// 将方法转为为请求方法和控制器方法
func (m *controller) parserMethod(method string) (reqMethod string, controlMethod string) {
    for _, s := range requestMethods {
        if strings.HasPrefix(method, s) {
            return s, snakeString(method[len(s):])
        }
    }
    return DefaultRequestMethod, snakeString(method)
}

// 根据请求方法和控制器方法构建methods的key
func (m *controller) makeMethodKey(reqMethod, controlMethod string) string {
    return fmt.Sprintf("%s/%s", strings.ToLower(reqMethod), controlMethod)
}

func (m *controller) handler(ctx iris.Context) {
    reqMethod := ctx.Method()
    rawParams := ctx.Params().Get(ParamsFieldName)
    rawParams = strings.Trim(rawParams, "/")

    controlMethod, params := rawParams, ""

    // 分离参数
    if k := strings.Index(rawParams, "/"); k != -1 {
        controlMethod, params = rawParams[:k], rawParams[k+1:]
    }

    // 如果没有该方法, 并且存在空方法时, 则方法为空
    if _, ok := m.methods[m.makeMethodKey(reqMethod, controlMethod)]; !ok {
        if _, ok = m.methods[m.makeMethodKey(reqMethod, "")]; ok {
            controlMethod, params = "", rawParams
        }
    }

    reqArg := &ReqArg{
        controlMethod: controlMethod,
        params:        params,
    }

    if m.factory != nil {
        ctx = m.factory(ctx)
        if ctx == nil {
            return
        }
    }

    // 中间件
    for _, handler := range m.reqHandlers {
        handler(ctx, reqArg)
        if reqArg.stop {
            return
        }
    }

    ctx.Params().Save(ParamsFieldName, reqArg.Params(), true)
    control, ok := m.methods[m.makeMethodKey(reqMethod, reqArg.ControlMethod())]
    if !ok {
        ctx.StatusCode(400)
        _, _ = ctx.WriteString(fmt.Sprintf("未定义的路由: [%s] <%s/%s/%s>", reqMethod, m.parentPath, m.name, reqArg.ControlMethod()))
        return
    }

    control.Handler(m, ctx)
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
func RegistryController(party iris.Party, a interface{}, handler ...ReqMiddleware) {
    RegistryControllerWithCustom(party, a, "", defaultCustomContextFactory, handler...)
}

// 注册控制器并设置控制器名
func RegistryControllerWithName(party iris.Party, a interface{}, name string, handler ...ReqMiddleware) {
    RegistryControllerWithCustom(party, a, name, defaultCustomContextFactory, handler...)
}

// 注册控制器并设置自定义上下文生成器
func RegistryControllerWithFactory(party iris.Party, a interface{}, factory CustomContextFactory, handler ...ReqMiddleware) {
    RegistryControllerWithCustom(party, a, "", factory, handler...)
}

// 注册控制器并设置控制器名和自定义上下文生成器
func RegistryControllerWithCustom(party iris.Party, a interface{}, name string, factory CustomContextFactory, handler ...ReqMiddleware) {
    service := NewControllerWithCustom(a, name, factory)
    service.Registry(party, handler...)
}
