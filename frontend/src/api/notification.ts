import axios from 'axios';

export interface ServiceHealth {
  service: string;
  status: string;
}

export const notificationApi = {
  getHealth: async (): Promise<ServiceHealth> => {
    const response = await axios.get<ServiceHealth>('/health/notification');
    return response.data;
  },

  getReady: async (): Promise<ServiceHealth> => {
    const response = await axios.get<ServiceHealth>('/health/notification/ready');
    return response.data;
  },
};
