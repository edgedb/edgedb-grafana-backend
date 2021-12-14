import {
  DataSourceInstanceSettings,
  ScopedVars,
  VariableModel,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { Options, Query } from './types';

export class DataSource extends DataSourceWithBackend<Query, Options> {
  constructor(instanceSettings: DataSourceInstanceSettings<Options>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: Query, scopedVars: ScopedVars) {
    const templateSrv = getTemplateSrv();
    const template = JSON.stringify(
      Object.fromEntries(
        templateSrv
          .getVariables()
          .map((v: VariableModel) => [v.name, `$${v.name}`])
      )
    );
    const args = JSON.parse(templateSrv.replace(template));
    query.args = {
      ...query.args,
      ...args,
    };
    return query;
  }
}
