import defaults from 'lodash/defaults';

import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, JsonData, Query } from './types';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, Query, JsonData>;

export class QueryEditor extends PureComponent<Props> {
  onQueryTextChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, queryText: event.target.value });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { queryText } = query;

    return (
      <div className="gf-form">
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
