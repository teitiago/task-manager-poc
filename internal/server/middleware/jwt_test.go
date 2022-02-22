//go:build unit

package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// TestJWTMiddlewareExtract Validates the correct behavior of JWTMiddlewareExtract
// Given a context it will be able to collect correct header and process the token.
// This test will integrate everything on the middleware.
func TestJWTMiddlewareExtract(t *testing.T) {
	testingMap := []struct {
		name           string
		inputHeader    string
		expectedCode   int
		expectedOutput string
	}{
		{
			name:           "no auth header",
			inputHeader:    "",
			expectedCode:   http.StatusUnauthorized,
			expectedOutput: "",
		},
		{
			name:           "invalid auth header",
			inputHeader:    "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9",
			expectedCode:   http.StatusUnauthorized,
			expectedOutput: "",
		},
		{
			name:           "no role test",
			inputHeader:    "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NDgwMCwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSJ9.5KbN2ywQ1zIqC11xfrtO65FZA9DJGWztaWGmyRGV9gk",
			expectedCode:   http.StatusUnauthorized,
			expectedOutput: "",
		},
		{
			name:           "single role test",
			inputHeader:    "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NDgwMCwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSIsInJvbGVzIjoiTWFuYWdlciJ9.l8cVqY0dOXj5-jVlGPjl2B5CQ0q4QDgDfO9z_7XOvwo",
			expectedCode:   http.StatusOK,
			expectedOutput: "Manager,69c10abf-487f-42b3-8730-a58947298fe5",
		},
		{
			name:           "no claim test",
			inputHeader:    "Bearer eyJhbGciOiJSUzI1NiIsImFhYSI6dHJ1ZX0.eyJpc3MiOiJteV90aGluZyIsImF1ZCI6InRhc2siLCJpYXQiOjE2NDU1Mzg5MjEsImV4cCI6MTY0NTUzOTUyMSwicm9sZXMiOiJNYW5hZ2VyIn0.jLCn_IwAXEYMxMheJKzGTSRCs8sSeJHSBfmzfHG8JNL9rDv3EFLpJTF4La-cBzwC8Ddkj5PgO7RdPMwL4koE2KE3RhoDAk9xtWy_GmCZ6Xxo24jnVDxdQUIrOpaD2qz4HNufW_2geOoRltbT3gdcq1hOSymLh9H5ijTyl3IJP35WtBHJWf1f5EkGdxj45FfJz76U8-a4tp69goCNeHhJLoBWVLxZ00WtX5960nKmwd3n8j69j9Fhno8wXCiF5mTRjzmFojjS_7mPtFOE8TaDNXhZLPkG-f5yyO-D_S4sRiY6bS2BeSVQ8N9QxLjiKeXOAhkPqMfp65kdvqf9ziCZ8w",
			expectedCode:   http.StatusUnauthorized,
			expectedOutput: "",
		},
	}

	router := gin.New()

	router.Use(JWTMiddlewareExtract())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("%v,%v", c.MustGet(RoleClaim).(string), c.MustGet(SubClaim)))
	})

	for _, test := range testingMap {
		t.Run(t.Name(), func(t *testing.T) {

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)
			if test.inputHeader != "" {
				req.Header.Set(authHeader, test.inputHeader)
			}
			router.ServeHTTP(w, req)

			if test.expectedCode != w.Code {
				t.Errorf("expected %v, got %v", test.expectedCode, w.Code)
			}
			if test.expectedOutput != w.Body.String() {
				t.Errorf("expected %v, got %v", test.expectedOutput, w.Body.String())
			}
		})
	}

}

