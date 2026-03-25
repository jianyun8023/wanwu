package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga"
	wga_persistent "github.com/UnicomAI/wanwu/pkg/wga-persistent"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GeneralAgentConversationChat(ctx *gin.Context, userId, orgId string, req request.GeneralAgentConversationChatReq) error {
	agentID := config.WgaCfg().AgentID
	runID := uuid.NewString()

	opts := buildWgaOptions(ctx, config.WgaCfg(), req.ThreadID, runID, req.Messages)

	_, iter, err := wga.Run(ctx.Request.Context(), agentID, opts...)
	if err != nil {
		return err
	}

	tr := ag_ui_util.NewEinoMultiAgentTranslator(req.ThreadID, runID)
	eventCh := tr.TranslateStream(ctx.Request.Context(), iter)

	outputCh := injectWgaWorkspaceActivity(
		ctx.Request.Context(),
		eventCh,
		req.ThreadID,
		runID,
		config.WgaCfg().Persistent.BaseDir,
		config.WgaCfg().Persistent.Enabled,
	)

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	ctx.Stream(func(w io.Writer) bool {
		select {
		case line, ok := <-outputCh:
			if !ok {
				return false
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", line)
			return true
		case <-ctx.Request.Context().Done():
			return false
		}
	})
	return nil
}

func buildWgaOptions(ctx *gin.Context, cfg *config.WgaConfig, threadID, runID string, messages []request.GeneralAgentConversationMessage) []wga_option.Option {
	opts := []wga_option.Option{
		wga_option.WithModelConfig(wga_option.ModelConfig{
			Provider:     cfg.Model.Provider,
			ProviderName: cfg.Model.ProviderName,
			BaseURL:      cfg.Model.BaseURL,
			APIKey:       cfg.Model.APIKey,
			Model:        cfg.Model.Model,
			ModelName:    cfg.Model.ModelName,
		}),
		wga_option.WithRunSession(wga_option.RunSession{
			ThreadID: threadID,
			RunID:    runID,
		}),
	}

	if cfg.Persistent.Enabled {
		// 持久化存储
		mode := wga_persistent.ModeVersioned
		if cfg.Persistent.Mode == string(wga_persistent.ModeOverwrite) {
			mode = wga_persistent.ModeOverwrite
		}
		store, err := wga_persistent.NewStore(mode, cfg.Persistent.BaseDir, threadID)
		if err == nil {
			// 创建目录并从上一次输出复制
			_, info, err := store.GetRunDir(runID, wga_persistent.WithMkdir(true))
			if err == nil {
				opts = append(opts,
					wga_option.WithInputDir(filepath.Clean(info.Dir)+"/."),
					wga_option.WithOutputDir(info.Dir),
				)
			}
		}
	}

	// 传递历史消息
	if len(messages) > 0 {
		msgs := make([]adk.Message, len(messages))
		for i, msg := range messages {
			msgs[i] = convertWgaMessage(ctx, msg.Role, msg.Content)
		}
		opts = append(opts, wga_option.WithMessages(msgs))
	}

	for _, tool := range cfg.Tools {
		opts = append(opts, wga_option.WithToolConfig(wga_option.ToolConfig{
			Title:   tool.Title,
			APIAuth: &tool.APIAuth,
		}))
	}

	return opts
}

func injectWgaWorkspaceActivity(
	ctx context.Context,
	eventCh <-chan aguievents.Event,
	threadID, runID, baseDir string,
	persistentEnabled bool,
) <-chan string {
	out := make(chan string, 1024)
	go func() {
		defer util.PrintPanicStack()
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-eventCh:
				if !ok {
					return
				}

				if evt.Type() == aguievents.EventTypeRunFinished {
					if wsEvent := buildWgaWorkspaceEvent(threadID, runID, baseDir, persistentEnabled); wsEvent != nil {
						if data, err := json.Marshal(wsEvent); err == nil {
							select {
							case out <- string(data):
							case <-ctx.Done():
								return
							}
						}
					}
				}

				if data, err := json.Marshal(evt); err == nil {
					select {
					case out <- string(data):
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}

func buildWgaWorkspaceEvent(threadID, runID, baseDir string, persistentEnabled bool) aguievents.Event {
	if !persistentEnabled {
		return nil
	}

	store, err := wga_persistent.NewStore(wga_persistent.ModeVersioned, baseDir, threadID)
	if err != nil {
		return nil
	}

	ok, info, err := store.GetRunDir(runID)
	if err != nil || !ok {
		return nil
	}

	statInfo, err := os.Stat(info.Dir)
	if err != nil || !statInfo.IsDir() {
		return nil
	}

	totalSize, fileCount, err := getWgaWorkspaceInfo(info.Dir, info.Dir)
	if err != nil || fileCount == 0 {
		return nil
	}

	content := map[string]interface{}{
		"runId":     runID,
		"threadId":  threadID,
		"fileCount": fileCount,
		"totalSize": totalSize,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return aguievents.NewActivitySnapshotEvent(
		aguievents.GenerateStepID(),
		"workspace",
		content,
	)
}

func getWgaWorkspaceInfo(rootDir, currentDir string) (int64, int, error) {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return 0, 0, err
	}

	var totalSize int64
	var fileCount int

	for _, entry := range entries {
		// 跳过隐藏文件
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(currentDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			log.Warnf("[wga] failed to get file info: %s: %v", fullPath, err)
			continue
		}

		if entry.IsDir() {
			dirSize, dirFileCount, err := getWgaWorkspaceInfo(rootDir, fullPath)
			if err != nil {
				log.Warnf("[wga] failed to build file tree for dir: %s: %v", fullPath, err)
				continue
			}
			totalSize += dirSize
			fileCount += dirFileCount
		} else {
			totalSize += info.Size()
			fileCount++
		}
	}

	return totalSize, fileCount, nil
}

func convertWgaMessage(ctx *gin.Context, role string, content interface{}) *schema.Message {
	switch v := content.(type) {
	case string:
		return &schema.Message{
			Role:    schema.RoleType(role),
			Content: v,
		}
	case []interface{}:
		parts := make([]schema.MessageInputPart, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				parts = append(parts, convertWgaMessageInputPart(ctx, m))
			}
		}
		if len(parts) == 0 {
			return &schema.Message{Role: schema.RoleType(role)}
		}
		return &schema.Message{
			Role:                  schema.RoleType(role),
			UserInputMultiContent: parts,
		}
	default:
		return &schema.Message{Role: schema.RoleType(role)}
	}
}

func convertWgaMessageInputPart(ctx *gin.Context, m map[string]interface{}) schema.MessageInputPart {
	part := schema.MessageInputPart{}

	typ, _ := m["type"].(string)
	switch typ {
	case "text":
		part.Type = schema.ChatMessagePartTypeText
		if text, ok := m["text"].(string); ok {
			part.Text = text
		}
	case "binary":
		mimeType, _ := m["mimeType"].(string)
		url, _ := m["url"].(string)
		base64Data, _ := FileUrlConvertBase64(ctx, &request.FileUrlConvertBase64Req{
			FileUrl: url,
		})
		switch {
		case strings.HasPrefix(mimeType, "image/"):
			part.Type = schema.ChatMessagePartTypeImageURL
			part.Image = &schema.MessageInputImage{
				MessagePartCommon: schema.MessagePartCommon{
					Base64Data: &base64Data,
					MIMEType:   mimeType,
				},
			}
		case strings.HasPrefix(mimeType, "audio/"):
			part.Type = schema.ChatMessagePartTypeAudioURL
			part.Audio = &schema.MessageInputAudio{
				MessagePartCommon: schema.MessagePartCommon{
					Base64Data: &base64Data,
					MIMEType:   mimeType,
				},
			}
		case strings.HasPrefix(mimeType, "video/"):
			part.Type = schema.ChatMessagePartTypeVideoURL
			part.Video = &schema.MessageInputVideo{
				MessagePartCommon: schema.MessagePartCommon{
					Base64Data: &base64Data,
					MIMEType:   mimeType,
				},
			}
		default:
			part.Type = schema.ChatMessagePartTypeFileURL
			part.File = &schema.MessageInputFile{
				MessagePartCommon: schema.MessagePartCommon{
					Base64Data: &base64Data,
					MIMEType:   mimeType,
				},
			}
		}
	default:
		part.Type = schema.ChatMessagePartType(typ)
		if text, ok := m["text"].(string); ok {
			part.Text = text
		}
	}
	return part
}
