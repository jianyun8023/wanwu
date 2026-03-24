# AG-UI Util

AG-UI 协议事件转换，将不同来源的事件流转换为 AG-UI 格式。

## AG-UI 协议规范

### 核心概念

AG-UI 采用**事件流驱动**的架构，通过事件流逐步构建消息对象。

#### 消息（Message）

消息是最终呈现给用户的内容单元，通过事件流逐步构建：

| 消息类型 | role 字段 | 说明 |
|---------|----------|------|
| AssistantMessage | `assistant` | AI 回复，可包含 content 和 toolCalls |
| UserMessage | `user` | 用户输入 |
| ToolMessage | `tool` | 工具执行结果 |
| ReasoningMessage | `reasoning` | AI 思考过程 |
| SystemMessage | `system` | 系统指令 |
| DeveloperMessage | `developer` | 开发者消息 |

消息结构：
```typescript
// 基础消息接口
interface Message {
  id: string;              // 消息唯一标识（messageId）
  role: string;            // 消息角色，决定消息类型
  content?: string;        // 文本内容
}

// UserMessage - 用户输入
interface UserMessage extends Message {
  role: "user";
  content: string | InputContent[];  // 纯文本或多模态内容（文本+文件）
}

// AssistantMessage - AI 回复
interface AssistantMessage extends Message {
  role: "assistant";
  content?: string;
  toolCalls?: ToolCall[];  // 可包含工具调用
}

// ToolMessage - 工具执行结果
interface ToolMessage extends Message {
  role: "tool";
  toolCallId: string;      // 关联到对应的工具调用
  content: string;         // 工具返回内容
}

// ReasoningMessage - AI 思考过程
interface ReasoningMessage extends Message {
  role: "reasoning";
  content: string;
}

// ToolCall - 工具调用（嵌入在 AssistantMessage 中）
interface ToolCall {
  id: string;              // toolCallId
  type: "function";
  function: {
    name: string;          // 工具名称
    arguments: string;     // JSON 格式的参数
  };
}
```

#### 用户上传文件

UserMessage 的 `content` 可以是多模态内容数组，支持文本和文件：

```typescript
// InputContent 联合类型
type InputContent = TextInputContent | BinaryInputContent;

// 文本内容
interface TextInputContent {
  type: "text";
  text: string;
}

// 二进制内容（文件/图片）
interface BinaryInputContent {
  type: "binary";
  mimeType: string;        // 如 "image/png", "application/pdf"
  id?: string;             // 文件引用 ID
  url?: string;            // 文件 URL
  data?: string;           // Base64 编码数据
  filename?: string;       // 文件名
}
```

**必须提供 `id`、`url` 或 `data` 中至少一个。**

示例：
```json
// 用户上传图片
{
  "id": "msg-123",
  "role": "user",
  "content": [
    { "type": "text", "text": "请分析这张图片" },
    { 
      "type": "binary", 
      "mimeType": "image/png", 
      "url": "https://example.com/image.png" 
    }
  ]
}

// 用户上传 PDF 文件（Base64）
{
  "id": "msg-456",
  "role": "user",
  "content": [
    { "type": "text", "text": "请总结这个文档" },
    { 
      "type": "binary", 
      "mimeType": "application/pdf",
      "filename": "report.pdf",
      "data": "JVBERi0xLjQK..."
    }
  ]
}
```

#### 事件（Event）

事件是流式传输的基本单元，用于增量更新消息状态：

| 事件类型 | 说明 | 关键字段 |
|---------|------|---------|
| `TEXT_MESSAGE_START` | 开始文本消息 | `messageId`, `role` |
| `TEXT_MESSAGE_CONTENT` | 追加文本内容 | `messageId`, `delta` |
| `TEXT_MESSAGE_END` | 结束文本消息 | `messageId` |
| `TOOL_CALL_START` | 开始工具调用 | `toolCallId`, `toolCallName`, `parentMessageId` |
| `TOOL_CALL_ARGS` | 追加工具参数 | `toolCallId`, `delta` |
| `TOOL_CALL_END` | 结束工具调用 | `toolCallId` |
| `TOOL_CALL_RESULT` | 工具执行结果 | `messageId`, `toolCallId`, `content` |
| `REASONING_START` | 开始推理过程 | `messageId` |
| `REASONING_MESSAGE_START` | 开始推理消息 | `messageId`, `role: "reasoning"` |
| `REASONING_MESSAGE_CONTENT` | 追加推理内容 | `messageId`, `delta` |
| `REASONING_MESSAGE_END` | 结束推理消息 | `messageId` |
| `REASONING_END` | 结束推理过程 | `messageId` |
| `RUN_STARTED` / `RUN_FINISHED` | 运行生命周期 | `threadId`, `runId` |

