import apiClient from './client';

export interface SupportTicket {
    name: string;
    email: string;
    reason: string;
}

export interface AdminSupportTicket extends SupportTicket {
    id: string;
    status: 'pending' | 'resolved' | 'rejected';
    created_at: string;
}

export const supportApi = {
    submitReactivationTicket: async (data: SupportTicket) => {
        const response = await apiClient.post('/support/tickets', data);
        return response.data;
    },
    
    listTickets: async (): Promise<AdminSupportTicket[]> => {
        const response = await apiClient.get('/support/tickets');
        if (response.data && response.data.data) return response.data.data;
        return response.data || [];
    },

    resolveTicket: async (id: string) => {
        const response = await apiClient.put(`/support/tickets/${id}/resolve`);
        return response.data;
    }
};
