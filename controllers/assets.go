package controllers

import "aws-web-server/models"

type AssetController struct {
	sb models.S3Bucket
}

func NewAssetController(sb *models.S3Bucket) *AssetController {
	return &AssetController{
		sb: *sb,
	}
}
