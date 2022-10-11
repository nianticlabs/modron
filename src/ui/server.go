package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/golang/glog"
	"google.golang.org/api/idtoken"
)

const (
	distPathEnvVar    = "DIST_PATH"
	environmentEnvVar = "ENVIRONMENT"
	portEnvVar        = "PORT"

	e2eTestEnvironment = "E2E_TESTING"
)

const (
	defaultPort      = 8080
	e2eTestUserEmail = "e2e-test@example.com"
)

var (
	port     int64
	distPath string
)

func main() {
	flag.Parse()
	validateEnvironment()
	mux := http.NewServeMux()
	mux.HandleFunc("/", handle)
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func isE2eTest() bool {
	return os.Getenv(environmentEnvVar) == e2eTestEnvironment
}

func validateEnvironment() error {
	portEnv := os.Getenv(portEnvVar)
	if portEnv != "" {
		if parsedPort, err := strconv.ParseInt(portEnv, 10, 32); err != nil {
			return fmt.Errorf("%s contains an invalid port number %s: %v", portEnvVar, portEnv, err)
		} else {
			port = parsedPort
		}
	} else {
		port = defaultPort
	}
	if distPathEnv := os.Getenv(distPathEnvVar); distPathEnv != "" {
		distPath = distPathEnv
	} else {
		return fmt.Errorf("environment variable %q is not set", distPathEnvVar)
	}
	return nil
}

func validateIapJwt(ctx context.Context, iapJwt string) (*idtoken.Payload, error) {
	// We skip audience validation because we are allowing both internal and load balancer ingress.
	if payload, err := idtoken.Validate(ctx, iapJwt, ""); err != nil {
		return nil, fmt.Errorf("idtoken.Validate: %v", err)
	} else {
		return payload, nil
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	// Validate IAP token
	ctx := context.Background()
	userEmail := e2eTestUserEmail
	if !isE2eTest() {
		iapJwt := r.Header.Get("X-Goog-IAP-JWT-Assertion")
		if payload, err := validateIapJwt(ctx, iapJwt); err != nil {
			glog.Errorf("invalid IAP token: %v", err)
			w.WriteHeader(403)
			return
		} else {
			userEmail = payload.Claims["email"].(string)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Path:  "/",
		Name:  "modron-user-email",
		Value: userEmail,
	})

	// Serve app
	dir, file := path.Split(r.URL.RequestURI())
	ext := filepath.Ext(file)
	filePath := ""
	if file == "" || ext == "" {
		filePath = path.Join(distPath, "index.html")
	} else {
		filePath = path.Join(distPath, dir, file)
	}
	http.ServeFile(w, r, filePath)
}
