import { DataSourceInstanceSettings } from '@grafana/data';

import { MyQuery, MyDataSourceOptions } from './types';
import { DataSourceWithBackend } from '@grafana/runtime';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  options: MyDataSourceOptions;

  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.options = instanceSettings.jsonData;
  }
}
