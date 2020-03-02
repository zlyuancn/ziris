/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/2
   Description :
-------------------------------------------------
*/

package ziris

import (
    "fmt"
)

type ColorType uint8

const (
    ColorDefault = ColorType(iota)      // 默认
    ColorRed     = ColorType(iota + 30) // 红
    ColorGreen                          // 绿
    ColorYellow                         // 黄
    ColorBlue                           // 蓝
    ColorMagenta                        // 紫
    ColorCyan                           // 深绿
    ColorWhite                          // 白
)

// 构建彩色文本
func MakeColorText(color ColorType, a string) string {
    return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, a)
}
