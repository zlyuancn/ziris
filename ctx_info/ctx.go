/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/1/17
   Description :
-------------------------------------------------
*/

package ctx_info

import (
    "fmt"
    "regexp"
    "strconv"
    "time"

    "github.com/kataras/iris/v12"
)

const StartTimeField = "ctx_start"
const DefaultLayout = "[%(status)s] %(latency)s %(ip)s %(method)s %(fullpath)s"

type FormatFlag string

const (
    StatusFlag   FormatFlag = "status"
    LatencyFlag  FormatFlag = "latency"
    IPFlag       FormatFlag = "ip"
    MethodFlag   FormatFlag = "method"
    PathFlag     FormatFlag = "path"
    FullPathFlag FormatFlag = "fullpath"
)

var formatParser = regexp.MustCompile(`%\(.*?\)s`)

// 设置开始时间
func SetStartTime(ctx iris.Context) bool {
    return SetStartTimeOf(ctx, time.Now().UnixNano())
}

// 设置指定的开始时间(纳秒)
func SetStartTimeOf(ctx iris.Context, a interface{}) bool {
    _, b := ctx.Values().Set(StartTimeField, a)
    return b
}

// 获取延迟时间
func GetLatency(ctx iris.Context) time.Duration {
    latency := time.Duration(-1)

    if start_time := ctx.Values().Get(StartTimeField); start_time != nil {
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
func GetInfo(ctx iris.Context) string {
    return GetInfoOfLayout(ctx, DefaultLayout)
}

// 获取描述信息并指定样式
func GetInfoOfLayout(ctx iris.Context, layout string) string {
    status := ctx.GetStatusCode()
    latency := GetLatency(ctx)
    ip := ctx.RemoteAddr()
    method := ctx.Method()
    path := ctx.Path()
    fullpath := ctx.Request().URL.RequestURI()

    s := formatParser.ReplaceAllStringFunc(layout, func(flag string) string {
        flag = flag[2 : len(flag)-2]
        switch FormatFlag(flag) {
        case StatusFlag:
            return strconv.Itoa(status)
        case LatencyFlag:
            return latency.String()
        case IPFlag:
            return ip
        case MethodFlag:
            return method
        case PathFlag:
            return path
        case FullPathFlag:
            return fullpath
        default:
            return fmt.Sprintf("(%%(%s)s)invalid)", flag)
        }
    })
    return s
}

// 日志信息中间件, 用于输出当前请求信息
func LogMiddleware(log interface{ Info(v ...interface{}) }) func(ctx iris.Context) {
    return func(ctx iris.Context) {
        SetStartTime(ctx)
        ctx.Next()
        log.Info(GetInfo(ctx))
    }
}
