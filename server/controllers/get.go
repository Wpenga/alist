package controllers

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"strings"
)

// handle get request
// 因为下载地址有时效，所以去掉了文件请求和直链的缓存
func Get(c *gin.Context) {
	var get alidrive.GetReq
	if err := c.ShouldBindJSON(&get); err != nil {
		c.JSON(200, MetaResponse(400,"Bad Request"))
		return
	}
	log.Debugf("get:%+v",get)
	// cache
	cacheKey:=fmt.Sprintf("%s-%s","g",get.FileId)
	if conf.Conf.Cache.Enable {
		file,exist:=conf.Cache.Get(cacheKey)
		if exist {
			log.Debugf("使用了缓存:%s",cacheKey)
			c.JSON(200,DataResponse(file))
			return
		}
	}
	file,err:=alidrive.GetFile(get.FileId)
	if err !=nil {
		c.JSON(200, MetaResponse(500,err.Error()))
		return
	}
	paths,err:=alidrive.GetPaths(get.FileId)
	if err!=nil {
		c.JSON(200, MetaResponse(500,err.Error()))
		return
	}
	file.Paths=*paths
	download,err:=alidrive.GetDownLoadUrl(get.FileId)
	if err!=nil {
		c.JSON(200, MetaResponse(500,err.Error()))
		return
	}
	file.DownloadUrl=download.Url
	if conf.Conf.Cache.Enable {
		conf.Cache.Set(cacheKey,file,cache.DefaultExpiration)
	}
	c.JSON(200, DataResponse(file))
}

func Down(c *gin.Context) {
	fileIdParam:=c.Param("file_id")
	log.Debugf("down:%s",fileIdParam)
	fileId:=strings.Split(fileIdParam,"/")[1]
	cacheKey:=fmt.Sprintf("%s-%s","d",fileId)
	if conf.Conf.Cache.Enable {
		downloadUrl,exist:=conf.Cache.Get(cacheKey)
		if exist {
			log.Debugf("使用了缓存:%s",cacheKey)
			message := "<script>location.href='" + downloadUrl.(string) + "';</script>"
			c.String(http.StatusOK, message)
			return
		}
	}
	file,err:=alidrive.GetDownLoadUrl(fileId)
	if err != nil {
		c.JSON(200, MetaResponse(500,err.Error()))
		return
	}
	if conf.Conf.Cache.Enable {
		conf.Cache.Set(cacheKey,file.Url,cache.DefaultExpiration)
	}
	message := "<script>location.href='" + file.Url + "';</script>"
	c.String(http.StatusOK, message)
	return
}
