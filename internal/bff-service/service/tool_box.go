package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// 所有工具的创建/更新时间统一使用 2026-01-01 00:00:00 UTC（纳秒）
const toolBoxFixedTimeNs int64 = 1767225600_000_000_000

// 工具创建/更新用户：内置工具固定为 system，自定义工具固定为 user
const (
	toolBoxBuiltinUser = "system"
	toolBoxCustomUser  = "user"
)

// toolBoxSource 一份工具箱的 schema 与时间/作者/鉴权元数据
type toolBoxSource struct {
	schema       string
	createTimeNs int64
	updateTimeNs int64
	createUser   string
	updateUser   string
	apiKey       string
	apiAuth      response.ToolBoxAPIAuth
}

// GetToolBoxDetail 工具箱明细查询：按 box_id + box_type 解析 schema 摊平成 tools[]
func GetToolBoxDetail(ctx *gin.Context, userID, orgID string, req *request.ToolBoxDetailReq) (*response.ToolBoxDetail, error) {
	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}

	if req.BoxID == "" || (req.BoxType != constant.ToolTypeBuiltIn && req.BoxType != constant.ToolTypeCustom) {
		return emptyToolBoxDetail(req.BoxID, page, pageSize), nil
	}

	src, err := fetchToolBoxSource(ctx, userID, orgID, req.BoxID, req.BoxType)
	if err != nil {
		return nil, err
	}
	tools, err := parseSchema2ToolBoxItems(ctx.Request.Context(), src)
	if err != nil {
		return nil, err
	}
	if req.ToolID != "" {
		tools = filterToolsByID(tools, req.ToolID)
	}
	return &response.ToolBoxDetail{
		Total:      len(tools),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
		BoxID:      req.BoxID,
		APIKey:     src.apiKey,
		APIAuth:    src.apiAuth,
		Tools:      tools,
	}, nil
}

func emptyToolBoxDetail(boxID string, page, pageSize int) *response.ToolBoxDetail {
	return &response.ToolBoxDetail{
		Page:     page,
		PageSize: pageSize,
		BoxID:    boxID,
		Tools:    []response.ToolBoxToolItem{},
	}
}

func fetchToolBoxSource(ctx *gin.Context, userID, orgID, boxID, boxType string) (toolBoxSource, error) {
	switch boxType {
	case constant.ToolTypeBuiltIn:
		resp, err := mcp.GetSquareTool(ctx.Request.Context(), &mcp_service.GetSquareToolReq{
			ToolSquareId: boxID,
			Identity: &mcp_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		if err != nil {
			return toolBoxSource{}, err
		}
		auth := toToolBoxAPIAuth(resp.GetBuiltInTools().GetApiAuth())
		return toolBoxSource{
			schema:       resp.Schema,
			createTimeNs: toolBoxFixedTimeNs,
			updateTimeNs: toolBoxFixedTimeNs,
			createUser:   toolBoxBuiltinUser,
			updateUser:   toolBoxBuiltinUser,
			apiKey:       auth.APIKeyValue,
			apiAuth:      auth,
		}, nil
	case constant.ToolTypeCustom:
		resp, err := mcp.GetCustomToolInfo(ctx.Request.Context(), &mcp_service.GetCustomToolInfoReq{
			CustomToolId: boxID,
		})
		if err != nil {
			return toolBoxSource{}, err
		}
		auth := toToolBoxAPIAuth(resp.GetApiAuth())
		return toolBoxSource{
			schema:       resp.Schema,
			createTimeNs: toolBoxFixedTimeNs,
			updateTimeNs: toolBoxFixedTimeNs,
			createUser:   toolBoxCustomUser,
			updateUser:   toolBoxCustomUser,
			apiKey:       auth.APIKeyValue,
			apiAuth:      auth,
		}, nil
	}
	return toolBoxSource{}, nil
}

// --- internal ---

// docContext OpenAPI 文档级别共享上下文（同一 box 内多个 action 复用）
type docContext struct {
	version    string
	serverURL  string
	components any
}

// parseSchema2ToolBoxItems 将 OpenAPI schema 摊平成 tools[]
//
// openapi3 库做严格 schema 验证 + 取顶层字段；raw map 保留原始 $ref 供对接方使用。
func parseSchema2ToolBoxItems(ctx context.Context, src toolBoxSource) ([]response.ToolBoxToolItem, error) {
	if strings.TrimSpace(src.schema) == "" {
		return []response.ToolBoxToolItem{}, nil
	}
	doc, err := openapi3_util.LoadFromData(ctx, []byte(src.schema))
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("parse openapi schema: %v", err))
	}
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(src.schema), &raw); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("parse openapi schema raw: %v", err))
	}
	rawPaths, _ := raw["paths"].(map[string]any)

	docCtx := docContext{
		version:    doc.Info.Version,
		serverURL:  firstServerURL(doc),
		components: raw["components"],
	}

	pathKeys := make([]string, 0, len(doc.Paths))
	for p := range doc.Paths {
		pathKeys = append(pathKeys, p)
	}
	sort.Strings(pathKeys)

	tools := make([]response.ToolBoxToolItem, 0)
	usedIDs := map[string]int{}
	for _, path := range pathKeys {
		pathItem := doc.Paths[path]
		if pathItem == nil {
			continue
		}
		rawPathItem, _ := rawPaths[path].(map[string]any)
		tools = append(tools, buildToolItemsForPath(pathItem, rawPathItem, path, docCtx, src, usedIDs)...)
	}
	return tools, nil
}

