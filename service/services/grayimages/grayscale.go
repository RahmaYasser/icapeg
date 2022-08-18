package grayimages

import (
	"bytes"
	"fmt"
	"icapeg/utils"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func (g *GrayImages) Processing(partial bool) (int, interface{}, map[string]string) {
	//TODO implement me

	// no need to scan part of the file, this service needs all the file at one time
	if partial {
		return utils.Continue, nil, nil
	}

	// ICAP response headers
	serviceHeaders := make(map[string]string)
	//extracting the file from http message
	file, reqContentType, err := g.generalFunc.CopyingFileToTheBuffer(g.methodName)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
	}

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
	fmt.Println(fileExtension)
	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	fmt.Println(g.extArrs)
	for i := 0; i < 3; i++ {
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
				errPage := g.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, g.serviceName, "NO ID", g.httpMsg.Request.RequestURI)
				g.httpMsg.Response = g.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
				g.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
				return utils.OkStatusCodeStr, g.httpMsg.Response, serviceHeaders
			}
		} else if g.extArrs[i].Name == "bypass" {
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
					return utils.NoModificationStatusCodeStr, msg, serviceHeaders
				}
				return utils.NoModificationStatusCodeStr, nil, serviceHeaders
			}
		}
	}

	// TODO check if file is image
	//grayImg, err := g.ConvertImgToGrayScale(fileExtension)
	if err != nil {
		scannedFile := g.generalFunc.PreparingFileAfterScanning(file.Bytes(), reqContentType, g.methodName)
		return utils.OkStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
	}

	// getting response body
	fmt.Println("return file")
	scannedFile := g.generalFunc.PreparingFileAfterScanning(file.Bytes(), reqContentType, g.methodName)
	return utils.OkStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders

	/*b, err := ioutil.ReadAll(grayImg)
	scannedFile := g.generalFunc.PreparingFileAfterScanning(b, reqContentType, g.methodName)
	return utils.OkStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
	*/
}

func (g *GrayImages) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}

func (g *GrayImages) ConvertImgToGrayScale(imgExtension string) (*os.File, error) {

	/*img, _, err := image.Decode(resp.Body)
	if err != nil {
		// handle error
		log.Println(err)
		return nil, err
	}
	log.Printf("Image type: %T", img)*/

	// Converting image to grayscale
	img, err := g.generalFunc.GetDecodedImage(g.methodName)
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
		if err := png.Encode(newImg, grayImg); err != nil {
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
		if err := jpeg.Encode(newImg, grayImg, nil); err != nil {
			return nil, err
		}
		return newImg, nil
	}
	return nil, err
}
