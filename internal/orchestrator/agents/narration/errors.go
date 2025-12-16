package narration

import "errors"

var (
	// ErrUnsupportedRequestType 表示不支援的請求類型
	ErrUnsupportedRequestType = errors.New("unsupported request type for NarrationAgent")

	// ErrInvalidSkeletonResponse 表示 Skeleton 響應無效
	ErrInvalidSkeletonResponse = errors.New("invalid skeleton response from LLM")

	// ErrMissingRequiredField 表示缺少必填字段
	ErrMissingRequiredField = errors.New("missing required field in response")

	// ErrJSONParseFailed 表示 JSON 解析失敗
	ErrJSONParseFailed = errors.New("failed to parse JSON response")
)