// TestExtractClaims Validates the correct behavior of extractClaims.
// Given a raw jwt string it's able to create the jwt map claims.
func TestExtractClaims(t *testing.T) {

	// Given
	testingMap := []struct {
		name           string
		input          string
		expectedClaims jwt.MapClaims
		expectingErr   bool
	}{
		{
			name:  "multiple roles and sub",
			input: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NzU0MSwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSIsInJvbGVzIjpbIk1hbmFnZXIiLCJUZWNobmljaWFuIl19.ZRmU23SX0m3FpRv7ptEc4252pZYRn_dGPC-ayB8OMEs",
			expectedClaims: map[string]interface{}{
				"iss":   "Online JWT Builder",
				"iat":   1.6452288e+09,
				"exp":   1.676767541e+09,
				"aud":   "task-manager-poc",
				"sub":   "69c10abf-487f-42b3-8730-a58947298fe5",
				"roles": []interface{}{"Manager", "Technician"},
			},
			expectingErr: false,
		},
		{
			name:  "single roles and sub",
			input: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NDgwMCwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSIsInJvbGVzIjoiTWFuYWdlciJ9.l8cVqY0dOXj5-jVlGPjl2B5CQ0q4QDgDfO9z_7XOvwo",
			expectedClaims: map[string]interface{}{
				"iss":   "Online JWT Builder",
				"iat":   1.6452288e+09,
				"exp":   1.6767648e+09,
				"aud":   "task-manager-poc",
				"sub":   "69c10abf-487f-42b3-8730-a58947298fe5",
				"roles": "Manager",
			},
			expectingErr: false,
		},
		{
			name:  "no roles and sub",
			input: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE2NDUyMjg4MDAsImV4cCI6MTY3Njc2NDgwMCwiYXVkIjoidGFzay1tYW5hZ2VyLXBvYyIsInN1YiI6IjY5YzEwYWJmLTQ4N2YtNDJiMy04NzMwLWE1ODk0NzI5OGZlNSJ9.5KbN2ywQ1zIqC11xfrtO65FZA9DJGWztaWGmyRGV9gk",
			expectedClaims: map[string]interface{}{
				"iss": "Online JWT Builder",
				"iat": 1.6452288e+09,
				"exp": 1.6767648e+09,
				"aud": "task-manager-poc",
				"sub": "69c10abf-487f-42b3-8730-a58947298fe5",
			},
			expectingErr: false,
		},
		{
			name:           "invalid token",
			input:          "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9",
			expectedClaims: map[string]interface{}{},
			expectingErr:   true,
		},
		{
			name:           "invalid token",
			input:          "test strinv invalid",
			expectedClaims: map[string]interface{}{},
			expectingErr:   true,
		},
	}

	for _, test := range testingMap {
		t.Run(t.Name(), func(t *testing.T) {

			// When
			gotClaims, gotError := extractClaims(test.input)

			// Then
			if !reflect.DeepEqual(gotClaims, test.expectedClaims) {
				t.Errorf("expected %v, got %v", test.expectedClaims, gotClaims)
			}
			if gotError != nil && test.expectingErr == false {
				t.Errorf("not expecting error, %v", gotError.Error())
			}
			if gotError == nil && test.expectingErr == true {
				t.Error("expecting error")
			}
		})
	}

}

// TestExtractRoleInfo Validates the correct behavior of extractRoleInfo.
// A map of claims is correctly converted from array to a comma separated string.
func TestExtractRoleInfo(t *testing.T) {

	// Given
	testingMap := []struct {
		name          string
		input         jwt.MapClaims
		expectedRoles string
		expectingErr  bool
	}{
		{
			name:          "multiple roles test",
			input:         map[string]interface{}{RoleClaim: []interface{}{"Manager", "Technician"}},
			expectedRoles: "Manager,Technician",
			expectingErr:  false,
		},
		{
			name:          "single role test",
			input:         map[string]interface{}{RoleClaim: []interface{}{"Manager"}},
			expectedRoles: "Manager",
			expectingErr:  false,
		},
		{
			name:          "invalid role test",
			input:         map[string]interface{}{RoleClaim: []interface{}{12345}},
			expectedRoles: "",
			expectingErr:  true,
		},
		{
			name:          "invalid role test",
			input:         map[string]interface{}{RoleClaim: 12345},
			expectedRoles: "",
			expectingErr:  true,
		},
		{
			name:          "no role test",
			input:         map[string]interface{}{"sub": "69c10abf-487f-42b3-8730-a58947298fe5"},
			expectedRoles: "",
			expectingErr:  true,
		},
		{
			name:          "invalid claim",
			input:         nil,
			expectedRoles: "",
			expectingErr:  true,
		},
	}

	for _, test := range testingMap {
		t.Run(t.Name(), func(t *testing.T) {

			// When
			gotRoles, gotError := extractRoleInfo(test.input)

			// Then
			if gotRoles != test.expectedRoles {
				t.Errorf("expected %v, got %v", test.expectedRoles, gotRoles)
			}
			if gotError != nil && test.expectingErr == false {
				t.Errorf("not expecting error, %v", gotError.Error())
			}
			if gotError == nil && test.expectingErr == true {
				t.Error("expecting error")
			}
		})
	}

}
