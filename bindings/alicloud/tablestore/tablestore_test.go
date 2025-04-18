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

package tablestore

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/components-contrib/metadata"
	"github.com/dapr/kit/logger"
)

func TestTableStoreMetadata(t *testing.T) {
	m := bindings.Metadata{}
	m.Properties = map[string]string{"accessKeyID": "ACCESSKEYID", "accessKey": "ACCESSKEY", "instanceName": "INSTANCENAME", "tableName": "TABLENAME", "endpoint": "ENDPOINT"}
	aliCloudTableStore := AliCloudTableStore{}

	meta, err := aliCloudTableStore.parseMetadata(m)

	require.NoError(t, err)
	assert.Equal(t, "ACCESSKEYID", meta.AccessKeyID)
	assert.Equal(t, "ACCESSKEY", meta.AccessKey)
	assert.Equal(t, "INSTANCENAME", meta.InstanceName)
	assert.Equal(t, "TABLENAME", meta.TableName)
	assert.Equal(t, "ENDPOINT", meta.Endpoint)
}

func TestDataEncodeAndDecode(t *testing.T) {
	if !isLiveTest() {
		return
	}

	aliCloudTableStore := NewAliCloudTableStore(logger.NewLogger("test"))

	metadata := bindings.Metadata{Base: metadata.Base{
		Properties: getTestProperties(),
	}}
	aliCloudTableStore.Init(t.Context(), metadata)

	// test create
	putData := map[string]interface{}{
		"pk1":     "data1",
		"column1": "the string value of column1",
		"column2": int64(2),
	}
	data, err := json.Marshal(putData)
	require.NoError(t, err)
	putRowReq := &bindings.InvokeRequest{
		Operation: bindings.CreateOperation,
		Metadata: map[string]string{
			tableName:   "dapr_test_table2",
			primaryKeys: "pk1",
		},
		Data: data,
	}

	putInvokeResp, err := aliCloudTableStore.Invoke(t.Context(), putRowReq)

	require.NoError(t, err)
	assert.NotNil(t, putInvokeResp)

	putRowReq.Data, _ = json.Marshal(map[string]interface{}{
		"pk1":     "data2",
		"column1": "the string value of column1",
		"column2": int64(2),
	})

	putInvokeResp, err = aliCloudTableStore.Invoke(t.Context(), putRowReq)

	require.NoError(t, err)
	assert.NotNil(t, putInvokeResp)

	// test get
	getData, err := json.Marshal(map[string]interface{}{
		"pk1": "data1",
	})
	require.NoError(t, err)
	getInvokeReq := &bindings.InvokeRequest{
		Operation: bindings.GetOperation,
		Metadata: map[string]string{
			tableName:   "dapr_test_table2",
			primaryKeys: "pk1",
			columnToGet: "column1,column2,column3",
		},
		Data: getData,
	}

	getInvokeResp, err := aliCloudTableStore.Invoke(t.Context(), getInvokeReq)

	require.NoError(t, err)
	assert.NotNil(t, getInvokeResp)

	respData := make(map[string]interface{})
	err = json.Unmarshal(getInvokeResp.Data, &respData)

	require.NoError(t, err)

	assert.Equal(t, putData["column1"], respData["column1"])
	assert.Equal(t, putData["column2"], int64(respData["column2"].(float64)))

	// test list
	listData, err := json.Marshal([]map[string]interface{}{
		{
			"pk1": "data1",
		},
		{
			"pk1": "data2",
		},
	})
	require.NoError(t, err)

	listReq := &bindings.InvokeRequest{
		Operation: bindings.ListOperation,
		Metadata: map[string]string{
			tableName:   "dapr_test_table2",
			primaryKeys: "pk1",
			columnToGet: "column1,column2,column3",
		},
		Data: listData,
	}

	listResp, err := aliCloudTableStore.Invoke(t.Context(), listReq)
	require.NoError(t, err)
	assert.NotNil(t, listResp)

	listRespData := make([]map[string]interface{}, len(listData))
	err = json.Unmarshal(listResp.Data, &listRespData)

	require.NoError(t, err)
	assert.Len(t, listRespData, 2)

	assert.Equal(t, listRespData[0]["column1"], putData["column1"])
	assert.Equal(t, "data2", listRespData[1]["pk1"])

	// test delete
	deleteData, err := json.Marshal(map[string]interface{}{
		"pk1": "data1",
	})
	require.NoError(t, err)

	deleteReq := &bindings.InvokeRequest{
		Operation: bindings.DeleteOperation,
		Metadata: map[string]string{
			tableName:   "dapr_test_table2",
			primaryKeys: "pk1",
		},
		Data: deleteData,
	}

	deleteResp, err := aliCloudTableStore.Invoke(t.Context(), deleteReq)

	require.NoError(t, err)
	assert.NotNil(t, deleteResp)

	getInvokeResp, err = aliCloudTableStore.Invoke(t.Context(), getInvokeReq)

	require.NoError(t, err)
	assert.Nil(t, getInvokeResp.Data)
}

func getTestProperties() map[string]string {
	return map[string]string{
		"accessKeyID":  "****",
		"accessKey":    "****",
		"instanceName": "dapr-test",
		"tableName":    "dapr_test_table2",
		"endpoint":     "https://dapr-test.cn-hangzhou.ots.aliyuncs.com",
	}
}

func isLiveTest() bool {
	return os.Getenv("RUN_LIVE_ROCKETMQ_TEST") == "true"
}
