// +build integration

package eyeshade

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/brave-intl/bat-go/datastore/grantserver"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type ControllersTestSuite struct {
	suite.Suite
	ctx     context.Context
	db      Datastore
	rodb    Datastore
	mock    sqlmock.Sqlmock
	mockRO  sqlmock.Sqlmock
	server  *httptest.Server
	service *Service
}

func TestControllersTestSuite(t *testing.T) {
	suite.Run(t, new(ControllersTestSuite))
}

func (suite *ControllersTestSuite) SetupSuite() {
	ctx := context.Background()
	// setup mock DB we will inject into our pg
	mockDB, mock, err := sqlmock.New()
	suite.Require().NoError(err, "failed to create a sql mock")
	mockRODB, mockRO, err := sqlmock.New()
	suite.Require().NoError(err, "failed to create a sql mock")

	suite.ctx = ctx
	name := "sqlmock"
	suite.db = NewFromConnection(&grantserver.Postgres{
		DB: sqlx.NewDb(mockDB, name),
	}, name)
	suite.mock = mock
	suite.rodb = NewFromConnection(&grantserver.Postgres{
		DB: sqlx.NewDb(mockRODB, name),
	}, name)
	suite.mockRO = mockRO

	service, err := SetupService(
		WithRouter,
		WithConnections(suite.db, suite.rodb),
	)
	suite.Require().NoError(err)
	suite.service = service
	server := httptest.NewServer(service.Router())
	suite.server = server
}

func (suite *ControllersTestSuite) TearDownSuite() {
	defer suite.server.Close()
}

func (suite *ControllersTestSuite) DoRequest(method string, path string, body io.Reader) (*http.Response, []byte) {
	req, err := http.NewRequest(method, suite.server.URL+path, body)
	suite.Require().NoError(err)
	resp, err := http.DefaultClient.Do(req)
	suite.Require().NoError(err)
	respBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	suite.Require().NoError(err)
	return resp, respBody
}