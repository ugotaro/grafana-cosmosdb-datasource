import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  database?: string;
  container?: string;
  partitionKey?: string;
  columns?: string;
}

export const defaultQuery: Partial<MyQuery> = {
  database: "",
  container: "",
  partitionKey: "",
  columns: "*",
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  defaultDatabase?: string;
  defaultContainer?: string;
  defaultPartitionKey?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  endpointUri?: string;
  primaryKey?: string;
}
