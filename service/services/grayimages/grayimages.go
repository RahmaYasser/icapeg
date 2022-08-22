package grayimages

import (
	"bytes"
	"errors"
	"fmt"
	"icapeg/utils"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strconv"
	"time"
)

const GrayImagesIdentifier = "GRAYIMAGES ID"

func (g *GrayImages) Processing(partial bool) (int, interface{}, map[string]string) {
	serviceHeaders := make(map[string]string)
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		return utils.Continue, nil, nil
	}

	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := g.generalFunc.CopyingFileToTheBuffer(g.methodName)
	if err != nil {
		log.Println("30")
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
	}

	// check if file is compressed
	isGzip = g.generalFunc.IsBodyGzipCompressed(g.methodName)
	//if it's compressed, we decompress it to send it to Glasswall service
	if isGzip {
		if file, err = g.generalFunc.DecompressGzipBody(file); err != nil {
			return utils.InternalServerErrStatusCodeStr, nil, nil
		}
	}
	//getting the extension of the file
	contentType := g.httpMsg.Response.Header["Content-Type"]
	var fileName string
	if g.methodName == utils.ICAPModeReq {
		fileName = utils.GetFileName(g.httpMsg.Request)
	} else {
		fileName = utils.GetFileName(g.httpMsg.Response)
	}
	fileExtension := utils.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	isProcess, icapStatus, httpMsg := g.generalFunc.CheckTheExtension(fileExtension, g.extArrs,
		g.processExts, g.rejectExts, g.bypassExts, g.return400IfFileExtRejected, isGzip,
		g.serviceName, g.methodName, GrayImagesIdentifier, g.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		return icapStatus, httpMsg, serviceHeaders
	}

}

func (g *GrayImages) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}

func (g *GrayImages) ConvertImgToGrayScale(imgExtension string, file *bytes.Buffer) (*os.File, error) {
	log.Println(imgExtension)
	log.Println(g.methodName)
	// convert HTTP file to image object
	img, err := g.generalFunc.GetDecodedImage(file)
	if err != nil {
		return nil, err
	}
	// convert the image to grayscale
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}
	// Working with grayscale image, convert to png
	if imgExtension == "png" {
		// create new temporarily png file
		newImg, err := os.CreateTemp(g.imagesDir, "*.png")
		log.Println(newImg.Name())
		if err != nil {
			return nil, err
		}
		// encode gray image data and save it into the created png file
		if err = png.Encode(newImg, grayImg); err != nil {
			return nil, err
		}
		// return the png file after converting it to gray image
		return newImg, nil
	} else if imgExtension == "jpeg" || imgExtension == "jpg" {
		// Working with grayscale image, convert to png
		pattern := fmt.Sprintf("*.%s", imgExtension)
		// create new temporarily (jpeg or jpg) file
		newImg, err := os.CreateTemp(g.imagesDir, pattern)
		if err != nil {
			return nil, err
		}
		// encode gray image data and save it into the created jpeg/jpg file
		if err = jpeg.Encode(newImg, grayImg, nil); err != nil {
			return nil, err
		}
		return newImg, nil
	} else {
		// if file isn't png or jpeg/jpg, return error
		return nil, errors.New("file is not a supported image")
	}
}
