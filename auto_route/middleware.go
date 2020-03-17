/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/17
   Description :
-------------------------------------------------
*/

package auto_route

import (
    "github.com/kataras/iris/v12"
)

// 请求中间件, 会在构建自定义上下文之前调用
type ReqMiddleware func(ctx iris.Context, arg *ReqArg)

// 请求参数
type ReqArg struct {
    // 控制器方法
    controlMethod string
    // 请求参数
    params string
    // 是否停止
    stop bool
}

// 停止请求
func (m *ReqArg) Stop() {
    m.stop = true
}

// 返回是否调用了Stop()
func (m *ReqArg) IsStop() bool {
    return m.stop
}

// 返回控制器方法
func (m *ReqArg) ControlMethod() string {
    return m.controlMethod
}

// 返回路径参数
func (m *ReqArg) Params() string {
    return m.params
}

// 设置控制器方法, 注意, 它应该是蛇形的
func (m *ReqArg) SetControlMethod(method string) {
    m.controlMethod = method
}

// 设置路径参数
func (m *ReqArg) SetParams(params string) {
    m.params = params
}
