package grayimages

import (
	"fmt"
	"icapeg/config"
	"icapeg/readValues"
	general_functions "icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"sync"
)

var doOnce sync.Once
var grayImagesConfig *GrayImages

type GrayImages struct {
	httpMsg                    *utils.HttpMsg
	serviceName                string
	methodName                 string
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []config.Extension
	maxFileSize                int
	FailThreshold              int
	policy                     string
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
}

func InitGrayImages(serviceName string) {
	doOnce.Do(func() {
		grayImagesConfig = &GrayImages{
			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
			FailThreshold:              readValues.ReadValuesInt(serviceName + ".fail_threshold"),
			policy:                     readValues.ReadValuesString(serviceName + ".policy"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
		}
		process := config.Extension{Name: "process", Exts: grayImagesConfig.processExts}
		reject := config.Extension{Name: "reject", Exts: grayImagesConfig.rejectExts}
		bypass := config.Extension{Name: "bypass", Exts: grayImagesConfig.bypassExts}
		extArrs := make([]config.Extension, 3)
		ind := 0
		if len(process.Exts) == 1 && process.Exts[0] == "*" {
			extArrs[2] = process
		} else {
			extArrs[ind] = process
			ind++
		}
		if len(reject.Exts) == 1 && reject.Exts[0] == "*" {
			extArrs[2] = reject
		} else {
			extArrs[ind] = reject
			ind++
		}
		if len(bypass.Exts) == 1 && bypass.Exts[0] == "*" {
			extArrs[2] = bypass
		} else {
			extArrs[ind] = bypass
			ind++
		}
		fmt.Println(extArrs)
		grayImagesConfig.extArrs = extArrs
	})
}

func NewGrayImagesService(serviceName, methodName string, httpMsg *utils.HttpMsg) *GrayImages {
	return &GrayImages{
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		maxFileSize:                grayImagesConfig.maxFileSize,
		FailThreshold:              grayImagesConfig.FailThreshold,
		policy:                     grayImagesConfig.policy,
		returnOrigIfMaxSizeExc:     grayImagesConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: grayImagesConfig.return400IfFileExtRejected,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg),
		bypassExts:                 grayImagesConfig.bypassExts,
		processExts:                grayImagesConfig.processExts,
		rejectExts:                 grayImagesConfig.rejectExts,
		extArrs:                    grayImagesConfig.extArrs,
	}
}
