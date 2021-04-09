package main

import (
	"context"
	"encoding/json"

	"github.com/edgedb/edgedb-go"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

type Instance struct {
	URI  string `json:"uri"`
	pool edgedb.Pool
}

func newInstanceWithContext(ctx context.Context, settings backend.DataSourceInstanceSettings) (*Instance, error) {
	var s Instance
	err := json.Unmarshal(settings.JSONData, &s)
	if err != nil {
		return nil, err
	}

	s.pool, err = edgedb.ConnectDSN(ctx, s.URI, edgedb.Options{})
	if err != nil {
		return nil, err
	}

	return &s, err
}

func newInstance(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	ctx := context.Background()
	return newInstanceWithContext(ctx, settings)
}

func (i *Instance) Dispose() {
	if i.pool != nil {
		i.pool.Close()
		i.pool = nil
	}
}
