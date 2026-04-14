import apiClient from './client';

export interface DoctorProfile {
  id: number;
  user_id: string;
  name: string;
  email: string;
  specialization: string | null;
  experience: number | null;
  hospital: string | null;
  nic: string | null;
  slmc_no: string | null;
  created_at: string;
  updated_at: string;
}

export const doctorApi = {
  getProfile: async () => {
    // The main Nginx proxy handles /api/doctors/me
    // Note: The doctor-service handler mounts /doctors group 
    // and RegisterRoutes has protected.GET("/me", ...)
    // So the URL is /api/doctors/me
    const response = await apiClient.get('/doctors/me');
    return response.data?.data as DoctorProfile;
  },

  updateProfile: async (id: number, data: Partial<DoctorProfile>) => {
    const response = await apiClient.put(`/doctors/${id}`, data);
    return response.data?.data as DoctorProfile;
  }
};
