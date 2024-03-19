import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface Query extends DataQuery {
  queryText?: string;
  args?: Record<string, string>;
}

export const defaultQuery: Partial<Query> = {
  queryText: `WITH
    # query arguments provided by grafana
    from := <datetime>$from,
    to := <datetime>$to,
    max_data_points := <int64>$max_data_points,
    # interval_ms := <int64>$interval_ms,

    # dasboard variables are always strings
    # my_dashboard_variable := <str>$my_dashboard_variable

# the query shape must be {
#   value <float64>,
#   time <datetime>,
#   label <str>,  # optional, enables long format dataframes
# }
SELECT frame := {
    (time := from, value :=  1.0, label := 'a'),
    (time := to,   value := 10.0, label := 'a'),
    (time := from, value :=  7.0, label := 'b'),
    (time := to,   value := -2.3, label := 'b'),
}

# limit results to the current chart window
FILTER
    frame.time >= from AND
    frame.time <= to

# must be orderd by time ASC for labels to work
ORDER BY frame.time ASC

# this seems to be optional, play with it and see what happens.
LIMIT max_data_points;
`,
};

/**
 * These are options configured for each DataSource instance
 */
export interface Options extends DataSourceJsonData {
  host?: string;
  port?: string;
  user?: string;
  database?: string;
  instance?: string;
  tlsCA?: string;
  tlsSecurity?: string;
}

export interface SecureJsonData {
  password?: string;
  secretKey?: string;
}
