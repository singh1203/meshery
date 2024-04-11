package credentials

import (
	"flag"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
)

var update = flag.Bool("update", false, "update golden files")

func TestCredentialCmd(t *testing.T) {
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

	// test scenrios for fetching data
	testcase := []struct {
		Name             string
		Args             []string
		ExpectedResponse string
		URLs             []utils.MockURL
		Token            string
		ExpectError      bool
	}{
		{
			Name:             "credential listcmd",
			Args:             []string{"list"},
			ExpectedResponse: "credential.list.output.golden",
			URLs: []utils.MockURL{
				{
					Method:       "GET",
					URL:          testContext.BaseURL + "/api/integrations/credentials",
					Response:     "credential.list.api.response.golden",
					ResponseCode: 200,
				},
			},
			Token:       filepath.Join(fixturesDir, "token.golden"),
			ExpectError: false,
		},
	}
	for _, test := range testcase {
		t.Run(test.Name, func(t *testing.T) {
			//set token
			utils.TokenFlag = test.Token

			for _, url := range test.URLs {
				apiResponse := utils.NewGoldenFile(t, url.Response, fixturesDir).Load()
				httpmock.RegisterResponder(url.Method, url.URL,
					httpmock.NewStringResponder(url.ResponseCode, apiResponse))
			}

			// Expected output from golden files
			testdataDir := filepath.Join(currDir, "testdata")
			golden := utils.NewGoldenFile(t, test.ExpectedResponse, testdataDir)

			buff := utils.SetupMeshkitLoggerTesting(t, false)
			CredentialCmd.SetOutput(buff)
			CredentialCmd.SetArgs(test.Args)

			err := CredentialCmd.Execute()
			if err != nil {
				if test.ExpectError {

					if *update {
						golden.Write(err.Error())
					}
					expectedResponse := golden.Load()

					utils.Equals(t, expectedResponse, err.Error())
					return
				}
				t.Error(err)
			}
			//print response string to console
			actualResponse := buff.String()
			// write it in file
			if *update {
				golden.Write(actualResponse)
			}
			expectedResponse := golden.Load()

			utils.Equals(t, expectedResponse, actualResponse)
		})
		t.Log("Credentials tests Passed")
	}
	// stop mock server
	utils.StopMockery(t)
}