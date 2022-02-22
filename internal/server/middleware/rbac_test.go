//go:build integration

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRBACMiddleware Validates the correct behavior of RBACMiddleware.
// Given a path and a list of roles, if the user is the resource owner or
// has the correct role to access the endpoint it wil, otherwise it will be
// retrieved a forbidden code.
func TestRBACMiddleware(t *testing.T) {

	testingMap := []struct {
		name            string
		endpointAllowed string
		restPath        string
		endpointPath    string
		inputHeader     string
		expectedCode    int
	}{

		{
			name:            "user cant access resource",
			endpointAllowed: "",
			restPath:        "/technician/:owner/tasks",
			endpointPath:    "/technician/7963069b-f321-433e-ab29-e4fb1946ea0e/tasks",
			inputHeader:     "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NDgwMCwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSIsInJvbGVzIjoiTWFuYWdlciJ9.l8cVqY0dOXj5-jVlGPjl2B5CQ0q4QDgDfO9z_7XOvwo",
			expectedCode:    http.StatusForbidden,
		},
		{
			name:            "user can access resource by role",
			endpointAllowed: "Manager",
			restPath:        "/technician/:owner/tasks",
			endpointPath:    "/technician/7963069b-f321-433e-ab29-e4fb1946ea0e/tasks",
			inputHeader:     "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NDgwMCwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSIsInJvbGVzIjoiTWFuYWdlciJ9.l8cVqY0dOXj5-jVlGPjl2B5CQ0q4QDgDfO9z_7XOvwo",
			expectedCode:    http.StatusOK,
		},
		{
			name:            "user can access because owns the resource",
			endpointAllowed: "",
			restPath:        "/technician/:owner/tasks",
			endpointPath:    "/technician/69c10abf-487f-42b3-8730-a58947298fe5/tasks",
			inputHeader:     "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NDgwMCwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSIsInJvbGVzIjoiTWFuYWdlciJ9.l8cVqY0dOXj5-jVlGPjl2B5CQ0q4QDgDfO9z_7XOvwo",
			expectedCode:    http.StatusOK,
		},

		{
			name:            "user can access by role",
			endpointAllowed: "Manager",
			restPath:        "/tasks",
			endpointPath:    "/tasks",
			inputHeader:     "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NDgwMCwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSIsInJvbGVzIjoiTWFuYWdlciJ9.l8cVqY0dOXj5-jVlGPjl2B5CQ0q4QDgDfO9z_7XOvwo",
			expectedCode:    http.StatusOK,
		},
	}

	for _, test := range testingMap {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.GET(test.restPath, JWTMiddlewareExtract(), RBACMiddleware(test.endpointAllowed), func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", test.endpointPath, nil)
			req.Header.Set(authHeader, test.inputHeader)
			router.ServeHTTP(w, req)

			if test.expectedCode != w.Code {
				t.Errorf("expected %v, got %v", test.expectedCode, w.Code)
			}
		})

	}

}
