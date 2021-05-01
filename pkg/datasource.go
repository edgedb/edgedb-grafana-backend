package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

func (td *SampleDatasource) getPool(pluginContext backend.PluginContext) (*edgedb.Pool, error) {
	instance, err := td.im.Get(pluginContext)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("------------------- im.Get failed: %q --------------------", err.Error()))
		return nil, err
	}

	settings, ok := instance.(*PoolWrapper)
	if !ok {
		log.DefaultLogger.Error("------------------- *PoolWrapper type cast failed --------------------")
		return nil, fmt.Errorf("*PoolWrapper type cast failed")
	}

	return settings.pool, nil
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	pool, err := td.getPool(req.PluginContext)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("------------------- could not get pool: %q --------------------", err.Error()))
		return nil, err
	}

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
	IntervalMS    int64  `json:"intervalMs"`
	MaxDataPoints int64  `json:"maxDataPoints"`
}

type QueryResult struct {
	ID    edgedb.UUID `edgedb:"id"`
	Time  time.Time   `edgedb:"time"`
	Value float64     `edgedb:"value"`
	Label string      `edgedb:"label"`
}

func (td *SampleDatasource) query(ctx context.Context, pool *edgedb.Pool, query backend.DataQuery) backend.DataResponse {
	// Unmarshal the json into our queryModel
	var qm queryModel
	response := backend.DataResponse{}
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		log.DefaultLogger.Error(fmt.Sprintf("------------------ json decode query failed: %q ------------", response.Error.Error()))
		return response
	}

	// Log a warning if `Format` is empty.
	if qm.Format == "" {
		log.DefaultLogger.Warn("format is empty. defaulting to time series")
	}

	args := map[string]interface{}{
		"from":            query.TimeRange.From,
		"to":              query.TimeRange.To,
		"interval_ms":     qm.IntervalMS,
		"max_data_points": query.MaxDataPoints,
	}

	var results []QueryResult
	response.Error = pool.Query(ctx, qm.QueryText, &results, args)
	if response.Error != nil {
		log.DefaultLogger.Error(fmt.Sprintf("------------------ query failed: %q ------------", response.Error.Error()))
		return response
	}

	times := make([]time.Time, len(results))
	values := make([]float64, len(results))
	labels := make([]string, len(results))

	for i, row := range results {
		times[i] = row.Time
		values[i] = row.Value
		labels[i] = row.Label
	}

	// create data frame response
	long := data.NewFrame("response")
	long.Fields = append(long.Fields, data.NewField("time", nil, times))
	long.Fields = append(long.Fields, data.NewField("value", nil, values))
	long.Fields = append(long.Fields, data.NewField("label", nil, labels))

	wide, err := data.LongToWide(long, &data.FillMissing{Mode: data.FillModeNull})
	if err != nil {
		response.Error = err
		response.Frames = append(response.Frames, long)
	} else {
		response.Frames = append(response.Frames, wide)
	}

	return response
}

func getOptions(s *backend.DataSourceInstanceSettings) (edgedb.Options, error) {
	var settings struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Database string `json:"database"`
	}

	err := json.Unmarshal(s.JSONData, &settings)
	if err != nil {
		return edgedb.Options{}, err
	}

	port, err := strconv.Atoi(settings.Port)
	if err != nil {
		return edgedb.Options{}, err
	}

	opts := edgedb.Options{
		Hosts:    []string{settings.Host},
		Ports:    []int{port},
		Database: settings.Database,
		User:     settings.User,
		Password: s.DecryptedSecureJSONData["password"],
		MaxConns: 1,
		MinConns: 1,
	}

	log.DefaultLogger.Debug(fmt.Sprintf("----------- connecting to edgedb server using: %#v -----------", opts))

	return opts, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SampleDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var result int64

	pool, err := td.getPool(req.PluginContext)
	if err != nil {
		goto Error
	}

	err = pool.QueryOne(ctx, "SELECT 1", &result)
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
	pool *edgedb.Pool
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	opts, err := getOptions(&setting)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("----------- could not get options: %q -----------", err.Error()))
		return nil, err
	}

	ctx := context.Background()
	pool, err := edgedb.Connect(ctx, opts)
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
