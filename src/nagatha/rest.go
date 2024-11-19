package nagatha

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/nianticlabs/modron/src/constants"
	modronmetric "github.com/nianticlabs/modron/src/metric"
	"github.com/nianticlabs/modron/src/model"
	"github.com/nianticlabs/modron/src/proto/generated/nagatha"
)

type ContextKey string

const (
	authorizationHeader = "Authorization"
	httpRequestTimeout  = 10 * time.Second
	apiVersion          = "v2"
)

type clientMetrics struct {
	RequestDuration metric.Float64Histogram
}

type Client struct {
	addr        string
	client      *http.Client
	tokenSource oauth2.TokenSource
	metrics     clientMetrics
}

var (
	opts = make([]http.Transport, 0)
)

func apiPath(path string) string {
	return "/" + apiVersion + path
}

func newNagathaClient(addr string, tokenSource oauth2.TokenSource) (*Client, error) {
	if tokenSource == nil {
		return nil, fmt.Errorf("tokenSource cannot be nil")
	}
	cp, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("cert pool: %w", err)
	}
	opts = append(opts, http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            cp,
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS13,
		},
	})
	c := &Client{
		client: &http.Client{
			Timeout:   httpRequestTimeout,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		addr:        addr,
		tokenSource: tokenSource,
	}
	err = c.initMetrics()
	return c, err
}

func (c *Client) CreateNotification(ctx context.Context, notification *nagatha.Notification) (model.Notification, error) {
	var resultNotification nagatha.Notification
	err := c.sendRequest(ctx, http.MethodPost, apiPath("/notifications"), notification, &resultNotification)
	return notificationFromProto(&resultNotification), err
}

func (c *Client) BatchCreateNotifications(ctx context.Context, notifications []*nagatha.Notification) ([]model.Notification, error) {
	ctx, span := tracer.Start(ctx, "BatchCreateNotifications")
	defer span.End()
	var resp longrunningpb.Operation
	err := c.sendRequest(ctx,
		http.MethodPost,
		apiPath("/notifications:batchCreate"),
		&nagatha.BatchCreateNotificationsRequest{
			Notifications: notifications,
		},
		&resp,
	)
	if err != nil {
		return nil, err
	}
	operationID := resp.Name

	deadline := time.Now().Add(5 * time.Minute) //nolint:mnd
	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("operation %q timeout", operationID)
		}
		op, err := c.GetOperation(ctx, &longrunningpb.GetOperationRequest{
			Name: operationID,
		})
		if err != nil {
			return nil, fmt.Errorf("get operation %q: %w", operationID, err)
		}
		if op.Done {
			if op.GetError() != nil {
				return nil, fmt.Errorf("operation %q failed: %v", operationID, op.GetError())
			}
			var result nagatha.BatchCreateNotificationsResponse
			if err := op.GetResponse().UnmarshalTo(&result); err != nil {
				return nil, fmt.Errorf("unmarshal response: %w", err)
			}
			var resultNotifications []model.Notification
			for _, n := range result.Notifications {
				resultNotifications = append(resultNotifications, notificationFromProto(n))
			}
			return resultNotifications, nil
		}
		log.Debugf("waiting for operation %q to complete", operationID)
		time.Sleep(1 * time.Second)
	}
}

func (c *Client) GetOperation(ctx context.Context, req *longrunningpb.GetOperationRequest) (*longrunningpb.Operation, error) {
	var operation longrunningpb.Operation
	err := c.sendRequest(ctx, http.MethodGet, apiPath("/"+url.PathEscape(req.Name)), nil, &operation)
	return &operation, err
}

