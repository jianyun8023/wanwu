package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerSetting(apiV1 *gin.RouterGroup) {
	mid.Sub("setting").Reg(apiV1, "/custom/tab", http.MethodPost, v1.UploadCustomTab, "标签页自定义配置")
	mid.Sub("setting").Reg(apiV1, "/custom/login", http.MethodPost, v1.UploadCustomLogin, "登录页自定义配置")
	mid.Sub("setting").Reg(apiV1, "/custom/home", http.MethodPost, v1.UploadCustomHome, "平台自定义配置")
	mid.Sub("setting").Reg(apiV1, "/custom/general-agent", http.MethodPost, v1.UploadCustomGeneralAgent, "通用智能体自定义配置")
}
