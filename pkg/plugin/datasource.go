package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &Datasource{}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// when logging at a non-Debug level, make sure you don't include sensitive information in the message
	// (like the *backend.QueryDataRequest)
	log.DefaultLogger.Debug("QueryData called", "numQueries", len(req.Queries))

	// create response struct
	response := backend.NewQueryDataResponse()

	settings := req.PluginContext.DataSourceInstanceSettings

	//Get Secure data
	if primaryKey, primaryKeyExists := settings.DecryptedSecureJSONData["primaryKey"]; primaryKeyExists {
		if endpointUri, endpointUriExists := settings.DecryptedSecureJSONData["endpointUri"]; endpointUriExists {

			cred, err := azcosmos.NewKeyCredential(primaryKey)
			if err != nil {
				log.DefaultLogger.Error("Failed to create a credential: ", err)
			}

			// Create a CosmosDB client
			client, err := azcosmos.NewClientWithKey(endpointUri, cred, nil)
			if err != nil {
				log.DefaultLogger.Error("Failed to create Azure Cosmos DB client: ", err)
			}
			// loop over queries and execute them individually.
			for _, q := range req.Queries {
				res := d.query(ctx, req.PluginContext, q, client)

				// save the response in a hashmap
				// based on with RefID as identifier
				response.Responses[q.RefID] = res
			}
		}
	}

	return response, nil
}

type queryModel struct {
	Database     string
	Container    string
	PartitionKey string
	Columns      string
}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery, cosmosClient *azcosmos.Client) backend.DataResponse {
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm queryModel
	err := json.Unmarshal(query.JSON, &qm)
	log.DefaultLogger.Debug("QueryData Cosmos DB", "QueryModel", qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, "json unmarshal: "+err.Error())
	}

	containerClient, err := cosmosClient.NewContainer(qm.Database, qm.Container)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, "Cosmos Client Error: "+err.Error())
	}

	// create data frame response.
	frame := data.NewFrame("response")

	//サーバーに問い合わせる
	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(qm.PartitionKey)
	//アイテムを取得する
	ctx := context.TODO()

	//フィールドの共通部分を取り出して、フィールドに挿入する
	//時刻で制限できるようにする
	queryText := "select "
	splittedColumns := strings.Split(qm.Columns, ",")
	for i := 0; i < len(splittedColumns); i++ {
		splittedColumns[i] = strings.Trim(splittedColumns[i], " ")
	}
	useWildcard := false
	for idx, str := range splittedColumns {
		if strings.Contains(str, "*") {
			useWildcard = true
			break
		} else {
			queryText += "c." + str
		}
		if idx < len(splittedColumns)-1 {
			queryText += ","
		}
	}

	if useWildcard {
		queryText = "select *"
	} else {
		queryText += ",c._ts"
	}

	//Currently, _ts is used for time series
	//Convert time to unix for _ts
	fromTs := query.TimeRange.From.Unix() - 1
	toTs := query.TimeRange.To.Unix() + 1
	queryText += fmt.Sprintf(" from docs c where c._ts > %d and c._ts < %d", fromTs, toTs)

	log.DefaultLogger.Debug("QueryData Cosmos DB", "Query", queryText)
	queryPager := containerClient.NewQueryItemsPager(queryText, pk, nil)
	log.DefaultLogger.Debug("Response:", queryPager)
	timeData := []time.Time{}

	//Get Column data from response
	columnData := make(map[string]interface{})
	columns := []string{}

	for queryPager.More() {
		queryResponse, err := queryPager.NextPage(ctx)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, "Query Pager error: "+err.Error())
		}
		log.DefaultLogger.Debug("Result Cosmos DB", "num of response", len(queryResponse.Items))

		for _, item := range queryResponse.Items {
			var itemResponseBody map[string]interface{}
			json.Unmarshal(item, &itemResponseBody)

			if len(columns) == 0 {
				for key := range itemResponseBody {
					if key != "_ts" {
						columns = append(columns, key)
					}
				}
				for _, str := range columns {
					columnData[str] = nil
				}
			}

			//Get time Data
			timeRaw := fmt.Sprintf("%v", itemResponseBody["_ts"])
			log.DefaultLogger.Debug("Result Cosmos DB", "time", timeRaw)
			tsf, _ := strconv.ParseFloat(timeRaw, 64)
			ts := int64(tsf)
			timeData = append(timeData, time.Unix(ts, 0))
			log.DefaultLogger.Debug("Result Cosmos DB", "time", ts)

			for _, str := range columns {
				raw := fmt.Sprintf("%v", itemResponseBody[str])
				log.DefaultLogger.Debug("Result Cosmos DB", "value", raw)
				value, err := strconv.ParseFloat(raw, 64)

				if columnData[str] == nil {
					if err == nil {
						columnData[str] = []float64{}
					} else {
						columnData[str] = []string{}
					}
				}
				if _, isFloat := columnData[str].([]float64); err == nil && isFloat {
					floatArray := columnData[str].([]float64)
					columnData[str] = append(floatArray, value)
				} else {
					stringArray := columnData[str].([]string)
					columnData[str] = append(stringArray, raw)
				}
			}
		}
	}

	log.DefaultLogger.Debug("Result Cosmos DB", "RawData", columnData)
	// add fields.
	frame.Fields = append(frame.Fields, data.NewField("time", nil, timeData))
	for _, str := range columns {
		frame.Fields = append(frame.Fields, data.NewField(str, nil, columnData[str]))
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// when logging at a non-Debug level, make sure you don't include sensitive information in the message
	// (like the *backend.QueryDataRequest)
	log.DefaultLogger.Debug("CheckHealth called")

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	settings := req.PluginContext.DataSourceInstanceSettings

	//Get Secure data
	if primaryKey, primaryKeyExists := settings.DecryptedSecureJSONData["primaryKey"]; primaryKeyExists {
		if endpointUri, endpointUriExists := settings.DecryptedSecureJSONData["endpointUri"]; endpointUriExists {

			cred, err := azcosmos.NewKeyCredential(primaryKey)
			if err != nil {
				log.DefaultLogger.Error("Failed to create a credential: ", err)
				message = "Failed to create a credential: " + err.Error()
				status = backend.HealthStatusError
			}

			// Create a CosmosDB client
			client, err := azcosmos.NewClientWithKey(endpointUri, cred, nil)
			if err != nil {
				log.DefaultLogger.Error("Failed to create Azure Cosmos DB client: ", err)
				message = "Failed to create Azure Cosmos DB client:" + err.Error()
				status = backend.HealthStatusError
			}
			message = "Azure Cosmos DB Client is connected : " + client.Endpoint()
		}
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
