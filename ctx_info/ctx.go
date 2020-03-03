/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/1/17
   Description :
-------------------------------------------------
*/

package ctx_info

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "regexp"
    "strconv"
    "strings"
    "time"

    "github.com/kataras/iris/v12"

    "github.com/zlyuancn/ziris"
)

// 开始时间标记
const StartTimeField = "ctx_start"

const DefaultLayout = "[%(cstatus)s] %(clatency)s %(ip)s %(cmethod)s %(fullpath)s%(brbody)s"
const DefaultLayoutWithHeader = "[%(cstatus)s] %(clatency)s %(ip)s %(cmethod)s %(fullpath)s%(brheader)s%(brbody)s"

type FormatFlag string

const (
    // http状态码
    StatusFlag FormatFlag = "status"
    // 彩色的http状态码
    CStatusFlag FormatFlag = "cstatus"
    // 延迟时间(处理时间)
    LatencyFlag FormatFlag = "latency"
    // 彩色的延迟时间(处理时间), 达到1秒黄色, 达到2秒红色
    CLatencyFlag FormatFlag = "clatency"
    // 客户端ip
    IPFlag FormatFlag = "ip"
    // 请求方法
    MethodFlag FormatFlag = "method"
    // 彩色的请求方法
    CMethodFlag FormatFlag = "cmethod"
    // 请求路径
    PathFlag FormatFlag = "path"
    // 请求路径和请求参数(get参数)
    FullPathFlag FormatFlag = "fullpath"
    // 请求体, 注意设置 iris.WithoutBodyConsumptionOnUnmarshal 选项, 否则无法读出body
    BodyFlag FormatFlag = "body"
    // 和BodyFlag相同, 但是在输出body之前会输出换行符号"\n"
    BrBodyFlag FormatFlag = "brbody"
    // header
    HeaderFlag FormatFlag = "header"
    // 和HeaderFlag相同, 但是在输出header之前会输出换行符号"\n"
    BrHeaderFlag FormatFlag = "brheader"
)

var formatParser = regexp.MustCompile(`%\(.*?\)s`)

// 设置开始时间, 将当前时间戳放入 ctx.Values() 的 StartTimeField 字段中
func SetStartTime(ctx iris.Context) bool {
    return SetStartTimeOf(ctx, time.Now().UnixNano())
}

// 设置指定的开始时间(纳秒)
func SetStartTimeOf(ctx iris.Context, a interface{}) bool {
    _, b := ctx.Values().SetImmutable(StartTimeField, a)
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
    s := formatParser.ReplaceAllStringFunc(layout, func(flag string) string {
        flag = flag[2 : len(flag)-2]
        switch FormatFlag(flag) {
        case StatusFlag:
            return strconv.Itoa(ctx.GetStatusCode())
        case CStatusFlag:
            code := ctx.GetStatusCode()
            code_text := strconv.Itoa(code)
            if code < 200 {
                return ziris.MakeColorText(ziris.ColorBlue, code_text)
            } else if code < 300 {
                return ziris.MakeColorText(ziris.ColorCyan, code_text)
            } else if code < 400 {
                return ziris.MakeColorText(ziris.ColorYellow, code_text)
            } else if code < 600 {
                return ziris.MakeColorText(ziris.ColorRed, code_text)
            }
            return code_text
        case LatencyFlag:
            return GetLatency(ctx).String()
        case CLatencyFlag:
            latency := GetLatency(ctx)
            latency_text := latency.String()
            if latency < 1e9 {
                return latency_text
            } else if latency < 2e9 {
                return ziris.MakeColorText(ziris.ColorYellow, latency_text)
            }
            return ziris.MakeColorText(ziris.ColorRed, latency_text)
        case IPFlag:
            return ctx.RemoteAddr()
        case MethodFlag:
            return ctx.Method()
        case CMethodFlag:
            method := ctx.Method()
            switch strings.ToLower(method) {
            case "get":
                return ziris.MakeColorText(ziris.ColorCyan, method)
            case "post":
                return ziris.MakeColorText(ziris.ColorBlue, method)
            case "put":
                return ziris.MakeColorText(ziris.ColorYellow, method)
            case "delete":
                return ziris.MakeColorText(ziris.ColorRed, method)
            case "head":
                return ziris.MakeColorText(ziris.ColorYellow, method)
            case "patch":
                return ziris.MakeColorText(ziris.ColorYellow, method)
            case "options":
                return ziris.MakeColorText(ziris.ColorYellow, method)
            }
            return method
        case PathFlag:
            return ctx.Path()
        case FullPathFlag:
            return ctx.Request().URL.RequestURI()
        case HeaderFlag, BrHeaderFlag:
            var buff bytes.Buffer
            if FormatFlag(flag) == BrHeaderFlag {
                buff.WriteString("\n")
            }

            header := ctx.Request().Header
            hm := make(map[string]interface{}, len(header))
            for k, v := range header {
                switch len(v) {
                case 1:
                    hm[k] = v[0]
                case 0:
                    hm[k] = ""
                default:
                    hm[k] = v
                }
            }

            h, _ := json.MarshalIndent(hm, "", "    ")
            buff.Write(h)

            return buff.String()
        case BodyFlag, BrBodyFlag:
            body := ctx.Request().Body
            if body == nil {
                return ""
            }

            bodyCopy, _ := ioutil.ReadAll(body)
            if len(bodyCopy) == 0 {
                return ""
            }

            var buff bytes.Buffer
            if FormatFlag(flag) == BrBodyFlag {
                buff.WriteString("\n")
            }
            buff.Write(bodyCopy)

            ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyCopy))
            return buff.String()
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
