package qm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"qm-mcp-server/api/market/market"
)

// CallMarketAPI 通用的市场API调用方法
func (c *Client) CallMarketAPI(method, path string, data map[string]interface{}) (int, map[string]string, map[string]interface{}, error) {
	// 构建完整URL
	url := c.BuildURL(path)

	// 创建HTTP请求
	httpReq, err := createHTTPRequest(method, url, data)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加认证头
	authHeaders := c.GenerateAuthHeaders()
	for key, value := range authHeaders {
		httpReq.Header.Set(key, value)
	}

	// 设置Content-Type
	if method != "GET" && data != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应体
	var responseBody map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &responseBody); err != nil {
			// 如果不是JSON格式，将原始内容作为字符串返回
			responseBody = map[string]interface{}{
				"raw_content": string(body),
			}
		}
	}

	// 构建响应头
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	return resp.StatusCode, responseHeaders, responseBody, nil
}

// ListServices 获取服务列表
// /api/mcp/services post 获取服务列表
func (c *Client) ListServices(req *market.ListRequest) (*market.ListResponse, error) {
	// 构建请求数据
	requestData := map[string]interface{}{
		"offset":       0,
		"type":         "MCP_SERVICE",
		"officialOnly": false,
		"keyword":      "",
		"kind":         0,
		"categoryIds":  []string{},
		"industryId":   "",
	}

	// 调用市场API
	_, _, responseBody, err := c.CallMarketAPI("POST", "/market/items", requestData)
	if err != nil {
		return nil, err
	}

	// 转换为pb结构
	return c.convertToListResponse(responseBody), nil
}

// GetCategories 获取服务分类
// /api/mcp/categories get 获取分类列表
func (c *Client) GetCategories(req *market.CategoryRequest) (*market.CategoryResponse, error) {
	// 调用市场API获取分类
	_, _, responseBody, err := c.CallMarketAPI("GET", "/market/publish/category", nil)
	if err != nil {
		return nil, err
	}

	// 转换为pb结构
	return c.convertToCategoryResponse(responseBody), nil
}

// GetServiceDetail 获取服务详情
// /api/mcp/services/{serviceId} get 获取服务详情
func (c *Client) GetServiceDetail(req *market.DetailRequest) (*market.DetailResponse, error) {
	// 调用市场API
	_, _, responseBody, err := c.CallMarketAPI("GET", fmt.Sprintf("/market/mcp/%d", req.Id), nil)
	if err != nil {
		return nil, err
	}

	// 转换为pb结构
	return c.convertToDetailResponse(responseBody), nil
}

// convertToListResponse 将qm API响应转换为pb ListResponse结构
func (c *Client) convertToListResponse(response map[string]interface{}) *market.ListResponse {
	pbResponse := &market.ListResponse{}

	// 优先解析新结构：{"code":200,"message":"操作成功","data":{ "nextOffset":..., "hasMore":..., "items":[...] }}
	if data, ok := response["data"].(map[string]interface{}); ok {
		if items, ok := data["items"].([]interface{}); ok {
			for _, item := range items {
				if serviceMap, ok := item.(map[string]interface{}); ok {
					serviceInfo := c.convertToListServiceInfo(serviceMap)
					pbResponse.List = append(pbResponse.List, serviceInfo)
				}
			}
			if total, ok := data["total"].(float64); ok {
				pbResponse.Total = int32(total)
			} else {
				// 当未返回总数时，使用当前返回数量作为兜底
				pbResponse.Total = int32(len(pbResponse.List))
			}
			return pbResponse
		}
		// 兼容 data 下为 list 的旧结构
		if listData, ok := data["list"].([]interface{}); ok {
			for _, item := range listData {
				if serviceMap, ok := item.(map[string]interface{}); ok {
					serviceInfo := c.convertToListServiceInfo(serviceMap)
					pbResponse.List = append(pbResponse.List, serviceInfo)
				}
			}
			if total, ok := data["total"].(float64); ok {
				pbResponse.Total = int32(total)
			} else {
				pbResponse.Total = int32(len(pbResponse.List))
			}
			return pbResponse
		}
	}

	// 兼容最旧结构：根级包含 total 和 list
	if total, ok := response["total"].(float64); ok {
		pbResponse.Total = int32(total)
	}
	if listData, ok := response["list"].([]interface{}); ok {
		for _, item := range listData {
			if serviceMap, ok := item.(map[string]interface{}); ok {
				serviceInfo := c.convertToListServiceInfo(serviceMap)
				pbResponse.List = append(pbResponse.List, serviceInfo)
			}
		}
	}

	return pbResponse
}

