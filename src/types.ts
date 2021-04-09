import { DataQuery, DataSourceJsonData } from '@grafana/data';

export enum FieldType {
  int16 = 'int16',
  int32 = 'int32',
  int64 = 'int64',
  float32 = 'float32',
  float64 = 'float64',
  bool = 'bool',
}

export interface MyQuery extends DataQuery {
  queryText?: string;
  valueType: string;
}

export const defaultQuery: Partial<MyQuery> = {
  queryText: `WITH 
    from := <datetime>$from,
    to := <datetime>$to,
    max_data_points := <int64>$max_data_points,
SELECT
    {(time := from, value := 1), (time := to, value := 10)}
LIMIT max_data_points;
`,
  valueType: 'int64',
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  uri?: string;
}
