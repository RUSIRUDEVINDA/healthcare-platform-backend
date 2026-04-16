import apiClient from './client';

export const authApi = {
    deactivateAccount: async () => {
        const response = await apiClient.post('/auth/deactivate');
        return response.data;
    }
};
