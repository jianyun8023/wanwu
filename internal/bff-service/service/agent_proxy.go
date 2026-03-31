package service

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

const (
	mainAgentEventType = 0
)

func AgentProxyChat(ctx *gin.Context, req *request.AgentProxyChatReq) (string, error) {
	agentConfig := config.Cfg().AgentService
	url := fmt.Sprintf("http://%s:%d/agent/chat", agentConfig.Host, agentConfig.Port)

	agentReq := map[string]interface{}{
		"assistantId": req.AssistantId,
		"input":       req.Input,
		"uploadFile":  req.UploadFile,
		"stream":      true,
		"userId":      req.UserId,
		"orgId":       req.OrgId,
	}

	retCh, errCh := agentProxyStream(ctx.Request.Context(), url, agentReq)
	if err := <-errCh; err != nil {
		return "", grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}

	var aggregatedResponse strings.Builder
	for data := range retCh {
		aggregatedResponse.WriteString(data)
	}

	return aggregatedResponse.String(), nil
}

func agentProxyStream(ctx context.Context, url string, req map[string]interface{}) (<-chan string, <-chan error) {
	ret := make(chan string, 1024)
	errCh := make(chan error, 1)

	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		defer close(errCh)

		var resp *resty.Response
		var err error

		request := resty.New().
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
			R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetBody(req).
			SetDoNotParseResponse(true)

		resp, err = request.Post(url)
		if err != nil {
			wrappedErr := fmt.Errorf("agent proxy stream post request failed | url: %s | error: %v", url, err)
			log.Errorf("%v", wrappedErr.Error())
			errCh <- wrappedErr
			return
		}
		defer func() {
			if resp != nil && resp.RawResponse != nil {
				_ = resp.RawResponse.Body.Close()
			}
		}()

		if resp.StatusCode() >= 300 {
			b, err := io.ReadAll(resp.RawResponse.Body)
			if err != nil {
				wrappedErr := fmt.Errorf("agent proxy stream read response body failed | url: %s: %w", url, err)
				log.Errorf("%v", wrappedErr)
				errCh <- wrappedErr
				return
			}
			wrappedErr := fmt.Errorf("agent proxy stream request failed | url: %s | status: %d | message: %s", url, resp.StatusCode(), string(b))
			log.Errorf("%v", wrappedErr.Error())
			errCh <- wrappedErr
			return
		}

		close(errCh)

		scan := bufio.NewScanner(resp.RawResponse.Body)
		for scan.Scan() {
			sseData := scan.Text()
			data := parseAgentSSEData(sseData)
			if data == "" {
				continue
			}

			select {
			case ret <- data:
			case <-ctx.Done():
				log.Warnf("agent proxy stream ctx canceled | url: %s", url)
				return
			}
		}

		if scanErr := scan.Err(); scanErr != nil {
			log.Errorf("agent proxy stream scan err | url: %s | error: %v", url, scanErr)
		}
	}()

	return ret, errCh
}

func parseAgentSSEData(sseData string) string {
	sseData = strings.TrimSpace(sseData)
	if sseData == "" || !strings.HasPrefix(sseData, "data:") {
		return ""
	}

	dataStr := strings.TrimPrefix(sseData, "data:")
	dataStr = strings.TrimSpace(dataStr)
	if dataStr == "" {
		return ""
	}

	log.Infof("agent proxy response data: %s", dataStr)

	var agentResp response.AgentProxyChatResp
	if err := json.Unmarshal([]byte(dataStr), &agentResp); err != nil {
		log.Errorf("unmarshal agent response error: %v, data: %s", err, dataStr)
		return ""
	}

	if agentResp.EventType == mainAgentEventType {
		return agentResp.Response
	}

	return ""
}
