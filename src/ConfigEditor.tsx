import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { JsonData, SecureJsonData } from './types';

const { FormField, SecretFormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<JsonData> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = { ...options.jsonData, host: event.target.value };
    onOptionsChange({ ...options, jsonData });
  };

  onPortChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = { ...options.jsonData, port: event.target.value };
    onOptionsChange({ ...options, jsonData });
  };

  onUserChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = { ...options.jsonData, user: event.target.value };
    onOptionsChange({ ...options, jsonData });
  };

  onDatabaseChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = { ...options.jsonData, database: event.target.value };
    onOptionsChange({ ...options, jsonData });
  };

  onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const secureJsonData = { ...options.secureJsonData, password: event.target.value };
    onOptionsChange({ ...options, secureJsonData });
  };

  onResetPassword = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: { ...options.secureJsonFields, password: false },
      secureJsonData: { ...options.secureJsonData, password: '' },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as SecureJsonData;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="Host"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onHostChange}
            value={jsonData.host || ''}
            placeholder="EdgeDB Host"
          />

          <FormField
            type="number"
            label="Port"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPortChange}
            value={jsonData.port || ''}
            placeholder="EdgeDB Port"
          />

          <FormField
            label="Database"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPortChange}
            value={jsonData.database || ''}
            placeholder="EdgeDB Database"
          />

          <FormField
            label="User"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onUserChange}
            value={jsonData.user || ''}
            placeholder="EdgeDB User"
          />

          <SecretFormField
            isConfigured={(secureJsonFields && secureJsonFields.password) as boolean}
            label="Password"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPasswordChange}
            value={secureJsonData.password || ''}
            placeholder="EdgeDB User"
            onReset={this.onResetPassword}
          />
        </div>
      </div>
    );
  }
}
