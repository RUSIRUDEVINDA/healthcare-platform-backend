import apiClient from './client';

export interface CheckoutRequest {
    appointment_id: string;
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

export const paymentApi = {
    checkout: async (data: CheckoutRequest): Promise<CheckoutResponse> => {
        const response = await apiClient.post<CheckoutResponse>('v1/payments/checkout', data);
        return response.data;
    },
    
    getPaymentStatus: async (paymentId: string) => {
        const response = await apiClient.get(`v1/payments/${paymentId}`);
        return response.data;
    }
};
