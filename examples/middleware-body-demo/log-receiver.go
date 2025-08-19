package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// DataMessage 数据消息结构
type DataMessage struct {
	Timestamp    string      `json:"timestamp"`
	ModuleCode   string      `json:"moduleCode"`
	ActionCode   string      `json:"actionCode"`
	RequestBody  interface{} `json:"requestBody,omitempty"`
	ResponseBody interface{} `json:"responseBody,omitempty"`
	Method       string      `json:"method"`
	Path         string      `json:"path"`
	Headers      interface{} `json:"headers,omitempty"`
}

// 注意：HTTP binding output 直接发送 data 内容作为 body，不需要 BindingRequest 包装

func main() {
	// 创建HTTP服务器来接收日志
	http.HandleFunc("/api/logs", handleLogs)

	fmt.Println("日志接收服务启动在端口 8090")
	fmt.Println("接收端点: http://localhost:8090/api/logs")

	log.Fatal(http.ListenAndServe(":8090", nil))
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	// 先读取原始数据进行调试
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求体失败", http.StatusBadRequest)
		return
	}

	fmt.Printf("\n=== 接收到的原始数据 ===\n")
	fmt.Printf("Content-Type: %s\n", r.Header.Get("Content-Type"))
	fmt.Printf("Headers:\n")
	for k, v := range r.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}
	fmt.Printf("Raw Body: %s\n", string(body))
	fmt.Printf("========================\n")

	// HTTP binding output 直接发送 data 内容，直接解析为 DataMessage
	var dataMsg DataMessage
	if err := json.Unmarshal(body, &dataMsg); err != nil {
		fmt.Printf("无法解析为 DataMessage 格式: %v\n", err)
		http.Error(w, "无效的JSON数据", http.StatusBadRequest)
		return
	}

	fmt.Printf("解析为 DataMessage: %+v\n", dataMsg)
	fmt.Printf("详细字段检查:\n")
	fmt.Printf("  Timestamp: '%s'\n", dataMsg.Timestamp)
	fmt.Printf("  ModuleCode: '%s'\n", dataMsg.ModuleCode)
	fmt.Printf("  ActionCode: '%s'\n", dataMsg.ActionCode)
	fmt.Printf("  Method: '%s'\n", dataMsg.Method)
	fmt.Printf("  Path: '%s'\n", dataMsg.Path)
	// 显示请求体信息
	if dataMsg.RequestBody != nil {
		if reqBodyBytes, err := json.Marshal(dataMsg.RequestBody); err == nil {
			fmt.Printf("  RequestBody (JSON): %s\n", string(reqBodyBytes))
		} else {
			fmt.Printf("  RequestBody: %v\n", dataMsg.RequestBody)
		}
	} else {
		fmt.Printf("  RequestBody: nil\n")
	}

	// 显示响应体信息
	if dataMsg.ResponseBody != nil {
		if respBodyBytes, err := json.Marshal(dataMsg.ResponseBody); err == nil {
			fmt.Printf("  ResponseBody (JSON): %s\n", string(respBodyBytes))
		} else {
			fmt.Printf("  ResponseBody: %v\n", dataMsg.ResponseBody)
		}
	} else {
		fmt.Printf("  ResponseBody: nil\n")
	}
	// 显示 Headers 信息
	if dataMsg.Headers != nil {
		if headersBytes, err := json.Marshal(dataMsg.Headers); err == nil {
			fmt.Printf("  Headers (JSON): %s\n", string(headersBytes))
		} else {
			fmt.Printf("  Headers: %v\n", dataMsg.Headers)
		}
	} else {
		fmt.Printf("  Headers: nil\n")
	}

	// 处理数据消息
	processDataMessage(dataMsg)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("日志已接收"))
}

func processDataMessage(dataMsg DataMessage) {
	// 在这里实现您的数据处理逻辑
	fmt.Printf("\n=== 收到数据消息 ===\n")
	fmt.Printf("时间戳: %s\n", dataMsg.Timestamp)
	fmt.Printf("模块码: %s\n", dataMsg.ModuleCode)
	fmt.Printf("动作码: %s\n", dataMsg.ActionCode)
	fmt.Printf("请求方法: %s\n", dataMsg.Method)
	fmt.Printf("请求路径: %s\n", dataMsg.Path)
	// 显示请求体
	if dataMsg.RequestBody != nil {
		if reqBodyBytes, err := json.Marshal(dataMsg.RequestBody); err == nil {
			fmt.Printf("请求体: %s\n", string(reqBodyBytes))
		} else {
			fmt.Printf("请求体: %v\n", dataMsg.RequestBody)
		}
	} else {
		fmt.Printf("请求体: <空>\n")
	}

	// 显示响应体
	if dataMsg.ResponseBody != nil {
		if respBodyBytes, err := json.Marshal(dataMsg.ResponseBody); err == nil {
			fmt.Printf("响应体: %s\n", string(respBodyBytes))
		} else {
			fmt.Printf("响应体: %v\n", dataMsg.ResponseBody)
		}
	} else {
		fmt.Printf("响应体: <空>\n")
	}
	// 显示请求头
	if dataMsg.Headers != nil {
		if headersBytes, err := json.Marshal(dataMsg.Headers); err == nil {
			fmt.Printf("请求头: %s\n", string(headersBytes))
		} else {
			fmt.Printf("请求头: %v\n", dataMsg.Headers)
		}
	}
	fmt.Printf("处理时间: %s\n", time.Now().Format(time.RFC3339))
	fmt.Println("==================")

	// 示例：根据模块码和动作码进行不同的处理
	switch dataMsg.ModuleCode {
	case "USER_MODULE":
		handleUserModuleLogs(dataMsg)
	case "ORDER_MODULE":
		handleOrderModuleLogs(dataMsg)
	default:
		handleGenericLogs(dataMsg)
	}
}

func handleUserModuleLogs(dataMsg DataMessage) {
	switch dataMsg.ActionCode {
	case "CREATE_USER":
		fmt.Println("📝 用户创建操作已记录")
		// 可以在这里实现用户创建的特殊处理逻辑
		// 例如：发送通知、更新统计、触发其他流程等
	case "UPDATE_USER":
		fmt.Println("✏️ 用户更新操作已记录")
	case "DELETE_USER":
		fmt.Println("🗑️ 用户删除操作已记录")
	}
}

func handleOrderModuleLogs(dataMsg DataMessage) {
	switch dataMsg.ActionCode {
	case "CREATE_ORDER":
		fmt.Println("🛒 订单创建操作已记录")
	case "UPDATE_ORDER":
		fmt.Println("📦 订单更新操作已记录")
	case "CANCEL_ORDER":
		fmt.Println("❌ 订单取消操作已记录")
	}
}

func handleGenericLogs(dataMsg DataMessage) {
	fmt.Printf("📊 通用数据处理: %s.%s\n", dataMsg.ModuleCode, dataMsg.ActionCode)
}
