package nagatha

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang/glog"
	"google.golang.org/api/idtoken"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type ContextKey string

const (
	authorizationHeader = "Authorization"
)

type NagathaClient struct {
	client *http.Client

	addr string
}

var (
	opts = make([]http.Transport, 0)
)

func NewNagathaClient(addr string) (*NagathaClient, error) {
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
	return &NagathaClient{
		client: &http.Client{Timeout: 10 * time.Second},
		addr:   addr,
	}, nil
}

func (c *NagathaClient) CreateNotification(ctx context.Context, notification *Notification) error {
	return c.sendRequest(ctx, http.MethodPost, "/v1/notification", notification, nil)
}

func (c *NagathaClient) GetException(ctx context.Context, uuid string) (*Exception, error) {
	request := &GetExceptionRequest{
		Uuid: uuid,
	}
	response := &Exception{}
	if err := c.sendRequest(ctx, http.MethodGet, "/v1/exception", request, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *NagathaClient) CreateException(ctx context.Context, exception *Exception) error {
	return c.sendRequest(ctx, http.MethodPost, "/v1/exception", exception, nil)
}

func (c *NagathaClient) UpdateException(ctx context.Context, exception *Exception, updateMask *field_mask.FieldMask) error {
	request := &UpdateExceptionRequest{
		Exception:  exception,
		UpdateMask: updateMask,
	}
	return c.sendRequest(ctx, http.MethodPatch, "/v1/exception", request, nil)
}

func (c *NagathaClient) DeleteException(ctx context.Context, uuid string) error {
	request := &DeleteExceptionRequest{
		Uuid: uuid,
	}
	return c.sendRequest(ctx, http.MethodDelete, "/v1/exception", request, nil)
}

func (c *NagathaClient) ListExceptions(ctx context.Context, userEmail string, pageSize int32, pageToken string) (*ListExceptionsResponse, error) {
	request := &ListExceptionsRequest{
		UserEmail: userEmail,
		PageSize:  pageSize,
		PageToken: pageToken,
	}
	response := &ListExceptionsResponse{}
	if err := c.sendRequest(ctx, http.MethodGet, "/v1/exceptions", request, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *NagathaClient) sendRequest(ctx context.Context, method, path string, request proto.Message, response proto.Message) error {
	addr := c.addr + path
	var httpRequest *http.Request
	var requestBody []byte
	var err error

	// Serialize request message to JSON
	if method == http.MethodPost {
		requestBody, err = protoToJSON(request)
		if err != nil {
			return fmt.Errorf("failed to serialize request message: %v", err)
		}
		glog.V(10).Infof("nagatha request: %+v", string(requestBody))
		httpRequest, err = http.NewRequest(method, addr, bytes.NewReader(requestBody))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %v", err)
		}
		httpRequest.Header.Set("Content-Type", "application/json")
	} else if method == http.MethodGet {
		httpRequest, err = http.NewRequest(method, addr, nil)
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %v", err)
		}
	}

	httpResponse, err := c.client.Do(addAuthentication(ctx, httpRequest))
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer httpResponse.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Check HTTP status code
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return fmt.Errorf("request failed with status code %d: %s", httpResponse.StatusCode, string(responseBody))
	}

	// Parse response JSON
	if response != nil {
		if err := protojson.Unmarshal(responseBody, response); err != nil {
			return fmt.Errorf("failed to parse response JSON: %v", err)
		}
	}

	return nil
}

func protoToJSON(message proto.Message) ([]byte, error) {
	return protojson.Marshal(message)
}

func addAuthentication(ctx context.Context, req *http.Request) *http.Request {
	// Create an identity token.
	// With a global TokenSource tokens would be reused and auto-refreshed at need.
	// A given TokenSource is specific to the audience.
	tokenSource, err := idtoken.NewTokenSource(ctx, clientID)
	if err != nil {
		glog.Warningf("idtoken.NewTokenSource: %v", err)
	} else {
		token, err := tokenSource.Token()
		if err != nil {
			glog.Warningf("TokenSource.Token: %v", err)
		} else {
			req.Header.Set(authorizationHeader, "Bearer "+token.AccessToken)
			return req
		}
	}
	glog.Warningf("no authentication added for context")
	return req
}