事件结构示例：
```typescript
// TEXT_MESSAGE_START - 开始文本消息
interface TextMessageStartEvent {
  type: "TEXT_MESSAGE_START";
  messageId: string;       // 消息唯一标识
  role: "assistant" | "user" | "system" | "developer";  // 默认 "assistant"
}

// TEXT_MESSAGE_CONTENT - 追加文本内容
interface TextMessageContentEvent {
  type: "TEXT_MESSAGE_CONTENT";
  messageId: string;       // 关联到 TEXT_MESSAGE_START
  delta: string;           // 增量文本，追加到消息末尾
}

// TOOL_CALL_START - 开始工具调用
interface ToolCallStartEvent {
  type: "TOOL_CALL_START";
  toolCallId: string;      // 工具调用唯一标识
  toolCallName: string;    // 工具名称
  parentMessageId?: string; // 关联到所属的 AssistantMessage（可选）
}

// TOOL_CALL_RESULT - 工具执行结果
interface ToolCallResultEvent {
  type: "TOOL_CALL_RESULT";
  messageId: string;       // 结果消息的唯一标识（新消息）
  toolCallId: string;      // 关联到 TOOL_CALL_START
  content: string;         // 工具返回内容
}

// REASONING_MESSAGE_START - 开始推理消息
interface ReasoningMessageStartEvent {
  type: "REASONING_MESSAGE_START";
  messageId: string;       // 推理消息唯一标识
  role: "reasoning";
}
```

#### 关键字段说明

- **messageId**: 消息唯一标识，用于关联同一消息的所有事件
- **toolCallId**: 工具调用唯一标识，用于关联工具调用的 START/ARGS/END/RESULT
- **parentMessageId**: 工具调用所属的消息ID，用于将工具调用关联到 AssistantMessage
- **role**: 消息角色，决定消息类型（assistant/user/tool/reasoning 等）
- **delta**: 增量内容，用于流式追加文本

### 事件与消息的关系

事件通过 `messageId` 关联，最终聚合为消息对象。

#### 文本消息示例

```
事件流:
TEXT_MESSAGE_START (messageId: "msg-1", role: "assistant")
TEXT_MESSAGE_CONTENT (messageId: "msg-1", delta: "Hello")
TEXT_MESSAGE_CONTENT (messageId: "msg-1", delta: " world")
TEXT_MESSAGE_END (messageId: "msg-1")

↓ 聚合为消息:

AssistantMessage {
  id: "msg-1",
  role: "assistant",
  content: "Hello world"
}
```

#### 工具调用关联示例

`parentMessageId` 用于将工具调用关联到已有的 AssistantMessage：

```
事件流:
TEXT_MESSAGE_START (messageId: "msg-1", role: "assistant")
TEXT_MESSAGE_CONTENT (messageId: "msg-1", delta: "Let me search")
TEXT_MESSAGE_END (messageId: "msg-1")

TOOL_CALL_START (toolCallId: "call-1", toolCallName: "search", parentMessageId: "msg-1")
TOOL_CALL_ARGS (toolCallId: "call-1", delta: "{\"query\": \"test\"}")
TOOL_CALL_END (toolCallId: "call-1")

TOOL_CALL_RESULT (messageId: "result-1", toolCallId: "call-1", content: "found 3 items")

↓ 聚合为消息:

AssistantMessage {
  id: "msg-1",
  role: "assistant",
  content: "Let me search",
  toolCalls: [{ id: "call-1", function: { name: "search", arguments: "{\"query\": \"test\"}" } }]
}

ToolMessage {
  id: "result-1",
  role: "tool",
  toolCallId: "call-1",
  content: "found 3 items"
}
```

**关键点：**
- `TOOL_CALL_START` 的 `parentMessageId` 匹配 `TEXT_MESSAGE_START` 的 `messageId` → 工具调用嵌入到该 AssistantMessage
- `TOOL_CALL_RESULT` 创建独立的 ToolMessage，通过 `toolCallId` 关联到工具调用
- 不提供 `parentMessageId` 时，`TOOL_CALL_START` 会创建新的 AssistantMessage

### 事件类型

| 事件类型 | 事件序列 | 说明 |
|---------|---------|------|
| Run | `RUN_STARTED` → ... → `RUN_FINISHED` | 一次完整的 AI 运行 |
| TextMessage | `TEXT_MESSAGE_START` → `TEXT_MESSAGE_CONTENT*` → `TEXT_MESSAGE_END` | 文本消息 |
| Reasoning | `REASONING_START` → `REASONING_MESSAGE_START` → `REASONING_MESSAGE_CONTENT*` → `REASONING_MESSAGE_END` → `REASONING_END` | 推理过程 |
| ToolCall (发起) | `TOOL_CALL_START` → `TOOL_CALL_ARGS*` → `TOOL_CALL_END` | Assistant 发起工具调用 |
| ToolCall (结果) | `TOOL_CALL_RESULT` | 工具执行返回结果（独立事件） |

