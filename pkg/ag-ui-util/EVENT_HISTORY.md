# AG-UI 事件历史管理

本包提供了 AG-UI 协议事件的历史管理和数据清洗功能。

## 概述

AG-UI 协议定义了一套标准的事件流格式，用于智能体与前端之间的通信。本包提供了三个核心功能：

1. **消息类型定义** (`message_types.go`) - 定义了历史消息的结构体和常量
2. **事件流处理** (`stream_processor.go`) - 提供事件流的数据清洗、聚合和格式化功能

## 核心功能

### 1. 消息类型定义

#### 常量定义

```go
const (
    RoleAssistant = "assistant"
    RoleUser      = "user"
    RoleTool      = "tool"
    RoleReasoning = "reasoning"
    RoleSystem    = "system"

    ToolCallTypeFunction = "function"
    ActivityTypeSubAgent = "sub_agent"

    ActivityStatusStarted  = "started"
    ActivityStatusFinished = "finished"
)
```

#### TextMessage - 文本消息

```go
type TextMessage struct {
    MessageID string `json:"messageId"`
    Role      string `json:"role"`
    Content   string `json:"content"`
}
```

用于存储文本消息的完整内容，由 `TEXT_MESSAGE_START`、`TEXT_MESSAGE_CONTENT`、`TEXT_MESSAGE_END` 事件序列聚合而成。

#### ReasoningMessage - 推理消息

```go
type ReasoningMessage struct {
    MessageID string `json:"messageId"`
    Role      string `json:"role"`
    Content   string `json:"content"`
}
```

用于存储推理过程的完整内容，由 `REASONING_MESSAGE_START`、`REASONING_MESSAGE_CONTENT`、`REASONING_MESSAGE_END` 事件序列聚合而成。

#### ToolCall - 工具调用

```go
type ToolCall struct {
    ID       string           `json:"id"`
    Type     string           `json:"type"`
    Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}
```

用于存储工具调用的定义，符合 AG-UI 协议规范。

#### ToolMessage - 工具执行结果

```go
type ToolMessage struct {
    MessageID  string `json:"messageId"`
    Role       string `json:"role"`
    ToolCallID string `json:"toolCallId"`
    Content    string `json:"content"`
}
```

用于存储工具执行的结果，通过 `toolCallId` 关联到对应的工具调用。

#### Activity - 智能体活动

```go
type Activity struct {
    ActivityID   string                 `json:"activityId"`
    ActivityType string                 `json:"activityType"`
    AgentName    string                 `json:"agentName"`
    InstanceNum  int                    `json:"instanceNum"`
    Status       string                 `json:"status"`
    Content      map[string]interface{} `json:"content,omitempty"`
}
```

用于存储智能体活动记录，如子智能体的启动和结束。

### 2. 事件流处理

#### ProcessorConfig - 处理器配置

```go
type ProcessorConfig struct {
    ToolNameMapper     map[string]string          // 工具名映射
    ExcludedAgentNames []string                   // 要排除的智能体名称
    ResultFormatters   map[string]func(string) string // 结果格式化器
}
```

#### StreamProcessor - 事件流处理器

```go
processor := NewStreamProcessor(config)
cleanedEventCh, historyCh := processor.Process(ctx, eventStream)
```

**核心特性：**
- 工具名映射：将内部工具名映射为用户友好的名称
- Activity 过滤：删除指定智能体的活动事件
- 结果格式化：自定义工具结果的格式化
- 消息聚合：将同一 ID 的多个事件聚合成完整消息
- 时间顺序：按消息完成时间顺序存储
- 并发支持：支持多个工具调用并发执行

### 3. 工具结果格式化

#### WebSearchResult - 网页搜索结果

```go
type WebSearchResult struct {
    Query    string    `json:"query"`
    WebCount int       `json:"webCount"`
    WebPages []WebPage `json:"webPages"`
}

type WebPage struct {
    Title    string `json:"title"`
    SiteName string `json:"siteName"`
    Icon     string `json:"icon"`
    Summary  string `json:"summary"`
    URL      string `json:"url"`
}
```

#### 格式化函数

包内提供的通用格式化函数：

```go
// 格式化 JSON 字符串，美化输出
func FormatJSONResult(result string) string

// 创建截断函数，限制结果长度
func TruncateResult(maxLen int) func(string) string

// 创建脱敏函数，隐藏敏感字段
func MaskSensitiveFields(sensitiveFields []string) func(string) string

// 创建前缀移除函数
func RemovePrefixes(prefixes []string) func(string) string
```

业务相关的格式化函数（定义在 `internal/bff-service/service/wga_tool_formatter.go`）：

