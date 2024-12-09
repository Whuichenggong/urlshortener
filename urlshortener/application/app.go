package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Whuichenggong/urlshortener/urlshortener/config"
	"github.com/Whuichenggong/urlshortener/urlshortener/database"
	"github.com/Whuichenggong/urlshortener/urlshortener/internal/api"
	"github.com/Whuichenggong/urlshortener/urlshortener/internal/cache"
	"github.com/Whuichenggong/urlshortener/urlshortener/internal/service"
	"github.com/Whuichenggong/urlshortener/urlshortener/pkg/shortcode"
	"github.com/gin-gonic/gin"
)

type Application struct {
	g                  *gin.Engine
	db                 *sql.DB
	redisClient        *cache.RedisCache
	urlService         *service.UrlService
	urlHandler         *api.URLHandler
	cfg                *config.Config
	shortCodeGenerator *shortcode.ShortCode
}

func (a *Application) Init(filePath string) error {
	cfg, err := config.LoadConfig(filePath)
	if err != nil {
		return fmt.Errorf("加载配置错误: %v", err)
	}
	a.cfg = cfg
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		return err
	}
	a.db = db

	redisClient, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		return err
	}
	a.redisClient = redisClient
	a.shortCodeGenerator = shortcode.NewShortCode(cfg.ShortCode.Length)

	a.urlService = service.NewURLService(db, a.shortCodeGenerator, cfg.APP.DefaultDuration, redisClient, cfg.APP.BaseURL)

	a.urlHandler = api.NewURLHandler(a.urlService)

	// 添加中间件
	g := gin.New()
	g.Use(gin.Logger())
	g.Use(gin.Recovery())
	g.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Next()
	})

	g.POST("/api/url/", a.urlHandler.CreateURL)
	g.GET("/:shortCode/", a.urlHandler.RedirectURL)
	a.g = g
	return nil
}

// 修改 RunServer 方法，支持优雅关闭
func (a *Application) RunServer() error {
	server := &http.Server{
		Addr:         a.cfg.Server.Addr,
		Handler:      a.g,
		ReadTimeout:  a.cfg.Server.ReadTimeout,
		WriteTimeout: a.cfg.Server.WriteTimeout,
	}

	// 优雅关闭
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		a.Close() // 清理资源
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v\n", err)
		}
	}()

	return server.ListenAndServe()
}

// 修改 application/app.go
func (a *Application) Close() {
	if a.db != nil {
		a.db.Close()
	}
	if a.redisClient != nil {
		a.redisClient.Close()
	}
}
