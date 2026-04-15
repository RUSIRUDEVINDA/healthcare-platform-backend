import apiClient from './client';

export interface SymptomCheckRequest {
  symptoms: string;
  optional_context?: string;
}

export interface SymptomCheckResponse {
  suggested_specialty: string;
  preliminary_notes: string;
  disclaimer: string;
}

interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: string;
  message?: string;
}

export const symptomApi = {
  checkSymptoms: async (data: SymptomCheckRequest): Promise<SymptomCheckResponse> => {
    const response = await apiClient.post<ApiResponse<SymptomCheckResponse>>('ai/symptom/check', data);
    return response.data.data;
  },
};
