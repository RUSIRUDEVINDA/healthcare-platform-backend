import apiClient from './client';


type ApiEnvelope<T> = {
  success?: boolean;
  data?: T;
  error?: string;
  message?: string;
};

export interface AdminUser {
  id: string;
  email: string;
  role: 'patient' | 'doctor' | 'admin' | string;
  first_name: string;
  last_name: string;
  is_verified: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AdminAppointment {
  id: string;
  patient_id: string;
  doctor_id: string;
  status: string;
  scheduled_at: string;
  reason?: string;
  created_at: string;
  updated_at: string;
}

export interface AdminTransaction {
  id: string;
  user_id: string;
  amount: number;
  currency: string;
  status: string;
  provider: string;
  reference?: string;
  created_at: string;
  updated_at: string;
}

export interface DoctorVerification {
  doctor_id: string;
  verified_by?: string;
  notes?: string;
  status: string;
  verified_at?: string;
  created_at: string;
  updated_at: string;
}

function unwrap<T>(body: ApiEnvelope<T> | T): T {
  if (body && typeof body === 'object' && 'data' in (body as ApiEnvelope<T>)) {
    return (body as ApiEnvelope<T>).data as T;
  }
  return body as T;
}

export const adminApi = {
  listUsers: async (): Promise<AdminUser[]> => {
    const response = await apiClient.get<ApiEnvelope<AdminUser[]> | AdminUser[]>('/admin/users');
    const data = unwrap(response.data);
    return Array.isArray(data) ? data : [];
  },

  listAppointments: async (): Promise<AdminAppointment[]> => {
    const response = await apiClient.get<ApiEnvelope<AdminAppointment[]> | AdminAppointment[]>('/admin/appointments');
    const data = unwrap(response.data);
    return Array.isArray(data) ? data : [];
  },

  listTransactions: async (): Promise<AdminTransaction[]> => {
    const response = await apiClient.get<ApiEnvelope<AdminTransaction[]> | AdminTransaction[]>('/admin/transactions');
    const data = unwrap(response.data);
    return Array.isArray(data) ? data : [];
  },

  verifyDoctor: async (doctorId: string, notes?: string): Promise<DoctorVerification> => {
    const response = await apiClient.put<ApiEnvelope<DoctorVerification> | DoctorVerification>(`/admin/doctors/${doctorId}/verify`, {
      notes: notes ?? '',
    });
    return unwrap(response.data);
  },

  deactivateUser: async (userId: string): Promise<string> => {
    const response = await apiClient.delete<ApiEnvelope<unknown>>(`/admin/users/${userId}`);
    return response.data?.message ?? 'User deactivated';
  },

  reactivateUser: async (userId: string): Promise<string> => {
    const response = await apiClient.post<ApiEnvelope<unknown>>(`/admin/users/${userId}/reactivate`);
    return response.data?.message ?? 'User reactivated';
  },

  syncData: async (): Promise<void> => {
    await apiClient.post('/admin/sync');
  },

  createUser: async (user: Partial<AdminUser>): Promise<AdminUser> => {
    const response = await apiClient.post<ApiEnvelope<AdminUser>>('/admin/users', user);
    return unwrap(response.data);
  },

  updateUser: async (id: string, user: Partial<AdminUser>): Promise<void> => {
    await apiClient.put(`/admin/users/${id}`, user);
  },

  createAppointment: async (appt: Partial<AdminAppointment>): Promise<AdminAppointment> => {
    const response = await apiClient.post<ApiEnvelope<AdminAppointment>>('/admin/appointments', appt);
    return unwrap(response.data);
  },

  updateAppointment: async (id: string, appt: Partial<AdminAppointment>): Promise<void> => {
    await apiClient.put(`/admin/appointments/${id}`, appt);
  },

  cancelAppointment: async (id: string): Promise<void> => {
    await apiClient.delete(`/admin/appointments/${id}`);
  },
};
