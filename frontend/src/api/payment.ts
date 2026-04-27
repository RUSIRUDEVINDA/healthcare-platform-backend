import apiClient from './client';

export interface CheckoutRequest {
    payment_id?: string;
    appointment_id: string;
    patient_id: string;
    amount: number;
    currency: string;
    items: string;
    customer: {
        first_name: string;
        last_name: string;
        email: string;
        phone: string;
        address: string;
        city: string;
        country: string;
    };
}

export interface CheckoutResponse {
    payment_id: string;
    provider: string;
    checkout_url: string;
    fields: Record<string, string>;
}

export interface Payment {
    id: string;
    appointment_id: string;
    patient_id: string;
    amount: number;
    currency: string;
    status: 'pending' | 'completed' | 'failed' | 'refunded' | 'partially_refunded';
    provider: string;
    provider_id: string;
    created_at: string;
    updated_at: string;
}

export const paymentApi = {
    createPayment: async (data: {
        appointment_id: string;
        patient_id: string;
        amount: number;
        currency: string;
    }) => {
        const response = await apiClient.post<{ payment_id: string; status: string; provider?: string }>('v1/payments/', data);
        return response.data;
    },

    checkout: async (data: CheckoutRequest): Promise<CheckoutResponse> => {
        const response = await apiClient.post<CheckoutResponse>('v1/payments/checkout', data);
        return response.data;
    },

    completePayment: async (paymentId: string) => {
        const response = await apiClient.post(`v1/payments/${paymentId}/complete`);
        return response.data;
    },
    
    getPaymentStatus: async (paymentId: string) => {
        const response = await apiClient.get<Payment>(`v1/payments/${paymentId}`);
        return response.data;
    },

    listPayments: async (patientId: string) => {
        const response = await apiClient.get<Payment[]>(`v1/payments/patient/${patientId}`);
        return response.data;
    }
};
