package callback

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// AgentProxyChat
//
//	@Tags			callback
//	@Summary		智能体代理问答
//	@Description	智能体代理问答，固定流式返回，提取eventType=0的数据聚合返回
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AgentProxyChatReq	true	"智能体代理问答请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/agent/proxy/chat [post]
func AgentProxyChat(ctx *gin.Context) {
	var req request.AgentProxyChatReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	data, err := service.AgentProxyChat(ctx, &req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}

	gin_util.Response(ctx, data, err)
}
