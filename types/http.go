package types

type BaseResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GetResponse struct {
	BaseResponse
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

type ListKeysResponse struct {
	BaseResponse
	Keys []string `json:"keys"`
}

type StatsResponse struct {
	BaseResponse
	TotalKeys int   `json:"total_keys"`
	TotalSize int64 `json:"total_size"`
	Segments  int   `json:"segments"`
}
