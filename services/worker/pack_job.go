package worker

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/xEasy/baker/services/cacher"
	"github.com/xEasy/baker/services/painter"
)

func (payload *Payload) UploadPackToUpyun() (err error) {

	cacher.SetCache(payload.CacheKey, "runing")
	zipFile, err := ioutil.TempFile("tmp", "ubaker")
	defer os.Remove(zipFile.Name())

	if err != nil {
		cacher.SetCache(payload.CacheKey, "生成zip临时文件出错:"+err.Error())
		return
	}
	zipWriter := zip.NewWriter(zipFile)
	merchantQrcodeConfing := &painter.MergeImageConfig{
		Top:     payload.PackTop,
		Left:    payload.PackLeft,
		QrWidth: payload.PackQrWidth,
	}
	for index, c := range payload.PackContents {
		var img *os.File
		var err error
		if payload.NoBack {
			img, err = painter.GenQrcodeImg(c, payload.PackQrWidth)
		} else {
			img, err = painter.GenMerchantQrcode(c, payload.BackgroudFile, merchantQrcodeConfing)
		}
		if err != nil {
			continue
		}
		img, err = os.Open(img.Name())
		if err != nil {
			continue
		}
		zipFile, err := zipWriter.Create("qrcode_" + strconv.FormatInt(int64(index), 10) + ".jpg")
		io.Copy(zipFile, img)
		defer os.Remove(img.Name())
	}
	err = zipWriter.Close()
	if err != nil {
		cacher.SetCache(payload.CacheKey, "压缩文件到zip出错:"+err.Error())
		return
	}

	formResp, err := uploadToUpyun(zipFile.Name(), "ubakers/packs/{filemd5}.zip")
	if err != nil {
		fmt.Println("[UPYUN] worker packs UploadToUpyun FAIL:", err)
		cacher.SetCache(payload.CacheKey, "保存到upyun出错："+err.Error())
	} else {
		fmt.Println("[UPYUN] worker packs set CacheKey:", payload.CacheKey, formResp.Url)
		cacher.SetCache(payload.CacheKey, formResp.Url)
	}
	return
}
