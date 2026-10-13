// Package utils 提供通用工具函数与全局变量初始化。
// 本文件定义了 HTTP 接口统一响应格式工具，采用大写 Key（Code/Msg/Rows/Total/Data），
// 与项目中的登录/注册接口（小写 key）形成双轨约定——详见下文 Resp 系列函数注释。
package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H 是 HTTP 接口统一响应体结构。
//
// 字段说明：
//   - Code  业务状态码，0 表示成功，-1 表示失败
//   - Msg   业务信息，成功时通常为 "xxx成功"，失败时为 "xxx失败|错误原因"
//   - Data  单条数据（用于查询/更新等非列表接口）
//   - Rows  列表数据（用于列表类接口，如好友列表、群列表）
//   - Total 列表条数或历史消息字符串数组（Redis 历史消息接口会把消息 JSON 字符串整体塞进 Total）
//
// 双轨制约定：本结构体导出字段首字母均大写，对应前端用 res.Code / res.Msg 读取；
// 但 /user/login 与 /user/createUser 两个早期接口因前端遗留代码约定，返回小写 code/msg/data，
// 那两个接口直接在 service 层写 gin.H 硬编码，不走本工具函数。
type H struct {
	Code  int
	Msg   string
	Data  interface{}
	Rows  interface{}
	Total interface{}
}

// Resp 向客户端输出统一格式的 JSON 响应（使用 Data 字段承载单条数据）。
//
// 参数：
//   - w    标准库 http.ResponseWriter
//   - code 业务状态码，0 表示成功，-1 表示失败
//   - data 要返回的数据体，可为 nil
//   - msg  业务提示信息
//
// 输出示例：{"Code":0, "Data":{...}, "Msg":"查询用户成功"}
//
// 注意：HTTP 状态码固定为 200，业务层用 Code 区分成功/失败。
func Resp(w http.ResponseWriter, code int, data interface{}, msg string) {
	// 设置响应头为JSON格式
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := H{
		Code: code,
		Data: data,
		Msg:  msg,
	}
	ret, err := json.Marshal(h)
	if err != nil {
		fmt.Println("Resp:", err)
	}
	w.Write(ret)
}

// RespList 向客户端输出列表类接口的 JSON 响应（使用 Rows 字段承载列表数据）。
//
// 参数：
//   - w     标准库 http.ResponseWriter
//   - code  业务状态码
//   - data  列表数据（切片），放入响应体 Rows 字段
//   - total 列表条数或额外统计信息，放入响应体 Total 字段
//
// 典型场景：好友列表、群列表、Redis 历史消息拉取。
// 其中历史消息接口把消息 JSON 字符串数组整体塞进 Total，需前端逐条 JSON.parse。
func RespList(w http.ResponseWriter, code int, data interface{}, total interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := H{
		Code:  code,
		Rows:  data,
		Total: total,
	}
	ret, err := json.Marshal(h)
	if err != nil {
		fmt.Println("RespList:", err)
	}
	w.Write(ret)
}
// RespFail 向客户端输出失败响应的简写，等价于 Resp(w, -1, nil, msg)。
// 常用于参数校验失败、业务逻辑返回错误等场景。
func RespFail(w http.ResponseWriter, msg string) {
	Resp(w, -1, nil, msg)
}

// RespOK 向客户端输出成功响应的简写，等价于 Resp(w, 0, data, msg)。
// 常用于单条数据的查询、更新、上传成功等场景。
func RespOK(w http.ResponseWriter, data interface{}, msg string) {
	Resp(w, 0, data, msg)
}

// RespOKList 向客户端输出成功列表响应的简写，等价于 RespList(w, 0, data, total)。
// 常用于好友列表、群列表等返回多条数据的接口。
func RespOKList(w http.ResponseWriter, data interface{}, total interface{}) {
	RespList(w, 0, data, total)
}
