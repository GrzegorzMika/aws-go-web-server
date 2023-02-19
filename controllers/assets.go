package controllers

import (
	"aws-web-server/models"
	"golang.org/x/net/context"
)

type AssetController struct {
	appContext context.Context
	sb         models.S3Bucket
}

func NewAssetController(ctx context.Context, sb *models.S3Bucket) *AssetController {
	return &AssetController{
		appContext: ctx,
		sb:         *sb,
	}
}
