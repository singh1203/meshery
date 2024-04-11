package credentials

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
)

func TestListCredentialCmd(t *testing.T) {
	// setup current context
	utils.SetupContextEnv(t)

	// initialize mock server for handling requests
	utils.StartMockery(t)

	// create a test helper
	testContext := utils.NewTestHelper(t)

	// get current directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Not able to get current working directory")
	}
	currDir := filepath.Dir(filename)
	fixturesDir := filepath.Join(currDir, "fixtures")

	// test scenarios for fetching data
	tests := []struct {
		Name             string
		Args             []string
		ExpectedResponse string
		URLs             []utils.MockURL
		Token            string
		ExpectError      bool
	}{
		{
			Name:             "Fetch Credentials List",
			Args:             []string{"list"},
			ExpectedResponse: "list.credentials.output.golden",
			URLs: []utils.MockURL{
				{
					Method:       "GET",
					URL:          testContext.BaseURL + "/api/integrations/credentials",
					Response:     "list.credentials.api.response.golden",
					ResponseCode: 200,
				},
			},
			Token:       filepath.Join(fixturesDir, "token.golden"),
			ExpectError: false,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			for _, url := range tt.URLs {
				apiResponse := utils.NewGoldenFile(t, url.Response, fixturesDir).Load()
				httpmock.RegisterResponder(url.Method, url.URL,
					httpmock.NewStringResponder(url.ResponseCode, apiResponse))
			}

			// set token
			utils.TokenFlag = tt.Token

			// Expected response
			testdataDir := filepath.Join(currDir, "testdata")
			golden := utils.NewGoldenFile(t, tt.ExpectedResponse, testdataDir)

			buff := utils.SetupMeshkitLoggerTesting(t, false)
			listCredentialCmd.SetArgs(tt.Args)
			listCredentialCmd.SetOutput(buff)
			err := listCredentialCmd.Execute()
			if err != nil {
				// if we're supposed to get an error
				if tt.ExpectError {
					// write it in file
					if *update {
						golden.Write(err.Error())
					}
					expectedResponse := golden.Load()

					utils.Equals(t, expectedResponse, err.Error())
					return
				}
				t.Fatal(err)
			}
			// response being printed in console
			actualResponse := buff.String()

			// write it in file
			if *update {
				golden.Write(actualResponse)
			}
			expectedResponse := golden.Load()
			utils.Equals(t, expectedResponse, actualResponse)
		})
		t.Log("List Credentials test Passed")
	}
	// stop mock server
	utils.StopMockery(t)
}