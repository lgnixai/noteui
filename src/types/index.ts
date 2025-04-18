export interface Base {
  ID: string;
  Name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Table {
  ID: string;
  Name: string;
  baseId: string;
  fields: Field[];
  createdAt: string;
  updatedAt: string;
}

export type FieldType = 'text' | 'number' | 'boolean' | 'date';

export interface Field {
  ID: string;
  Name: string;
  KeyName: string;
  type: FieldType;
  tableId: string;
  createdAt: string;
  updatedAt: string;
}

export interface Record {
  Id: string;
  TableId: string;
  Data: { [key: string]: any };
  createdAt: string;
  updatedAt: string;
}

export interface PaginationParams {
  limit: number;
  offset: number;
}

export interface SortParams {
  field: string;
  direction: 'asc' | 'desc';
}

export interface FilterParams {
  field: string;
  operator: string;
  value: any;
}

export interface WebSocketMessage {
  type: 'record_created' | 'record_updated' | 'record_deleted';
  tableId: string;
  recordId: string;
  data?: any;
}