// convertToListServiceInfo 将map数据转换为pb ListResponse_ServiceInfo结构
func (c *Client) convertToListServiceInfo(data map[string]interface{}) *market.ListResponse_ServiceInfo {
	serviceInfo := &market.ListResponse_ServiceInfo{}

	// 基础字段转换 - 根据market.proto中的ListResponse_ServiceInfo定义
	if id, ok := data["id"].(float64); ok {
		serviceInfo.Id = int32(id)
	}
	if name, ok := data["name"].(string); ok {
		serviceInfo.Name = name
	}
	if iconUrl, ok := data["iconUrl"].(string); ok {
		serviceInfo.IconUrl = iconUrl
	}
	// brief 优先读取 brief，若无则尝试用 description 兜底
	if brief, ok := data["brief"].(string); ok {
		serviceInfo.Brief = brief
	} else if desc, ok := data["description"].(string); ok {
		serviceInfo.Brief = desc
	}
	if typeId, ok := data["typeId"].(float64); ok {
		serviceInfo.TypeId = int32(typeId)
	}

	// categoryIds 数组字段（兼容不同返回结构：categoryIds 或 categories）
	if categoryIds, ok := data["categoryIds"].([]interface{}); ok {
		for _, categoryId := range categoryIds {
			if id, ok := categoryId.(float64); ok {
				serviceInfo.CategoryIds = append(serviceInfo.CategoryIds, int32(id))
			}
		}
	} else if categories, ok := data["categories"].([]interface{}); ok {
		for _, cat := range categories {
			switch v := cat.(type) {
			case float64:
				serviceInfo.CategoryIds = append(serviceInfo.CategoryIds, int32(v))
			case map[string]interface{}:
				if rawID, ok := v["id"]; ok {
					switch id := rawID.(type) {
					case float64:
						serviceInfo.CategoryIds = append(serviceInfo.CategoryIds, int32(id))
					case int:
						serviceInfo.CategoryIds = append(serviceInfo.CategoryIds, int32(id))
					case int32:
						serviceInfo.CategoryIds = append(serviceInfo.CategoryIds, id)
					}
				}
			}
		}
	}

	// typeName 若无，尝试从 "type" 映射
	if typeName, ok := data["typeName"].(string); ok {
		serviceInfo.TypeName = typeName
	} else if t, ok := data["type"].(string); ok {
		serviceInfo.TypeName = t
	}
	if publishTime, ok := data["publishTime"].(float64); ok {
		serviceInfo.PublishTime = int64(publishTime)
	}
	if offlineTime, ok := data["offlineTime"].(float64); ok {
		serviceInfo.OfflineTime = int64(offlineTime)
	}
	if toolCount, ok := data["toolCount"].(float64); ok {
		serviceInfo.ToolCount = int32(toolCount)
	}
	if providerName, ok := data["providerName"].(string); ok {
		serviceInfo.ProviderName = providerName
	}
	if providerContact, ok := data["providerContact"].(string); ok {
		serviceInfo.ProviderContact = providerContact
	}
	if status, ok := data["status"].(float64); ok {
		serviceInfo.Status = int32(status)
	}
	if deployMode, ok := data["deployMode"].(string); ok {
		serviceInfo.DeployMode = deployMode
	}

	return serviceInfo
}

// convertToCategoryResponse 将qm API响应转换为pb CategoryResponse结构
func (c *Client) convertToCategoryResponse(response map[string]interface{}) *market.CategoryResponse {
	// 创建响应结构
	categoryResponse := &market.CategoryResponse{
		List: []*market.CategoryInfo{},
	}

	// 解析分类数据
	if data, ok := response["data"]; ok {
		if dataArray, ok := data.([]interface{}); ok {
			for _, item := range dataArray {
				if categoryData, ok := item.(map[string]interface{}); ok {
					categoryInfo := c.convertToCategoryInfo(categoryData)
					if categoryInfo != nil {
						categoryResponse.List = append(categoryResponse.List, categoryInfo)
					}
				}
			}
		}
	}

	return categoryResponse
}

// convertToCategoryInfo 将map数据转换为pb CategoryInfo结构
func (c *Client) convertToCategoryInfo(data map[string]interface{}) *market.CategoryInfo {
	categoryInfo := &market.CategoryInfo{}

	// 转换ID
	if id, ok := data["id"]; ok {
		switch v := id.(type) {
		case float64:
			categoryInfo.Id = int32(v)
		case int:
			categoryInfo.Id = int32(v)
		case int32:
			categoryInfo.Id = v
		}
	}

	// 转换名称
	if name, ok := data["name"]; ok {
		if nameStr, ok := name.(string); ok {
			categoryInfo.Name = nameStr
		}
	}

	return categoryInfo
}

