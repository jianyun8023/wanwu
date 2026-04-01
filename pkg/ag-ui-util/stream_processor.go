package ag_ui_util

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/UnicomAI/wanwu/pkg/log"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
)

// ============================================================================
// 类型定义
// ============================================================================

type ProcessorConfig struct {
	ToolNameMapper     map[string]string
	ExcludedAgentNames []string
	ResultFormatters   map[string]func(string) string
}

type RunStartedEventWithInput struct {
	*aguievents.BaseEvent
	ThreadIDValue string      `json:"threadId"`
	RunIDValue    string      `json:"runId"`
	Input         interface{} `json:"input,omitempty"`
}

func NewRunStartedEventWithInput(threadID, runID string, input interface{}) *RunStartedEventWithInput {
	return &RunStartedEventWithInput{
		BaseEvent:     aguievents.NewBaseEvent(aguievents.EventTypeRunStarted),
		ThreadIDValue: threadID,
		RunIDValue:    runID,
		Input:         input,
	}
}

func (e *RunStartedEventWithInput) ThreadID() string { return e.ThreadIDValue }
func (e *RunStartedEventWithInput) RunID() string    { return e.RunIDValue }
func (e *RunStartedEventWithInput) Validate() error  { return nil }
func (e *RunStartedEventWithInput) ToJSON() ([]byte, error) {
	type alias RunStartedEventWithInput
	return json.Marshal((*alias)(e))
}

type StreamProcessor struct {
	mu     sync.RWMutex
	config *ProcessorConfig

	currentTextMsgID           string
	currentTextMsgContent      string
	currentReasoningMsgID      string
	currentReasoningMsgContent string
	currentToolMessage         *ToolMessage
	toolCallMap                map[string]*ToolCall

	currentTextMsgBackup      *TextMessage
	currentReasoningMsgBackup *ReasoningMessage
}

// ============================================================================
// 构造函数
// ============================================================================

func NewStreamProcessor(config *ProcessorConfig) *StreamProcessor {
	if config == nil {
		config = &ProcessorConfig{}
	}
	return &StreamProcessor{
		config:      config,
		toolCallMap: make(map[string]*ToolCall),
	}
}

// ============================================================================
// 公开方法
// ============================================================================

