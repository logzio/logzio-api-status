package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang/snappy"
	"github.com/jarcoal/httpmock"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getMetrics(request *http.Request) ([]map[string]interface{}, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	defer func(body io.ReadCloser) {
		if err = body.Close(); err != nil {
			panic(err)
		}

	}(request.Body)

	uncompressedBody, err := snappy.Decode(nil, body)
	if err != nil {
		return nil, err
	}

	writeRequest := &prompb.WriteRequest{}
	if err = writeRequest.Unmarshal(uncompressedBody); err != nil {
		return nil, err
	}

	metrics := make([]map[string]interface{}, 0)

	for _, timeseries := range writeRequest.Timeseries {
		metric := make(map[string]interface{})

		metric["value"] = timeseries.Samples[0].Value

		for _, label := range timeseries.Labels {
			metric[label.Name] = label.Value
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func TestNewLogzioApiStatus_Success(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	apiStatus, err := newLogzioApiStatus(context.Background())
	require.NoError(t, err)
	require.NotNil(t, apiStatus)

	assert.Equal(t, "https://example.api:1234", apiStatus.url)
	assert.Equal(t, http.MethodGet, apiStatus.method)
	assert.Equal(t, map[string]string{"Content-Type": "text/application", "Accept": "text/application"}, apiStatus.headers)
	assert.Equal(t, "test", apiStatus.body)
	assert.Equal(t, 10*time.Second, apiStatus.responseTimeout)
	assert.Empty(t, apiStatus.bearerToken)
	assert.Empty(t, apiStatus.username)
	assert.Empty(t, apiStatus.password)
	assert.Equal(t, http.StatusOK, apiStatus.expectedResponseStatusCode)
	assert.Equal(t, "success", apiStatus.expectedResponseBody)
	assert.Equal(t, "https://listener.logz.io:8053", apiStatus.logzioMetricsListener)
	assert.Equal(t, "123456789a", apiStatus.logzioMetricsToken)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoApiURL(t *testing.T) {
	err := os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoMethod(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_BadHeaders(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type:text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoApiResponseTimeout(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoApiResponseTimeoutNumber(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "apiResponseTimeout")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoApiResponseTimeoutPositiveNumber(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "0")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoExpectedStatusCode(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoExpectedStatusCodeNumber(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "expectedStatusCode")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoExpectedStatusCodeValidStatusCode(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "25")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoLogzioMetricsListener(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestNewLogzioApiStatus_NoLogzioMetricsToken(t *testing.T) {
	err := os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	_, err = newLogzioApiStatus(context.Background())
	require.Error(t, err)

	os.Clearenv()
}

func TestCreateApiHttpRequest_Success(t *testing.T) {
	apiStatus := &logzioApiStatus{
		ctx:                        context.Background(),
		logzioMetricsListener:      "https://listener.logz.io:8053",
		logzioMetricsToken:         "123456789a",
		url:                        "https://example.api:1234",
		method:                     http.MethodGet,
		headers:                    map[string]string{"Content-Type": "text/application", "Accept": "text/application"},
		body:                       "test",
		responseTimeout:            10,
		bearerToken:                "",
		username:                   "",
		password:                   "",
		expectedResponseStatusCode: http.StatusOK,
		expectedResponseBody:       "success",
	}

	request, err := apiStatus.createApiHttpRequest()
	require.NoError(t, err)
	require.NotNil(t, request)

	bodyBytes, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.NotNil(t, bodyBytes)

	defer closeResponseBody(request.Body)

	assert.Equal(t, "example.api:1234", request.Host)
	assert.Equal(t, http.MethodGet, request.Method)
	assert.Equal(t, []string{"text/application"}, request.Header["Content-Type"])
	assert.Equal(t, []string{"text/application"}, request.Header["Accept"])
	assert.Equal(t, apiStatus.body, string(bodyBytes))
}

func TestGetApiHttpResponse(t *testing.T) {
	apiStatus := &logzioApiStatus{
		ctx:                        context.Background(),
		logzioMetricsListener:      "https://listener.logz.io:8053",
		logzioMetricsToken:         "123456789a",
		url:                        "https://example.api:1234",
		method:                     http.MethodGet,
		headers:                    map[string]string{"Content-Type": "text/application", "Accept": "text/application"},
		body:                       "test",
		responseTimeout:            10,
		bearerToken:                "",
		username:                   "",
		password:                   "",
		expectedResponseStatusCode: http.StatusOK,
		expectedResponseBody:       "success",
	}

	request, err := apiStatus.createApiHttpRequest()
	require.NoError(t, err)
	require.NotNil(t, request)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, "https://example.api:1234",
		httpmock.NewStringResponder(http.StatusOK, "success"))

	response, responseTime, err := apiStatus.getApiHttpResponse(request)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, responseTime)

	bodyBytes, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotNil(t, bodyBytes)

	defer closeResponseBody(response.Body)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "success", string(bodyBytes))
}

func TestCreateController_Success(t *testing.T) {
	apiStatus := &logzioApiStatus{
		ctx:                        context.Background(),
		logzioMetricsListener:      "https://listener.logz.io:8053",
		logzioMetricsToken:         "123456789a",
		url:                        "https://example.api:1234",
		method:                     http.MethodGet,
		headers:                    map[string]string{"Content-Type": "text/application", "Accept": "text/application"},
		body:                       "test",
		responseTimeout:            10,
		bearerToken:                "",
		username:                   "",
		password:                   "",
		expectedResponseStatusCode: http.StatusOK,
		expectedResponseBody:       "success",
	}

	cont, err := apiStatus.createController()
	require.NoError(t, err)
	require.NotNil(t, cont)
}

func TestCollectMetrics_AllMetrics(t *testing.T) {
	err := os.Setenv(awsRegionEnvName, "us-east-1")
	require.NoError(t, err)

	err = os.Setenv(awsLambdaFunctionNameEnvName, "test")
	require.NoError(t, err)

	apiStatus := &logzioApiStatus{
		ctx:                        context.Background(),
		logzioMetricsListener:      "https://listener.logz.io:8053",
		logzioMetricsToken:         "123456789a",
		url:                        "https://example.api:1234",
		method:                     http.MethodGet,
		headers:                    map[string]string{"Content-Type": "text/application", "Accept": "text/application"},
		body:                       "test",
		responseTimeout:            10,
		bearerToken:                "",
		username:                   "",
		password:                   "",
		expectedResponseStatusCode: http.StatusOK,
		expectedResponseBody:       "success",
	}

	request, err := apiStatus.createApiHttpRequest()
	require.NoError(t, err)
	require.NotNil(t, request)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, "https://example.api:1234",
		httpmock.NewStringResponder(http.StatusOK, "success"))

	response, responseTime, err := apiStatus.getApiHttpResponse(request)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, responseTime)

	bodyBytes, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotNil(t, bodyBytes)

	defer closeResponseBody(response.Body)

	gaugeObservers := make([]metricRegister, 0)
	statusGaugeObserver := apiStatus.getSuccessStatusGaugeObserver(http.StatusOK)
	responseTimeGaugeObserver := apiStatus.getResponseTimeGaugeObserver(responseTime)
	responseBodyLengthGaugeObserver := apiStatus.getResponseBodyLengthGaugeObserver(len(bodyBytes))
	gaugeObservers = append(gaugeObservers, statusGaugeObserver, responseTimeGaugeObserver, responseBodyLengthGaugeObserver)

	httpmock.RegisterResponder(http.MethodPost, "https://listener.logz.io:8053",
		func(request *http.Request) (*http.Response, error) {
			metrics, err := getMetrics(request)
			require.NoError(t, err)
			require.NotNil(t, metrics)

			assert.Len(t, metrics, 3)

			for _, metric := range metrics {
				assert.Contains(t, []string{statusMetricName, responseTimeMetricName, responseBodyLengthMetricName}, metric["__name__"])

				if metric["__name__"] == statusMetricName {
					assert.Len(t, metric, 8)
					assert.Equal(t, float64(statusMetricValue), metric["value"])
					assert.Equal(t, successStatusMetricStatusLabelValue, metric[statusMetricStatusLabelName])
					assert.Equal(t, "200", metric[statusMetricResponseStatusCodeLabelName])
				} else if metric["__name__"] == responseTimeMetricName {
					assert.Len(t, metric, 7)
					assert.Equal(t, responseTime, metric["value"])
					assert.Equal(t, responseTimeMetricUnitLabelValue, metric[unitLabelName])
				} else if metric["__name__"] == responseBodyLengthMetricName {
					assert.Len(t, metric, 7)
					assert.Equal(t, float64(len(bodyBytes)), metric["value"])
					assert.Equal(t, responseBodyLengthMetricUnitLabelValue, metric[unitLabelName])
				}

				assert.Equal(t, apiStatus.url, metric[urlLabelName])
				assert.Equal(t, apiStatus.method, metric[methodLabelName])
				assert.Equal(t, "us-east-1", metric[awsRegionLabelName])
				assert.Equal(t, "test", metric[awsLambdaFunctionLabelName])
			}

			return httpmock.NewStringResponse(http.StatusOK, ""), nil
		})

	err = apiStatus.collectMetrics(gaugeObservers)
	require.NoError(t, err)

	os.Clearenv()
}

func TestRun_SuccessStatus(t *testing.T) {
	err := os.Setenv(awsRegionEnvName, "us-east-1")
	require.NoError(t, err)

	err = os.Setenv(awsLambdaFunctionNameEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, "https://example.api:1234",
		httpmock.NewStringResponder(http.StatusOK, "success"))

	httpmock.RegisterResponder(http.MethodPost, "https://listener.logz.io:8053",
		func(request *http.Request) (*http.Response, error) {
			metrics, err := getMetrics(request)
			require.NoError(t, err)
			require.NotNil(t, metrics)

			assert.Len(t, metrics, 3)

			for _, metric := range metrics {
				assert.Contains(t, []string{statusMetricName, responseTimeMetricName, responseBodyLengthMetricName}, metric["__name__"])

				if metric["__name__"] == statusMetricName {
					assert.Len(t, metric, 8)
					assert.Equal(t, float64(statusMetricValue), metric["value"])
					assert.Equal(t, successStatusMetricStatusLabelValue, metric[statusMetricStatusLabelName])
					assert.Equal(t, "200", metric[statusMetricResponseStatusCodeLabelName])
				} else if metric["__name__"] == responseTimeMetricName {
					assert.Len(t, metric, 7)
					assert.NotEmpty(t, metric["value"])
					assert.Equal(t, responseTimeMetricUnitLabelValue, metric[unitLabelName])
				} else if metric["__name__"] == responseBodyLengthMetricName {
					assert.Len(t, metric, 7)
					assert.Equal(t, float64(len("success")), metric["value"])
					assert.Equal(t, responseBodyLengthMetricUnitLabelValue, metric[unitLabelName])
				}

				assert.Equal(t, "https://example.api:1234", metric[urlLabelName])
				assert.Equal(t, http.MethodGet, metric[methodLabelName])
				assert.Equal(t, "us-east-1", metric[awsRegionLabelName])
				assert.Equal(t, "test", metric[awsLambdaFunctionLabelName])
			}

			return httpmock.NewStringResponse(http.StatusOK, ""), nil
		})

	err = run(context.Background())
	require.NoError(t, err)

	os.Clearenv()
}

func TestRun_ConnectionFailedStatus(t *testing.T) {
	err := os.Setenv(awsRegionEnvName, "us-east-1")
	require.NoError(t, err)

	err = os.Setenv(awsLambdaFunctionNameEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodPost, "https://listener.logz.io:8053",
		func(request *http.Request) (*http.Response, error) {
			metrics, err := getMetrics(request)
			require.NoError(t, err)
			require.NotNil(t, metrics)

			assert.Len(t, metrics, 1)

			metric := metrics[0]

			assert.Len(t, metric, 8)

			assert.Equal(t, statusMetricName, metric["__name__"])
			assert.Equal(t, float64(statusMetricValue), metric["value"])
			assert.Equal(t, "https://example.api:1234", metric[urlLabelName])
			assert.Equal(t, http.MethodGet, metric[methodLabelName])
			assert.Equal(t, connectionFailedStatusMetricStatusLabelValue, metric[statusMetricStatusLabelName])
			assert.NotEmpty(t, metric[statusMetricErrorLabelName])
			assert.Equal(t, "us-east-1", metric[awsRegionLabelName])
			assert.Equal(t, "test", metric[awsLambdaFunctionLabelName])

			return httpmock.NewStringResponse(http.StatusOK, ""), nil
		})

	err = run(context.Background())
	require.NoError(t, err)

	os.Clearenv()
}

func TestRun_NoMatchStatusCodeStatus(t *testing.T) {
	err := os.Setenv(awsRegionEnvName, "us-east-1")
	require.NoError(t, err)

	err = os.Setenv(awsLambdaFunctionNameEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "success")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, "https://example.api:1234",
		httpmock.NewStringResponder(http.StatusUnauthorized, "success"))

	httpmock.RegisterResponder(http.MethodPost, "https://listener.logz.io:8053",
		func(request *http.Request) (*http.Response, error) {
			metrics, err := getMetrics(request)
			require.NoError(t, err)
			require.NotNil(t, metrics)

			assert.Len(t, metrics, 3)

			for _, metric := range metrics {
				assert.Contains(t, []string{statusMetricName, responseTimeMetricName, responseBodyLengthMetricName}, metric["__name__"])

				if metric["__name__"] == statusMetricName {
					assert.Len(t, metric, 9)
					assert.Equal(t, float64(statusMetricValue), metric["value"])
					assert.Equal(t, noMatchStatusCodeStatusMetricStatusLabelValue, metric[statusMetricStatusLabelName])
					assert.Equal(t, "401", metric[statusMetricResponseStatusCodeLabelName])
					assert.Equal(t, "200", metric[statusMetricExpectedResponseStatusCodeLabelName])
				} else if metric["__name__"] == responseTimeMetricName {
					assert.Len(t, metric, 7)
					assert.NotEmpty(t, metric["value"])
					assert.Equal(t, responseTimeMetricUnitLabelValue, metric[unitLabelName])
				} else if metric["__name__"] == responseBodyLengthMetricName {
					assert.Len(t, metric, 7)
					assert.Equal(t, float64(len("success")), metric["value"])
					assert.Equal(t, responseBodyLengthMetricUnitLabelValue, metric[unitLabelName])
				}

				assert.Equal(t, "https://example.api:1234", metric[urlLabelName])
				assert.Equal(t, http.MethodGet, metric[methodLabelName])
				assert.Equal(t, "us-east-1", metric[awsRegionLabelName])
				assert.Equal(t, "test", metric[awsLambdaFunctionLabelName])
			}

			return httpmock.NewStringResponse(http.StatusOK, ""), nil
		})

	err = run(context.Background())
	require.NoError(t, err)

	os.Clearenv()
}

func TestRun_NoMatchResponseBodyStatus(t *testing.T) {
	err := os.Setenv(awsRegionEnvName, "us-east-1")
	require.NoError(t, err)

	err = os.Setenv(awsLambdaFunctionNameEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiUrlEnvName, "https://example.api:1234")
	require.NoError(t, err)

	err = os.Setenv(methodEnvName, http.MethodGet)
	require.NoError(t, err)

	err = os.Setenv(headersEnvName, "Content-Type=text/application,Accept=text/application")
	require.NoError(t, err)

	err = os.Setenv(bodyEnvName, "test")
	require.NoError(t, err)

	err = os.Setenv(apiResponseTimeoutEnvName, "10")
	require.NoError(t, err)

	err = os.Setenv(expectedStatusCodeEnvName, "200")
	require.NoError(t, err)

	err = os.Setenv(expectedBodyEnvName, "API is working")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsListenerEnvName, "https://listener.logz.io:8053")
	require.NoError(t, err)

	err = os.Setenv(logzioMetricsTokenEnvName, "123456789a")
	require.NoError(t, err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, "https://example.api:1234",
		httpmock.NewStringResponder(http.StatusOK, "success"))

	httpmock.RegisterResponder(http.MethodPost, "https://listener.logz.io:8053",
		func(request *http.Request) (*http.Response, error) {
			metrics, err := getMetrics(request)
			require.NoError(t, err)
			require.NotNil(t, metrics)

			assert.Len(t, metrics, 3)

			for _, metric := range metrics {
				assert.Contains(t, []string{statusMetricName, responseTimeMetricName, responseBodyLengthMetricName}, metric["__name__"])

				if metric["__name__"] == statusMetricName {
					assert.Len(t, metric, 10)
					assert.Equal(t, float64(statusMetricValue), metric["value"])
					assert.Equal(t, noMatchResponseBodyStatusMetricStatusLabelValue, metric[statusMetricStatusLabelName])
					assert.Equal(t, "200", metric[statusMetricResponseStatusCodeLabelName])
					assert.Equal(t, "success", metric[statusMetricResponseBodyLabelName])
					assert.Equal(t, "API is working", metric[statusMetricExpectedResponseBodyLabelName])
				} else if metric["__name__"] == responseTimeMetricName {
					assert.Len(t, metric, 7)
					assert.NotEmpty(t, metric["value"])
					assert.Equal(t, responseTimeMetricUnitLabelValue, metric[unitLabelName])
				} else if metric["__name__"] == responseBodyLengthMetricName {
					assert.Len(t, metric, 7)
					assert.Equal(t, float64(len("success")), metric["value"])
					assert.Equal(t, responseBodyLengthMetricUnitLabelValue, metric[unitLabelName])
				}

				assert.Equal(t, "https://example.api:1234", metric[urlLabelName])
				assert.Equal(t, http.MethodGet, metric[methodLabelName])
				assert.Equal(t, "us-east-1", metric[awsRegionLabelName])
				assert.Equal(t, "test", metric[awsLambdaFunctionLabelName])
			}

			return httpmock.NewStringResponse(http.StatusOK, ""), nil
		})

	err = run(context.Background())
	require.NoError(t, err)

	os.Clearenv()
}
