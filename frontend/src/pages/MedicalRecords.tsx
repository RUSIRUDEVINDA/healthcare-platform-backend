import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { AlertCircle, ClipboardList, Download, FileText, FileUp, Loader2, Shield } from 'lucide-react';
import { appointmentApi, type Appointment } from '../api/appointments';
import {
    filesApi,
    isClinicalFile,
    patientLabelFromAppointment,
    type DocumentCategory,
    type FileRecord,
} from '../api/files';

function formatBytes(n: number): string {
    if (n < 1024) return `${n} B`;
    if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
    return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

function categoryLabel(c?: string): string {
    switch (c) {
        case 'prescription':
            return 'Prescription';
        case 'medical_report':
            return 'Medical report';
        default:
            return 'General';
    }
}

function categoryBadgeClass(c?: string): string {
    switch (c) {
        case 'prescription':
            return 'bg-emerald-50 text-emerald-800 ring-emerald-600/15';
        case 'medical_report':
            return 'bg-sky-50 text-sky-800 ring-sky-600/15';
        default:
            return 'bg-gray-50 text-gray-700 ring-gray-500/10';
    }
}

function formatFileKind(kind: string): string {
    return kind === 'image' ? 'Image' : 'Document';
}

function formatDocTimestamp(iso: string): string {
    return new Date(iso).toLocaleString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
    });
}

