package qm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kymo-mcp/mcpcan/api/market/market"
)

// CallMarketAPI general market API call method
func (c *Client) CallMarketAPI(method, path string, data map[string]interface{}) (int, map[string]string, map[string]interface{}, error) {
	// Build full URL
	url := c.BuildURL(path)

	// Create HTTP request
	httpReq, err := createHTTPRequest(method, url, data)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add authentication headers
	authHeaders := c.GenerateAuthHeaders()
	for key, value := range authHeaders {
		httpReq.Header.Set(key, value)
	}

	// Set Content-Type
	if method != "GET" && data != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Send request
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response body
	var responseBody map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &responseBody); err != nil {
			// If not JSON format, return raw content as string
			responseBody = map[string]interface{}{
				"raw_content": string(body),
			}
		}
	}

	// Build response headers
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	return resp.StatusCode, responseHeaders, responseBody, nil
}

// ListServices gets the list of services
// /api/mcp/services POST - get service list
func (c *Client) ListServices(req *market.ListRequest) (*market.ListResponse, error) {
	// Build request data
	requestData := map[string]interface{}{
		"offset":       0,
		"type":         "MCP_SERVICE",
		"officialOnly": false,
		"keyword":      "",
		"kind":         0,
		"categoryIds":  []string{},
		"industryId":   "",
	}

	// Call market API
	_, _, responseBody, err := c.CallMarketAPI("POST", "/market/items", requestData)
	if err != nil {
		return nil, err
	}

	// Convert to pb structure
	return c.convertToListResponse(responseBody), nil
}

// GetCategories gets service categories
// /api/mcp/categories GET - get category list
func (c *Client) GetCategories(req *market.CategoryRequest) (*market.CategoryResponse, error) {
	// Call market API to get categories
	_, _, responseBody, err := c.CallMarketAPI("GET", "/market/publish/category", nil)
	if err != nil {
		return nil, err
	}

	// Convert to pb structure
	return c.convertToCategoryResponse(responseBody), nil
}

// GetServiceDetail gets service details
// /api/mcp/services/{serviceId} GET - get service details
func (c *Client) GetServiceDetail(req *market.DetailRequest) (*market.DetailResponse, error) {
	// Call market API
	_, _, responseBody, err := c.CallMarketAPI("GET", fmt.Sprintf("/market/mcp/%d", req.Id), nil)
	if err != nil {
		return nil, err
	}

	// Convert to pb structure
	return c.convertToDetailResponse(responseBody), nil
}

// convertToListResponse converts qm API response to pb ListResponse structure
func (c *Client) convertToListResponse(response map[string]interface{}) *market.ListResponse {
	pbResponse := &market.ListResponse{}

	// Prioritize parsing new structure: {"code":200,"message":"Operation successful","data":{ "nextOffset":..., "hasMore":..., "items":[...] }}
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
				// If total is not returned, use current count as fallback
				pbResponse.Total = int32(len(pbResponse.List))
			}
			return pbResponse
		}
		// Compatible with old structure where data is a list
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

	// Compatible with oldest structure: total and list at root level
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

// convertToListServiceInfo converts map data to pb ListResponse_ServiceInfo structure
func (c *Client) convertToListServiceInfo(data map[string]interface{}) *market.ListResponse_ServiceInfo {
	serviceInfo := &market.ListResponse_ServiceInfo{}

	// Basic field conversion - according to ListResponse_ServiceInfo definition in market.proto
	if id, ok := data["id"].(float64); ok {
		serviceInfo.Id = int32(id)
	}
	if name, ok := data["name"].(string); ok {
		serviceInfo.Name = name
	}
	if iconUrl, ok := data["iconUrl"].(string); ok {
		serviceInfo.IconUrl = iconUrl
	}
	// brief: prioritize reading brief, if not available, try description as fallback
	if brief, ok := data["brief"].(string); ok {
		serviceInfo.Brief = brief
	} else if desc, ok := data["description"].(string); ok {
		serviceInfo.Brief = desc
	}
	if typeId, ok := data["typeId"].(float64); ok {
		serviceInfo.TypeId = int32(typeId)
	}

	// categoryIds array field (compatible with different return structures: categoryIds or categories)
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

	// typeName: if not available, try mapping from "type"
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

// convertToCategoryResponse converts qm API response to pb CategoryResponse structure
func (c *Client) convertToCategoryResponse(response map[string]interface{}) *market.CategoryResponse {
	// Create response structure
	categoryResponse := &market.CategoryResponse{
		List: []*market.CategoryInfo{},
	}

	// Parse category data
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

// convertToCategoryInfo converts map data to pb CategoryInfo structure
func (c *Client) convertToCategoryInfo(data map[string]interface{}) *market.CategoryInfo {
	categoryInfo := &market.CategoryInfo{}

	// Convert ID
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

	// Convert name
	if name, ok := data["name"]; ok {
		if nameStr, ok := name.(string); ok {
			categoryInfo.Name = nameStr
		}
	}

	return categoryInfo
}

// convertToDetailResponse converts qm API response to pb DetailResponse structure
func (c *Client) convertToDetailResponse(response map[string]interface{}) *market.DetailResponse {
	pbResponse := &market.DetailResponse{}

	// Prioritize parsing new structure: {"code":200,"message":"Operation successful","data":{...}}
	if data, ok := response["data"].(map[string]interface{}); ok {
		pbResponse.Service = c.convertToServiceInfo(data)
		return pbResponse
	}

	// Compatible with old structure: convert service details
	if serviceData, ok := response["service"].(map[string]interface{}); ok {
		pbResponse.Service = c.convertToServiceInfo(serviceData)
	} else {
		// If response is directly service data
		pbResponse.Service = c.convertToServiceInfo(response)
	}

	return pbResponse
}

// convertToServiceInfo converts map data to pb ServiceInfo structure
func (c *Client) convertToServiceInfo(data map[string]interface{}) *market.ServiceInfo {
	serviceInfo := &market.ServiceInfo{}

	// Basic field conversion - based on ServiceInfo definition in market.proto
	if id, ok := data["id"].(float64); ok {
		serviceInfo.Id = int32(id)
	}
	if name, ok := data["name"].(string); ok {
		serviceInfo.Name = name
	}

	// categoryIds array field
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
	// status field supports string and numeric types
	if status, ok := data["status"].(string); ok {
		serviceInfo.Status = status
	} else if status, ok := data["status"].(float64); ok {
		// If it is a numeric type, convert to string
		serviceInfo.Status = fmt.Sprintf("%.0f", status)
	}
	if subscribeStatus, ok := data["subscribeStatus"].(string); ok {
		serviceInfo.SubscribeStatus = subscribeStatus
	}
	if categoryIdsJson, ok := data["categoryIdsJson"].(string); ok {
		serviceInfo.CategoryIdsJson = categoryIdsJson
	}

	// tools array field conversion
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

				// params array field conversion
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

// createHTTPRequest creates an HTTP request
func createHTTPRequest(method, url string, data map[string]interface{}) (*http.Request, error) {
	var body io.Reader

	if method != "GET" && data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize request data: %v", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	return req, nil
}
