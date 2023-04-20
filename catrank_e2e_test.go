//go:build e2e

package main_test

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CatRankE2ETestSuite struct {
	suite.Suite
	baseURL    string
	restClient *resty.Client
}

func (s *CatRankE2ETestSuite) SetupTest() {
	s.baseURL = "http://localhost:8080"
	s.restClient = resty.New()
}

func (s *CatRankE2ETestSuite) TestGetAPI() {
	resp, err := s.restClient.R().
		SetHeader("Accept", "application/json").
		Get(s.baseURL)
	assert.NoError(s.T(), err, "An error during setup of test. "+
		"Please ensure your environment is setup correct")

	assert.Equal(s.T(), 200, resp.StatusCode(), "Unexpected status code")
}

func TestAccountRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(CatRankE2ETestSuite))
}
