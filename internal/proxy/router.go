package proxy

import "trae-proxy-go/pkg/models"

// selectBackendByModel 根据请求的模型ID选择后端API
func selectBackendByModel(config *models.Config, requestedModel string) *models.API {
	apis := config.APIs

	// 首先尝试根据模型ID精确匹配
	for i := range apis {
		if apis[i].Active && apis[i].CustomModelID == requestedModel {
			return &apis[i]
		}
	}

	// 如果没有精确匹配，使用第一个激活的API
	for i := range apis {
		if apis[i].Active {
			return &apis[i]
		}
	}

	// 如果都没有激活的，使用第一个
	if len(apis) > 0 {
		return &apis[0]
	}

	return nil
}

