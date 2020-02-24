/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/11/5
   Description :
-------------------------------------------------
*/

package ziris

import (
    "fmt"
    "io"
)

type ErrorWithCode struct {
    err  error
    code int
}

func WrapErrorWithCode(err error, code int) error {
    return &ErrorWithCode{
        err:  err,
        code: code,
    }
}

func (m *ErrorWithCode) Err() error {
    return m.err
}
func (m *ErrorWithCode) Code() int {
    return m.code
}

func (m *ErrorWithCode) Error() string {
    return m.err.Error()
}

func (m *ErrorWithCode) Format(s fmt.State, verb rune) {
    switch verb {
    case 'v':
        if s.Flag('+') {
            _, _ = fmt.Fprintf(s, "%+v", m.err)
            return
        }
        fallthrough
    case 's':
        _, _ = io.WriteString(s, m.err.Error())
    case 'q':
        _, _ = fmt.Fprintf(s, "%q", m.err.Error())
    }
}
