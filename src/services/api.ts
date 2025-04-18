import axios from 'axios';
import { Base, Table, Field, Record, PaginationParams, SortParams, FilterParams } from '../types';

const API_BASE_URL = 'http://localhost:8080/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Base API
export const baseApi = {
  getAll: () => api.get<Base[]>('/bases'),
  getById: (id: string) => api.get<Base>(`/bases/${id}`),
  create: (data: { name: string; description?: string }) => api.post<Base>('/bases', data),
  update: (id: string, data: { name: string; description?: string }) => api.put<Base>(`/bases/${id}`, data),
  delete: (id: string) => api.delete(`/bases/${id}`),
};

// Table API
export const tableApi = {
  getAll: (baseId: string) => api.get<Table[]>(`/bases/${baseId}/tables`),
  getById: (baseId: string, tableId: string) => api.get<Table>(`/bases/${baseId}/tables/${tableId}`),
  create: (baseId: string, data: { name: string }) => api.post<Table>(`/bases/${baseId}/tables`, data),
  update: (baseId: string, tableId: string, data: { name: string }) =>
    api.put<Table>(`/bases/${baseId}/tables/${tableId}`, data),
  delete: (baseId: string, tableId: string) => api.delete(`/bases/${baseId}/tables/${tableId}`),
};

// Field API
export const fieldApi = {
  getAll: (baseId: string, tableId: string) => api.get<Field[]>(`/bases/${baseId}/tables/${tableId}/fields`),
  create: (baseId: string, tableId: string, data: { Name: string; Type: string; KeyName: string }) =>
    api.post<Field>(`/bases/${baseId}/tables/${tableId}/fields`, data),
  update: (baseId: string, tableId: string, fieldId: string, data: { name: string; type: string }) =>
    api.put<Field>(`/bases/${baseId}/tables/${tableId}/fields/${fieldId}`, data),
  delete: (baseId: string, tableId: string, fieldId: string) =>
    api.delete(`/bases/${baseId}/tables/${tableId}/fields/${fieldId}`),
};

// Record API
export const recordApi = {
  getAll: (
    baseId: string,
    tableId: string,
    params: {
      pagination?: PaginationParams;
      sort?: SortParams[];
      filter?: FilterParams[];
    }
  ) => {
    const queryParams = new URLSearchParams();
    if (params.pagination) {
      queryParams.append('limit', params.pagination.limit.toString());
      queryParams.append('offset', params.pagination.offset.toString());
    }
    if (params.sort) {
      queryParams.append('sort', JSON.stringify(params.sort));
    }
    if (params.filter) {
      queryParams.append('filter', JSON.stringify(params.filter));
    }
    return api.get<{ records: Record[]; total: number }>(
      `/bases/${baseId}/tables/${tableId}/records?${queryParams.toString()}`
    );
  },
  getById: (baseId: string, tableId: string, recordId: string) =>
    api.get<Record>(`/bases/${baseId}/tables/${tableId}/records/${recordId}`),
  create: (baseId: string, tableId: string, data: Record<string, any>) =>
    api.post<Record>(`/bases/${baseId}/tables/${tableId}/records`, { data }),
  update: (baseId: string, tableId: string, recordId: string, data: Record<string, any>) =>
    api.put<Record>(`/bases/${baseId}/tables/${tableId}/records/${recordId}`, { data }),
  delete: (baseId: string, tableId: string, recordId: string) =>
    api.delete(`/bases/${baseId}/tables/${tableId}/records/${recordId}`),
};
