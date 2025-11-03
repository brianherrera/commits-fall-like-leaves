package api

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/brianherrera/commits-fall-like-leaves/internal/service/haiku"
	"github.com/gin-gonic/gin"
)

type HaikuService interface {
	CreateHaiku(ctx context.Context, request haiku.HaikuCommitRequest) (haiku.HaikuCommitResponse, error)
}

type HaikuAPI struct {
	haikuService HaikuService
}

func NewHaikuAPI(haikuService HaikuService) *HaikuAPI {
	return &HaikuAPI{
		haikuService: haikuService,
	}
}

func NewDefaultHaikuAPI(cfg aws.Config) *HaikuAPI {
	return NewHaikuAPI(haiku.NewDefaultHaikuService(cfg))
}

func (api *HaikuAPI) SetupMiddleware(router *gin.Engine) {
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
}

func (api *HaikuAPI) SetupRoutes(router *gin.Engine) {
	router.POST("/haiku", api.postHaiku)
}
