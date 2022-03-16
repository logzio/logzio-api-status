package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	metricsExporter "github.com/logzio/go-metrics-sdk"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const (
	apiUrlEnvName                                      = "API_URL"
	methodEnvName                                      = "METHOD"
	headersEnvName                                     = "HEADERS"
	bodyEnvName                                        = "BODY"
	bearerTokenEnvName                                 = "BEARER_TOKEN"
	usernameEnvName                                    = "USERNAME"
	passwordEnvName                                    = "PASSWORD"
	apiResponseTimeoutEnvName                          = "API_RESPONSE_TIMEOUT"
	expectedStatusCodeEnvName                          = "EXPECTED_STATUS_CODE"
	expectedBodyEnvName                                = "EXPECTED_BODY"
	logzioMetricsListenerEnvName                       = "LOGZIO_METRICS_LISTENER"
	logzioMetricsTokenEnvName                          = "LOGZIO_METRICS_TOKEN"
	awsRegionEnvName                                   = "AWS_REGION"
	awsLambdaFunctionNameEnvName                       = "AWS_LAMBDA_FUNCTION_NAME"
	meterName                                          = "api_status"
	statusMetricName                                   = meterName + "_status"
	responseTimeMetricName                             = meterName + "_response_time"
	responseBodyLengthMetricName                       = meterName + "_response_body_length"
	statusObserverDescription                          = "API status"
	statusMetricValue                                  = 1
	awsRegionLabelName                                 = "aws_region"
	awsLambdaFunctionLabelName                         = "aws_lambda_function"
	urlLabelName                                       = "url"
	methodLabelName                                    = "method"
	statusMetricStatusLabelName                        = "status"
	responseTimeoutStatusMetricStatusLabelValue        = "response_timeout"
	connectionFailedStatusMetricStatusLabelValue       = "connection_failed"
	readResponseBodyFailedStatusMetricStatusLabelValue = "read_response_body_failed"
	noMatchStatusCodeStatusMetricStatusLabelValue      = "no_match_status_code"
	noMatchResponseBodyStatusMetricStatusLabelValue    = "no_match_response_body"
	successStatusMetricStatusLabelValue                = "success"
	statusMetricResponseTimeoutLabelName               = "response_timeout"
	statusMetricResponseTimeoutUnitLabelName           = "response_timeout_unit"
	statusMetricResponseTimeoutUnitLabelValue          = "seconds"
	statusMetricErrorLabelName                         = "error"
	statusMetricResponseStatusCodeLabelName            = "response_status_code"
	statusMetricExpectedResponseStatusCodeLabelName    = "expected_response_status_code"
	statusMetricResponseBodyLabelName                  = "response_body"
	statusMetricExpectedResponseBodyLabelName          = "expected_response_body"
	unitLabelName                                      = "unit"
	responseTimeMetricUnitLabelValue                   = "milliseconds"
	responseBodyLengthMetricUnitLabelValue             = "bytes"
)

var (
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
)

type logzioApiStatus struct {
	ctx                        context.Context
	logzioMetricsListener      string
	logzioMetricsToken         string
	url                        string
	method                     string
	headers                    map[string]string
	body                       string
	responseTimeout            time.Duration
	bearerToken                string
	username                   string
	password                   string
	expectedResponseStatusCode int
	expectedResponseBody       string
}

type int64GaugeObserver struct {
	name                  string
	int64ObserverCallback func(context.Context, metric.Int64ObserverResult)
	description           string
}

type float64GaugeObserver struct {
	name                    string
	float64ObserverCallback func(context.Context, metric.Float64ObserverResult)
	description             string
}

type metricRegister interface {
	registerMetric(metric.Meter)
}

