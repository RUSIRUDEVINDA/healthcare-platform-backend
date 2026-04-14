import apiClient from './client';

export interface Doctor {
    id: number;
    user_id: string;
    email: string;
    name: string;
    specialization: string;
    experience: number;
    hospital: string;
    nic: string;
    slmc_no: string;
    created_at: string;
    updated_at: string;
}

interface DoctorApiResponse {
    success: boolean;
    data: Doctor[];
    error?: string;
}

export const doctorApi = {
    listDoctors: async (specialization?: string): Promise<Doctor[]> => {
        const params = specialization ? `?specialization=${encodeURIComponent(specialization)}` : '';
        const response = await apiClient.get<DoctorApiResponse>(`/doctors${params}`);
        // doctor-service wraps in { success, data }
        const body = response.data as any;
        if (body?.data && Array.isArray(body.data)) return body.data;
        if (Array.isArray(body)) return body;
        return [];
    },

    getDoctor: async (id: number): Promise<Doctor | null> => {
        const response = await apiClient.get<any>(`/doctors/${id}`);
        const body = response.data;
        return body?.data ?? body ?? null;
    },
};
