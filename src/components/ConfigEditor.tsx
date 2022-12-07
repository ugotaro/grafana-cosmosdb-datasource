import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onPathChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      path: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  onEndpointUriChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        endpointUri: event.target.value,
      },
    });
  };

  onResetEndpointUri = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        endpointUri: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        endpointUri: '',
      },
    });
  };

  onPrimaryKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        primaryKey: event.target.value,
      },
    });
  };

  onResetPrimaryKey = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        primaryKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        primaryKey: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="Path"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onPathChange}
            value={jsonData.database || ''}
            placeholder="json field returned to frontend"
          />
        </div>

        <div className="gf-form-inline">
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.endpointUri) as boolean}
              value={secureJsonData.endpointUri || ''}
              label="Endpoint URI"
              placeholder="secure json field (backend only)"
              labelWidth={6}
              inputWidth={20}
              onReset={this.onResetEndpointUri}
              onChange={this.onEndpointUriChange}
            />
          </div>
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.primaryKey) as boolean}
              value={secureJsonData.primaryKey || ''}
              label="Primary Key"
              placeholder="secure json field (backend only)"
              labelWidth={6}
              inputWidth={20}
              onReset={this.onResetPrimaryKey}
              onChange={this.onPrimaryKeyChange}
            />
          </div>
        </div>
      </div>
    );
  }
}