// buildToolItemsForPath 把单个 path 下的所有 HTTP method 摊平成 tools[] 元素。
// method 来自 pathItem.Operations()（大写）；raw schema 里 method key 是小写，需要做一次大小写映射。
func buildToolItemsForPath(pathItem *openapi3.PathItem, rawPathItem map[string]any, path string,
	doc docContext, src toolBoxSource, usedIDs map[string]int) []response.ToolBoxToolItem {
	ops := pathItem.Operations()
	methods := make([]string, 0, len(ops))
	for m := range ops {
		methods = append(methods, m)
	}
	sort.Strings(methods) // 排序保证返回稳定
	out := make([]response.ToolBoxToolItem, 0, len(methods))
	for _, method := range methods {
		rawOp, _ := rawPathItem[strings.ToLower(method)].(map[string]any)
		if rawOp == nil {
			continue
		}
		out = append(out, buildToolItem(rawOp, path, method, doc, src, usedIDs))
	}
	return out
}

func buildToolItem(op map[string]any, path, method string, doc docContext, src toolBoxSource, usedIDs map[string]int) response.ToolBoxToolItem {
	operationID := pickOperationID(op, path, method)
	toolID := uniqueToolID(operationID, usedIDs)
	desc, _ := op["description"].(string)
	summary, _ := op["summary"].(string)
	description := firstNonEmpty(desc, summary)
	params, _ := op["parameters"].([]any)
	if params == nil {
		params = []any{}
	}

	return response.ToolBoxToolItem{
		ToolID:       toolID,
		Name:         operationID,
		Description:  description,
		Status:       "enabled",
		MetadataType: "openapi",
		Metadata: response.ToolBoxMetadata{
			Version:     doc.version,
			Summary:     operationID,
			Description: description,
			ServerURL:   doc.serverURL,
			Path:        path,
			Method:      method, // 来自 pathItem.Operations()，已经是大写
			CreateTime:  src.createTimeNs,
			UpdateTime:  src.updateTimeNs,
			CreateUser:  src.createUser,
			UpdateUser:  src.updateUser,
			APISpec: response.ToolBoxAPISpec{
				Parameters:   params,
				RequestBody:  op["requestBody"],
				Responses:    openapiResponsesToArray(op["responses"]),
				Components:   doc.components,
				Callbacks:    op["callbacks"],
				Security:     op["security"],
				Tags:         stringSliceAt(op, "tags"),
				ExternalDocs: op["externalDocs"],
			},
		},
		UseRule:          "",
		GlobalParameters: response.ToolBoxGlobalParams{},
		CreateTime:       src.createTimeNs,
		UpdateTime:       src.updateTimeNs,
		CreateUser:       src.createUser,
		UpdateUser:       src.updateUser,
		ExtendInfo:       map[string]any{},
		ResourceObject:   "tool",
	}
}

// filterToolsByID 按 action 的 tool_id 精确过滤
func filterToolsByID(tools []response.ToolBoxToolItem, toolID string) []response.ToolBoxToolItem {
	out := make([]response.ToolBoxToolItem, 0, 1)
	for _, t := range tools {
		if t.ToolID == toolID {
			out = append(out, t)
		}
	}
	return out
}

// toToolBoxAPIAuth 把下游 common.ApiAuthWebRequest 转成对外的 snake_case 结构
func toToolBoxAPIAuth(a *common.ApiAuthWebRequest) response.ToolBoxAPIAuth {
	if a == nil {
		return response.ToolBoxAPIAuth{}
	}
	return response.ToolBoxAPIAuth{
		AuthType:           a.AuthType,
		APIKeyHeaderPrefix: a.ApiKeyHeaderPrefix,
		APIKeyHeader:       a.ApiKeyHeader,
		APIKeyQueryParam:   a.ApiKeyQueryParam,
		APIKeyValue:        a.ApiKeyValue,
	}
}

// pickOperationID 取 operationId，空则用 method_path 兜底
func pickOperationID(op map[string]any, path, method string) string {
	id, _ := op["operationId"].(string)
	if id != "" {
		return id
	}
	id = fmt.Sprintf("%s_%s", strings.ToLower(method), strings.Trim(path, "/"))
	return strings.NewReplacer("/", "_", "{", "", "}", "").Replace(id)
}

// uniqueToolID 同一 box 内若 operationId 重复，追加 _2 / _3 后缀
func uniqueToolID(id string, used map[string]int) string {
	n, ok := used[id]
	if !ok {
		used[id] = 1
		return id
	}
	used[id] = n + 1
	return fmt.Sprintf("%s_%d", id, n+1)
}

func openapiResponsesToArray(raw any) []response.ToolBoxResponseItem {
	out := []response.ToolBoxResponseItem{}
	m, _ := raw.(map[string]any)
	if m == nil {
		return out
	}
	codes := make([]string, 0, len(m))
	for code := range m {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	for _, code := range codes {
		body, _ := m[code].(map[string]any)
		if body == nil {
			continue
		}
		desc, _ := body["description"].(string)
		out = append(out, response.ToolBoxResponseItem{
			StatusCode:  code,
			Description: desc,
			Content:     body["content"],
		})
	}
	return out
}

func firstServerURL(doc *openapi3.T) string {
	if doc == nil || len(doc.Servers) == 0 {
		return ""
	}
	return doc.Servers[0].URL
}

func stringSliceAt(m map[string]any, key string) []string {
	raw, _ := m[key].([]any)
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}