func newLogzioApiStatus(ctx context.Context) (*logzioApiStatus, error) {
	logzioMetricsListener := os.Getenv(logzioMetricsListenerEnvName)
	if logzioMetricsListener == "" {
		return nil, fmt.Errorf("%s must not be empty", logzioMetricsListenerEnvName)
	}

	logzioMetricsToken := os.Getenv(logzioMetricsTokenEnvName)
	if logzioMetricsToken == "" {
		return nil, fmt.Errorf("%s must not be empty", logzioMetricsTokenEnvName)
	}

	apiURL := os.Getenv(apiUrlEnvName)
	if apiURL == "" {
		return nil, fmt.Errorf("%s must not be empty", apiUrlEnvName)
	}

	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s: %v", apiURL, err)
	}

	method := os.Getenv(methodEnvName)
	if method != http.MethodGet && method != http.MethodPost {
		return nil, fmt.Errorf("%s must be GET or POST", methodEnvName)
	}

	headers, err := getApiRequestHeaders()
	if err != nil {
		return nil, fmt.Errorf("error getting api headers: %v", err)
	}

	responseTimeout, err := strconv.Atoi(os.Getenv(apiResponseTimeoutEnvName))
	if err != nil {
		return nil, fmt.Errorf("%s must be a number", apiResponseTimeoutEnvName)
	}

	if responseTimeout < 1 {
		return nil, fmt.Errorf("%s must be a positive number", apiResponseTimeoutEnvName)
	}

	expectedResponseStatusCode, err := strconv.Atoi(os.Getenv(expectedStatusCodeEnvName))
	if err != nil {
		return nil, fmt.Errorf("%s must be a number", expectedStatusCodeEnvName)
	}

	if expectedResponseStatusCode < 100 || expectedResponseStatusCode > 599 {
		return nil, fmt.Errorf("%s must be a between 100 and 599 (inclusive)", apiResponseTimeoutEnvName)
	}

	return &logzioApiStatus{
		ctx:                        ctx,
		logzioMetricsListener:      logzioMetricsListener,
		logzioMetricsToken:         logzioMetricsToken,
		url:                        parsedURL.String(),
		method:                     method,
		headers:                    headers,
		body:                       os.Getenv(bodyEnvName),
		responseTimeout:            time.Duration(responseTimeout) * time.Second,
		bearerToken:                os.Getenv(bearerTokenEnvName),
		username:                   os.Getenv(usernameEnvName),
		password:                   os.Getenv(passwordEnvName),
		expectedResponseStatusCode: expectedResponseStatusCode,
		expectedResponseBody:       os.Getenv(expectedBodyEnvName),
	}, nil
}

func newInt64GaugeObserver(name string, observerCallback func(context.Context, metric.Int64ObserverResult), description string) *int64GaugeObserver {
	return &int64GaugeObserver{
		name:                  name,
		int64ObserverCallback: observerCallback,
		description:           description,
	}
}

func newFloat64GaugeObserver(name string, observerCallback func(context.Context, metric.Float64ObserverResult), description string) *float64GaugeObserver {
	return &float64GaugeObserver{
		name:                    name,
		float64ObserverCallback: observerCallback,
		description:             description,
	}
}

func (igo *int64GaugeObserver) registerMetric(meter metric.Meter) {
	_ = metric.Must(meter).NewInt64GaugeObserver(
		igo.name,
		igo.int64ObserverCallback,
		metric.WithDescription(igo.description),
	)
}

func (fgo *float64GaugeObserver) registerMetric(meter metric.Meter) {
	_ = metric.Must(meter).NewFloat64GaugeObserver(
		fgo.name,
		fgo.float64ObserverCallback,
		metric.WithDescription(fgo.description),
	)
}

func (las *logzioApiStatus) createApiHttpRequest() (*http.Request, error) {
	debugLogger.Println("Creating API HTTP request...")

	var bodyReader io.Reader

	if las.body != "" {
		bodyReader = strings.NewReader(las.body)
	}

	request, err := http.NewRequest(las.method, las.url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	if las.bearerToken != "" {
		bearer := "Bearer " + strings.Trim(las.bearerToken, "\n")
		request.Header.Add("Authorization", bearer)
	}

	for key, value := range las.headers {
		request.Header.Add(key, value)
		if key == "Host" {
			request.Host = value
		}
	}

	if las.username != "" || las.password != "" {
		request.SetBasicAuth(las.username, las.password)
	}

	return request, nil
}

func (las *logzioApiStatus) getApiHttpResponse(request *http.Request) (*http.Response, float64, error) {
	debugLogger.Println("Getting API HTTP response...")

	client := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   las.responseTimeout * time.Second,
	}
	start := time.Now()
	response, err := client.Do(request)
	end := time.Now()
	responseTime := float64(end.Sub(start)) / float64(time.Millisecond)

	return response, responseTime, err
}

func (las *logzioApiStatus) getResponseErrorStatusGaugeObserver(responseError error) *int64GaugeObserver {
	if responseError == nil {
		debugLogger.Println("No response error status")
		return nil
	}

	if timeoutError, ok := responseError.(net.Error); ok && timeoutError.Timeout() {
		observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
			debugLogger.Println("Running response timeout status observer callback...")

			result.Observe(statusMetricValue,
				attribute.String(urlLabelName, las.url),
				attribute.String(methodLabelName, las.method),
				attribute.String(statusMetricStatusLabelName, responseTimeoutStatusMetricStatusLabelValue),
				attribute.Float64(statusMetricResponseTimeoutLabelName, float64(las.responseTimeout/time.Second)),
				attribute.String(statusMetricResponseTimeoutUnitLabelName, statusMetricResponseTimeoutUnitLabelValue),
				attribute.String(statusMetricErrorLabelName, responseError.Error()))
		}

		return newInt64GaugeObserver(statusMetricName, observerCallback, statusObserverDescription)
	}

	observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running connection failed status observer callback...")

		result.Observe(statusMetricValue,
			attribute.String(urlLabelName, las.url),
			attribute.String(methodLabelName, las.method),
			attribute.String(statusMetricStatusLabelName, connectionFailedStatusMetricStatusLabelValue),
			attribute.String(statusMetricErrorLabelName, responseError.Error()))
	}

	return newInt64GaugeObserver(statusMetricName, observerCallback, statusObserverDescription)
}

