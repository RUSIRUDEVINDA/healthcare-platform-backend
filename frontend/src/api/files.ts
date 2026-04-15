import apiClient from './client';

export interface FileRecord {
  id: string;
  owner_id: string;
  uploader_id: string;
  kind: 'image' | 'document' | string;
  storage_provider: 'cloudinary' | 'r2' | string;
  original_name: string;
  stored_name: string;
  mime_type: string;
  size_bytes: number;
  checksum: string;
  cloudinary_public_id?: string;
  cloudinary_url?: string;
  r2_bucket?: string;
  r2_object_key?: string;
  is_public: boolean;
  created_at: string;
  updated_at: string;
  deleted_at?: string | null;
}

export interface FileListResponse {
  files: FileRecord[];
}

export interface FileResponse {
  file: FileRecord;
  message?: string;
}

export const fileApi = {
  listMyFiles: async () => {
    const response = await apiClient.get<FileListResponse>('v1/files');
    return response.data.files ?? [];
  },

  listPatientFiles: async (patientId: string) => {
    const response = await apiClient.get<FileListResponse>(`v1/files/patients/${patientId}/files`);
    return response.data.files ?? [];
  },

  uploadDocument: async (file: File) => {
    const formData = new FormData();
    formData.append('file', file);

    const response = await apiClient.post<FileResponse>('v1/files/documents', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    return response.data;
  },

  uploadPatientFile: async (patientId: string, file: File) => {
    const formData = new FormData();
    formData.append('file', file);

    const response = await apiClient.post<FileResponse>(`v1/files/patients/${patientId}/files`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    return response.data;
  },

  getFile: async (fileId: string) => {
    const response = await apiClient.get<{ file: FileRecord }>(`v1/files/${fileId}`);
    return response.data.file;
  },

  downloadFile: async (fileId: string) => {
    const response = await apiClient.get<Blob>(`v1/files/${fileId}/download`, {
      responseType: 'blob',
    });
    return response.data;
  },

  deleteFile: async (fileId: string) => {
    const response = await apiClient.delete(`v1/files/${fileId}`);
    return response.data;
  },
};
