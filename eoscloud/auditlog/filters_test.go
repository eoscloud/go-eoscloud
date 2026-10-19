package auditlog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/auditlog"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestIdentityClient_List(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":2,"results":[{"id":"id-1","email":"a@b.com","name":"A","type":"user"},{"id":"id-2","name":"ci-bot","type":"client"}]}`))
	}))
	defer srv.Close()

	c := auditlog.NewIdentityClient(api.NewClient(api.WithBaseURL(srv.URL)))
	result, err := c.List(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/v1/identities", gotPath)
	require.Len(t, result.Results, 2)
	assert.Equal(t, "id-1", result.Results[0].ID)
	assert.Equal(t, auditlog.IdentityTypeUser, result.Results[0].Type)
	assert.Equal(t, "id-2", result.Results[1].ID)
	assert.Equal(t, auditlog.IdentityTypeClient, result.Results[1].Type)
}

func TestServiceClient_List(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"compute","name":"compute"}]}`))
	}))
	defer srv.Close()

	c := auditlog.NewServiceClient(api.NewClient(api.WithBaseURL(srv.URL)))
	result, err := c.List(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, "/v1/services", gotPath)
	require.Len(t, result.Results, 1)
	assert.Equal(t, "compute", result.Results[0].Name)
}

func TestResourceKindClient_List(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":1,"results":[{"name":"vms","service":"compute"}]}`))
	}))
	defer srv.Close()

	c := auditlog.NewResourceKindClient(api.NewClient(api.WithBaseURL(srv.URL)))
	result, err := c.List(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, "/v1/resource-kinds", gotPath)
	require.Len(t, result.Results, 1)
	assert.Equal(t, "vms", result.Results[0].Name)
}

func TestResourceClient_List(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"res-1","name":"my-vm","service":"compute","resource_kind":"vms"}]}`))
	}))
	defer srv.Close()

	c := auditlog.NewResourceClient(api.NewClient(api.WithBaseURL(srv.URL)))
	result, err := c.List(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, "/v1/resources", gotPath)
	require.Len(t, result.Results, 1)
	assert.Equal(t, "res-1", result.Results[0].ID)
}

func TestActionTypeClient_List(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":1,"results":[{"name":"create","service":"compute","resource_kind":"vms"}]}`))
	}))
	defer srv.Close()

	c := auditlog.NewActionTypeClient(api.NewClient(api.WithBaseURL(srv.URL)))
	result, err := c.List(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, "/v1/action-types", gotPath)
	require.Len(t, result.Results, 1)
	assert.Equal(t, "create", result.Results[0].Name)
}
