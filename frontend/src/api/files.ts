import apiClient from './client';
import type { Appointment } from './appointments';

export type DocumentCategory = 'prescription' | 'medical_report' | 'general';

export interface FileRecord {
  id: string;
  owner_id: string;
  uploader_id: string;
  kind: 'image' | 'document' | string;
  document_category?: DocumentCategory | string;
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

  uploadPatientFile: async (patientId: string, file: File, documentCategory?: DocumentCategory) => {
    const formData = new FormData();
    formData.append('file', file);
    if (documentCategory) {
      formData.append('document_category', documentCategory);
    }

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

const FALLBACK_PATIENT_NAME = 'Unknown patient';

function normalizeDocumentCategory(value?: string | null): DocumentCategory {
  const raw = String(value ?? '').trim().toLowerCase();
  if (raw === 'prescription') return 'prescription';
  if (raw === 'medical_report' || raw === 'report') return 'medical_report';
  return 'general';
}

export function isClinicalFile(file: FileRecord, patientId?: string): boolean {
  if (file.deleted_at) return false;
  if (patientId && file.owner_id && file.owner_id !== patientId) return false;
  if (file.kind !== 'document' && file.kind !== 'image') return false;

  const category = normalizeDocumentCategory(file.document_category);
  return category === 'prescription' || category === 'medical_report' || category === 'general';
}

export function patientLabelFromAppointment(appointment: Appointment): string {
  const first = appointment.patient_first_name?.trim() ?? '';
  const last = appointment.patient_last_name?.trim() ?? '';
  const fullName = [first, last].filter(Boolean).join(' ').trim();
  if (fullName) return fullName;

  const patientId = appointment.patient_id?.trim() ?? '';
  if (!patientId) return FALLBACK_PATIENT_NAME;
  return `Patient ${patientId.slice(0, 8)}`;
}

export const filesApi = {
  listForPatient: async (patientId: string) => fileApi.listPatientFiles(patientId),

  uploadForPatient: async (patientId: string, file: File, documentCategory: DocumentCategory = 'general') =>
    fileApi.uploadPatientFile(patientId, file, documentCategory),

  downloadBlob: async (fileId: string) => fileApi.downloadFile(fileId),
};
