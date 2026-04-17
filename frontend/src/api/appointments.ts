import apiClient from './client';

export interface Slot {
    id: string;
    doctor_id: string;
    owner_user_id?: string;
    hospital?: string;
    start_time: string;
    end_time: string;
    is_booked: boolean;
    status?: 'available' | 'booked'; // frontend compat
}

export interface CreateSlotRequest {
    doctor_id?: string;
    start_time: string;
    end_time: string;
    hospital?: string;
}

export interface UpdateSlotRequest {
    start_time?: string;
    end_time?: string;
    hospital?: string;
}

export interface Appointment {
    id: string;
    patient_id: string;
    patient_first_name?: string;
    patient_last_name?: string;
    doctor_id: string;
    doctor_owner_user_id?: string;
    slot_id: string;
    consultation_mode: 'physical' | 'jitsi' | 'video' | 'in-person';
    room_name?: string;
    join_url?: string;
    scheduled_at: string;
    scheduled_time?: string; // alias kept for compat
    duration_minutes?: number;
    status: 'pending' | 'confirmed' | 'cancelled' | 'completed';
    payment_status?: 'pending' | 'paid' | 'overdue' | 'failed' | 'expired';
    payment_due_at?: string;
    paid_at?: string;
    notes?: string;
    consult_fee?: number;
    reason?: string;
    created_at?: string;
    updated_at?: string;
}

export interface BookAppointmentRequest {
    appointment_id?: string;
    doctor_id: string;
    slot_id: string;
    scheduled_at?: string;
    notes?: string;
    payment_mode?: 'pay_now' | 'pay_later';
    consultation_mode?: 'jitsi' | 'physical';
    payment_completed?: boolean;
}

export const appointmentApi = {
    getDoctorSlots: async (doctorId: string, status?: string) => {
        const response = await apiClient.get<Slot[]>(`v1/appointments/doctor/${doctorId}${status ? `?status=${status}` : ''}`);
        return response.data;
    },

    listAppointments: async () => {
        const response = await apiClient.get<Appointment[]>('v1/appointments');
        return response.data;
    },

    bookAppointment: async (data: BookAppointmentRequest) => {
        const response = await apiClient.post<Appointment>('v1/appointments', data);
        return response.data;
    },

    cancelAppointment: async (id: string) => {
        const response = await apiClient.put(`v1/appointments/${id}/cancel`);
        return response.data;
    },

    getAppointmentStatus: async (id: string) => {
        const response = await apiClient.get<Appointment>(`v1/appointments/${id}`);
        return response.data;
    },

    listMySlots: async () => {
        const response = await apiClient.get<Slot[]>('v1/slots');
        return response.data;
    },

    createSlot: async (data: CreateSlotRequest) => {
        const response = await apiClient.post<Slot>('v1/slots', data);
        return response.data;
    },

    updateSlot: async (id: string, data: UpdateSlotRequest) => {
        const response = await apiClient.put<Slot>(`v1/slots/${id}`, data);
        return response.data;
    },

    deleteSlot: async (id: string) => {
        const response = await apiClient.delete(`v1/slots/${id}`);
        return response.data;
    },
};
