export interface Field {
  ID: string;
  TableID: string;
  Name: string;
  KeyName: string;
  Type: 'text' | 'number' | 'boolean' | 'date';
  Required: boolean;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  Table: {
    ID: string;
    Name: string;
    BaseID: string;
    Fields: Field[];
  };
}

export interface FieldFormData {
  name: string;
  keyName: string;
  type: Field['Type'];
  required: boolean;
} 