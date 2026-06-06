import axios from 'axios';
import { SensorType, GatewayTemplate, RegisterBinding } from './types';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const sensorApi = {
  getAll: () => api.get<SensorType[]>('/sensors'),
  create: (data: Omit<SensorType, 'id'>) => api.post<SensorType>('/sensors', data),
  update: (id: number, data: SensorType) => api.put<SensorType>(`/sensors/${id}`, data),
  delete: (id: number) => api.delete(`/sensors/${id}`),
};

export const templateApi = {
  getAll: () => api.get<GatewayTemplate[]>('/templates'),
  create: (data: Omit<GatewayTemplate, 'id'>) => api.post<GatewayTemplate>('/templates', data),
  update: (id: number, data: GatewayTemplate) => api.put<GatewayTemplate>(`/templates/${id}`, data),
  delete: (id: number) => api.delete(`/templates/${id}`),
};

export const bindingApi = {
  getAll: (templateId?: number) =>
    api.get<RegisterBinding[]>('/bindings', { params: { gatewayTemplateId: templateId } }),
  create: (data: {
    gatewayTemplateId: number;
    sensorTypeId: number;
    startAddress: number;
    label?: string;
  }) => api.post<RegisterBinding>('/bindings', data),
  update: (
    id: number,
    data: {
      gatewayTemplateId: number;
      sensorTypeId: number;
      startAddress: number;
      label?: string;
    }
  ) => api.put<RegisterBinding>(`/bindings/${id}`, data),
  delete: (id: number) => api.delete(`/bindings/${id}`),
  saveAll: (data: {
    gatewayTemplateId: number;
    bindings: {
      sensorTypeId: number;
      startAddress: number;
      label?: string;
    }[];
  }) => api.post<RegisterBinding[]>('/bindings/save-all', data),
};
