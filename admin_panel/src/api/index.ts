import axios from 'axios';
import type {
  Drone,
  Good,
  ParcelAutomat,
  LockerCell,
  CreateDroneRequest,
  CreateGoodRequest,
  UpdateCellRequest,
  CreateParcelAutomatRequest,
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export const dronesAPI = {
  getAll: () => apiClient.get<Drone[]>('/drones'),
  getById: (id: string) => apiClient.get<Drone>(`/drones/${id}`),
  getStatus: (id: string) => apiClient.get(`/drones/${id}/status`),
  sendCommand: (id: string, command: string) => 
    apiClient.post(`http://localhost:8001/api/drones/${id}/command`, { command }),
  create: (data: CreateDroneRequest) => apiClient.post<Drone>('/drones', data),
  update: (id: string, data: { model: string; ip_address?: string }) =>
    apiClient.put<Drone>(`/drones/${id}`, data),
  updateStatus: (id: string, data: { status: string }) =>
    apiClient.patch<Drone>(`/drones/${id}/status`, data),
  delete: (id: string) => apiClient.delete(`/drones/${id}`),
};

export const goodsAPI = {
  getAll: () => apiClient.get<Good[]>('/goods'),
  getById: (id: string) => apiClient.get<Good>(`/goods/${id}`),
  create: (data: CreateGoodRequest) => apiClient.post<Good[]>('/goods', data),
  update: (id: string, data: Omit<CreateGoodRequest, 'quantity'>) =>
    apiClient.patch<Good>(`/goods/${id}`, data),
  delete: (id: string) => apiClient.delete(`/goods/${id}`),
};

export const parcelAutomatsAPI = {
  getAll: () => apiClient.get<ParcelAutomat[]>('/automats/'),
  getWorking: () => apiClient.get<ParcelAutomat[]>('/automats/working'),
  getById: (id: string) => apiClient.get<ParcelAutomat>(`/automats/${id}`),
  getCells: (id: string) => apiClient.get<LockerCell[]>(`/automats/${id}/cells`),
  updateCell: (automatId: string, cellId: string, data: UpdateCellRequest) =>
    apiClient.patch<LockerCell>(`/automats/${automatId}/cells/${cellId}`, data),
  create: (data: CreateParcelAutomatRequest) =>
    apiClient.post<ParcelAutomat>('/automats/', data),
  update: (id: string, data: { city: string; address: string; ip_address: string; coordinates: string }) =>
    apiClient.put<ParcelAutomat>(`/automats/${id}`, data),
  updateStatus: (id: string, isWorking: boolean) =>
    apiClient.patch<ParcelAutomat>(`/automats/${id}/status`, { is_working: isWorking }),
  delete: (id: string) => apiClient.delete(`/automats/${id}`),
};

export default apiClient;
