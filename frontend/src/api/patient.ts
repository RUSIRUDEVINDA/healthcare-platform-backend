import apiClient from './client';

export interface PatientProfile {
  id: string;
  user_id: string;
  email: string;
  first_name: string;
  last_name: string;
  date_of_birth?: string;
  gender?: string;
  phone_number?: string;
  address?: string;
  emergency_contact?: string;
  blood_group?: string;
  nationality?: string;
  nic?: string;
  created_at: string;
  updated_at: string;
}

export interface UpdateProfileRequest {
  date_of_birth?: string;
  gender?: string;
  phone_number?: string;
  address?: string;
  emergency_contact?: string;
  blood_group?: string;
  nationality?: string;
  nic?: string;
}

export const patientApi = {
  getProfile: async () => {
    const response = await apiClient.get<PatientProfile>('v1/patient/profile');
    return response.data;
  },
  
  updateProfile: async (data: UpdateProfileRequest) => {
    const response = await apiClient.put('v1/patient/profile', data);
    return response.data;
  },
  
  patchProfile: async (data: UpdateProfileRequest) => {
    const response = await apiClient.patch('v1/patient/profile', data);
    return response.data;
  },

  deleteProfile: async () => {
    const response = await apiClient.delete('v1/patient/profile');
    return response.data;
  }
};
