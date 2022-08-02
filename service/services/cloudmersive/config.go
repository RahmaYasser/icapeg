package cloudmersive

import (
	"icapeg/readValues"
	general_functions "icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"sync"
	"time"
)

var doOnce sync.Once
var cloudMersiveConfig *CloudMersive

type CloudMersive struct {
	httpMsg                     *utils.HttpMsg
	serviceName                 string
	methodName                  string
	allowExecutables            bool
	allowInvalidFiles           bool
	allowScripts                bool
	allowPasswordProtectedFiles bool
	allowMacros                 bool
	allowXmlExternalEntities    bool
	maxFileSize                 int
	bypassExts                  []string
	processExts                 []string
	rejectExts                  []string
	ScanEndPoint                string
	Timeout                     time.Duration
	APIKey                      string
	FailThreshold               int
	policy                      string
	returnOrigIfMaxSizeExc      bool
	return400IfFileExtRejected  bool
	generalFunc                 *general_functions.GeneralFunc
}

func InitCloudMersiveConfig(serviceName string) {
	doOnce.Do(func() {
		cloudMersiveConfig = &CloudMersive{
			maxFileSize:                 readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                  readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                 readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                  readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			ScanEndPoint:                readValues.ReadValuesString(serviceName + ".scan_endpoint"),
			Timeout:                     readValues.ReadValuesDuration(serviceName + ".timeout"),
			APIKey:                      readValues.ReadValuesString(serviceName + ".api_key"),
			FailThreshold:               readValues.ReadValuesInt(serviceName + ".fail_threshold"),
			policy:                      readValues.ReadValuesString(serviceName + ".policy"),
			returnOrigIfMaxSizeExc:      readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected:  readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
			allowScripts:                readValues.ReadValuesBool(serviceName + ".allow_scripts"),
			allowExecutables:            readValues.ReadValuesBool(serviceName + ".allow_executables"),
			allowMacros:                 readValues.ReadValuesBool(serviceName + ".allow_macros"),
			allowInvalidFiles:           readValues.ReadValuesBool(serviceName + ".allow_invalid_files"),
			allowXmlExternalEntities:    readValues.ReadValuesBool(serviceName + ".allow_xml_external_ntities"),
			allowPasswordProtectedFiles: readValues.ReadValuesBool(serviceName + ".allow_password_protected_files"),
		}
	})
}

func NewCloudMersiveService(serviceName, methodName string, httpMsg *utils.HttpMsg) *CloudMersive {
	return &CloudMersive{
		httpMsg:                     httpMsg,
		serviceName:                 serviceName,
		methodName:                  methodName,
		allowExecutables:            cloudMersiveConfig.allowExecutables,
		allowXmlExternalEntities:    cloudMersiveConfig.allowXmlExternalEntities,
		allowMacros:                 cloudMersiveConfig.allowMacros,
		allowScripts:                cloudMersiveConfig.allowScripts,
		allowInvalidFiles:           cloudMersiveConfig.allowInvalidFiles,
		allowPasswordProtectedFiles: cloudMersiveConfig.allowPasswordProtectedFiles,
		maxFileSize:                 cloudMersiveConfig.maxFileSize,
		bypassExts:                  cloudMersiveConfig.bypassExts,
		processExts:                 cloudMersiveConfig.processExts,
		rejectExts:                  cloudMersiveConfig.rejectExts,
		ScanEndPoint:                cloudMersiveConfig.ScanEndPoint,
		Timeout:                     cloudMersiveConfig.Timeout * time.Second,
		APIKey:                      cloudMersiveConfig.APIKey,
		FailThreshold:               cloudMersiveConfig.FailThreshold,
		policy:                      cloudMersiveConfig.policy,
		returnOrigIfMaxSizeExc:      cloudMersiveConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected:  cloudMersiveConfig.return400IfFileExtRejected,
		generalFunc:                 general_functions.NewGeneralFunc(httpMsg),
	}
}
