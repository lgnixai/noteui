import { Record, RecordFilter, RecordSort, RecordListResponse, RecordFormData } from '../types/record';
import { API_BASE_URL } from '../config';

export class RecordService {
  private static instance: RecordService;
  private baseUrl: string;

  private constructor() {
    this.baseUrl = API_BASE_URL;
  }

  public static getInstance(): RecordService {
    if (!RecordService.instance) {
      RecordService.instance = new RecordService();
    }
    return RecordService.instance;
  }

  async getRecords(
    tableId: string,
    filter?: RecordFilter,
    sort?: RecordSort[],
    limit: number = 50,
    offset: number = 0
  ): Promise<RecordListResponse> {
    const params = new URLSearchParams({
      limit: limit.toString(),
      offset: offset.toString(),
    });

    if (filter) {
      params.append('filter', JSON.stringify(filter));
    }

    if (sort) {
      params.append('sort', JSON.stringify(sort));
    }

    const response = await fetch(`${this.baseUrl}/tables/${tableId}/records?${params}`);
    if (!response.ok) {
      throw new Error('Failed to fetch records');
    }

    return response.json();
  }

  async getRecord(recordId: string): Promise<Record> {
    const response = await fetch(`${this.baseUrl}/records/${recordId}`);
    if (!response.ok) {
      throw new Error('Failed to fetch record');
    }

    return response.json();
  }

  async createRecord(tableId: string, data: RecordFormData): Promise<Record> {
    const response = await fetch(`${this.baseUrl}/tables/${tableId}/records`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ data }),
    });

    if (!response.ok) {
      throw new Error('Failed to create record');
    }

    return response.json();
  }

  async updateRecord(recordId: string, data: RecordFormData): Promise<Record> {
    const response = await fetch(`${this.baseUrl}/records/${recordId}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ data }),
    });

    if (!response.ok) {
      throw new Error('Failed to update record');
    }

    return response.json();
  }

  async deleteRecord(recordId: string): Promise<void> {
    const response = await fetch(`${this.baseUrl}/records/${recordId}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      throw new Error('Failed to delete record');
    }
  }
}

export const recordService = RecordService.getInstance(); 