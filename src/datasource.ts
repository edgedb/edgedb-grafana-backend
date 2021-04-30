import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { JsonData, Query } from './types';

export class DataSource extends DataSourceWithBackend<Query, JsonData> {
  constructor(instanceSettings: DataSourceInstanceSettings<JsonData>) {
    super(instanceSettings);
  }
}