> **说明**：`TOOL_CALL_START/ARGS/END` 和 `TOOL_CALL_RESULT` 是两个独立的事件序列。前者由 Assistant 发起调用，后者由工具执行返回结果。它们通过 `toolCallId` 关联。

### 事件穿插规则（AG-UI 完整规范）

- TEXT_MESSAGE、TOOL_CALL、REASONING_MESSAGE 是**独立的事件流**，可以任意穿插
- 通过 `messageId`/`parentMessageId` 关联到同一个消息对象：
  - `TOOL_CALL_START` 的 `parentMessageId` 匹配 `TEXT_MESSAGE_START` 的 `messageId` → 同一个 AssistantMessage
  - 不匹配则创建独立的消息
- 多个 TEXT_MESSAGE 可同时活跃（不同 messageId）
- 多个 TOOL_CALL 可同时活跃（不同 toolCallId）
- 唯一约束：REASONING 同一时刻只能有一个活跃

### 消息类型

| 消息类型 | 角色 | 来源事件 |
|---------|------|---------|
| AssistantMessage | assistant | TEXT_MESSAGE_* + TOOL_CALL_* (via parentMessageId) |
| ReasoningMessage | reasoning | REASONING_MESSAGE_* |
| ToolMessage | tool | TOOL_CALL_RESULT |

## 本实现规则（AG-UI 规范子集）

本实现采用**串行处理模式**，不穿插事件。

### 活跃状态说明

"活跃"指一个事件序列已经发送了 START 事件但尚未发送 END 事件：

| 状态 | 活跃条件 | AG-UI 完整规范 | 本实现 |
|-----|---------|---------------|-------|
| TEXT_MESSAGE 活跃 | 存在 messageId，已发送 `TEXT_MESSAGE_START` 但未发送 `TEXT_MESSAGE_END` | 多个可同时活跃（不同 messageId） | 最多 1 个活跃（单个 messageId） |
| TOOL_CALL 活跃 | 存在 toolCallId，已发送 `TOOL_CALL_START` 但未发送 `TOOL_CALL_END` | 多个可同时活跃（不同 toolCallId） | 最多 1 个活跃（串行处理） |
| REASONING 活跃 | 已发送 `REASONING_START` 但未发送 `REASONING_END` | 最多 1 个活跃 | 最多 1 个活跃 ✅ |

### 事件发送顺序

收到不同类型内容时，按以下顺序发送事件：

#### Tool 消息（Role=Tool）

收到工具执行结果时，直接发送 `TOOL_CALL_RESULT`（独立事件，不需要 START/END）：

```
REASONING_MESSAGE_END (如果 REASONING 活跃)
REASONING_END (如果 REASONING 活跃)
TEXT_MESSAGE_END (如果 TEXT_MESSAGE 活跃)
TOOL_CALL_RESULT
```

> **说明**：`TOOL_CALL_RESULT` 是工具执行返回的结果，与 `TOOL_CALL_START/ARGS/END` 是独立的事件序列。前者由 Assistant 发起，后者由工具执行返回。

#### ToolCalls（Assistant 消息中的工具调用）

收到工具调用请求时：

```
parentMsgID = 当前 messageId              ← 保存当前消息 ID
REASONING_MESSAGE_END (如果 REASONING 活跃)
REASONING_END (如果 REASONING 活跃)
TEXT_MESSAGE_END (如果 TEXT_MESSAGE 活跃)
TOOL_CALL_START (parentMessageId: parentMsgID)
TOOL_CALL_ARGS
TOOL_CALL_END                    ← 每个 ToolCall 完整处理
TOOL_CALL_START (parentMessageId: parentMsgID)
TOOL_CALL_ARGS
TOOL_CALL_END
...
```

> **说明**：
> - 所有 ToolCalls 处理完毕后，不存在活跃的 TOOL_CALL。后续收到 ReasoningContent 或 Content 时无需发送 TOOL_CALL_END。
> - **重要**：ToolCall 通过 `parentMessageId` 关联到已关闭的 AssistantMessage。即使 TEXT_MESSAGE 已 END，`parentMessageId` 仍可正确关联。

#### ReasoningContent

收到推理内容时（所有 ToolCalls 已处理完毕，无活跃的 TOOL_CALL）：

```
TEXT_MESSAGE_END (如果 TEXT_MESSAGE 活跃)
REASONING_START (如果 REASONING 未活跃)
REASONING_MESSAGE_START (如果 REASONING_MESSAGE 未活跃)
REASONING_MESSAGE_CONTENT
```

#### Content

收到文本内容时（所有 ToolCalls 已处理完毕，无活跃的 TOOL_CALL）：

