package application

import (
	"database/sql"
	"fmt"
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
	cfg, err := config.LodeConfig(filePath)
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
	g := gin.New()
	a.g = g

	g.POST("api/url", a.urlHandler.CreateURL)
	g.GET("/:code", a.urlHandler.RedirectURL)
	a.g = g
	return nil
}

func (a *Application) RunServer() {
	err := a.g.Run(":8080")
	if err != nil {
		panic(err)
	}
}
