import React, { ChangeEvent, PureComponent, SyntheticEvent } from 'react';
import { LegacyForms, TextArea } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { Options, SecureJsonData } from './types';

const { SecretFormField, FormField, Switch } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<Options> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      host: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onPortChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      port: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onUserChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      user: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onDatabaseChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      database: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onTLSCAChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      tlsCA: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onStrictTLSChange = (event: SyntheticEvent<HTMLInputElement, Event>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      tlsSecurity: (event.target as HTMLInputElement).checked
        ? 'secure'
        : 'insecure',
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        password: event.target.value,
      },
    });
  };

  onResetPasssword = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as SecureJsonData;

    return (
      <div className="gf-form-group">
        <div className="gf-form--alt">
          <FormField
            label="Host"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onHostChange}
            value={jsonData.host || ''}
            placeholder="127.0.0.1"
          />

          <FormField
            type="number"
            label="Port"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPortChange}
            value={jsonData.port || ''}
            placeholder="5656"
          />

          <FormField
            label="Database"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onDatabaseChange}
            value={jsonData.database || ''}
            placeholder="edgedb"
          />

          <FormField
            label="User"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onUserChange}
            value={jsonData.user || ''}
            placeholder="edgedb"
          />

          <SecretFormField
            isConfigured={
              (secureJsonFields && secureJsonFields.password) as boolean
            }
            value={secureJsonData.password || ''}
            label="Password"
            placeholder="Password"
            labelWidth={6}
            inputWidth={20}
            onReset={this.onResetPasssword}
            onChange={this.onPasswordChange}
          />

          <Switch
            label="Strict TLS"
            labelClass="width-6"
            onChange={this.onStrictTLSChange}
            checked={
              jsonData.tlsSecurity === undefined ||
              jsonData.tlsSecurity === null ||
              jsonData.tlsSecurity === '' ||
              jsonData.tlsSecurity === 'secure'
            }
          />

          <FormField
            label="TLS CA"
            onChange={this.onTLSCAChange}
            inputEl={
              <TextArea
                label="TLS CA"
                className="width-30"
                value={jsonData.tlsCA || ''}
                placeholder="TLS CA"
              />
            }
          />
        </div>
      </div>
    );
  }
}
