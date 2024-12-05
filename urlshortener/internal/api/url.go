package api

import (
	"context"
	"fmt"
	"github.com/Whuichenggong/urlshortener/urlshortener/internal/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

//业务功能

// 抽象出来的接口

// 所有符合这个接口的实现类（例如内存存储、数据库存储）都可以提供 CreateURL 和 GetOriginalURL 方法。
// 然后我们可以根据不同的需求实现 URLSERVICE 接口。例如：
type URLService interface {
	//创建短URL
	CreateURL(ctx context.Context, req model.CreateURLRequest) (*model.CreateURLResponse, error)
	//根据短 URL 重定向原始URL
	GetURL(ctx context.Context, shortCode string) (string, error)
}

//根据需求开发商编写api
//第一个需求 POST /api/url original_url, custom_code duration => 短URL 过期时间
//第二个需求 GET /:code 把短链接重定向到长URL

// 为什么要传入接口
type URLHandler struct {
	urlService URLService
}

func NewURLHandler(urlService URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

// h URLHandler 实现了接口
func (h *URLHandler) CreateURL(ctx *gin.Context) {
	//数据提取
	var req model.CreateURLRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, fmt.Sprintf("error: %s", err.Error()))
	}
	//验证数据格式 校验器 判断传进来的json
	// 如果验证成功，可以继续执行其他操作
	ctx.JSON(http.StatusOK, gin.H{"message": "Account created successfully!"})
	//调用业务函数
	resq, err := h.urlService.CreateURL(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
	}

	//响应
	ctx.JSON(http.StatusOK, resq)
}

// GET /:code 把短URL重定向到长URL

func (h *URLHandler) RedirectURL(ctx *gin.Context) {
	// 获取短码
	shortCode := ctx.Param("shortCode")

	// 调用业务函数获取原始 URL
	originalURL, err := h.urlService.GetURL(ctx.Request.Context(), shortCode)
	if err != nil {
		// 如果获取 URL 失败，返回 500 错误
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return // 错误发生后应该立即返回
	}

	if originalURL == "" {
		// 如果 URL 为空，返回 404 错误
		ctx.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// 执行重定向
	ctx.Redirect(http.StatusFound, originalURL) // 302 临时重定向
}
