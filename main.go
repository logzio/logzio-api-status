package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/exporters/metric/cortex"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const (
	meterName = "api_status"
	statusObserverValue = 1
	statusObserverName = meterName + ".status"
	statusObserverDescription = "api status"
	responseTimeObserverName = meterName + ".response_time"
	statusCodeObserverName = meterName + ".status_code"
	responseBodyLengthObserverName = meterName + ".response_body_length"
)

type logzioApiStatus struct {
	url string
	method string
	headers map[string]string
	body string
	responseTimeout time.Duration
	bearerToken string
	username string
	password string
	expectedResponseStatusCode int
	expectedResponseBody string
	urlToSend string
}

type int64GaugeObserver struct {
	name string
	int64ObserverCallback func(context.Context, metric.Int64ObserverResult)
	description string
}

type float64GaugeObserver struct {
	name string
	float64ObserverCallback func(context.Context, metric.Float64ObserverResult)
	description string
}

type metricRegister interface {
	registerMetric(metric.Meter)
}

func newLogzioApiStatus() (*logzioApiStatus, error) {
	log.Debug("Creating LogzioApiStatus object...")

	apiURL := os.Getenv("API_URL")
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing url %s: %v", apiURL, err)
	}

	method := os.Getenv("METHOD")
	if method != "GET" && method != "POST" {
		return nil, fmt.Errorf("API_HTTP_REQUEST_METHOD must be GET or POST")
	}

	headers, err := getApiRequestHeaders()
	if err != nil {
		return nil, fmt.Errorf("error getting api headers: %v", err)
	}

	responseTimeoutSeconds, err := strconv.Atoi(os.Getenv("API_RESPONSE_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("API_RESPONSE_TIMEOUT must be a number")
	}

	expectedResponseStatusCode, err := strconv.Atoi(os.Getenv("EXPECTED_STATUS_CODE"))
	if err != nil {
		return nil, fmt.Errorf("API_HTTP_RESPONSE_STATUS_CODE must be a number")
	}

	urlToSend := parsedURL.String()
	if os.Getenv("SEND_API_URL_WITHOUT_PARAMS") == "true" {
		if strings.Contains(urlToSend, "/?") {
			urlToSend = strings.Split(urlToSend, "/?")[0]
		}
	}

	return &logzioApiStatus{
		url: parsedURL.String(),
		method: method,
		headers: headers,
		body: os.Getenv("BODY"),
		responseTimeout: time.Duration(responseTimeoutSeconds) * time.Second,
		bearerToken: os.Getenv("BEARER_TOKEN"),
		username: os.Getenv("USERNAME"),
		password: os.Getenv("PASSWORD"),
		expectedResponseStatusCode: expectedResponseStatusCode,
		expectedResponseBody: os.Getenv("EXPECTED_BODY"),
		urlToSend: urlToSend,
	}, nil
}

func newInt64GaugeObserver(name string, observerCallback func(context.Context, metric.Int64ObserverResult), description string) *int64GaugeObserver {
	return &int64GaugeObserver{
		name: name,
		int64ObserverCallback: observerCallback,
		description: description,
	}
}

func newFloat64GaugeObserver(name string, observerCallback func(context.Context, metric.Float64ObserverResult), description string) *float64GaugeObserver {
	return &float64GaugeObserver{
		name: name,
		float64ObserverCallback: observerCallback,
		description: description,
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
	log.Debug("Creating api http request...")

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

func (las *logzioApiStatus) getApiHttpResponse(request *http.Request) (*http.Response, error, float64) {
	log.Debug("Getting api http response...")

	client := &http.Client{
		Timeout: las.responseTimeout * time.Second,
	}
	start := time.Now()
	response, err := client.Do(request)
	responseTime := time.Since(start).Seconds()

	return response, err, responseTime
}

func (las *logzioApiStatus) getResponseErrorStatusGaugeObserver(responseError error) *int64GaugeObserver {
	if responseError == nil {
		log.Debug("No response error status")
		return nil
	}

	if timeoutError, ok := responseError.(net.Error); ok && timeoutError.Timeout() {
		log.Debug("Getting response timeout status observer...")

		observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(statusObserverValue,
				attribute.String("url", las.urlToSend),
				attribute.String("method", las.method),
				attribute.String("status", "timeout"),
				attribute.Float64("response_timeout", float64(las.responseTimeout/time.Second)),
				attribute.String("response_timeout_unit", "seconds"),
				attribute.String("error", responseError.Error()))
		}

		return newInt64GaugeObserver(statusObserverName, observerCallback, statusObserverDescription)

	}

	log.Debug("Getting response connection failed status observer...")

	observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
		result.Observe(statusObserverValue,
			attribute.String("url", las.urlToSend),
			attribute.String("method", las.method),
			attribute.String("status", "connection_failed"),
			attribute.String("error", responseError.Error()))
	}

	return newInt64GaugeObserver(statusObserverName, observerCallback, statusObserverDescription)
}

