//
// Author:: Darren Murray (<darren.murray@lacework.net>)
// Copyright:: Copyright 2020, Lacework Inc.
// License:: Apache License, Version 2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lacework/go-sdk/api"
	"github.com/lacework/go-sdk/internal/intgguid"
	"github.com/lacework/go-sdk/internal/lacework"
)

func TestIntegrationsNewServiceNowAlertChannel(t *testing.T) {
	subject := api.NewServiceNowAlertChannel("integration_name",
		api.ServiceNowChannelData{
			InstanceURL:   "snow-lacework.com",
			Username:      "snow-user",
			Password:      "snow-pass",
			IssueGrouping: "Events",
		},
	)
	assert.Equal(t, api.ServiceNowChannelIntegration.String(), subject.Type)
}

func TestIntegrationsCreateServiceNowAlertChannel(t *testing.T) {
	var (
		intgGUID   = intgguid.New()
		fakeServer = lacework.MockServer()
	)
	fakeServer.MockAPI("external/integrations", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "CreateServiceNowAlertChannel should be a POST method")

		if assert.NotNil(t, r.Body) {
			body := httpBodySniffer(r)
			assert.Contains(t, body, "integration_name", "integration name is missing")
			assert.Contains(t, body, "SERVICE_NOW_REST", "wrong integration type")
			assert.Contains(t, body, "snow-user", "wrong username")
			assert.Contains(t, body, "snow-lacework.com", "wrong instance url")
			assert.Contains(t, body, "snow-pass", "wrong password")
			assert.Contains(t, body, "Events", "wrong issue grouping")
			assert.Contains(t, body, "ENABLED\":1", "integration is not enabled")
		}

		fmt.Fprintf(w, serviceNowChannelIntegrationJsonResponse(intgGUID))
	})
	defer fakeServer.Close()

	c, err := api.NewClient("test",
		api.WithToken("TOKEN"),
		api.WithURL(fakeServer.URL()),
	)
	assert.Nil(t, err)

	data := api.NewServiceNowAlertChannel("integration_name",
		api.ServiceNowChannelData{
			InstanceURL:   "snow-lacework.com",
			Username:      "snow-user",
			Password:      "snow-pass",
			IssueGrouping: "Events",
		},
	)
	assert.Equal(t, "integration_name", data.Name, "ServiceNowChannel integration name mismatch")
	assert.Equal(t, "SERVICE_NOW_REST", data.Type, "a new ServiceNowChannel integration should match its type")
	assert.Equal(t, 1, data.Enabled, "a new ServiceNowChannel integration should be enabled")

	response, err := c.Integrations.CreateServiceNowAlertChannel(data)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Ok)
	if assert.Equal(t, 1, len(response.Data)) {
		resData := response.Data[0]
		assert.Equal(t, intgGUID, resData.IntgGuid)
		assert.Equal(t, "integration_name", resData.Name)
		assert.True(t, resData.State.Ok)
		assert.Equal(t, "snow-lacework.com", resData.Data.InstanceURL)
		assert.Equal(t, "snow-user", resData.Data.Username)
		assert.Equal(t, "snow-pass", resData.Data.Password)
		assert.Equal(t, "Events", resData.Data.IssueGrouping)
	}
}

func TestIntegrationsGetServiceNowAlertChannel(t *testing.T) {
	var (
		intgGUID   = intgguid.New()
		apiPath    = fmt.Sprintf("external/integrations/%s", intgGUID)
		fakeServer = lacework.MockServer()
	)
	fakeServer.MockAPI(apiPath, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "GetServiceNowAlertChannel should be a GET method")
		fmt.Fprintf(w, serviceNowChannelIntegrationJsonResponse(intgGUID))
	})
	defer fakeServer.Close()

	c, err := api.NewClient("test",
		api.WithToken("TOKEN"),
		api.WithURL(fakeServer.URL()),
	)
	assert.Nil(t, err)

	response, err := c.Integrations.GetServiceNowAlertChannel(intgGUID)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Ok)
	if assert.Equal(t, 1, len(response.Data)) {
		resData := response.Data[0]
		assert.Equal(t, intgGUID, resData.IntgGuid)
		assert.Equal(t, "integration_name", resData.Name)
		assert.True(t, resData.State.Ok)
		assert.Equal(t, "snow-lacework.com", resData.Data.InstanceURL)
		assert.Equal(t, "snow-user", resData.Data.Username)
		assert.Equal(t, "snow-pass", resData.Data.Password)
		assert.Equal(t, "Events", resData.Data.IssueGrouping)
	}
}