// convertToDetailResponse 将qm API响应转换为pb DetailResponse结构
func (c *Client) convertToDetailResponse(response map[string]interface{}) *market.DetailResponse {
	pbResponse := &market.DetailResponse{}

	// 优先解析新结构：{"code":200,"message":"操作成功","data":{...}}
	if data, ok := response["data"].(map[string]interface{}); ok {
		pbResponse.Service = c.convertToServiceInfo(data)
		return pbResponse
	}

	// 兼容旧结构：转换服务详情
	if serviceData, ok := response["service"].(map[string]interface{}); ok {
		pbResponse.Service = c.convertToServiceInfo(serviceData)
	} else {
		// 如果response直接是服务数据
		pbResponse.Service = c.convertToServiceInfo(response)
	}

	return pbResponse
}

// convertToServiceInfo 将map数据转换为pb ServiceInfo结构
func (c *Client) convertToServiceInfo(data map[string]interface{}) *market.ServiceInfo {
	serviceInfo := &market.ServiceInfo{}

	// 基础字段转换 - 根据market.proto中的ServiceInfo定义
	if id, ok := data["id"].(float64); ok {
		serviceInfo.Id = int32(id)
	}
	if name, ok := data["name"].(string); ok {
		serviceInfo.Name = name
	}

	// categoryIds 数组字段
	if categoryIds, ok := data["categoryIds"].([]interface{}); ok {
		for _, categoryId := range categoryIds {
			if id, ok := categoryId.(float64); ok {
				serviceInfo.CategoryIds = append(serviceInfo.CategoryIds, int32(id))
			}
		}
	}

	if description, ok := data["description"].(string); ok {
		serviceInfo.Description = description
	}
	if detail, ok := data["detail"].(string); ok {
		serviceInfo.Detail = detail
	}
	if secretTips, ok := data["secretTips"].(string); ok {
		serviceInfo.SecretTips = secretTips
	}
	if cloudApiUrl, ok := data["cloudApiUrl"].(string); ok {
		serviceInfo.CloudApiUrl = cloudApiUrl
	}
	if iconUrl, ok := data["iconUrl"].(string); ok {
		serviceInfo.IconUrl = iconUrl
	}
	if providerName, ok := data["providerName"].(string); ok {
		serviceInfo.ProviderName = providerName
	}
	if providerContact, ok := data["providerContact"].(string); ok {
		serviceInfo.ProviderContact = providerContact
	}
	if overview, ok := data["overview"].(string); ok {
		serviceInfo.Overview = overview
	}
	if configTemplate, ok := data["configTemplate"].(string); ok {
		serviceInfo.ConfigTemplate = configTemplate
	}
	if deployMode, ok := data["deployMode"].(string); ok {
		serviceInfo.DeployMode = deployMode
	}
	if codeUrl, ok := data["codeUrl"].(string); ok {
		serviceInfo.CodeUrl = codeUrl
	}
	if imageUrl, ok := data["imageUrl"].(string); ok {
		serviceInfo.ImageUrl = imageUrl
	}
	if publishTime, ok := data["publishTime"].(float64); ok {
		serviceInfo.PublishTime = int64(publishTime)
	}
	// status 字段支持字符串和数字类型
	if status, ok := data["status"].(string); ok {
		serviceInfo.Status = status
	} else if status, ok := data["status"].(float64); ok {
		// 如果是数字类型，转换为字符串
		serviceInfo.Status = fmt.Sprintf("%.0f", status)
	}
	if subscribeStatus, ok := data["subscribeStatus"].(string); ok {
		serviceInfo.SubscribeStatus = subscribeStatus
	}
	if categoryIdsJson, ok := data["categoryIdsJson"].(string); ok {
		serviceInfo.CategoryIdsJson = categoryIdsJson
	}

	// tools 数组字段转换
	if tools, ok := data["tools"].([]interface{}); ok {
		for _, tool := range tools {
			if toolMap, ok := tool.(map[string]interface{}); ok {
				toolInfo := &market.ServiceInfo_ToolInfo{}

				if name, ok := toolMap["name"].(string); ok {
					toolInfo.Name = name
				}
				if description, ok := toolMap["description"].(string); ok {
					toolInfo.Description = description
				}

				// params 数组字段转换
				if params, ok := toolMap["params"].([]interface{}); ok {
					for _, param := range params {
						if paramMap, ok := param.(map[string]interface{}); ok {
							paramInfo := &market.ServiceInfo_ParamInfo{}

							if name, ok := paramMap["name"].(string); ok {
								paramInfo.Name = name
							}
							if paramType, ok := paramMap["type"].(string); ok {
								paramInfo.Type = paramType
							}
							if required, ok := paramMap["required"].(bool); ok {
								paramInfo.Required = required
							}

							toolInfo.Params = append(toolInfo.Params, paramInfo)
						}
					}
				}

				serviceInfo.Tools = append(serviceInfo.Tools, toolInfo)
			}
		}
	}

	return serviceInfo
}

// createHTTPRequest 创建HTTP请求
func createHTTPRequest(method, url string, data map[string]interface{}) (*http.Request, error) {
	var body io.Reader

	if method != "GET" && data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("序列化请求数据失败: %v", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	return req, nil
}
