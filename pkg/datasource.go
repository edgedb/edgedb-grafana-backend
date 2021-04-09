package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edgedb/edgedb-go"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	log.DefaultLogger.Info("------------------- newDataSource --------------------")

	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	ds := &SampleDatasource{
		im: datasource.NewInstanceManager(newDataSourceInstance),
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

// SampleDatasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type SampleDatasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("------------------- query --------------------")
	log.DefaultLogger.Info("QueryData", "request", req)

	instance, err := td.im.Get(req.PluginContext)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("------------------- im.Get failed: %q --------------------", err.Error()))
		return nil, err
	}

	settings, ok := instance.(*PoolWrapper)
	if !ok {
		log.DefaultLogger.Error("------------------- *PoolWrapper type cast failed --------------------")
		return nil, fmt.Errorf("*PoolWrapper type cast failed")
	}

	pool := settings.pool

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := td.query(ctx, pool, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	Format        string `json:"format"`
	QueryText     string `json:"queryText"`
	MaxDataPoints int64  `json:"maxDataPoints"`
}

type QueryResult struct {
	ID    edgedb.UUID `edgedb:"id"`
	Time  time.Time   `edgedb:"time"`
	Value int64       `edgedb:"value"`
}

func (td *SampleDatasource) query(ctx context.Context, pool edgedb.Pool, query backend.DataQuery) backend.DataResponse {
	log.DefaultLogger.Info("------------------- query() ----------------")

	// Unmarshal the json into our queryModel
	var qm queryModel
	response := backend.DataResponse{}
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		log.DefaultLogger.Error(fmt.Sprintf("------------------ json decode query failed: %q ------------", response.Error.Error()))
		return response
	}
	log.DefaultLogger.Info(qm.QueryText)

	// Log a warning if `Format` is empty.
	if qm.Format == "" {
		log.DefaultLogger.Warn("format is empty. defaulting to time series")
	}

	log.DefaultLogger.Info(fmt.Sprintf("------------------ from: %v, to: %v, mdp: %v ------------", query.TimeRange.From, query.TimeRange.To, query.MaxDataPoints))

	args := map[string]interface{}{
		"from":            query.TimeRange.From,
		"to":              query.TimeRange.To,
		"max_data_points": query.MaxDataPoints,
	}

	var results []QueryResult
	response.Error = pool.Query(ctx, qm.QueryText, &results, args)
	if response.Error != nil {
		log.DefaultLogger.Error(fmt.Sprintf("------------------ query failed: %q ------------", response.Error.Error()))
		return response
	}

	log.DefaultLogger.Info(fmt.Sprintf("------------------ query results: %v ------------", len(results)))

	times := make([]time.Time, len(results))
	values := make([]float64, len(results))

	for i, row := range results {
		times[i] = row.Time
		values[i] = float64(row.Value)
		log.DefaultLogger.Info(fmt.Sprintf("%v, %v", row.Time, row.Value))
	}

	// create data frame response
	log.DefaultLogger.Info("------------------- check ----------------")
	frame := data.NewFrame("response")
	log.DefaultLogger.Info("------------------- check ----------------")
	frame.Fields = append(frame.Fields, data.NewField("time", nil, times))
	log.DefaultLogger.Info("------------------- check ----------------")
	frame.Fields = append(frame.Fields, data.NewField("values", nil, values))
	log.DefaultLogger.Info("------------------- check ----------------")
	response.Frames = append(response.Frames, frame)

	log.DefaultLogger.Info(fmt.Sprintf("------------------- response: %#v ----------------", response.Frames[0]))
	log.DefaultLogger.Info("------------------- query() done ----------------")
	return response
}

func getURI(s *backend.DataSourceInstanceSettings) (string, error) {
	log.DefaultLogger.Info("------------------- getURI --------------------")

	var settings struct {
		URI string `json:"uri"`
	}

	err := json.Unmarshal(s.JSONData, &settings)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("---------- json decode uri failed: %q ----------", err.Error()))
		return "", err
	}

	return settings.URI, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SampleDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("------------------- CheckHealth --------------------")
	var (
		pool   edgedb.Pool
		result int64
	)

	uri, err := getURI(req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		goto Error
	}

	pool, err = edgedb.ConnectDSN(ctx, uri, edgedb.Options{MaxConns: 1, MinConns: 1})
	if err != nil {
		goto Error
	}

	err = pool.QueryOne(ctx, "SELECT 1", &result)
	if err != nil {
		goto Error
	}

	err = pool.Close()
	if err != nil {
		goto Error
	}

	return &backend.CheckHealthResult{
		Status:      backend.HealthStatusOk,
		Message:     "Data source is working",
		JSONDetails: []byte{},
	}, nil

Error:
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusError,
		Message: err.Error(),
	}, nil
}

type PoolWrapper struct {
	pool edgedb.Pool
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info(fmt.Sprintf("---------------- newDataSourceInstance: %#v", setting))

	uri, err := getURI(&setting)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	pool, err := edgedb.ConnectDSN(ctx, uri, edgedb.Options{MaxConns: 1, MinConns: 1})
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("----------- could not connect: %q -----------", err.Error()))
		return nil, err
	}

	return &PoolWrapper{pool}, nil
}

func (s *PoolWrapper) Dispose() {
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
	if s.pool != nil {
		err := s.pool.Close()
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("----------- could not close pool: %q -----------", err.Error()))
			log.DefaultLogger.Error(err.Error())
		}
		s.pool = nil
	}
}
