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

// 全局自定义上下文生成器
var defaultCustomContextFactory CustomContextFactory = nil

type CustomContexter interface {
    iris.Context
    // 请求处理完毕后会调用这个方法, 如果请求处理函数没有返回值会传入nil
    SetResult(a interface{})
}

type CustomContextFactory func(ctx iris.Context) CustomContexter

// 设置全局自定义上下文生成器
func SetDefaultCustomContextFactory(factory CustomContextFactory) {
    defaultCustomContextFactory = factory
}
