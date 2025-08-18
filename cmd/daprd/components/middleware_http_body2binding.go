//go:build allcomponents || stablecomponents

/*
Copyright 2024 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package components

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	contribmiddleware "github.com/dapr/components-contrib/middleware"
	httpMiddlewareLoader "github.com/dapr/dapr/pkg/components/middleware/http"
	"github.com/dapr/dapr/pkg/middleware"
	runtimev1pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/kit/logger"
)

// DataMessage 定义发送到 binding 的数据消息结构
type DataMessage struct {
	Timestamp    string `json:"timestamp"`
	ModuleCode   string `json:"moduleCode"`
	ActionCode   string `json:"actionCode"`
	RequestBody  string `json:"requestBody"`
	ResponseBody string `json:"responseBody"`
	Method       string `json:"method"`
	Path         string `json:"path"`
}

// sendDataToBinding 通过 gRPC API 发送数据消息到指定的 binding
func sendDataToBinding(bindingName, daprGRPCPort string, dataMsg DataMessage, log logger.Logger) {
	if bindingName == "" {
		return
	}

	// 异步发送以避免阻塞请求
	go func() {
		// 序列化数据消息
		jsonData, err := json.Marshal(dataMsg)
		if err != nil {
			log.Errorf("Marshal data failed: %v", err)
			return
		}

		// 连接到 Dapr gRPC API
		conn, err := grpc.Dial(
			fmt.Sprintf("localhost:%s", daprGRPCPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Errorf("gRPC connect failed: %v", err)
			return
		}
		defer conn.Close()

		client := runtimev1pb.NewDaprClient(conn)

		// 调用 binding
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := &runtimev1pb.InvokeBindingRequest{
			Name:      bindingName,
			Data:      jsonData,
			Operation: "create",
			Metadata:  map[string]string{},
		}

		log.Debugf("Sending data to binding %s via gRPC", bindingName)

		_, err = client.InvokeBinding(ctx, req)
		if err != nil {
			log.Errorf("Binding %s invoke failed: %v", bindingName, err)
		}
	}()
}

func init() {
	httpMiddlewareLoader.DefaultRegistry.RegisterComponent(func(log logger.Logger) httpMiddlewareLoader.FactoryMethod {
		return func(metadata contribmiddleware.Metadata) (middleware.HTTP, error) {
			// 获取日志文件路径，默认为 middleware_body.log
			logFile := metadata.Properties["logFile"]

			// 获取 binding 名称，如果配置了则使用 binding 发送日志
			bindingName := metadata.Properties["bindingName"]

			// 获取 Dapr gRPC 端口，优先从环境变量获取
			daprGRPCPort := os.Getenv("DAPR_GRPC_PORT")
			if daprGRPCPort == "" {
				daprGRPCPort = metadata.Properties["daprGRPCPort"]
				if daprGRPCPort == "" {
					daprGRPCPort = "50001"
				}
			}

			// 获取是否记录请求体，默认为 true
			logRequest := metadata.Properties["logRequest"] != "false"

			// 获取是否记录响应体，默认为 true
			logResponse := metadata.Properties["logResponse"] != "false"

			// 获取最大记录体大小，默认为 1MB
			maxBodySize := int64(1024 * 1024) // 1MB
			if size := metadata.Properties["maxBodySize"]; size != "" {
				if parsed, err := fmt.Sscanf(size, "%d", &maxBodySize); parsed != 1 || err != nil {
					log.Warnf("Invalid maxBodySize value: %s, using default 1MB", size)
					maxBodySize = 1024 * 1024
				}
			}

			// 获取模块码和操作码的 header 字段名，默认值
			moduleHeader := metadata.Properties["moduleHeader"]
			if moduleHeader == "" {
				moduleHeader = "X-Module-Code"
			}

			actionHeader := metadata.Properties["actionHeader"]
			if actionHeader == "" {
				actionHeader = "X-Action-Code"
			}

			// 获取需要记录日志的HTTP方法，默认为 POST,PUT,DELETE
			logMethods := metadata.Properties["logMethods"]
			if logMethods == "" {
				logMethods = "POST,PUT,DELETE"
			}

			// 解析允许的HTTP方法
			allowedMethods := make(map[string]bool)
			for _, method := range strings.Split(logMethods, ",") {
				method = strings.TrimSpace(strings.ToUpper(method))
				if method != "" {
					allowedMethods[method] = true
				}
			}

			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var requestBodyStr string

					// 记录请求体
					if logRequest && r.Body != nil {
						requestBody, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
						if err != nil {
							log.Errorf("Read request body failed: %v", err)
							requestBodyStr = fmt.Sprintf("Read error: %v", err)
						} else {
							requestBodyStr = string(requestBody)
							// 重置请求体供后续处理使用
							r.Body = io.NopCloser(bytes.NewReader(requestBody))
						}
					}

					var responseBodyStr string
					var bodyRecorder *bodyResponseWriter

					// 如果需要记录响应体，包装响应写入器
					if logResponse {
						bodyRecorder = &bodyResponseWriter{
							ResponseWriter: w,
							body:           &bytes.Buffer{},
							maxSize:        maxBodySize,
						}
						w = bodyRecorder
					}

					// 调用下一个处理器
					next.ServeHTTP(w, r)

					// 获取响应体
					if logResponse && bodyRecorder != nil {
						responseBodyStr = bodyRecorder.body.String()
					}

					// 记录到日志文件
					if logRequest || logResponse {
						// 检查当前请求方法是否在允许记录的方法列表中
						if !allowedMethods[r.Method] {
							return
						}

						// 从 header 中获取模块码和动作码
						moduleCode := r.Header.Get(moduleHeader)
						actionCode := r.Header.Get(actionHeader)

						// 如果模块码或动作码为空，则不记录日志
						if moduleCode == "" || actionCode == "" {
							log.Debugf("Skipping log due to missing headers: moduleCode=%s, actionCode=%s", moduleCode, actionCode)
							return
						}

						// 五列一行格式：模块码 | 动作码 | 请求体 | 响应体 | 时间戳
						var requestBody, responseBody string
						if logRequest {
							requestBody = requestBodyStr
						} else {
							requestBody = ""
						}

						if logResponse {
							responseBody = responseBodyStr
						} else {
							responseBody = ""
						}

						// 将换行符和分隔符替换为空格，确保一行记录
						requestBody = strings.ReplaceAll(strings.ReplaceAll(requestBody, "\n", " "), "|", "｜")
						responseBody = strings.ReplaceAll(strings.ReplaceAll(responseBody, "\n", " "), "|", "｜")

						// 获取当前时间戳
						timestamp := time.Now().Format(time.RFC3339)

						// 创建数据消息结构
						dataMsg := DataMessage{
							Timestamp:    timestamp,
							ModuleCode:   moduleCode,
							ActionCode:   actionCode,
							RequestBody:  requestBody,
							ResponseBody: responseBody,
							Method:       r.Method,
							Path:         r.URL.Path,
						}

						log.Debugf("Logging %s %s: %s/%s", dataMsg.Method, dataMsg.Path, dataMsg.ModuleCode, dataMsg.ActionCode)

						// 如果配置了 binding，则发送到 binding
						if bindingName != "" {
							sendDataToBinding(bindingName, daprGRPCPort, dataMsg, log)
						}

						// 同时写入日志文件（可选）
						if logFile != "" {
							logEntry := fmt.Sprintf("%s|%s|%s|%s|%s\n",
								timestamp,
								moduleCode,
								actionCode,
								requestBody,
								responseBody)

							// 异步写入日志文件以减少性能影响
							go func() {
								file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
								if err != nil {
									log.Errorf("Open log file failed: %v", err)
									return
								}
								defer file.Close()

								if _, err := file.WriteString(logEntry); err != nil {
									log.Errorf("Write log file failed: %v", err)
								}
							}()
						}
					}
				})
			}, nil
		}
	}, "body2binding")
}

// bodyResponseWriter 包装 http.ResponseWriter 以捕获响应体
type bodyResponseWriter struct {
	http.ResponseWriter
	body    *bytes.Buffer
	maxSize int64
}

func (brw *bodyResponseWriter) Write(p []byte) (int, error) {
	// 如果缓冲区大小未超过限制，则记录响应体
	if int64(brw.body.Len()+len(p)) <= brw.maxSize {
		brw.body.Write(p)
	} else if brw.body.Len() < int(brw.maxSize) {
		// 如果当前缓冲区未满但添加新数据会超过限制，则只记录部分数据
		remaining := brw.maxSize - int64(brw.body.Len())
		brw.body.Write(p[:remaining])
	}

	// 写入原始响应
	return brw.ResponseWriter.Write(p)
}

// Header 返回响应头
func (brw *bodyResponseWriter) Header() http.Header {
	return brw.ResponseWriter.Header()
}

// WriteHeader 写入响应状态码
func (brw *bodyResponseWriter) WriteHeader(statusCode int) {
	brw.ResponseWriter.WriteHeader(statusCode)
}
