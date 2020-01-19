/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/1/17
   Description :
-------------------------------------------------
*/

package ziris

import (
    "fmt"
    "regexp"
    "strconv"
    "time"

    "github.com/kataras/iris/v12"
)

const CtxStartTimeField = "ctx_start"
const DefaultCtxInfoLayout = "[%(status)s] %(latency)s %(ip)s %(method)s %(path)s"

type CtxFormatFlag string

const (
    CtxStatusFlag  CtxFormatFlag = "status"
    CtxLatencyFlag CtxFormatFlag = "latency"
    CtxIPFlag      CtxFormatFlag = "ip"
    CtxMethodFlag  CtxFormatFlag = "method"
    CtxPathFlag    CtxFormatFlag = "path"
)

var ctxInfoformatParser = regexp.MustCompile(`%\(.*?\)s`)

// 设置开始时间
func SetCtxStart(ctx iris.Context) bool {
    return SetCtxStartOf(ctx, time.Now().UnixNano())
}

// 设置指定的开始时间
func SetCtxStartOf(ctx iris.Context, a interface{}) bool {
    _, b := ctx.Values().Set(CtxStartTimeField, a)
    return b
}

// 获取延迟时间
func GetCtxLatency(ctx iris.Context) time.Duration {
    latency := time.Duration(-1)

    if start_time := ctx.Values().Get(CtxStartTimeField); start_time != nil {
        endTime := time.Duration(time.Now().UnixNano())
        start := time.Duration(-1)

        switch v := start_time.(type) {
        case int:
            start = time.Duration(v)
        case int64:
            start = time.Duration(v)

        case uint:
            start = time.Duration(v)
        case uint64:
            start = time.Duration(v)

        case time.Duration:
            start = v

        case time.Time:
            start = time.Duration(v.UnixNano())
        case *time.Time:
            start = time.Duration(v.UnixNano())
        }

        if start != -1 {
            latency = endTime - start
        }
    }
    return latency
}

// 获取描述信息
func GetCtxInfo(ctx iris.Context) string {
    return GetCtxInfoOfLayout(ctx, DefaultCtxInfoLayout)
}

// 获取描述信息并指定样式
func GetCtxInfoOfLayout(ctx iris.Context, layout string) string {
    status := ctx.GetStatusCode()
    latency := GetCtxLatency(ctx)
    ip := ctx.RemoteAddr()
    method := ctx.Method()
    path := ctx.Request().URL.RequestURI()

    s := ctxInfoformatParser.ReplaceAllStringFunc(layout, func(flag string) string {
        flag = flag[2 : len(flag)-2]
        switch CtxFormatFlag(flag) {
        case CtxStatusFlag:
            return strconv.Itoa(status)
        case CtxLatencyFlag:
            return latency.String()
        case CtxIPFlag:
            return ip
        case CtxMethodFlag:
            return method
        case CtxPathFlag:
            return path
        default:
            return fmt.Sprintf("(%%(%s)s)invalid)", flag)
        }
    })
    return s
}

// 日志信息中间件, 用于输出当前请求信息
func CtxLogMiddleware(log interface{ Info(v ...interface{}) }) func(ctx iris.Context) {
    return func(ctx iris.Context) {
        SetCtxStart(ctx)
        ctx.Next()
        log.Info(GetCtxInfo(ctx))
    }
}