func (las *logzioApiStatus) getReadResponseBodyErrorStatusGaugeObserver(responseStatusCode int, readResponseBodyError error) *int64GaugeObserver {
	if readResponseBodyError == nil {
		log.Debug("No read response body error status")
		return nil
	}

	log.Debug("Getting read response body error status observer...")

	observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
		result.Observe(statusObserverValue,
			attribute.String("url", las.urlToSend),
			attribute.String("method", las.method),
			attribute.String("status", "read_response_body_failed"),
			attribute.Int("response_status_code", responseStatusCode),
			attribute.String("error", readResponseBodyError.Error()))
	}

	return newInt64GaugeObserver(statusObserverName, observerCallback, statusObserverDescription)
}

func (las *logzioApiStatus) getNoMatchStatusGaugeObserver(responseStatusCode int, responseBodyBytes []byte) *int64GaugeObserver {
	if responseStatusCode != las.expectedResponseStatusCode {
		log.Debug("Getting no match status code status observer...")

		observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(statusObserverValue,
				attribute.String("url", las.urlToSend),
				attribute.String("method", las.method),
				attribute.String("status", "no_match_status_code"),
				attribute.Int("response_status_code", responseStatusCode),
				attribute.Int("expected_response_status_code", las.expectedResponseStatusCode))
		}

		return newInt64GaugeObserver(statusObserverName, observerCallback, statusObserverDescription)
	}

	if string(responseBodyBytes) != las.expectedResponseBody {
		log.Debug("Getting no match response body status observer...")

		observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
			result.Observe(statusObserverValue,
				attribute.String("url", las.urlToSend),
				attribute.String("method", las.method),
				attribute.String("status", "no_match_response_body"),
				attribute.Int("response_status_code", responseStatusCode),
				attribute.String("response_body", string(responseBodyBytes)),
				attribute.String("expected_response_body", las.expectedResponseBody))
		}

		return newInt64GaugeObserver(statusObserverName, observerCallback, statusObserverDescription)
	}

	log.Debug("No no match status")
	return nil
}

func (las *logzioApiStatus) getSuccessStatusGaugeObserver(responseStatusCode int) *int64GaugeObserver {
	log.Debug("Getting success status observer...")

	observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
		result.Observe(statusObserverValue,
			attribute.String("url", las.urlToSend),
			attribute.String("method", las.method),
			attribute.String("status", "success"),
			attribute.Int("response_status_code", responseStatusCode))
	}

	return newInt64GaugeObserver(statusObserverName, observerCallback, statusObserverDescription)
}

func (las *logzioApiStatus) getResponseTimeGaugeObserver(responseTime float64) *float64GaugeObserver {
	log.Debug("Getting response time observer...")

	observerCallback := func(_ context.Context, result metric.Float64ObserverResult) {
		result.Observe(responseTime,
			attribute.String("url", las.urlToSend),
			attribute.String("method", las.method),
			attribute.String("unit", "seconds"))
	}

	return newFloat64GaugeObserver(responseTimeObserverName, observerCallback, "api response time")
}

func (las *logzioApiStatus) getResponseStatusCodeGaugeObserver(response *http.Response) *int64GaugeObserver {
	log.Debug("Getting response status code observer...")

	observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
		result.Observe(int64(response.StatusCode),
			attribute.String("url", las.urlToSend),
			attribute.String("method", las.method))
	}

	return newInt64GaugeObserver(statusCodeObserverName, observerCallback, "api status code")
}

