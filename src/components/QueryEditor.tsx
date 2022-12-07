import defaults from 'lodash/defaults';

import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms, HorizontalGroup, VerticalGroup, Field, Input } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { defaultQuery, MyDataSourceOptions, MyQuery } from '../types';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {

  onDatabaseChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({...query, database: event.target.value });
  }

  onContainerChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, container: event.target.value });
  };

  onPartitionKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, partitionKey: event.target.value});
  }

  onColumnsChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query} = this.props;
    onChange({ ...query, columns: event.target.value });
  }

  

  render() {
    const { options } = this.props.datasource;
    defaultQuery.database = options.defaultDatabase;
    defaultQuery.container = options.defaultContainer;
    defaultQuery.partitionKey = options.defaultPartitionKey;
    const query = defaults(this.props.query, defaultQuery);
    let { columns, database, container, partitionKey } = query;


    return (
      <div className="gf-form">
        <VerticalGroup>
          <HorizontalGroup>
            <FormField
              width={4}
              value={database}
              onChange={this.onDatabaseChange}
              label="Database"
            />
            <FormField
              labelWidth={8}
              value={container}
              onChange={this.onContainerChange}
              label="Container"
              tooltip="Not used yet"
            />
            <FormField
              width={4}
              value={partitionKey}
              onChange={this.onPartitionKeyChange}
              label="PartitionKey"
              type="Text"
            />
          </HorizontalGroup>
          <Field
            label="Columns"
            description="columns for data, wildcard(*) is available"
          >
            <Input 
                width={80}
                value={columns}
                onChange={this.onColumnsChange}
                label="Data Columns"
                placeholder="column1, column2,...(Comma separated)"
                type="Text"/>
          </Field>
        </VerticalGroup>
      </div>
    );
  }
}
