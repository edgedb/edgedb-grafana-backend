package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/edgedb/edgedb-go"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Make sure EdgeDBDatasource implements required interfaces. This is important
// to do since otherwise we will only get a not implemented error response from
// plugin in runtime. In this example datasource instance implements
// backend.QueryDataHandler, backend.CheckHealthHandler, backend.StreamHandler
// interfaces. Plugin should not implement all these interfaces - only those
// which are required for a particular task.  For example if plugin does not
// need streaming functionality then you are free to remove methods that
// implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a
// new datasource instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*EdgeDBDatasource)(nil)
	_ backend.CheckHealthHandler    = (*EdgeDBDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*EdgeDBDatasource)(nil)
	// _ backend.StreamHandler         = (*EdgeDBDatasource)(nil)
)

// NewEdgeDBDatasource creates a new datasource instance.
func NewEdgeDBDatasource(
	ctx context.Context,
	settings backend.DataSourceInstanceSettings,
) (instancemgmt.Instance, error) {
	opts, cloudInstance, err := getOptions(&settings)
	if err != nil {
		log.DefaultLogger.Error(
			fmt.Sprintf("could not get options: %q", err.Error()))
		return nil, err
	}

	msg :=fmt.Sprintf("connecting to edgedb server using: instance=\"%s\", %#v",
		cloudInstance, opts)
	log.DefaultLogger.Debug(msg)

	// the client config options accepts a cloud instance as DSN if present
	client, err := edgedb.CreateClientDSN(ctx, cloudInstance, opts)
	if err != nil {
		log.DefaultLogger.Error(
			fmt.Sprintf("could not connect: %q", err.Error()))
		return nil, err
	}

	return &EdgeDBDatasource{client}, nil
}

// EdgeDBDatasource is an example datasource which can respond to data queries,
// reports its health and has streaming skills.
type EdgeDBDatasource struct {
	client *edgedb.Client
}

func (d *EdgeDBDatasource) Dispose() {
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
	if d.client != nil {
		err := d.client.Close()
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf(
				"could not close client: %q", err.Error()))
		}
		d.client = nil
	}
}

// QueryData handles multiple queries and returns multiple responses.  req
// contains the queries []DataQuery (where each query contains RefID as a unique
// identifier).  The QueryDataResponse contains a map of RefID to the response
// for each query, and each response contains Frames ([]*Frame).
func (d *EdgeDBDatasource) QueryData(
	ctx context.Context, req *backend.QueryDataRequest,
) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	QueryText     string                 `json:"queryText"`
	IntervalMS    int64                  `json:"intervalMs"`
	MaxDataPoints int64                  `json:"maxDataPoints"`
	Args          map[string]interface{} `json:"args"`
}

type QueryResult struct {
	ID    edgedb.UUID        `edgedb:"id"`
	Time  time.Time          `edgedb:"time"`
	Value float64            `edgedb:"value"`
	Label edgedb.OptionalStr `edgedb:"label"`
}

func (d *EdgeDBDatasource) query(
	ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery,
) backend.DataResponse {
	response := backend.DataResponse{}

	var qm queryModel
	// Unmarshal the json into our queryModel
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		log.DefaultLogger.Error(fmt.Sprintf(
			"json decode query failed: %q", response.Error.Error()))
		return response
	}

	args := qm.Args
	if args == nil {
		args = make(map[string]interface{})
	}
	for k, v := range args {
		if v == nil {
			args[k] = edgedb.OptionalStr{}
		}
	}
	args["from"] = query.TimeRange.From
	args["to"] = query.TimeRange.To
	args["interval_ms"] = qm.IntervalMS
	args["max_data_points"] = query.MaxDataPoints
	log.DefaultLogger.Info(fmt.Sprintf("args: %#v", args))

	var results []QueryResult
	response.Error = d.client.Query(ctx, qm.QueryText, &results, args)
	if response.Error != nil {
		log.DefaultLogger.Error(
			fmt.Sprintf("query failed: %q", response.Error.Error()))
		return response
	}

	times := make([]time.Time, len(results))
	values := make([]float64, len(results))
	labels := make([]string, len(results))

	for i, row := range results {
		times[i] = row.Time
		values[i] = row.Value
		if label, ok := row.Label.Get(); ok {
			labels[i] = label
		}
	}

	// create data frame response
	long := data.NewFrame("response")
	long.Fields = append(long.Fields, data.NewField("time", nil, times))
	long.Fields = append(long.Fields, data.NewField("value", nil, values))
	long.Fields = append(long.Fields, data.NewField("label", nil, labels))

	wide, err := data.LongToWide(
		long, &data.FillMissing{Mode: data.FillModeNull})
	if err != nil {
		response.Error = err
		response.Frames = append(response.Frames, long)
	} else {
		response.Frames = append(response.Frames, wide)
	}

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *EdgeDBDatasource) CheckHealth(
	ctx context.Context, req *backend.CheckHealthRequest,
) (*backend.CheckHealthResult, error) {
	var result int64

	err := d.client.QuerySingle(ctx, "SELECT 1", &result)
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

func getOptions(s *backend.DataSourceInstanceSettings) (edgedb.Options, string, error) {
	var settings struct {
		Host        string `json:"host"`
		Port        string `json:"port"`
		User        string `json:"user"`
		Database    string `json:"database"`
		Instance    string `json:"cloudInstance"`
		TlsCA       string `json:"tlsCA"`
		TlsSecurity string `json:"tlsSecurity"`
	}

	err := json.Unmarshal(s.JSONData, &settings)
	if err != nil {
		return edgedb.Options{}, "", err
	}

	var port int
	if settings.Port != "" {
		port, err = strconv.Atoi(settings.Port)
		if err != nil {
			return edgedb.Options{}, "", err
		}
	}

	if settings.TlsSecurity == "" {
		settings.TlsSecurity = "strict"
	}

	opts := edgedb.Options{
		Host:     settings.Host,
		Port:     port,
		Database: settings.Database,
		User:     settings.User,
		Password: edgedb.NewOptionalStr(s.DecryptedSecureJSONData["password"]),
		TLSOptions: edgedb.TLSOptions{
			CA:           []byte(settings.TlsCA),
			SecurityMode: edgedb.TLSSecurityMode(settings.TlsSecurity),
		},
		// below for edgedb cloud
		SecretKey: s.DecryptedSecureJSONData["secretKey"],
	}
	return opts, settings.Instance, nil
}