func (las *logzioApiStatus) getResponseBodyLengthGaugeObserver(responseBodyLength int) *int64GaugeObserver {
	log.Debug("Getting response body length observer...")

	observerCallback := func(_ context.Context, result metric.Int64ObserverResult) {
		result.Observe(int64(responseBodyLength),
			attribute.String("url", las.urlToSend),
			attribute.String("method", las.method),
			attribute.String("unit", "bytes"))
	}

	return newInt64GaugeObserver(responseBodyLengthObserverName, observerCallback, "api response body length")
}

func getApiRequestHeaders() (map[string]string, error) {
	var headers map[string]string

	if headersString := os.Getenv("API_HTTP_REQUEST_HEADERS"); headersString != "" {
		if !strings.Contains(headersString, ",") {
			return nil, fmt.Errorf("headers must be separated by comma")
		}

		headers = make(map[string]string)

		for _, header := range strings.Split(headersString, ",") {
			if !strings.Contains(header, ":") {
				return nil, fmt.Errorf("header's key and value must be separated by colon")
			}

			headerKeyAndValue := strings.Split(header, ":")
			headers[headerKeyAndValue[0]] = headerKeyAndValue[1]
		}
	}

	return headers, nil
}

func createController() (*controller.Controller, error) {
	config := cortex.Config{
		Endpoint:      os.Getenv("LOGZIO_URL"),
		RemoteTimeout: 30 * time.Second,
		PushInterval:  5 * time.Second,
		BearerToken:   os.Getenv("LOGZIO_TOKEN"),
	}

	return cortex.InstallNewPipeline(config,
		controller.WithCollectPeriod(5*time.Second),
		controller.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				attribute.String("aws_region", os.Getenv("AWS_REGION")),
			),
		),
	)
}

func collectMetrics(metricRegisters []metricRegister) error {
	log.Debug("Collecting metrics...")

	ctx := context.Background()
	cont, err := createController()
	if err != nil {
		return fmt.Errorf("error creating controller: %v", err)
	}

	defer handleErr(cont.Stop(ctx))

	meter := cont.Meter(meterName)
	err = cont.Start(ctx)
	if err != nil {
		return fmt.Errorf("error starting controller: %v", err)
	}

	for _, metricReg := range metricRegisters {
		metricReg.registerMetric(meter)
	}

	time.Sleep(10 * time.Second)
	return nil
}

func run() error {
	gaugeObservers := make([]metricRegister, 0)
	apiStatus, err := newLogzioApiStatus()
	if err != nil {
		return fmt.Errorf("error creating logzioApiStatus: %v", err)
	}

	request, err := apiStatus.createApiHttpRequest()
	if err != nil {
		return fmt.Errorf("error creating api http request: %v", err)
	}

	log.Debug("Api http request was created successfully")

	response, err, responseTime := apiStatus.getApiHttpResponse(request)
	if statusGaugeObserver := apiStatus.getResponseErrorStatusGaugeObserver(err); statusGaugeObserver != nil {
		gaugeObservers = append(gaugeObservers, statusGaugeObserver)
		return collectMetrics(gaugeObservers)
	}

	log.Debug("Api http response was received successfully")

	responseTimeGaugeObserver := apiStatus.getResponseTimeGaugeObserver(responseTime)
	gaugeObservers = append(gaugeObservers, responseTimeGaugeObserver)

	defer closeResponseBody(response.Body)

	bodyBytes, err := io.ReadAll(response.Body)
	if statusGaugeObserver := apiStatus.getReadResponseBodyErrorStatusGaugeObserver(response.StatusCode, err); statusGaugeObserver != nil {
		gaugeObservers = append(gaugeObservers, statusGaugeObserver)
		return collectMetrics(gaugeObservers)
	}

	responseBodyLengthGaugeObserver := apiStatus.getResponseBodyLengthGaugeObserver(len(bodyBytes))
	gaugeObservers = append(gaugeObservers, responseBodyLengthGaugeObserver)

	if statusGaugeObserver := apiStatus.getNoMatchStatusGaugeObserver(response.StatusCode, bodyBytes); statusGaugeObserver != nil {
		gaugeObservers = append(gaugeObservers, statusGaugeObserver)
		return collectMetrics(gaugeObservers)
	}

	statusGaugeObserver := apiStatus.getSuccessStatusGaugeObserver(response.StatusCode)
	gaugeObservers = append(gaugeObservers, statusGaugeObserver)

	return collectMetrics(gaugeObservers)
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

func main() {
	if err := run(); err != nil {
		panic(err)
	}

	log.Info("Your api status has been sent to Logz.io successfully")
}
