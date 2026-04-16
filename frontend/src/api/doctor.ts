import apiClient from './client';

export interface Doctor {
  id: number;
  user_id?: string;
  name: string;
  email?: string;
  specialization: string;
  experience: number;
  hospital: string;
  channeling_fee: number;
  nic?: string;
  slmc_no?: string;
  created_at?: string;
  updated_at?: string;
}

export interface DoctorProfile {
  id: number;
  user_id: string;
  name: string;
  email: string;
  specialization: string | null;
  experience: number | null;
  hospital: string | null;
  channeling_fee: number | null;
  nic: string | null;
  slmc_no: string | null;
  created_at: string;
  updated_at: string;
}

export const doctorApi = {
  listDoctors: async () => {
    // Public endpoint: GET /api/doctors -> doctor-service GET /doctors
    const response = await apiClient.get('doctors');
    const data = response.data?.data;
    return (Array.isArray(data) ? data : []) as Doctor[];
  },

  getProfile: async () => {
    // The main Nginx proxy handles /api/doctors/me
    // Note: The doctor-service handler mounts /doctors group 
    // and RegisterRoutes has protected.GET("/me", ...)
    // So the URL is /api/doctors/me
    const response = await apiClient.get('doctors/me');
    return response.data?.data as DoctorProfile;
  },

  updateProfile: async (id: number, data: Partial<DoctorProfile>) => {
    const response = await apiClient.put(`doctors/${id}`, data);
    return response.data?.data as DoctorProfile;
  }
};