func (las *logzioApiStatus) getReadResponseBodyErrorStatusGaugeObserver(responseStatusCode int, readResponseBodyError error) *int64GaugeObserver {
	if readResponseBodyError == nil {
		debugLogger.Println("No read response body error status")
		return nil
	}

	observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running read response body failed status observer callback...")

		result.Observe(statusMetricValue,
			attribute.String(urlLabelName, las.url),
			attribute.String(methodLabelName, las.method),
			attribute.String(statusMetricStatusLabelName, readResponseBodyFailedStatusMetricStatusLabelValue),
			attribute.Int(statusMetricResponseStatusCodeLabelName, responseStatusCode),
			attribute.String(statusMetricErrorLabelName, readResponseBodyError.Error()))
	}

	return newInt64GaugeObserver(statusMetricName, observerCallback, statusObserverDescription)
}

func (las *logzioApiStatus) getNoMatchStatusGaugeObserver(responseStatusCode int, responseBodyBytes []byte) *int64GaugeObserver {
	if responseStatusCode != las.expectedResponseStatusCode {
		observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
			debugLogger.Println("Running no match status code status observer callback...")

			result.Observe(statusMetricValue,
				attribute.String(urlLabelName, las.url),
				attribute.String(methodLabelName, las.method),
				attribute.String(statusMetricStatusLabelName, noMatchStatusCodeStatusMetricStatusLabelValue),
				attribute.Int(statusMetricResponseStatusCodeLabelName, responseStatusCode),
				attribute.Int(statusMetricExpectedResponseStatusCodeLabelName, las.expectedResponseStatusCode))
		}

		return newInt64GaugeObserver(statusMetricName, observerCallback, statusObserverDescription)
	}

	if string(responseBodyBytes) != las.expectedResponseBody {
		observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
			debugLogger.Println("Running no match response body status observer callback...")

			result.Observe(statusMetricValue,
				attribute.String(urlLabelName, las.url),
				attribute.String(methodLabelName, las.method),
				attribute.String(statusMetricStatusLabelName, noMatchResponseBodyStatusMetricStatusLabelValue),
				attribute.Int(statusMetricResponseStatusCodeLabelName, responseStatusCode),
				attribute.String(statusMetricResponseBodyLabelName, string(responseBodyBytes)),
				attribute.String(statusMetricExpectedResponseBodyLabelName, las.expectedResponseBody))
		}

		return newInt64GaugeObserver(statusMetricName, observerCallback, statusObserverDescription)
	}

	debugLogger.Println("No no match status")
	return nil
}

func (las *logzioApiStatus) getSuccessStatusGaugeObserver(responseStatusCode int) *int64GaugeObserver {
	observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running success status observer callback...")
		result.Observe(statusMetricValue,
			attribute.String(urlLabelName, las.url),
			attribute.String(methodLabelName, las.method),
			attribute.String(statusMetricStatusLabelName, successStatusMetricStatusLabelValue),
			attribute.Int(statusMetricResponseStatusCodeLabelName, responseStatusCode))
	}

	return newInt64GaugeObserver(statusMetricName, observerCallback, statusObserverDescription)
}

func (las *logzioApiStatus) getResponseTimeGaugeObserver(responseTime float64) *float64GaugeObserver {
	observerCallback := func(_ context.Context, result metric.Float64ObserverResult) {
		debugLogger.Println("Running response time observer callback...")

		result.Observe(responseTime,
			attribute.String(urlLabelName, las.url),
			attribute.String(methodLabelName, las.method),
			attribute.String(unitLabelName, responseTimeMetricUnitLabelValue))
	}

	return newFloat64GaugeObserver(responseTimeMetricName, observerCallback, "API response time")
}

func (las *logzioApiStatus) getResponseBodyLengthGaugeObserver(responseBodyLength int) *int64GaugeObserver {
	observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
		debugLogger.Println("Running response body length observer callback...")

		result.Observe(int64(responseBodyLength),
			attribute.String(urlLabelName, las.url),
			attribute.String(methodLabelName, las.method),
			attribute.String(unitLabelName, responseBodyLengthMetricUnitLabelValue))
	}

	return newInt64GaugeObserver(responseBodyLengthMetricName, observerCallback, "API response body length")
}