```go
// 格式化 bocha 搜索结果
func WgaFormatBochaWebSearchResult(result string) string

// 格式化 tavily 搜索结果
func WgaFormatTavilySearchResult(result string) string
```

> **注意**：业务相关的格式化函数需要在应用层定义，不在此包中。以上示例仅供参考。

## 使用示例

### 基础用法

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    
    ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"
)

func main() {
    // 1. 创建处理器配置
    config := &ag_ui_util.ProcessorConfig{
        ToolNameMapper: map[string]string{
            "transfer_to_agent": "正在交给专业智能体",
            "internal_search":   "搜索",
        },
        ExcludedAgentNames: []string{
            "default",
            "Supervisor Agent",
        },
        ResultFormatters: map[string]func(string) string{
            "bochaWebSearch":      formatBochaWebSearchResult,
            "tavily_basic_search": formatTavilySearchResult,
        },
    }
    
    // 2. 创建处理器
    processor := ag_ui_util.NewStreamProcessor(config)
    
    // 3. 处理事件流
    ctx := context.Background()
    cleanedEventCh, historyCh := processor.Process(ctx, rawEventStream)
    
    // 4. 实时处理清洗后的事件
    go func() {
        for event := range cleanedEventCh {
            // 发送给前端
            sendToClient(event)
        }
    }()
    
    // 5. 收集历史记录
    var historyMessages []interface{}
    for msg := range historyCh {
        historyMessages = append(historyMessages, msg)
    }
    
    // 6. 保存为 JSON 文件
    if len(historyMessages) > 0 {
        if jsonBytes, err := json.MarshalIndent(historyMessages, "", "  "); err == nil {
            filename := fmt.Sprintf("history_%s.json", runID)
            os.WriteFile(filename, jsonBytes, 0644)
        }
    }
}
```

### 与 EinoTranslator 集成

```go
func translateWithHistory(ctx context.Context, iter *adk.AsyncIterator[*adk.AgentEvent]) {
    // 创建翻译器
    translator := ag_ui_util.NewEinoTranslator(threadID, runID)
    eventCh := translator.TranslateStream(ctx, iter)
    
    // 创建处理器
    config := &ag_ui_util.ProcessorConfig{
        ToolNameMapper: map[string]string{
            "transfer_to_agent": "正在交给专业智能体",
        },
        ExcludedAgentNames: []string{"default"},
        ResultFormatters: map[string]func(string) string{
            "bochaWebSearch": formatBochaWebSearchResult,
        },
    }
    processor := ag_ui_util.NewStreamProcessor(config)
    
    // 处理并聚合事件
    cleanedEventCh, historyCh := processor.Process(ctx, eventCh)
    
    // 处理清洗后的事件流
    for event := range cleanedEventCh {
        // 实时发送给前端
        sendToClient(event)
    }
    
    // 收集历史记录
    var historyMessages []interface{}
    for msg := range historyCh {
        historyMessages = append(historyMessages, msg)
    }
    saveToDatabase(historyMessages)
}
```

### 自定义结果格式化

```go
config := &ag_ui_util.ProcessorConfig{
    ResultFormatters: map[string]func(string) string{
        // 格式化 JSON 结果
        "search": ag_ui_util.FormatJSONResult,
        
        // 截断长结果
        "code_exec": ag_ui_util.TruncateResult(500),
        
        // 脱敏敏感字段
        "user_query": ag_ui_util.MaskSensitiveFields([]string{"password", "token"}),
        
        // 移除前缀
        "error_handler": ag_ui_util.RemovePrefixes([]string{"[ERROR]", "[WARN]"}),
    },
}
```

## API 文档

### StreamProcessor

#### NewStreamProcessor

```go
func NewStreamProcessor(config *ProcessorConfig) *StreamProcessor
```

创建事件流处理器实例。

**参数：**
- `config`: 处理器配置，可以为 nil（使用默认配置）

**返回：**
- `*StreamProcessor`: 事件流处理器实例

#### Process

```go
func (p *StreamProcessor) Process(ctx context.Context, in <-chan aguievents.Event) (<-chan aguievents.Event, <-chan interface{})
```

处理并聚合事件流。

**参数：**
- `ctx`: 上下文
- `in`: 原始事件流

**返回：**
- `<-chan aguievents.Event`: 清洗后的事件流（实时）
- `<-chan interface{}`: 历史消息流（逐个发送）

### 辅助函数

#### FormatJSONResult

```go
func FormatJSONResult(result string) string
```

格式化 JSON 字符串，美化输出。

#### TruncateResult

```go
func TruncateResult(maxLen int) func(string) string
```

创建截断函数，限制结果长度。

**参数：**
- `maxLen`: 最大长度

**返回：**
- `func(string) string`: 截断函数

#### MaskSensitiveFields

```go
func MaskSensitiveFields(sensitiveFields []string) func(string) string
```

创建脱敏函数，隐藏敏感字段。

**参数：**
- `sensitiveFields`: 敏感字段名称列表

**返回：**
- `func(string) string`: 脱敏函数

#### RemovePrefixes

```go
func RemovePrefixes(prefixes []string) func(string) string
```

创建前缀移除函数。

**参数：**
- `prefixes`: 要移除的前缀列表

**返回：**
- `func(string) string`: 前缀移除函数

## 数据流处理

### 文本消息聚合流程

```
原始事件流
    ↓
