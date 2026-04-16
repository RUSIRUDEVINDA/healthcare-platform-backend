import apiClient from './client';

export interface SupportTicket {
    name: string;
    email: string;
    reason: string;
}

export const supportApi = {
    submitReactivationTicket: async (data: SupportTicket) => {
        const response = await apiClient.post('/support/tickets', data);
        return response.data;
    },
    
    listTickets: async () => {
        const response = await apiClient.get('/support/tickets');
        return response.data;
    }
};
