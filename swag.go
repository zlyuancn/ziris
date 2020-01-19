/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/1/19
   Description :
-------------------------------------------------
*/

package ziris

import (
    "github.com/kataras/iris/v12"
    httpSwagger "github.com/swaggo/http-swagger"
)

// 安装swagger
// swag i -g xxx.go -o ./docs
// swag i -g xxx.go -o ./docs --parseDependency   解析外部依赖, 注意, 非常慢
func SetupSwagger(ver iris.Party, path string) {
    p := ver.GetRelPath()
    if p == "/" {
        p = ""
    }

    ver.Get(path, func(ctx iris.Context) {
        ctx.Redirect(p+path+"/index.html", 301)
    })

    // fn := httpSwagger.Handler(httpSwagger.URL(p + path + "/doc.json"))
    fn := httpSwagger.Handler()
    ver.Get(path+"/{any:path}", func(ctx iris.Context) {
        fn(ctx.ResponseWriter(), ctx.Request())
    })
}