func TestIntegrationsUpdateServiceNowAlertChannel(t *testing.T) {
	var (
		intgGUID   = intgguid.New()
		apiPath    = fmt.Sprintf("external/integrations/%s", intgGUID)
		fakeServer = lacework.MockServer()
	)
	fakeServer.MockAPI(apiPath, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method, "UpdateServiceNowAlertChannel should be a PATCH method")

		if assert.NotNil(t, r.Body) {
			body := httpBodySniffer(r)
			assert.Contains(t, body, intgGUID, "INTG_GUID missing")
			assert.Contains(t, body, "integration_name", "integration name is missing")
			assert.Contains(t, body, "SERVICE_NOW_REST", "wrong integration type")
			assert.Contains(t, body, "snow-user", "wrong username")
			assert.Contains(t, body, "snow-lacework.com", "wrong instance url")
			assert.Contains(t, body, "snow-pass", "wrong password")
			assert.Contains(t, body, "Events", "wrong issue grouping")
			assert.Contains(t, body, "ENABLED\":1", "integration is not enabled")
		}

		fmt.Fprintf(w, serviceNowChannelIntegrationJsonResponse(intgGUID))
	})
	defer fakeServer.Close()

	c, err := api.NewClient("test",
		api.WithToken("TOKEN"),
		api.WithURL(fakeServer.URL()),
	)
	assert.Nil(t, err)

	data := api.NewServiceNowAlertChannel("integration_name",
		api.ServiceNowChannelData{
			InstanceURL:   "snow-lacework.com",
			Username:      "snow-user",
			Password:      "snow-pass",
			IssueGrouping: "Events",
		},
	)
	assert.Equal(t, "integration_name", data.Name, "ServiceNowChannel integration name mismatch")
	assert.Equal(t, "SERVICE_NOW_REST", data.Type, "a new ServiceNowChannel integration should match its type")
	assert.Equal(t, 1, data.Enabled, "a new ServiceNowChannel integration should be enabled")
	data.IntgGuid = intgGUID

	response, err := c.Integrations.UpdateServiceNowAlertChannel(data)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "SUCCESS", response.Message)
	assert.Equal(t, 1, len(response.Data))
	assert.Equal(t, intgGUID, response.Data[0].IntgGuid)
}

func TestIntegrationsListServiceNowAlertChannel(t *testing.T) {
	var (
		intgGUIDs  = []string{intgguid.New(), intgguid.New(), intgguid.New()}
		fakeServer = lacework.MockServer()
	)
	fakeServer.MockAPI("external/integrations/type/SERVICE_NOW_REST",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method, "ListServiceNowAlertChannel should be a GET method")
			fmt.Fprintf(w, serviceNowChanMultiIntegrationJsonResponse(intgGUIDs))
		},
	)
	defer fakeServer.Close()

	c, err := api.NewClient("test",
		api.WithToken("TOKEN"),
		api.WithURL(fakeServer.URL()),
	)
	assert.Nil(t, err)

	response, err := c.Integrations.ListServiceNowAlertChannel()
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Ok)
	assert.Equal(t, len(intgGUIDs), len(response.Data))
	for _, d := range response.Data {
		assert.Contains(t, intgGUIDs, d.IntgGuid)
	}
}

func serviceNowChannelIntegrationJsonResponse(intgGUID string) string {
	return `
{
  "data": [` + singleServiceNowChanIntegration(intgGUID) + `],
  "ok": true,
  "message": "SUCCESS"
}
`
}

func serviceNowChanMultiIntegrationJsonResponse(guids []string) string {
	integrations := []string{}
	for _, guid := range guids {
		integrations = append(integrations, singleServiceNowChanIntegration(guid))
	}
	return `
{
"data": [` + strings.Join(integrations, ", ") + `],
"ok": true,
"message": "SUCCESS"
}
`
}

func singleServiceNowChanIntegration(id string) string {
	return `
		{
			"INTG_GUID": "` + id + `",
			"NAME": "integration_name",
			"CREATED_OR_UPDATED_TIME": "2020-Mar-10 01:00:00 UTC",
			"CREATED_OR_UPDATED_BY": "user@email.com",
			"TYPE": "SERVICE_NOW_REST",
			"ENABLED": 1,
			"STATE": {
				"ok": true,
				"lastUpdatedTime": "2020-Mar-10 01:00:00 UTC",
				"lastSuccessfulTime": "2020-Mar-10 01:00:00 UTC"
			},
			"IS_ORG": 0,
			"DATA": {
				"INSTANCE_URL": "snow-lacework.com",
				"USERNAME": "snow-user",
				"PASSWORD": "snow-pass",
				"ISSUE_GROUPING": "Events"
			},
			"TYPE_NAME": "SERVICE_NOW_REST"
		}
	`
}
