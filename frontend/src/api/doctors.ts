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

interface ApiResponse<T> {
    success: boolean;
    data: T;
    error?: string;
}

export const doctorApi = {
    listDoctors: async (specialization?: string): Promise<Doctor[]> => {
        const params = specialization ? `?specialization=${encodeURIComponent(specialization)}` : '';
        const response = await apiClient.get<ApiResponse<Doctor[]> | Doctor[]>(`doctors${params}`);
        const body = response.data;
        
        if (body && typeof body === 'object' && 'data' in body && Array.isArray(body.data)) return body.data;
        if (Array.isArray(body)) return body;
        return [];
    },

    getDoctor: async (id: number): Promise<Doctor | null> => {
        const response = await apiClient.get<ApiResponse<Doctor> | Doctor>(`doctors/${id}`);
        const body = response.data;
        
        if (body && typeof body === 'object') {
            if ('data' in body && body.data) return body.data;
            if ('id' in body) return body as Doctor;
        }
        return null;
    },
};