export default function MedicalRecords() {
    const [role, setRole] = useState<string | null>(null);
    const [userId, setUserId] = useState<string | null>(null);
    const [appointments, setAppointments] = useState<Appointment[]>([]);
    const [selectedPatientId, setSelectedPatientId] = useState<string>('');
    const [files, setFiles] = useState<FileRecord[]>([]);
    const [loading, setLoading] = useState(true);
    const [listLoading, setListLoading] = useState(false);
    const [uploading, setUploading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [docCategory, setDocCategory] = useState<DocumentCategory | ''>('');

    const isDoctor = role === 'doctor';
    const activePatientId = isDoctor ? selectedPatientId : userId ?? '';

    const patientOptions = useMemo(() => {
        const m = new Map<string, string>();
        for (const a of appointments) {
            const id = a.patient_id;
            if (!id || m.has(id)) continue;
            m.set(id, patientLabelFromAppointment(a));
        }
        return Array.from(m.entries()).map(([id, name]) => ({ id, name }));
    }, [appointments]);

    const loadAppointmentsForDoctor = useCallback(async () => {
        try {
            const data = await appointmentApi.listAppointments();
            const list = Array.isArray(data) ? data : [];
            setAppointments(list);
            setSelectedPatientId((prev) => {
                if (prev && list.some((a) => a.patient_id === prev)) return prev;
                return '';
            });
        } catch (e) {
            console.error(e);
        }
    }, []);

    const loadFiles = useCallback(async () => {
        if (!activePatientId) {
            setFiles([]);
            return;
        }
        setListLoading(true);
        setError(null);
        try {
            const list = await filesApi.listForPatient(activePatientId);
            setFiles(list);
        } catch (e: unknown) {
            const msg =
                e && typeof e === 'object' && 'response' in e
                    ? (e as { response?: { data?: { error?: string } } }).response?.data?.error
                    : null;
            setError(msg || 'Could not load documents.');
            setFiles([]);
        } finally {
            setListLoading(false);
        }
    }, [activePatientId]);

    useEffect(() => {
        let cancelled = false;
        (async () => {
            let r: string | null = null;
            let uid: string | null = null;
            try {
                const raw = localStorage.getItem('user');
                const u = raw ? JSON.parse(raw) : null;
                r = u?.role ?? null;
                uid = u?.id ?? null;
            } catch {
                r = null;
                uid = null;
            }
            setRole(r);
            setUserId(uid);
            if (!uid) {
                setLoading(false);
                return;
            }
            setLoading(true);
            if (r === 'doctor') {
                await loadAppointmentsForDoctor();
            }
            if (!cancelled) setLoading(false);
        })();
        return () => {
            cancelled = true;
        };
    }, [loadAppointmentsForDoctor]);

    useEffect(() => {
        if (loading) return;
        void loadFiles();
    }, [loading, loadFiles]);

    const clinicalFiles = useMemo(() => {
        if (!activePatientId) return [];
        return files.filter((f) => isClinicalFile(f, activePatientId));
    }, [files, activePatientId]);

    const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const input = e.target;
        const file = input.files?.[0];
        input.value = '';
        if (!file || !activePatientId || !docCategory) return;
        setUploading(true);
        setError(null);
        try {
            await filesApi.uploadForPatient(activePatientId, file, docCategory);
            await loadFiles();
        } catch (err: unknown) {
            const msg =
                err && typeof err === 'object' && 'response' in err
                    ? (err as { response?: { data?: { error?: string } } }).response?.data?.error
                    : null;
            setError(msg || 'Upload failed.');
        } finally {
            setUploading(false);
        }
    };

    const handleDownload = async (f: FileRecord) => {
        try {
            const blob = await filesApi.downloadBlob(f.id);
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = f.original_name || 'download';
            a.click();
            window.URL.revokeObjectURL(url);
        } catch {
            setError('Download failed.');
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-[#f6f8fa] flex items-center justify-center">
                <div className="flex flex-col items-center gap-3">
                    <Loader2 className="h-10 w-10 text-brand animate-spin" aria-hidden />
                    <p className="text-sm text-gray-500">Loading clinical documents…</p>
                </div>
            </div>
        );
    }

    if (!userId || !role) {
        return (
            <div className="min-h-screen bg-[#f6f8fa] flex items-center justify-center p-6">
                <div className="max-w-md w-full bg-white rounded-2xl border border-gray-100 shadow-sm p-8 text-center">
                    <Shield className="h-10 w-10 text-gray-400 mx-auto mb-4" aria-hidden />
                    <h1 className="text-lg font-semibold text-gray-900">Session required</h1>
                    <p className="text-sm text-gray-500 mt-2 mb-6">Sign in to access medical records and uploads.</p>
                    <Link
                        to="/auth/login"
                        className="inline-flex items-center justify-center rounded-xl bg-gray-900 text-white text-sm font-semibold px-5 py-2.5 hover:bg-brand transition-colors"
                    >
                        Go to sign in
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="flex min-h-screen min-w-0 w-full flex-1 flex-col bg-[#f6f8fa] font-sans">
                <header className="sticky top-0 z-10 w-full border-b border-gray-100 bg-white px-6 sm:px-8 py-4">
                    <div className="w-full max-w-none">
                        <div className="flex flex-wrap items-start justify-between gap-3">
                            <div>
                                <p className="text-[11px] font-semibold uppercase tracking-wider text-gray-400 mb-1">
                                    Clinical documentation
                                </p>
                                <h1 className="text-lg font-semibold tracking-tight text-slate-900">Medical records</h1>
                                <p className="text-sm text-gray-500 mt-1 max-w-2xl leading-relaxed">
                                    {isDoctor
                                        ? 'Upload prescriptions and medical reports for patients with whom you have an active clinical relationship (shared appointments). Files are available to the patient in their portal.'
                                        : 'Review documents shared by your care team. Only clinical files such as prescriptions and reports are listed here.'}
                                </p>
                            </div>
                            <span
                                className={
                                    isDoctor
                                        ? 'shrink-0 inline-flex items-center rounded-full px-3 py-1 text-xs font-medium bg-violet-50 text-violet-800 ring-1 ring-violet-600/15'
                                        : 'shrink-0 inline-flex items-center rounded-full px-3 py-1 text-xs font-medium bg-teal-50 text-teal-800 ring-1 ring-teal-600/15'
                                }
                            >
                                {isDoctor ? 'Clinician view' : 'Patient view'}
                            </span>
                        </div>
                    </div>
                </header>
                <main className="min-w-0 w-full flex-1 overflow-y-auto overflow-x-hidden p-6 sm:p-8">
                    <div className="w-full max-w-none space-y-6">
                        {isDoctor && (
                            <section
                                className="bg-white rounded-2xl border border-gray-100 shadow-sm p-5 sm:p-6"
                                aria-labelledby="patient-context-heading"
                            >
                                <h2 id="patient-context-heading" className="text-sm font-semibold text-gray-900">
                                    Patient context
                                </h2>
                                <p className="text-xs text-gray-500 mt-1 mb-4">
                                    Select the patient you are documenting. Only patients from your appointment history appear
                                    here.
                                </p>
                                <label htmlFor="records-patient" className="sr-only">
                                    Patient
                                </label>
                                <select
                                    id="records-patient"
                                    value={selectedPatientId}
                                    onChange={(e) => {
                                        setSelectedPatientId(e.target.value);
                                        setDocCategory('');
                                    }}
                                    className="w-full max-w-lg px-3.5 py-2.5 rounded-xl border border-gray-200 text-sm bg-white text-gray-900 shadow-sm focus:outline-none focus:ring-2 focus:ring-brand/25 focus:border-brand"
                                >
                                    {patientOptions.length === 0 ? (
                                        <option value="">No eligible patients — book or complete an appointment first</option>
                                    ) : (
                                        <>
                                            <option value="">Select a patient…</option>
                                            {patientOptions.map((p) => (
                                                <option key={p.id} value={p.id}>
                                                    {p.name}
                                                </option>
                                            ))}
                                        </>
                                    )}
                                </select>
                            </section>
                        )}

                        {isDoctor && activePatientId && (
                            <section
                                className="bg-white rounded-2xl border border-gray-100 shadow-sm p-5 sm:p-6"
                                aria-labelledby="upload-heading"
                            >
                                <div className="flex items-start gap-3 mb-4">
                                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-brand/10 text-brand">
                                        <FileUp className="h-4 w-4" aria-hidden />
                                    </div>
                                    <div>
                                        <h2 id="upload-heading" className="text-sm font-semibold text-gray-900">
                                            Add document to patient record
                                        </h2>
                                        <p className="text-xs text-gray-500 mt-0.5">
                                            Classify the file before upload. Use prescriptions for medication orders and medical
                                            reports for lab or imaging summaries.
                                        </p>
                                    </div>
                                </div>
                                <div className="space-y-4 mb-4">
                                    <div>
                                        <label
                                            htmlFor="doc-category"
                                            className="block text-xs font-medium text-gray-600 mb-1.5"
                                        >
                                            Document type
                                        </label>
                                        <select
                                            id="doc-category"
                                            value={docCategory}
                                            onChange={(e) => setDocCategory(e.target.value as DocumentCategory | '')}
                                            className="w-full max-w-lg px-3.5 py-2.5 rounded-xl border border-gray-200 text-sm bg-white shadow-sm focus:outline-none focus:ring-2 focus:ring-brand/25 focus:border-brand"
                                        >
                                            <option value="">Select record type…</option>
                                            <option value="prescription">Prescription</option>
                                            <option value="medical_report">Medical report</option>
                                            <option value="general">General clinical document</option>
                                        </select>
                                    </div>
                                    <div>
                                        <span className="block text-xs font-medium text-gray-600 mb-1.5">File</span>
                                        <input
                                            type="file"
                                            accept=".pdf,.png,.jpg,.jpeg,.webp,application/pdf,image/*"
                                            disabled={uploading || !docCategory}
                                            onChange={handleUpload}
                                            className="sr-only"
                                            id="clinical-file-upload"
                                        />
                                        <div className="flex flex-wrap items-center gap-2">
                                            <label
                                                htmlFor="clinical-file-upload"
                                                className={`inline-flex items-center justify-center rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-sm font-medium text-gray-800 shadow-sm transition-colors ${uploading || !docCategory ? 'pointer-events-none cursor-not-allowed opacity-50' : 'cursor-pointer hover:bg-gray-100'}`}
                                            >
                                                Choose file
                                            </label>
                                            {uploading && (
                                                <span className="inline-flex items-center gap-2 text-sm text-gray-500">
                                                    <Loader2 className="h-4 w-4 animate-spin" aria-hidden />
                                                    Uploading…
                                                </span>
                                            )}
                                        </div>
                                    </div>
                                </div>
                                <p className="text-xs text-gray-400 border-t border-gray-50 pt-4 flex items-start gap-2">
                                    <Shield className="h-3.5 w-3.5 shrink-0 mt-0.5 text-gray-400" aria-hidden />
                                    <span>
                                        Accepted formats: PDF and common image types. Maximum size is enforced by the platform.
                                        Handle all documents in line with your organization&apos;s privacy and retention policies.
                                    </span>
                                </p>
                            </section>
                        )}

                        {error && (
                            <div
                                role="alert"
                                className="flex gap-3 p-4 rounded-xl bg-red-50 text-red-800 text-sm border border-red-100"
                            >
                                <AlertCircle className="h-5 w-5 shrink-0 text-red-600" aria-hidden />
                                <p>{error}</p>
                            </div>
                        )}

                        <section
                            className="bg-white rounded-2xl border border-gray-100 shadow-sm overflow-hidden"
                            aria-labelledby="documents-list-heading"
                        >
                            <div className="px-5 py-4 border-b border-gray-100 flex flex-wrap items-center justify-between gap-2 bg-gray-50/50">
                                <div>
                                    <h2 id="documents-list-heading" className="text-sm font-semibold text-gray-900">
                                        {isDoctor ? 'Documents for selected patient' : 'Your clinical documents'}
                                    </h2>
                                    {activePatientId && !listLoading && (
                                        <p className="text-xs text-gray-500 mt-0.5">
                                            {clinicalFiles.length}{' '}
                                            {clinicalFiles.length === 1 ? 'item' : 'items'}
                                            {isDoctor ? ' in this record' : ''}
                                        </p>
                                    )}
                                </div>
                                {listLoading && (
                                    <span className="inline-flex items-center gap-2 text-xs text-gray-500">
                                        <Loader2 className="h-3.5 w-3.5 animate-spin" aria-hidden />
                                        Refreshing…
                                    </span>
                                )}
                            </div>

                            {!activePatientId ? (
                                <div className="p-10 text-center">
                                    <FileText className="h-10 w-10 text-gray-300 mx-auto mb-3" aria-hidden />
                                    <p className="text-sm font-medium text-gray-800">
                                        {isDoctor ? 'Select a patient to view documents' : 'Unable to load your record'}
                                    </p>
                                    <p className="text-xs text-gray-500 mt-2 max-w-sm mx-auto">
                                        {isDoctor
                                            ? 'Choose a patient from the list above to see existing files or add new ones.'
                                            : 'Your account could not be verified. Try signing out and back in, or contact support.'}
                                    </p>
                                </div>
                            ) : clinicalFiles.length === 0 ? (
                                <div className="p-10 text-center">
                                    <ClipboardList className="h-10 w-10 text-gray-300 mx-auto mb-3" aria-hidden />
                                    <p className="text-sm font-medium text-gray-800">No clinical documents yet</p>
                                    <p className="text-xs text-gray-500 mt-2 max-w-sm mx-auto">
                                        {isDoctor
                                            ? 'When you upload a prescription or report, it will appear in this list for you and the patient.'
                                            : 'When your clinician shares a prescription or medical report, it will appear here for download.'}
                                    </p>
                                </div>
                            ) : (
                                <div className="w-full min-w-0 overflow-x-auto">
                                    <table className="w-full min-w-0 text-left text-sm">
                                        <thead>
                                            <tr className="border-b border-gray-100 text-xs font-semibold uppercase tracking-wide text-gray-400 bg-white">
                                                <th className="min-w-0 px-5 py-3 font-medium">Document</th>
                                                <th className="whitespace-nowrap px-5 py-3 font-medium hidden sm:table-cell">Category</th>
                                                <th className="whitespace-nowrap px-5 py-3 font-medium hidden md:table-cell">Format</th>
                                                <th className="whitespace-nowrap px-5 py-3 font-medium hidden lg:table-cell">Size</th>
                                                <th className="whitespace-nowrap px-5 py-3 font-medium hidden md:table-cell">Added</th>
                                                <th className="w-px whitespace-nowrap px-5 py-3 text-right font-medium">Action</th>
                                            </tr>
                                        </thead>
                                        <tbody className="divide-y divide-gray-50">
                                            {clinicalFiles.map((f) => (
                                                <tr key={f.id} className="hover:bg-gray-50/80 transition-colors">
                                                    <td className="min-w-0 px-5 py-3.5 align-top">
                                                        <div className="flex min-w-0 items-start gap-2">
                                                            <FileText className="mt-0.5 h-4 w-4 shrink-0 text-gray-400" aria-hidden />
                                                            <span
                                                                className="min-w-0 break-words font-medium text-gray-900"
                                                                title={f.original_name}
                                                            >
                                                                {f.original_name}
                                                            </span>
                                                        </div>
                                                        <span
                                                            className={`sm:hidden inline-flex mt-2 rounded-full px-2 py-0.5 text-[10px] font-semibold ring-1 ${categoryBadgeClass(f.document_category)}`}
                                                        >
                                                            {categoryLabel(f.document_category)}
                                                        </span>
                                                    </td>
                                                    <td className="px-5 py-3.5 hidden sm:table-cell">
                                                        <span
                                                            className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ${categoryBadgeClass(f.document_category)}`}
                                                        >
                                                            {categoryLabel(f.document_category)}
                                                        </span>
                                                    </td>
                                                    <td className="px-5 py-3.5 text-gray-600 hidden md:table-cell">
                                                        {formatFileKind(f.kind)}
                                                    </td>
                                                    <td className="px-5 py-3.5 text-gray-600 tabular-nums hidden lg:table-cell">
                                                        {formatBytes(f.size_bytes)}
                                                    </td>
                                                    <td className="px-5 py-3.5 text-gray-600 whitespace-nowrap hidden md:table-cell">
                                                        {formatDocTimestamp(f.created_at)}
                                                    </td>
                                                    <td className="px-5 py-3.5 text-right">
                                                        <button
                                                            type="button"
                                                            onClick={() => handleDownload(f)}
                                                            className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg border border-gray-200 bg-white text-gray-800 text-xs font-semibold shadow-sm hover:border-gray-300 hover:bg-gray-50 transition-colors"
                                                        >
                                                            <Download className="h-3.5 w-3.5" aria-hidden />
                                                            <span className="hidden sm:inline">Download</span>
                                                            <span className="sm:hidden">Get</span>
                                                        </button>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            )}
                        </section>
                    </div>
                </main>
        </div>
    );
}