func getApiRequestHeaders() (map[string]string, error) {
	var headers map[string]string

	if headersString := os.Getenv(headersEnvName); headersString != "" {
		headers = make(map[string]string)

		for _, header := range strings.Split(headersString, ",") {
			if !strings.Contains(header, "=") {
				return nil, fmt.Errorf("header's key and value must be separated by '='")
			}

			header = strings.Replace(header, " ", "", -1)
			headerKeyAndValue := strings.Split(header, "=")
			headers[headerKeyAndValue[0]] = headerKeyAndValue[1]

			debugLogger.Println("Got API HTTP request header:", headerKeyAndValue[0], "=", headerKeyAndValue[1])
		}
	}

	return headers, nil
}

func (las *logzioApiStatus) createController() (*controller.Controller, error) {
	config := metricsExporter.Config{
		LogzioMetricsListener: las.logzioMetricsListener,
		LogzioMetricsToken:    las.logzioMetricsToken,
		RemoteTimeout:         30 * time.Second,
		PushInterval:          15 * time.Second,
	}

	return metricsExporter.InstallNewPipeline(config,
		controller.WithCollectPeriod(5*time.Second),
		controller.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				attribute.String(awsRegionLabelName, os.Getenv(awsRegionEnvName)),
				attribute.String(awsLambdaFunctionLabelName, os.Getenv(awsLambdaFunctionNameEnvName)),
			),
		),
	)
}

func (las *logzioApiStatus) collectMetrics(metricRegisters []metricRegister) error {
	cont, err := las.createController()
	if err != nil {
		return fmt.Errorf("error creating controller: %v", err)
	}

	debugLogger.Println("Collecting metrics...")

	defer func() {
		handleErr(cont.Stop(las.ctx))
	}()

	meter := cont.Meter(meterName)

	for _, metricReg := range metricRegisters {
		metricReg.registerMetric(meter)
	}

	return nil
}

func run(ctx context.Context) error {
	gaugeObservers := make([]metricRegister, 0)
	apiStatus, err := newLogzioApiStatus(ctx)
	if err != nil {
		return fmt.Errorf("error creating logzioApiStatus instance: %v", err)
	}

	request, err := apiStatus.createApiHttpRequest()
	if err != nil {
		return fmt.Errorf("error creating API HTTP request: %v", err)
	}

	response, responseTime, err := apiStatus.getApiHttpResponse(request)
	if statusGaugeObserver := apiStatus.getResponseErrorStatusGaugeObserver(err); statusGaugeObserver != nil {
		gaugeObservers = append(gaugeObservers, statusGaugeObserver)
		return apiStatus.collectMetrics(gaugeObservers)
	}

	responseTimeGaugeObserver := apiStatus.getResponseTimeGaugeObserver(responseTime)
	gaugeObservers = append(gaugeObservers, responseTimeGaugeObserver)

	defer closeResponseBody(response.Body)

	bodyBytes, err := io.ReadAll(response.Body)
	if statusGaugeObserver := apiStatus.getReadResponseBodyErrorStatusGaugeObserver(response.StatusCode, err); statusGaugeObserver != nil {
		gaugeObservers = append(gaugeObservers, statusGaugeObserver)
		return apiStatus.collectMetrics(gaugeObservers)
	}

	responseBodyLengthGaugeObserver := apiStatus.getResponseBodyLengthGaugeObserver(len(bodyBytes))
	gaugeObservers = append(gaugeObservers, responseBodyLengthGaugeObserver)

	if statusGaugeObserver := apiStatus.getNoMatchStatusGaugeObserver(response.StatusCode, bodyBytes); statusGaugeObserver != nil {
		gaugeObservers = append(gaugeObservers, statusGaugeObserver)
		return apiStatus.collectMetrics(gaugeObservers)
	}

	statusGaugeObserver := apiStatus.getSuccessStatusGaugeObserver(response.StatusCode)
	gaugeObservers = append(gaugeObservers, statusGaugeObserver)

	return apiStatus.collectMetrics(gaugeObservers)
}

func handleErr(err error) {
	if err != nil {
		panic(fmt.Errorf("something went wrong: %v", err))
	}
}

func closeResponseBody(responseBody io.ReadCloser) {
	if err := responseBody.Close(); err != nil {
		panic(fmt.Errorf("error closing response body: %v", err))
	}
}

func HandleRequest(ctx context.Context) error {
	infoLogger.Println("Starting to get API status...")

	if err := run(ctx); err != nil {
		return err
	}

	infoLogger.Println("API status has been sent to Logz.io successfully")
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
