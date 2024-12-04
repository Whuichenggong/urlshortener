package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Whuichenggong/urlshortener/urlshortener/internal/model"
	"github.com/Whuichenggong/urlshortener/urlshortener/internal/repo"
	"time"
)

// 生成短url
type ShortCodeGenerator interface {
	GenerateIdShortCode() string
}

// 只要实现了接口这个方法 随便执行插拔
type Cache interface {
	SetURL(ctx context.Context, url repo.Url) error
}

type UrlService struct {
	querier            repo.Querier
	shortCodeGenerator ShortCodeGenerator
	defaultDuration    time.Duration
	cache              Cache
	baseURL            string // "http://localhost:8080"
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
	//插入数据库
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

	// 返回响应
	return &model.CreateURLResponse{
		ShortURL: s.baseURL + "/" + url.ShortCode,
		ExpireAt: url.ExpiredAt,
	}, nil
}

// 对数据库进行操作传入Context
func (s *UrlService) getShortCode(ctx context.Context, n int) (string, error) {
	if n > 5 {
		return "", errors.New("重试过多")
	}
	shortCode := s.shortCodeGenerator.GenerateIdShortCode()

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
