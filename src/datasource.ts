import { DataSourceInstanceSettings, DataQueryRequest, DataQueryResponse, VariableModel } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { JsonData, Query } from './types';
import { Observable } from 'rxjs';

export class DataSource extends DataSourceWithBackend<Query, JsonData> {
  constructor(instanceSettings: DataSourceInstanceSettings<JsonData>) {
    super(instanceSettings);
  }

  query(options: DataQueryRequest<Query>): Observable<DataQueryResponse> {
    const templateSrv = getTemplateSrv();
    const template = JSON.stringify(
      Object.fromEntries(templateSrv.getVariables().map((v: VariableModel) => [v.name, `$${v.name}`]))
    );
    const args = JSON.parse(templateSrv.replace(template));

    return super.query({
      ...options,
      targets: options.targets.map((q: Query) => {
        return { ...q, args: args };
      }),
    });
  }
}