```
REASONING_MESSAGE_END (如果 REASONING 活跃)
REASONING_END (如果 REASONING 活跃)
TEXT_MESSAGE_START (如果 TEXT_MESSAGE 未活跃)
TEXT_MESSAGE_CONTENT
```

### 与完整规范的差异

| 特性 | AG-UI 完整规范 | 本实现 |
|-----|---------------|-------|
| TEXT_MESSAGE 穿插 | 多个可同时活跃 | 单个活跃，串行处理 |
| TOOL_CALL 穿插 | 多个可同时活跃 | 单个活跃，串行处理 |
| TOOL_CALL 关联 | 可关联到活跃/关闭的消息 | 关联到已关闭的消息（parentMessageId） |
| Reasoning 穿插 | 只能一个活跃 | ✅ 符合规范 |

### ID 生成

使用 AG-UI SDK 提供的 ID 生成器：
- `GenerateMessageID()` → `msg-{uuid}`
- `GenerateToolCallID()` → `tool-{uuid}`
- `GenerateStepID()` → `step-{uuid}`

## 转换器

| 转换器 | 输入 | 使用场景 |
|--------|------|---------|
| OpencodeTranslator | opencode JSON 字符串 | wga-sandbox 输出转换 |
| EinoTranslator | eino AgentEvent | wga.Run() 输出转换（单智能体） |
| EinoMultiAgentTranslator | eino AgentEvent | wga.Run() 输出转换（多智能体） |

## 使用

```go
import ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"

// OpencodeTranslator - wga-sandbox 输出转换
runSession, outputCh, _ := wga_sandbox.Run(ctx, opts...)
tr := ag_ui_util.NewOpencodeTranslator(runSession.ThreadID, runSession.RunID)
eventCh := tr.TranslateStream(ctx, outputCh)
jsonCh := ag_ui_util.EventsToJSONChannel(ctx, eventCh)

// EinoTranslator - wga 单智能体输出转换
runSession, iter, _ := wga.Run(ctx, agentID, opts...)
tr := ag_ui_util.NewEinoTranslator(runSession.ThreadID, runSession.RunID)
eventCh := tr.TranslateStream(ctx, iter)
jsonCh := ag_ui_util.EventsToJSONChannel(ctx, eventCh)

// EinoMultiAgentTranslator - wga 多智能体输出转换
// 使用 ActivitySnapshot 标识当前运行的智能体
runSession, iter, _ := wga.Run(ctx, agentID, opts...)
tr := ag_ui_util.NewEinoMultiAgentTranslator(runSession.ThreadID, runSession.RunID)
eventCh := tr.TranslateStream(ctx, iter)
jsonCh := ag_ui_util.EventsToJSONChannel(ctx, eventCh)
```

## 多智能体模式

多智能体模式通过 AG-UI 协议的 `ACTIVITY_SNAPSHOT` 事件标识当前运行的智能体。

### ActivitySnapshot 结构

```json
{
  "type": "ACTIVITY_SNAPSHOT",
  "messageId": "step-xxx",
  "activityType": "sub_agent",
  "content": {
    "agentName": "Plan Agent",
    "instanceNum": 1,
    "status": "started"
  }
}
```

### 字段说明

| 字段 | 说明 |
|-----|------|
| `activityType` | 固定为 `"sub_agent"` |
| `content.agentName` | 智能体名称 |
| `content.instanceNum` | 智能体实例编号（同一智能体可能多次运行） |
| `content.status` | 状态：`"started"` 或 `"finished"` |

### 智能体切换流程

1. 检测到 `AgentEvent.AgentName` 变化
2. 发送前一个智能体的消息结束事件（TextMessageEnd 等）
3. 发送 `ACTIVITY_SNAPSHOT` 事件，`status: "finished"`
4. 为新智能体创建独立的消息 ID
5. 发送新智能体的 `ACTIVITY_SNAPSHOT` 事件，`status: "started"`
6. 发送新智能体的消息内容

## API

### 转换器

| 函数 | 说明 |
|------|------|
| `NewOpencodeTranslator(threadID, runID)` | 创建 opencode 转换器 |
| `NewEinoTranslator(threadID, runID)` | 创建 eino 单智能体转换器 |
| `NewEinoMultiAgentTranslator(threadID, runID)` | 创建 eino 多智能体转换器 |

### OpencodeTranslator

| 方法 | 说明 |
|------|------|
| `TranslateStream(ctx, <-chan string)` | 转换 opencode JSON 字符串流 |

### EinoTranslator

| 方法 | 说明 |
|------|------|
| `TranslateStream(ctx, *adk.AsyncIterator[*adk.AgentEvent])` | 转换 eino AgentEvent 迭代器 |

### 辅助函数

| 函数 | 说明 |
|------|------|
| `EventsToJSONChannel(ctx, events)` | 事件流 → JSON 字符串流 |
