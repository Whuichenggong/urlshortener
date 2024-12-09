package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Whuichenggong/urlshortener/urlshortener/internal/model"
	"github.com/Whuichenggong/urlshortener/urlshortener/internal/repo"
)

// 生成短url
type ShortCodeGenerator interface {
	GenerateShortCode() string
}

// 只要实现了接口这个方法 随便执行插拔
type Cache interface {
	SetURL(ctx context.Context, url repo.Url) error
	GetURL(ctx context.Context, shortCode string) (*repo.Url, error)
	DeleteURL(ctx context.Context, code string) error
}

type UrlService struct {
	querier            repo.Querier
	shortCodeGenerator ShortCodeGenerator
	defaultDuration    time.Duration //时间有效期
	cache              Cache
	baseURL            string // "http://localhost:8080"
}

func NewURLService(db *sql.DB, shortCodeGenerator ShortCodeGenerator, duration time.Duration, cache Cache, baseURL string) *UrlService {
	return &UrlService{
		querier:            repo.New(db),
		shortCodeGenerator: shortCodeGenerator,
		defaultDuration:    duration,
		cache:              cache,
		baseURL:            baseURL,
	}
}

func (s *UrlService) DeleteURL(ctx context.Context, shortCode string) error {
	result, err := s.querier.DeleteUrlByShortCode(ctx, shortCode)
	if err != nil {
		log.Printf("删除失败,数据库中没有这个短url: %v", err)
	}
	if result == 0 {
		return errors.New("数据库中没有这个短url")
	}

	//删除缓存
	s.cache.DeleteURL(ctx, shortCode)

	return nil

}

// 有了接口s *UrlService 只需要这一个实例实现接口方法
// 插入数据库 存入Redis 为什么用指针 因为指针返回nil 而如果不使用指针 则会返回 model.CreateURLResponse{}
func (s *UrlService) CreateURL(ctx context.Context, req model.CreateURLRequest) (*model.CreateURLResponse, error) {

	var shortCode string
	var isCustom bool
	var expiredAt time.Time
	//sql语句写错了导致 生成的参数少了一个到时候再看看sql
	if req.CustomCode != "" {
		//从数据库中查询custom是否存在
		isAvailabel, err := s.querier.IsShortCodeAvailable(ctx, req.CustomCode)
		if err != nil {
			return nil, err
		}
		//用户传进来的别名已经在数据库中存在
		if !isAvailabel {
			return nil, fmt.Errorf("别名存在了")

		}
		//别名可以得到 把shortCode 设置为用户传进来的
		shortCode = req.CustomCode
		isCustom = true

		//名字不存在 生成一个URL
	} else {
		code, err := s.getShortCode(ctx, 0)
		if err != nil {
			return nil, err
		}
		shortCode = code
	}

	if req.Duration == nil {
		expiredAt = time.Now().Add(s.defaultDuration)
	} else {
		expiredAt = time.Now().Add(time.Hour * time.Duration(*req.Duration))
	}
	//插入数据库 把整个信息返回
	url, err := s.querier.CreateURL(ctx, repo.CreateURLParams{
		OriginalUrl: req.OriginnalURL,
		ShortCode:   shortCode,
		ExpiredAt:   expiredAt,
		IsCustom:    isCustom,
	})
	if err != nil {
		return nil, err
	}

	//存入缓存redis
	if err := s.cache.SetURL(ctx, url); err != nil {
		return nil, err
	}

	// 返回响应重新
	return &model.CreateURLResponse{
		ShortURL: s.baseURL + "/" + url.ShortCode,
		ExpireAt: url.ExpiredAt,
	}, nil
}

// 先访问缓存查询结果如果有直接返回
func (s *UrlService) GetURL(ctx context.Context, shortCode string) (string, error) {
	//先访问cache
	url, err := s.cache.GetURL(ctx, shortCode)
	if err != nil {
		return "", err
	} //说明去除了URL
	if url != nil {
		return url.OriginalUrl, nil
	}

	//访问数据库
	url2, err := s.querier.GetUrlByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	//存入缓存
	if err := s.cache.SetURL(ctx, url2); err != nil {
		return "", err
	}
	return url2.OriginalUrl, nil
}

// 对数据库进行操作传入Context 只在包内可用
func (s *UrlService) getShortCode(ctx context.Context, n int) (string, error) {
	if n > 5 {
		return "", errors.New("重试过多")
	}
	shortCode := s.shortCodeGenerator.GenerateShortCode()

	isAvailable, err := s.querier.IsShortCodeAvailable(ctx, shortCode)
	if err != nil {
		return "", err
	}
	//数据库没有可以插进去
	if isAvailable {
		return shortCode, nil
	}
	//递归调用n+1次 重复机制就是这样
	return s.getShortCode(ctx, n+1)
}
