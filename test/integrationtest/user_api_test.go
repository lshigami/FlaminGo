package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"queue_system/internal/dto/request"
	"queue_system/internal/model"
	"queue_system/internal/service" // Để truy cập các hằng số lỗi
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserAPI_CreateAndGetUser(t *testing.T) {
	CheckTestEnv(t)
	require.NotNil(t, globalTestApp, "globalTestApp not initialized")
	require.NotNil(t, globalTestApp.DB, "globalTestApp.DB not initialized")
	require.NotNil(t, globalTestApp.Router, "globalTestApp.Router not initialized")

	ClearTables(t, globalTestApp.DB, &model.User{}, &model.Appointment{})

	// 1. Create User
	createUserReq := request.CreateUserRequest{
		Name:  "API Test User",
		Email: "api.user@example.com",
		Role:  "tester",
	}
	rr := MakeRequest(t, globalTestApp.Router, http.MethodPost, "/api/v1/users", createUserReq)
	require.Equal(t, http.StatusCreated, rr.Code, "Create User failed. Response: %s", rr.Body.String())

	var createdUser model.User
	err := json.Unmarshal(rr.Body.Bytes(), &createdUser)
	require.NoError(t, err)
	assert.Equal(t, createUserReq.Name, createdUser.Name)
	assert.Equal(t, createUserReq.Email, createdUser.Email)
	assert.NotZero(t, createdUser.ID)
	createdUserID := createdUser.ID

	// 2. Get User by ID
	rrGet := MakeRequest(t, globalTestApp.Router, http.MethodGet, fmt.Sprintf("/api/v1/users/%d", createdUserID), nil)
	require.Equal(t, http.StatusOK, rrGet.Code)

	var fetchedUser model.User
	err = json.Unmarshal(rrGet.Body.Bytes(), &fetchedUser)
	require.NoError(t, err)
	assert.Equal(t, createdUserID, fetchedUser.ID)
	assert.Equal(t, createUserReq.Name, fetchedUser.Name)

	// 3. Try to create user with same email (should fail with 409)
	rrConflict := MakeRequest(t, globalTestApp.Router, http.MethodPost, "/api/v1/users", createUserReq)
	require.Equal(t, http.StatusConflict, rrConflict.Code, "Expected 409 Conflict for duplicate email. Response: %s", rrConflict.Body.String())
	var errorResponseConflict map[string]string
	err = json.Unmarshal(rrConflict.Body.Bytes(), &errorResponseConflict)
	require.NoError(t, err)
	assert.Contains(t, errorResponseConflict["error"], service.ErrEmailExists.Error())

	// 4. Get non-existent user (should fail with 404)
	rrNotFound := MakeRequest(t, globalTestApp.Router, http.MethodGet, "/api/v1/users/999999", nil)
	require.Equal(t, http.StatusNotFound, rrNotFound.Code, "Expected 404 Not Found for non-existent user. Body: %s", rrNotFound.Body.String())
	var errorResponseNotFound map[string]string
	err = json.Unmarshal(rrNotFound.Body.Bytes(), &errorResponseNotFound)
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(errorResponseNotFound["error"]), strings.ToLower(service.ErrUserNotFound.Error()))

}
