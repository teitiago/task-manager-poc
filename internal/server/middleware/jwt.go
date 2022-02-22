package middleware

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	authHeader = "authorization"
	RoleClaim  = "roles"
	SubClaim   = "sub"
)

// JWTExtract Extracts the JWT from the header and place the role claim
// and the sub claim as part of the context to be used latter.
func JWTMiddlewareExtract() gin.HandlerFunc {

	return func(c *gin.Context) {
		bearerToken := c.Request.Header.Get(authHeader)
		if bearerToken == "" || bearerToken == "Bearer " {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authToken := strings.Split(bearerToken, "Bearer ")[1]
		// collect jwt claims
		claims, err := extractClaims(authToken)
		if err != nil {
			zap.L().Error(err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// extract roles list
		roles, err := extractRoleInfo(claims)
		if err != nil {
			zap.L().Error(err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// extract user id
		userID, ok := claims[SubClaim]
		if !ok {
			zap.L().Warn("no sub claim provided")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// setup context keys
		c.Set(SubClaim, userID)
		c.Set(RoleClaim, roles)

		c.Next()
	}

}

// extractClaims Extracts the claims from a given JWT token.
// The reason why this function don't validate the token signature
// is because the authn should be done by other service.
func extractClaims(rawToken string) (jwt.MapClaims, error) {

	// ignore parts
	token, _, err := new(jwt.Parser).ParseUnverified(rawToken, jwt.MapClaims{})

	if err != nil {
		return jwt.MapClaims{}, err
	}

	// cast the claims to map claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.MapClaims{}, errors.New("failed to convert map claims")
	}

	return claims, nil
}

// extractRoleInfo Extracts the claims from a JWT token.
// The JWT is used to store the role to avoid an extra call to collect it from the user auth info.
// Be aware when setting the validation time.
func extractRoleInfo(claims jwt.MapClaims) (string, error) {

	if claims == nil {
		return "", errors.New("no claims provided")
	}

	// collect the role claim
	roleClaim, ok := claims[RoleClaim]
	if !ok {
		return "", errors.New("role claim not provided")
	}

	// join the claim in a single comma separated string
	switch kind := reflect.TypeOf(roleClaim).Kind(); kind {
	case reflect.Slice:
		interfaceArray := roleClaim.([]interface{})
		roleMap := make([]string, len(interfaceArray))
		for i, role := range interfaceArray {
			switch v := role.(type) {
			case string:
				roleMap[i] = v
			default:
				return "", errors.New("unknown role type")
			}

		}
		return strings.Join(roleMap, ","), nil
	case reflect.String:
		return roleClaim.(string), nil
	}
	return "", errors.New("unknown map type")
}
