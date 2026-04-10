package service

import (
	"fmt"
	"net/http"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/gin-gonic/gin"
)

func ModelPdfParser(ctx *gin.Context, modelID string, req *mp_common.PdfParserReq) {
	// modelInfo by modelID
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelID})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if !modelInfo.IsActive {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFModelStatus, modelInfo.ModelId))
		return
	}

	// pdfParser config
	pdfParser, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, modelPdfParserErrMsg(modelInfo, fmt.Sprintf("pdfParser err: %v", err))))
		return
	}
	iPdfParser, ok := pdfParser.(mp.IPdfParser)
	if !ok {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, modelPdfParserErrMsg(modelInfo, "pdfParser err: invalid provider")))
		return
	}

	pdfParserReq, err := iPdfParser.NewReq(req)
	if err != nil {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, modelPdfParserErrMsg(modelInfo, fmt.Sprintf("pdfParser NewReq err: %v", err))))
		return
	}
	startTime := time.Now()
	resp, err := iPdfParser.PdfParser(ctx, pdfParserReq)
	if err != nil {
		recordModelStatistic(ctx, modelInfo, false, 0, 0, 0, 0, 0, false)
		gin_util.ResponseErrWithStatus(ctx, http.StatusBadGateway, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, modelPdfParserErrMsg(modelInfo, fmt.Sprintf("pdfParser err: %v", err))))
		return
	}
	data, err := resp.ConvertRespWithErr()
	if err == nil {
		status := http.StatusOK
		ctx.Set(gin_util.STATUS, status)
		//ctx.Set(config.RESULT, resp.String())
		ctx.JSON(status, data)
		costs := int(time.Since(startTime).Milliseconds())
		recordModelStatistic(ctx, modelInfo, true, 0, 0, 0, costs, 0, false)
		return
	}
	recordModelStatistic(ctx, modelInfo, false, 0, 0, 0, 0, 0, false)
	gin_util.ResponseErrWithStatus(ctx, http.StatusBadGateway, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, modelPdfParserErrMsg(modelInfo, fmt.Sprintf("pdfParser err: %v", err))))
}

func modelPdfParserErrMsg(modelInfo *model_service.ModelInfo, detail string) string {
	displayName := modelInfo.DisplayName
	if displayName == "" {
		displayName = modelInfo.Model
	}
	return fmt.Sprintf("modelId=%v model=%v displayName=%v provider=%v %v", modelInfo.ModelId, modelInfo.Model, displayName, modelInfo.Provider, detail)
}
