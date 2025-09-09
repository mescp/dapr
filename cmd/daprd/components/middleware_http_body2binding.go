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
	"regexp"
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
	Timestamp    string      `json:"timestamp"`
	FunctionCode string      `json:"functionCode"`
	ActionCode   string      `json:"actionCode"`
	RequestBody  interface{} `json:"requestBody,omitempty"`
	ResponseBody interface{} `json:"responseBody,omitempty"`
	Method       string      `json:"method"`
	Path         string      `json:"path"`
	Headers      interface{} `json:"headers,omitempty"`
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
		conn, err := grpc.NewClient(
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

			// 获取功能码和操作码的 header 字段名，默认值
			functionHeader := metadata.Properties["functionHeader"]
			if functionHeader == "" {
				functionHeader = "X-Function-Code"
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

			// 获取需要包含在日志中的 header keys，支持逗号分隔的列表
			includeHeaders := metadata.Properties["includeHeaders"]
			var headerKeys []string
			if includeHeaders != "" {
				for _, header := range strings.Split(includeHeaders, ",") {
					header = strings.TrimSpace(header)
					if header != "" {
						headerKeys = append(headerKeys, header)
					}
				}
			}

			// 获取路径过滤配置，支持正则表达式
			includePathsStr := metadata.Properties["includePaths"]
			excludePathsStr := metadata.Properties["excludePaths"]

			// 编译包含路径的正则表达式
			var includePathRegexes []*regexp.Regexp
			if includePathsStr != "" {
				for _, pathPattern := range strings.Split(includePathsStr, ",") {
					pathPattern = strings.TrimSpace(pathPattern)
					if pathPattern != "" {
						if regex, err := regexp.Compile(pathPattern); err != nil {
							log.Warnf("Invalid includePaths regex pattern '%s': %v", pathPattern, err)
						} else {
							includePathRegexes = append(includePathRegexes, regex)
						}
					}
				}
			}

			// 编译排除路径的正则表达式
			var excludePathRegexes []*regexp.Regexp
			if excludePathsStr != "" {
				for _, pathPattern := range strings.Split(excludePathsStr, ",") {
					pathPattern = strings.TrimSpace(pathPattern)
					if pathPattern != "" {
						if regex, err := regexp.Compile(pathPattern); err != nil {
							log.Warnf("Invalid excludePaths regex pattern '%s': %v", pathPattern, err)
						} else {
							excludePathRegexes = append(excludePathRegexes, regex)
						}
					}
				}
			}

			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var requestBodyObj interface{}

					// 记录请求体
					if logRequest && r.Body != nil {
						requestBody, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
						if err != nil {
							log.Errorf("Read request body failed: %v", err)
							requestBodyObj = map[string]string{"error": fmt.Sprintf("Read error: %v", err)}
						} else {
							// 尝试解析为 JSON
							var jsonObj interface{}
							if err := json.Unmarshal(requestBody, &jsonObj); err != nil {
								// 如果不是有效的 JSON，则作为字符串存储
								requestBodyObj = string(requestBody)
							} else {
								requestBodyObj = jsonObj
							}
							// 重置请求体供后续处理使用
							r.Body = io.NopCloser(bytes.NewReader(requestBody))
						}
					}

					var responseBodyObj interface{}
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
						responseBodyStr := bodyRecorder.body.String()
						// 尝试解析为 JSON
						var jsonObj interface{}
						if err := json.Unmarshal([]byte(responseBodyStr), &jsonObj); err != nil {
							// 如果不是有效的 JSON，则作为字符串存储
							responseBodyObj = responseBodyStr
						} else {
							responseBodyObj = jsonObj
						}
					}

					// 记录到日志文件
					if logRequest || logResponse {
						// 检查当前请求方法是否在允许记录的方法列表中
						if !allowedMethods[r.Method] {
							return
						}

						// 检查路径是否应该被记录
						requestPath := r.URL.RequestURI()
						shouldLog := true

						// 如果配置了包含路径，检查当前路径是否匹配任何包含模式
						if len(includePathRegexes) > 0 {
							shouldLog = false
							for _, regex := range includePathRegexes {
								if regex.MatchString(requestPath) {
									shouldLog = true
									break
								}
							}
						}

						// 如果路径通过包含检查，再检查是否在排除列表中
						if shouldLog && len(excludePathRegexes) > 0 {
							for _, regex := range excludePathRegexes {
								if regex.MatchString(requestPath) {
									shouldLog = false
									break
								}
							}
						}

						// 如果路径被过滤掉，则不记录日志
						if !shouldLog {
							log.Debugf("Skipping log due to path filtering: %s", requestPath)
							return
						}

						// 从 header 中获取功能码和动作码
						functionCode := r.Header.Get(functionHeader)
						actionCode := r.Header.Get(actionHeader)

						// 如果功能码或动作码为空，则不记录日志
						if functionCode == "" || actionCode == "" {
							log.Debugf("Skipping log due to missing headers: functionCode=%s, actionCode=%s", functionCode, actionCode)
							return
						}

						// 获取当前时间戳
						timestamp := time.Now().Format(time.RFC3339)

						// 收集指定的 headers
						var headers interface{}
						if len(headerKeys) > 0 {
							headersMap := make(map[string]interface{})
							for _, headerKey := range headerKeys {
								if headerValues := r.Header.Values(headerKey); len(headerValues) > 0 {
									if len(headerValues) == 1 {
										// 单个值：尝试解析为 JSON，失败则作为字符串
										headerValue := headerValues[0]
										var jsonObj interface{}
										if err := json.Unmarshal([]byte(headerValue), &jsonObj); err != nil {
											// 不是有效的 JSON，作为字符串存储
											headersMap[headerKey] = headerValue
										} else {
											// 是有效的 JSON，存储解析后的对象
											headersMap[headerKey] = jsonObj
										}
									} else {
										// 多个值：创建数组，每个值尝试解析为 JSON
										var valueArray []interface{}
										for _, value := range headerValues {
											var jsonObj interface{}
											if err := json.Unmarshal([]byte(value), &jsonObj); err != nil {
												// 不是有效的 JSON，作为字符串存储
												valueArray = append(valueArray, value)
											} else {
												// 是有效的 JSON，存储解析后的对象
												valueArray = append(valueArray, jsonObj)
											}
										}
										headersMap[headerKey] = valueArray
									}
								}
							}
							if len(headersMap) > 0 {
								headers = headersMap
							}
						}

						// 创建数据消息结构
						dataMsg := DataMessage{
							Timestamp:    timestamp,
							FunctionCode: functionCode,
							ActionCode:   actionCode,
							Method:       r.Method,
							Path:         requestPath,
							Headers:      headers,
						}

						// 根据配置设置请求体和响应体
						if logRequest {
							dataMsg.RequestBody = requestBodyObj
						}
						if logResponse {
							dataMsg.ResponseBody = responseBodyObj
						}

						log.Debugf("Logging %s %s: %s/%s", dataMsg.Method, dataMsg.Path, dataMsg.FunctionCode, dataMsg.ActionCode)

						// 如果配置了 binding，则发送到 binding
						if bindingName != "" {
							sendDataToBinding(bindingName, daprGRPCPort, dataMsg, log)
						}

						// 同时写入日志文件（可选）
						if logFile != "" {
							// 将 JSON 对象序列化为字符串用于日志文件
							var requestBodyStr, responseBodyStr string

							if dataMsg.RequestBody != nil {
								if reqBodyBytes, err := json.Marshal(dataMsg.RequestBody); err == nil {
									requestBodyStr = string(reqBodyBytes)
								} else {
									requestBodyStr = fmt.Sprintf("%v", dataMsg.RequestBody)
								}
							}

							if dataMsg.ResponseBody != nil {
								if respBodyBytes, err := json.Marshal(dataMsg.ResponseBody); err == nil {
									responseBodyStr = string(respBodyBytes)
								} else {
									responseBodyStr = fmt.Sprintf("%v", dataMsg.ResponseBody)
								}
							}

							// 将换行符和分隔符替换为空格，确保一行记录
							requestBodyStr = strings.ReplaceAll(strings.ReplaceAll(requestBodyStr, "\n", " "), "|", "｜")
							responseBodyStr = strings.ReplaceAll(strings.ReplaceAll(responseBodyStr, "\n", " "), "|", "｜")

							logEntry := fmt.Sprintf("%s|%s|%s|%s|%s\n",
								timestamp,
								functionCode,
								actionCode,
								requestBodyStr,
								responseBodyStr)

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