TEXT_MESSAGE_START (创建 TextMessage)
    ↓
TEXT_MESSAGE_CONTENT (追加内容)
    ↓
TEXT_MESSAGE_END (返回完整 TextMessage)
    ↓
历史消息流
```

### 工具调用聚合流程

```
TOOL_CALL_START (创建 ToolCall)
    ↓
TOOL_CALL_ARGS (设置参数)
    ↓
TOOL_CALL_END (返回完整 ToolCall)
    ↓
历史消息流
    ↓
TOOL_CALL_RESULT (创建 ToolMessage)
    ↓
历史消息流
```

### 并发工具调用处理

```
ToolCall 1 Start → ToolCall 1 Args → ToolCall 1 End
ToolCall 2 Start → ToolCall 2 Args → ToolCall 2 End
    ↓                    ↓
ToolCall 1 Result    ToolCall 2 Result
    ↓                    ↓
ToolMessage 1        ToolMessage 2
```

使用 `toolCallMap` 存储所有工具调用，支持并发执行。

## 输出示例

### JSON 输出格式

```json
[
  {
    "messageId": "msg-001",
    "role": "assistant",
    "content": "我来帮您分析这个问题..."
  },
  {
    "messageId": "reasoning-001",
    "role": "reasoning",
    "content": "首先需要理解用户意图..."
  },
  {
    "id": "tc-001",
    "type": "function",
    "function": {
      "name": "bochaWebSearch",
      "arguments": "{\"query\": \"杭州天气\"}"
    }
  },
  {
    "messageId": "msg-002",
    "role": "tool",
    "toolCallId": "tc-001",
    "content": "{\"query\": \"杭州天气\", \"webCount\": 5, \"webPages\": [...]}"
  },
  {
    "activityId": "act-001",
    "activityType": "sub_agent",
    "agentName": "Plan Agent",
    "status": "started"
  }
]
```

### 工具结果格式化示例

#### Bocha 搜索结果

```json
{
  "query": "杭州天气",
  "webCount": 3,
  "webPages": [
    {
      "title": "杭州天气预报",
      "siteName": "天气网",
      "icon": "https://example.com/icon.png",
      "summary": "杭州今天天气晴朗...",
      "url": "https://example.com/weather"
    }
  ]
}
```

#### Tavily 搜索结果

```json
{
  "query": "人工智能",
  "webCount": 5,
  "webPages": [
    {
      "title": "人工智能简介",
      "siteName": "Tavily",
      "icon": "https://imgbed-1303886329.cos.ap-nanjing.myqcloud.com/20260327144847.png",
      "summary": "人工智能是研究...",
      "url": "https://example.com/ai"
    }
  ]
}
```

## 注意事项

### 1. 线程安全
- `StreamProcessor` 的所有方法都是线程安全的
- 可以在多个 goroutine 中并发使用

### 2. 内存管理
- 历史消息逐个发送到 channel，避免内存堆积
- 使用 `toolCallMap` 存储工具调用，支持并发执行

### 3. 事件顺序
- 历史消息按完成时间顺序发送
- 同一 ID 的事件会被聚合成一个完整消息

### 4. 错误处理
- 处理器不会修改原始事件流
- 过滤的事件不会出现在历史记录中
- 格式化失败时返回原始结果

### 5. 并发工具调用
- 支持多个工具调用并发执行
- 通过 `toolCallID` 精确匹配工具调用和结果

## 性能优化

### 1. 使用缓冲 Channel

```go
cleanedOut := make(chan aguievents.Event, 1024)
historyOut := make(chan interface{}, 1024)
```

### 2. 避免频繁加锁
- 使用 `sync.RWMutex` 进行读写分离
- 聚合逻辑在单个 goroutine 中执行

### 3. 使用 Map 存储工具调用
```go
toolCallMap map[string]*ToolCall  // 支持并发工具调用
```

## 参考资料

- [AG-UI Protocol Specification](https://github.com/ag-ui-protocol/ag-ui)
- [AG-UI Events Definition](https://github.com/ag-ui-protocol/ag-ui/tree/main/sdks/community/go/pkg/core/events)
- [Eino Translator Implementation](./translator_eino.go)
