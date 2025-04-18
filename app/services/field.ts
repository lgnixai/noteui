import { Field, FieldFormData } from '../types/field';
import { API_BASE_URL } from '../config';

export class FieldService {
  private static instance: FieldService;
  private baseUrl: string;

  private constructor() {
    this.baseUrl = API_BASE_URL;
  }

  public static getInstance(): FieldService {
    if (!FieldService.instance) {
      FieldService.instance = new FieldService();
    }
    return FieldService.instance;
  }

  async getFieldsByTableId(tableId: string): Promise<Field[]> {
    const response = await fetch(`${this.baseUrl}/tables/${tableId}/fields`);
    if (!response.ok) {
      throw new Error('Failed to fetch fields');
    }

    return response.json();
  }

  async getField(fieldId: string): Promise<Field> {
    const response = await fetch(`${this.baseUrl}/fields/${fieldId}`);
    if (!response.ok) {
      throw new Error('Failed to fetch field');
    }

    return response.json();
  }

  async createField(tableId: string, data: FieldFormData): Promise<Field> {
    const response = await fetch(`${this.baseUrl}/tables/${tableId}/fields`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error('Failed to create field');
    }

    return response.json();
  }

  async updateField(fieldId: string, data: FieldFormData): Promise<Field> {
    const response = await fetch(`${this.baseUrl}/fields/${fieldId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error('Failed to update field');
    }

    return response.json();
  }

  async deleteField(fieldId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/fields/${fieldId}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      throw new Error('Failed to delete field');
    }
  }
}

export const fieldService = FieldService.getInstance(); 