func (c *Client) GetException(ctx context.Context, uuid string) (*nagatha.Exception, error) {
	request := &nagatha.GetExceptionRequest{
		Uuid: uuid,
	}
	response := &nagatha.Exception{}
	if err := c.sendRequest(ctx, http.MethodGet, apiPath("/exceptions/"+url.PathEscape(uuid)), request, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) CreateException(ctx context.Context, exception *nagatha.Exception) error {
	return c.sendRequest(ctx, http.MethodPost, apiPath("/exceptions"), exception, nil)
}

func (c *Client) UpdateException(ctx context.Context, exception *nagatha.Exception, updateMask *field_mask.FieldMask) error {
	request := &nagatha.UpdateExceptionRequest{
		Exception:  exception,
		UpdateMask: updateMask,
	}
	return c.sendRequest(ctx, http.MethodPatch, apiPath("/exceptions/"+url.PathEscape(exception.Uuid)), request, nil)
}

func (c *Client) DeleteException(ctx context.Context, uuid string) error {
	request := &nagatha.DeleteExceptionRequest{
		Uuid: uuid,
	}
	return c.sendRequest(ctx, http.MethodDelete, apiPath("/exceptions/"+url.PathEscape(uuid)), request, nil)
}

func (c *Client) ListExceptions(ctx context.Context, userEmail string, pageSize int32, pageToken string) (*nagatha.ListExceptionsResponse, error) {
	request := &nagatha.ListExceptionsRequest{
		UserEmail: userEmail,
		PageSize:  pageSize,
		PageToken: pageToken,
	}
	response := &nagatha.ListExceptionsResponse{}
	if err := c.sendRequest(ctx, http.MethodGet, apiPath("/exceptions"), request, response); err != nil {
		return nil, err
	}
	return response, nil
}

type RequestError struct {
	StatusCode int
	Message    string
}

func (r RequestError) Error() string {
	return fmt.Sprintf("request failed with status code %d: %s", r.StatusCode, r.Message)
}

var _ error = (*RequestError)(nil)

func (c *Client) sendRequest(ctx context.Context, method, path string, request, response proto.Message) error {
	ctx, span := tracer.Start(ctx, "sendRequest",
		trace.WithAttributes(
			attribute.String(constants.TraceKeyMethod, method),
			attribute.String(constants.TraceKeyPath, path),
		),
	)
	defer span.End()
	addr := c.addr + path
	var httpRequest *http.Request
	var requestBody []byte
	var err error

	reqStart := time.Now()
	// Serialize request message to JSON
	if method == http.MethodPost {
		requestBody, err = protoToJSON(request)
		if err != nil {
			return fmt.Errorf("failed to serialize request message: %w", err)
		}
		log.Tracef("nagatha request: %+v", string(requestBody))
		httpRequest, err = http.NewRequestWithContext(ctx, method, addr, bytes.NewReader(requestBody))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}
		httpRequest.Header.Set("Content-Type", "application/json")
	} else if method == http.MethodGet {
		httpRequest, err = http.NewRequestWithContext(ctx, method, addr, nil)
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}
	}

	httpRequest = c.addAuthentication(httpRequest)
	httpResponse, err := c.client.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		c.metrics.RequestDuration.
			Record(ctx, time.Since(reqStart).Seconds(),
				metric.WithAttributes(
					attribute.Int(modronmetric.KeyStatus, httpResponse.StatusCode),
					attribute.String(modronmetric.KeyMethod, method),
					attribute.String(modronmetric.KeyPath, path),
				),
			)
	}()
	defer httpResponse.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status code
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return RequestError{
			StatusCode: httpResponse.StatusCode,
			Message:    string(responseBody),
		}
	}

	// Parse response JSON
	if response != nil {
		if err := protojson.Unmarshal(responseBody, response); err != nil {
			return fmt.Errorf("failed to parse response JSON: %w", err)
		}
	}

	return nil
}

func (c *Client) addAuthentication(request *http.Request) *http.Request {
	token, err := c.tokenSource.Token()
	if err != nil {
		log.Errorf("TokenSource.Token: %v", err)
	} else {
		request.Header.Set(authorizationHeader, "Bearer "+token.AccessToken)
	}
	return request
}

func (c *Client) initMetrics() error {
	clientReqDurationHist, err := meter.Float64Histogram(
		constants.MetricsPrefix+"nagatha_client_request_duration_seconds",
		metric.WithDescription("Duration of Nagatha client requests"),
	)
	if err != nil {
		return err
	}
	c.metrics = clientMetrics{
		RequestDuration: clientReqDurationHist,
	}
	return nil
}

func protoToJSON(message proto.Message) ([]byte, error) {
	return protojson.Marshal(message)
}
