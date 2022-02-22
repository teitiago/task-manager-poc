package service

import (
	"context"
	"errors"
	"strings"

	"github.com/teitiago/task-manager-poc/internal/server/middleware"
	"github.com/teitiago/task-manager-poc/pkg/models"
)

type userInfo struct {
	roles string
	ID    string
}

// NewUserInfo Creates a new UserInfo instance
func NewUserInfo(ctx context.Context) (userInfo, error) {

	userID, ok := ctx.Value(middleware.SubClaim).(string)
	if !ok {
		return userInfo{}, errors.New("no user id provided")
	}
	roles, ok := ctx.Value(middleware.RoleClaim).(string)
	if !ok {
		return userInfo{}, errors.New("no roles provided")
	}

	return userInfo{roles: roles, ID: userID}, nil

}

// validateTask Validates the level of access to a given task
func (user *userInfo) validateTask(task models.Task) bool {

	if task.UserID.String() == user.ID {
		return true
	}

	return strings.Contains(user.roles, "Manager")

}
