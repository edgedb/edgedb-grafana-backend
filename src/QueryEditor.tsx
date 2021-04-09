import defaults from 'lodash/defaults';

import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from './types';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onQueryTextChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, queryText: event.target.value });
  };

  onValueTypeChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, valueType: event.target.value });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { queryText, valueType } = query;

    return (
      <div className="gf-form">
        <FormField
          width={8}
          value={valueType}
          onChange={this.onValueTypeChange}
          label="Value type"
          tooltip="an EdgeDB scalar type"
          type="string"
        />
        <FormField
          labelWidth={8}
          label="Query Text"
          tooltip="EdgeQL"
          type="string"
          inputEl={
            <textarea
              className="gf-form-input"
              cols={125}
              rows={12}
              value={queryText || ''}
              onChange={this.onQueryTextChange}
            />
          }
        />
      </div>
    );
  }
}
