import { Field } from './field';

export type { Field };

export interface Base {
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  ID: string;
  Name: string;
  UserID: string;
  Tables: null;
}

export interface Table {
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  ID: string;
  Name: string;
  BaseID: string;
  Base: Base;
  Fields: Field[];
}

export interface Record {
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  ID: string;
  TableID: string;
  Table: Table;
  Data: { [key: string]: string };
}

export interface RecordFilter {
  Operator: 'AND' | 'OR';
  Conditions: (RecordCondition | RecordFilter)[];
}

export interface RecordCondition {
  FieldID: string;
  Operator: '=' | '!=' | '>' | '<' | '>=' | '<=' | 'contains' | 'not_contains' | 'starts_with' | 'ends_with' | 'is_empty' | 'is_not_empty';
  Value: any;
}

export interface RecordSort {
  FieldID: string;
  Direction: 'asc' | 'desc';
}

export interface RecordUpdateMessage {
  Type: 'record_created' | 'record_updated' | 'record_deleted';
  TableID: string;
  RecordID: string;
  Record?: Record;
}

export interface RecordListResponse {
  Records: Record[];
  Total: number;
}

export interface RecordFormData {
  [key: string]: string;
}

export interface RecordValidationError {
  FieldID: string;
  Message: string;
}

export function validateRecord(data: RecordFormData, fields: Field[]): RecordValidationError[] {
  const errors: RecordValidationError[] = [];

  fields.forEach(field => {
    const value = data[field.KeyName];

    // Required field validation
    if (field.Required && (value === undefined || value === null || value === '')) {
      errors.push({
        FieldID: field.ID,
        Message: `${field.Name} is required`
      });
      return;
    }

    // Type-specific validation
    if (value !== undefined && value !== null && value !== '') {
      switch (field.Type) {
        case 'number':
          if (isNaN(Number(value))) {
            errors.push({
              FieldID: field.ID,
              Message: `${field.Name} must be a number`
            });
          }
          break;
        case 'boolean':
          if (typeof value !== 'boolean') {
            errors.push({
              FieldID: field.ID,
              Message: `${field.Name} must be a boolean`
            });
          }
          break;
        case 'date':
          if (isNaN(Date.parse(value))) {
            errors.push({
              FieldID: field.ID,
              Message: `${field.Name} must be a valid date`
            });
          }
          break;
      }
    }
  });

  return errors;
} 