import apiClient from './client';
import type { Appointment } from './appointments';

export interface FileRecord {
    id: string;
    owner_id: string;
    uploader_id: string;
    kind: 'image' | 'document';
    document_category?: string;
    storage_provider: string;
    original_name: string;
    mime_type: string;
    size_bytes: number;
    cloudinary_url?: string;
    created_at: string;
}

export type DocumentCategory = 'prescription' | 'medical_report' | 'general';

export const filesApi = {
    listForPatient: async (patientId: string): Promise<FileRecord[]> => {
        const res = await apiClient.get<{ files: FileRecord[] }>(`v1/files/patients/${encodeURIComponent(patientId)}/files`);
        return res.data.files ?? [];
    },

    uploadForPatient: async (patientId: string, file: File, documentCategory: DocumentCategory) => {
        const form = new FormData();
        form.append('file', file);
        form.append('document_category', documentCategory);
        const res = await apiClient.post<{ file: FileRecord; message?: string }>(
            `v1/files/patients/${encodeURIComponent(patientId)}/files`,
            form
        );
        return res.data;
    },

    downloadBlob: async (fileId: string): Promise<Blob> => {
        const res = await apiClient.get<Blob>(`v1/files/${encodeURIComponent(fileId)}/download`, {
            responseType: 'blob',
        });
        return res.data;
    },
};

/** Hide self-uploaded profile photos from the medical-records list. */
export function isClinicalFile(f: FileRecord, patientUserId: string): boolean {
    if (f.kind === 'document') return true;
    if (f.document_category === 'prescription' || f.document_category === 'medical_report') return true;
    if (f.kind === 'image' && f.uploader_id !== patientUserId) return true;
    return false;
}

export function patientLabelFromAppointment(appt: Appointment): string {
    const f = appt.patient_first_name?.trim() || '';
    const l = appt.patient_last_name?.trim() || '';
    if (f || l) return `${f} ${l}`.trim();
    const id = appt.patient_id || '';
    return id.length > 10 ? `Patient ${id.slice(0, 8)}…` : 'Patient';
}
