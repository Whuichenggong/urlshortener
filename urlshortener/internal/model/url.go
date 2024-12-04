package model

import "time"

// json的序列化和反序列化
type CreateURLRequest struct {
	OriginnalURL string `json:"original_url" binding:"required,url"`
	CustomCode   string `json:"custom_code,omitempty" binding:"omitempty,min=4,max=10,alphanum"`
	Duration     *int   `json:"duration,omitempty" binding:"omitempty,min=1,max=100"` //nil 0 区别
}

// 这是我们服务器相应的数据 不需要验证
type CreateURLResponse struct {
	ShortURL string    `json:"shout_url"`
	ExpireAt time.Time `json:"expire_at"`
}
