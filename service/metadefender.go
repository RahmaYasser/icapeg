package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"icapeg/dtos"
	"icapeg/readValues"
	"icapeg/transformers"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// scan_endpoint = "/file"
// report_endpoint = "/file"

// MetaDefender represents the informations regarding the MetaDefender service
type MetaDefender struct {
	BaseURL              string
	Timeout              time.Duration
	APIKey               string
	ScanEndpoint         string
	ReportEndpoint       string
	FailThreshold        int
	statusCheckInterval  time.Duration
	statusCheckTimeout   time.Duration
	badFileStatus        []string
	okFileStatus         []string
	statusEndPointExists bool
	respSupported        bool
	reqSupported         bool
}

// NewMetaDefenderService returns a new populated instance of the metadefender service
func NewMetaDefenderService(serviceName string) Service {
	return &MetaDefender{
		BaseURL:              readValues.ReadValuesString(serviceName + "base_url"),
		Timeout:              readValues.ReadValuesDuration(serviceName+"timeout") * time.Second,
		APIKey:               readValues.ReadValuesString(serviceName + "api_key"),
		ScanEndpoint:         readValues.ReadValuesString(serviceName + "scan_endpoint"),
		ReportEndpoint:       readValues.ReadValuesString(serviceName + "report_endpoint"),
		FailThreshold:        viper.GetInt(serviceName + "fail_threshold"),
		statusCheckInterval:  readValues.ReadValuesDuration(serviceName+"status_check_interval") * time.Second,
		statusCheckTimeout:   readValues.ReadValuesDuration(serviceName+"status_check_timeout") * time.Second,
		badFileStatus:        readValues.ReadValuesSlice(serviceName + "bad_file_status"),
		okFileStatus:         readValues.ReadValuesSlice(serviceName + "ok_file_status"),
		statusEndPointExists: false,
		respSupported:        readValues.ReadValuesBool(serviceName + ".resp_mode"),
		reqSupported:         readValues.ReadValuesBool(serviceName + ".req_mode"),
	}
}

// SubmitFile calls the submission api for metadefender
func (m *MetaDefender) SubmitFile(f *bytes.Buffer, filename string) (*dtos.SubmitResponse, error) {

	urlStr := m.BaseURL + m.ScanEndpoint

	// bodyBuf := &bytes.Buffer{}
	//
	// bodyWriter := multipart.NewWriter(bodyBuf)
	//
	// bodyWriter.WriteField("apikey", m.APIKey)
	//
	// part, err := bodyWriter.CreateFormFile("file", filename)
	//
	// if err != nil {
	// 	return nil, err
	// }
	//
	// io.Copy(part, bytes.NewReader(f.Bytes()))
	// if err := bodyWriter.Close(); err != nil {
	// 	errorLogger.LogToFile("failed to close writer", err.Error())
	// 	return nil, err
	// }
	//

	req, err := http.NewRequest(http.MethodPost, urlStr, f)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Apikey", m.APIKey)
	req.Header.Set("Content-Type", "application/octet-stream")
	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		errorLogger.LogToFile("service: metadefender: failed to do request:", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bdy))
	}

	scanResp := dtos.MetaDefenderScanFileResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		return nil, err
	}

	return transformers.TransformMetaDefenderToSubmitResponse(&scanResp), nil
}

// GetSampleFileInfo returns the submitted sample file's info
func (m *MetaDefender) GetSampleFileInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {

	urlStr := m.BaseURL + fmt.Sprintf(m.ReportEndpoint+"/"+sampleID)
	//urlStr := v.BaseURL + fmt.Sprintf(readValues.ReadValuesString("metadefender.report_endpoint"), readValues.ReadValuesString("metadefender.api_key"), sampleID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("apikey", m.APIKey)
	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		bdyStr := ""
		if string(bdy) == "" {
			bdyStr = fmt.Sprintf("Status code received:%d with no body", resp.StatusCode)
		} else {
			bdyStr = string(bdy)
		}
		return nil, errors.New(bdyStr)
	}

	sampleResp := dtos.MetaDefenderReportResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		return nil, err
	}

	fm := dtos.FileMetaInfo{}

	if len(filemetas) > 0 {
		fm = filemetas[0]
	}

	return transformers.TransformMetaDefenderToSampleInfo(&sampleResp, fm, m.FailThreshold), nil

}

// GetSubmissionStatus returns the submission status of a submitted sample
func (m *MetaDefender) GetSubmissionStatus(submissionID string) (*dtos.SubmissionStatusResponse, error) {

	urlStr := m.BaseURL + fmt.Sprintf(m.ReportEndpoint+"/"+submissionID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("apikey", readValues.ReadValuesString("metadefender.api_key"))
	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent {
			return nil, errors.New("No content receive from metadefender on status check, maybe request quota expired")
		}
		bdy, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bdy))
	}

	sampleResp := dtos.MetaDefenderReportResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		return nil, err
	}

	return transformers.TransformMetaDefenderToSubmissionStatusResponse(&sampleResp), nil
}

// SubmitURL calls the submission api for metadefender
func (m *MetaDefender) SubmitURL(fileURL, filename string) (*dtos.SubmitResponse, error) {
	return nil, nil
}

// GetSampleURLInfo returns the submitted sample url's info
func (m *MetaDefender) GetSampleURLInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {
	return nil, nil
}

// GetStatusCheckInterval returns the status_check_interval duration of the service
func (m *MetaDefender) GetStatusCheckInterval() time.Duration {
	return m.statusCheckInterval
}

// GetStatusCheckTimeout returns the status_check_timeout duraion of the service
func (m *MetaDefender) GetStatusCheckTimeout() time.Duration {
	return m.statusCheckTimeout
}

// GetBadFileStatus returns the bad_file_status slice of the service
func (m *MetaDefender) GetBadFileStatus() []string {
	return m.badFileStatus
}

// GetOkFileStatus returns the ok_file_status slice of the service
func (m *MetaDefender) GetOkFileStatus() []string {
	return m.okFileStatus
}

// StatusEndpointExists returns the status_endpoint_exists boolean value of the service
func (m *MetaDefender) StatusEndpointExists() bool {
	return m.statusEndPointExists
}

// RespSupported returns the respSupported field of the service
func (m *MetaDefender) RespSupported() bool {
	return m.respSupported
}

// ReqSupported returns the reqSupported field of the service
func (m *MetaDefender) ReqSupported() bool {
	return m.reqSupported
}
func (g *MetaDefender) SendFileApi(f *bytes.Buffer, filename string) (*http.Response, error) {

	urlStr := g.BaseURL + g.ScanEndpoint

	bodyBuf := &bytes.Buffer{}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
	//fmt.Println(req)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		errorLogger.LogToFile("service: Glasswall: failed to do request:", err.Error())
		return nil, err
	}
	return resp, err

}
