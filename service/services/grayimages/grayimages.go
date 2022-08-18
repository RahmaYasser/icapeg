package grayimages

import (
	"bytes"
	"errors"
	"fmt"
	"icapeg/utils"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Processing is a func used for to processing the http message
func (g *GrayImages) Processing(partial bool) (int, interface{}, map[string]string) {
	log.Println("processing")
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
	//getting the extension of the file
	var contentType []string
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	var fileName string
	if g.methodName == utils.ICAPModeReq {
		contentType = g.httpMsg.Request.Header["Content-Type"]
		fileName = utils.GetFileName(g.httpMsg.Request)
	} else {
		contentType = g.httpMsg.Response.Header["Content-Type"]
		fileName = utils.GetFileName(g.httpMsg.Response)
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	fileExtension := utils.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	/*	for i := 0; i < 3; i++ {
			if g.extArrs[i].Name == "process" {
				if g.generalFunc.IfFileExtIsX(fileExtension, g.processExts) {
					break
				}
			} else if g.extArrs[i].Name == "reject" {
				if g.generalFunc.IfFileExtIsX(fileExtension, g.rejectExts) {
					reason := "File rejected"
					if g.return400IfFileExtRejected {
						return utils.BadRequestStatusCodeStr, nil, serviceHeaders
					}
					errPage := g.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, g.serviceName, "ECHO ID", g.httpMsg.Request.RequestURI)
					g.httpMsg.Response = g.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
					g.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
					return utils.OkStatusCodeStr, g.httpMsg.Response, serviceHeaders
				}
			} else if g.extArrs[i].Name == "bypass" {
				log.Println("70")
				if g.generalFunc.IfFileExtIsX(fileExtension, g.bypassExts) {
					fileAfterPrep, httpMsg := g.generalFunc.IfICAPStatusIs204(g.methodName, utils.NoModificationStatusCodeStr,
						file, false, reqContentType, g.httpMsg)
					if fileAfterPrep == nil && httpMsg == nil {
						return utils.InternalServerErrStatusCodeStr, nil, nil
					}

					//returning the http message and the ICAP status code
					switch msg := httpMsg.(type) {
					case *http.Request:
						msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
						return utils.NoModificationStatusCodeStr, msg, serviceHeaders
					case *http.Response:
						msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
						log.Println("88")
						return utils.NoModificationStatusCodeStr, msg, serviceHeaders
					}
					return utils.NoModificationStatusCodeStr, nil, serviceHeaders
				}
			}
		}
	*/
	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if g.maxFileSize != 0 && g.maxFileSize < file.Len() {
		status, file, httpMsg := g.generalFunc.IfMaxFileSeizeExc(g.returnOrigIfMaxSizeExc, g.serviceName, file, g.maxFileSize)
		fileAfterPrep, httpMsg := g.generalFunc.IfStatusIs204WithFile(g.methodName, status, file, isGzip, reqContentType, httpMsg)
		if fileAfterPrep == nil && httpMsg == nil {
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			log.Println("104")
			return status, msg, nil
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			log.Println("108")
			return status, msg, nil
		}
		log.Println("111")
		return status, nil, nil
	}

	//check if the body of the http message is compressed in Gzip or not
	//isGzip = g.generalFunc.IsBodyGzipCompressed(g.methodName)
	////if it's compressed, we decompress it to send it to Glasswall service
	//if isGzip {
	//	if file, err = g.generalFunc.DecompressGzipBody(file); err != nil {
	//		fmt.Println("here")
	//		return utils.InternalServerErrStatusCodeStr, nil, nil
	//	}
	//}

	log.Printf("img processing")
	scannedFile := file.Bytes()

	//if the original file was compressed in GZIP, we will compress the scanned file in GZIP also
	//if isGzip {
	//	scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
	//	if err != nil {
	//		return utils.InternalServerErrStatusCodeStr, nil, nil
	//	}
	//}

	scale, err := g.ConvertImgToGrayScale(fileExtension)
	if err != nil {
		scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
		return utils.OkStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
	}

	//returning the scanned file if everything is ok
	scannedFile, err = os.ReadFile("./gray_images" + scale.Name()) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
	return utils.OkStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
}

func (g *GrayImages) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}

func (g *GrayImages) ConvertImgToGrayScale(imgExtension string) (*os.File, error) {
	fmt.Println(imgExtension)

	/*img, _, err := image.Decode(resp.Body)
	if err != nil {
		// handle error
		log.Println(err)
		return nil, err
	}
	log.Printf("Image type: %T", img)*/

	// Converting image to grayscale
	img, _ := g.generalFunc.GetDecodedImage(g.methodName)
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}

	// Working with grayscale image, e.g. convert to png
	if imgExtension == "png" {
		newImg, err := os.CreateTemp("./gray_images", "*.png")
		fmt.Println(newImg.Name())
		if err != nil {
			fmt.Println("err: ", err)
			return nil, err
		}
		if err = png.Encode(newImg, grayImg); err != nil {
			return nil, err
		}
		return newImg, nil
	} else if imgExtension == "jpeg" || imgExtension == "jpg" {
		pattern := fmt.Sprintf("*.%s", imgExtension)
		newImg, err := os.CreateTemp("./gray_images", pattern)
		fmt.Println(newImg.Name())
		if err != nil {
			fmt.Println("err: ", err)
			return nil, err
		}
		if err = jpeg.Encode(newImg, grayImg, nil); err != nil {
			return nil, err
		}
		return newImg, nil
	} else {
		return nil, errors.New("file is not a supported image")
	}
}