func (p *StreamProcessor) ProcessBackup(ctx context.Context, in <-chan aguievents.Event) (<-chan aguievents.Event, <-chan interface{}) {
	cleanedOut := make(chan aguievents.Event, 1024)
	historyOut := make(chan interface{}, 1024)

	go func() {
		defer close(cleanedOut)
		defer close(historyOut)

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-in:
				if !ok {
					return
				}

				cleanedEvent := p.cleanEvent(event)

				if cleanedEvent == nil {
					continue
				}

				if msg := p.aggregateEventBackup(cleanedEvent); msg != nil {
					select {
					case historyOut <- msg:
					case <-ctx.Done():
						return
					}
				}

				select {
				case cleanedOut <- cleanedEvent:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return cleanedOut, historyOut
}

func (p *StreamProcessor) Process(ctx context.Context, in <-chan aguievents.Event, input interface{}) (<-chan aguievents.Event, <-chan aguievents.Event) {
	cleanedOut := make(chan aguievents.Event, 1024)
	historyOut := make(chan aguievents.Event, 1024)

	go func() {
		defer close(cleanedOut)
		defer close(historyOut)

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-in:
				if !ok {
					return
				}

				cleanedEvent := p.cleanEvent(event)

				if cleanedEvent == nil {
					continue
				}

				if cleanedEvent.Type() == aguievents.EventTypeRunStarted {
					if runStarted, ok := cleanedEvent.(*aguievents.RunStartedEvent); ok {
						cleanedEvent = NewRunStartedEventWithInput(runStarted.ThreadID(), runStarted.RunID(), input)
					}
				}

				for _, aggregatedEvent := range p.aggregateEventForHistory(cleanedEvent) {
					if aggregatedEvent != nil {
						select {
						case historyOut <- aggregatedEvent:
						case <-ctx.Done():
							return
						}
					}
				}

				select {
				case cleanedOut <- cleanedEvent:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return cleanedOut, historyOut
}

// ============================================================================
// 私有方法 - 事件清洗
// ============================================================================

func (p *StreamProcessor) cleanEvent(event aguievents.Event) aguievents.Event {
	switch e := event.(type) {
	case *aguievents.ToolCallStartEvent:
		return p.cleanToolCallStart(e)

	case *aguievents.ToolCallResultEvent:
		return p.cleanToolCallResult(e)

	case *aguievents.ActivitySnapshotEvent:
		if p.shouldExcludeActivity(e) {
			return nil
		}
		return e

	default:
		return event
	}
}

func (p *StreamProcessor) cleanToolCallStart(event *aguievents.ToolCallStartEvent) *aguievents.ToolCallStartEvent {
	if p.config.ToolNameMapper == nil {
		return event
	}

	originalName := event.ToolCallName
	if newName, ok := p.config.ToolNameMapper[originalName]; ok {
		var parentMsgID string
		if event.ParentMessageID != nil {
			parentMsgID = *event.ParentMessageID
		}
		return aguievents.NewToolCallStartEvent(
			event.ToolCallID,
			newName,
			aguievents.WithParentMessageID(parentMsgID),
		)
	}

	return event
}

func (p *StreamProcessor) cleanToolCallResult(event *aguievents.ToolCallResultEvent) *aguievents.ToolCallResultEvent {
	if p.config.ResultFormatters == nil {
		return event
	}

	toolCallID := event.ToolCallID
	toolName := p.findToolNameByID(toolCallID)
	if toolName == "" {
		return event
	}

	formatter, ok := p.config.ResultFormatters[toolName]
	if !ok {
		return event
	}

	formattedResult := formatter(event.Content)
	return aguievents.NewToolCallResultEvent(event.MessageID, event.ToolCallID, formattedResult)
}

func (p *StreamProcessor) shouldExcludeActivity(event *aguievents.ActivitySnapshotEvent) bool {
	if len(p.config.ExcludedAgentNames) == 0 {
		return false
	}

	content := parseActivityContent(event.Content)
	if content == nil {
		return false
	}

	agentName, ok := content["agentName"].(string)
	if !ok {
		return false
	}

	for _, excluded := range p.config.ExcludedAgentNames {
		if agentName == excluded {
			return true
		}
	}

	return false
}

func (p *StreamProcessor) findToolNameByID(toolCallID string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if tc, ok := p.toolCallMap[toolCallID]; ok {
		return tc.Function.Name
	}
	return ""
}

// ============================================================================
// 私有方法 - 事件聚合
// ============================================================================

func (p *StreamProcessor) aggregateEventBackup(event aguievents.Event) interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch e := event.(type) {
	case *aguievents.TextMessageStartEvent:
		role := RoleAssistant
		if e.Role != nil {
			role = *e.Role
		}
		p.currentTextMsgBackup = &TextMessage{
			MessageID: e.MessageID,
			Role:      role,
		}
		return nil

	case *aguievents.TextMessageContentEvent:
		if p.currentTextMsgBackup != nil {
			p.currentTextMsgBackup.Content += e.Delta
		}
		return nil

	case *aguievents.TextMessageEndEvent:
		if p.currentTextMsgBackup != nil {
			msg := p.currentTextMsgBackup
			p.currentTextMsgBackup = nil
			return msg
		}
		return nil

	case *aguievents.ReasoningMessageStartEvent:
		p.currentReasoningMsgBackup = &ReasoningMessage{
			MessageID: e.MessageID,
			Role:      e.Role,
		}
		return nil

	case *aguievents.ReasoningMessageContentEvent:
		if p.currentReasoningMsgBackup != nil {
			p.currentReasoningMsgBackup.Content += e.Delta
		}
		return nil

	case *aguievents.ReasoningMessageEndEvent:
		if p.currentReasoningMsgBackup != nil {
			msg := p.currentReasoningMsgBackup
			p.currentReasoningMsgBackup = nil
			return msg
		}
		return nil

	case *aguievents.ToolCallStartEvent:
		tc := &ToolCall{
			ID:   e.ToolCallID,
			Type: ToolCallTypeFunction,
			Function: ToolCallFunction{
				Name: e.ToolCallName,
			},
		}
		p.toolCallMap[e.ToolCallID] = tc
		return nil

	case *aguievents.ToolCallArgsEvent:
		if tc, ok := p.toolCallMap[e.ToolCallID]; ok {
			tc.Function.Arguments += e.Delta
		}
		return nil

	case *aguievents.ToolCallEndEvent:
		if tc, ok := p.toolCallMap[e.ToolCallID]; ok {
			return tc
		}
		return nil

	case *aguievents.ToolCallResultEvent:
		p.currentToolMessage = &ToolMessage{
			MessageID:  e.MessageID,
			Role:       RoleTool,
			ToolCallID: e.ToolCallID,
			Content:    e.Content,
		}
		return p.currentToolMessage

	case *aguievents.ActivitySnapshotEvent:
		activity := &Activity{
			ActivityID:   e.MessageID,
			ActivityType: e.ActivityType,
		}
		content := parseActivityContent(e.Content)
		if content != nil {
			activity.Content = content
			if agentName, ok := content["agentName"].(string); ok {
				activity.AgentName = agentName
			}
			if status, ok := content["status"].(string); ok {
				activity.Status = status
			}
			if instanceNum, ok := content["instanceNum"].(int); ok {
				activity.InstanceNum = instanceNum
			}
		}
		return activity

	default:
		return nil
	}
}

func (p *StreamProcessor) aggregateEventForHistory(event aguievents.Event) []aguievents.Event {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch e := event.(type) {

	case *aguievents.TextMessageStartEvent:
		p.currentTextMsgID = e.MessageID
		p.currentTextMsgContent = ""
		return []aguievents.Event{event}

	case *aguievents.TextMessageContentEvent:
		if p.currentTextMsgID != "" {
			p.currentTextMsgContent += e.Delta
		}
		return nil

	case *aguievents.TextMessageEndEvent:
		if p.currentTextMsgID != "" {
			evt := aguievents.NewTextMessageContentEvent(p.currentTextMsgID, p.currentTextMsgContent)
			p.currentTextMsgID = ""
			p.currentTextMsgContent = ""
			return []aguievents.Event{evt, event}
		}
		return []aguievents.Event{event}

	case *aguievents.ReasoningMessageStartEvent:
		p.currentReasoningMsgID = e.MessageID
		p.currentReasoningMsgContent = ""
		return []aguievents.Event{event}

	case *aguievents.ReasoningMessageContentEvent:
		if p.currentReasoningMsgID != "" {
			p.currentReasoningMsgContent += e.Delta
		}
		return nil

	case *aguievents.ReasoningMessageEndEvent:
		if p.currentReasoningMsgID != "" {
			evt := aguievents.NewReasoningMessageContentEvent(p.currentReasoningMsgID, p.currentReasoningMsgContent)
			p.currentReasoningMsgID = ""
			p.currentReasoningMsgContent = ""
			return []aguievents.Event{evt, event}
		}
		return []aguievents.Event{event}

	case *aguievents.ToolCallStartEvent:
		tc := &ToolCall{
			ID:   e.ToolCallID,
			Type: ToolCallTypeFunction,
			Function: ToolCallFunction{
				Name: e.ToolCallName,
			},
		}
		p.toolCallMap[e.ToolCallID] = tc
		return []aguievents.Event{event}

	case *aguievents.ToolCallArgsEvent:
		if tc, ok := p.toolCallMap[e.ToolCallID]; ok {
			tc.Function.Arguments += e.Delta
		}
		return []aguievents.Event{event}

	case *aguievents.ToolCallEndEvent:
		return []aguievents.Event{event}

	default:
		return []aguievents.Event{event}
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

func parseActivityContent(content any) map[string]interface{} {
	if content == nil {
		return nil
	}

	switch v := content.(type) {
	case map[string]interface{}:
		return v
	default:
		data, err := json.Marshal(content)
		if err != nil {
			log.Warnf("[ag-ui-util] parseActivityContent marshal error: %v", err)
			return nil
		}
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			log.Warnf("[ag-ui-util] parseActivityContent unmarshal error: %v", err)
			return nil
		}
		return result
	}
}

func FormatJSONResult(result string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		log.Warnf("[ag-ui-util] FormatJSONResult unmarshal error: %v", err)
		return result
	}

	formatted, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Warnf("[ag-ui-util] FormatJSONResult marshal error: %v", err)
		return result
	}
	return string(formatted)
}

func TruncateResult(maxLen int) func(string) string {
	return func(result string) string {
		if len(result) <= maxLen {
			return result
		}
		return result[:maxLen] + "..."
	}
}

func MaskSensitiveFields(sensitiveFields []string) func(string) string {
	return func(result string) string {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(result), &obj); err != nil {
			log.Warnf("[ag-ui-util] MaskSensitiveFields unmarshal error: %v", err)
			return result
		}

		for _, field := range sensitiveFields {
			if _, ok := obj[field]; ok {
				obj[field] = "***MASKED***"
			}
		}

		formatted, err := json.Marshal(obj)
		if err != nil {
			log.Warnf("[ag-ui-util] MaskSensitiveFields marshal error: %v", err)
			return result
		}
		return string(formatted)
	}
}

func RemovePrefixes(prefixes []string) func(string) string {
	return func(result string) string {
		for _, prefix := range prefixes {
			if len(result) > len(prefix) && result[:len(prefix)] == prefix {
				return result[len(prefix):]
			}
		}
		return result
	}
}
