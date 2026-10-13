package service

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"chronchat/utils"
)

// Upload
// @Summary 上传文件(图片/语音/视频)
// @Tags 附件
// @Param file formData file false "文件"
// @Param filetype formData string false "文件类型(语音/视频二进制流上传时前端指定 如.mp3)"
// @Success 200 {string} json{"code","msg","data"}
// @Router /attach/upload [post]
func Upload(c *gin.Context) {
	w := c.Writer
	req := c.Request
	srcFile, head, err := req.FormFile("file")
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	defer srcFile.Close()
	//优先使用前端指定的文件类型(语音/视频为二进制流 无文件名后缀)
	suffix := ".png"
	if fileType := req.FormValue("filetype"); fileType != "" {
		suffix = fileType
	} else {
		ofilName := head.Filename
		tem := strings.Split(ofilName, ".")
		if len(tem) > 1 {
			suffix = "." + tem[len(tem)-1]
		}
	}
	//目录不存在时自动创建
	err = os.MkdirAll("./asset/upload", os.ModePerm)
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	fileName := fmt.Sprintf("%d%04d%s", time.Now().Unix(), rand.Int31(), suffix)
	dstFile, err := os.Create("./asset/upload/" + fileName)
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		utils.RespFail(w, err.Error())
		return
	}
	url := "/asset/upload/" + fileName
	utils.RespOK(w, url, "上传成功")
}
