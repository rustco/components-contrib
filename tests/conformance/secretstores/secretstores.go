/*
Copyright 2021 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secretstores

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapr/components-contrib/metadata"
	"github.com/dapr/components-contrib/secretstores"
	"github.com/dapr/components-contrib/tests/conformance/utils"
)

// creating this struct so that it can be expanded later.
type TestConfig struct {
	utils.CommonConfig
}

func NewTestConfig(name string, operations []string) TestConfig {
	tc := TestConfig{
		CommonConfig: utils.CommonConfig{
			ComponentType: "secretstores",
			ComponentName: name,
			Operations:    utils.NewStringSet(operations...),
		},
	}

	return tc
}

func ConformanceTests(t *testing.T, props map[string]string, store secretstores.SecretStore, config TestConfig) {
	// TODO add support for metadata
	// For local env var based component test
	t.Setenv("conftestsecret", "abcd")
	t.Setenv("secondsecret", "efgh")

	// Init
	t.Run("init", func(t *testing.T) {
		err := store.Init(t.Context(), secretstores.Metadata{Base: metadata.Base{
			Properties: props,
		}})
		require.NoError(t, err, "expected no error on initializing store")
	})

	t.Run("ping", func(t *testing.T) {
		err := secretstores.Ping(t.Context(), store)
		// TODO: Ideally, all stable components should implement a ping function,
		// so will only assert require.NoError(t, err) finally, i.e. when current implementation
		// implements ping in existing stable components
		if err != nil {
			require.EqualError(t, err, "ping is not implemented by this secret store")
		} else {
			require.NoError(t, err)
		}
	})

	// Get
	t.Run("get", func(t *testing.T) {
		getSecretRequest := secretstores.GetSecretRequest{
			Name: "conftestsecret",
		}
		getSecretResponse := secretstores.GetSecretResponse{
			Data: map[string]string{
				"conftestsecret": "abcd",
			},
		}

		t.Run("get", func(t *testing.T) {
			resp, err := store.GetSecret(t.Context(), getSecretRequest)
			require.NoError(t, err, "expected no error on getting secret %v", getSecretRequest)
			assert.NotNil(t, resp, "expected value to be returned")
			assert.NotNil(t, resp.Data, "expected value to be returned")
			assert.Equal(t, getSecretResponse.Data, resp.Data, "expected values to be equal")
		})
	})

	// Bulkget
	t.Run("bulkGet", func(t *testing.T) {
		bulkReq := secretstores.BulkGetSecretRequest{}
		expectedData := map[string]map[string]string{
			"conftestsecret": {
				"conftestsecret": "abcd",
			},
			"secondsecret": {
				"secondsecret": "efgh",
			},
		}

		t.Run("bulkget", func(t *testing.T) {
			resp, err := store.BulkGetSecret(t.Context(), bulkReq)
			require.NoError(t, err, "expected no error on getting secret %v", bulkReq)
			assert.NotNil(t, resp, "expected value to be returned")
			assert.NotNil(t, resp.Data, "expected value to be returned")

			// Many secret stores don't allow us to start with an
			// empty set of secrets.  For example, every Kubernetes
			// namespace will contain a secret token.
			//
			// As a result, here we can only confirm that the secret
			// store contains all that we expected, but it is possible that
			// it may have more.
			for k, m := range expectedData {
				assert.Equal(t, m, resp.Data[k], "expected values to be equal")
			}
		})
	})
}
