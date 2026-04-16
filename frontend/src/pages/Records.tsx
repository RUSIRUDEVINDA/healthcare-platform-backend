import { useEffect, useMemo, useState, type ChangeEvent, type DragEvent } from 'react';
import {
  Activity,
  AlertCircle,
  ArrowDownToLine,
  Calendar,
  CheckCircle2,
  CloudUpload,
  Eye,
  FileText,
  Image as ImageIcon,
  LogOut,
  Search,
  Shield,
  Trash2,
  Upload,
  User,
  X,
  Stethoscope,
  Scale,
  ChevronDown,
  ClipboardList,
  CreditCard,
} from 'lucide-react';
import axios from 'axios';
import { Link, useLocation } from 'react-router-dom';
import { fileApi, type FileRecord } from '../api/files';

type FilterKey = 'all' | 'prescriptions' | 'reports';
type RecordBucket = Exclude<FilterKey, 'all'>;

const filterLabels: Record<FilterKey, string> = {
  all: 'All',
  prescriptions: 'Prescriptions',
  reports: 'Reports',
};

function sortNewestFirst(items: FileRecord[]) {
  return [...items].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
}

function formatDate(dateStr: string) {
  const date = new Date(dateStr);
  if (Number.isNaN(date.getTime())) return 'Unknown date';
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

function formatFileSize(bytes: number) {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 KB';
  const units = ['B', 'KB', 'MB', 'GB'];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }
  return `${size.toFixed(size >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}

function classifyRecord(file: FileRecord): RecordBucket {
  const name = `${file.original_name} ${file.stored_name}`.toLowerCase();
  if (name.includes('prescription') || name.includes('rx') || name.includes('medication')) return 'prescriptions';
  return 'reports';
}

function getFileTypeLabel(file: FileRecord) {
  return classifyRecord(file) === 'prescriptions' ? 'Prescription' : 'Report';
}

function getFileIcon(file: FileRecord) {
  return file.kind === 'image' || file.mime_type.startsWith('image/')
    ? <ImageIcon className="h-4 w-4" />
    : <FileText className="h-4 w-4" />;
}

function isPreviewable(file: FileRecord) {
  return file.mime_type.startsWith('image/') || file.mime_type === 'application/pdf';
}

export default function Records() {
  const location = useLocation();
  const [files, setFiles] = useState<FileRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeFilter, setActiveFilter] = useState<FilterKey>('all');
  const [dragActive, setDragActive] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewFile, setPreviewFile] = useState<FileRecord | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [role, setRole] = useState<string | null>(null);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);

  const loadFiles = async () => {
    setLoading(true);
    setUploadError(null);
    try {
      const userString = localStorage.getItem('user');
      const user = userString ? JSON.parse(userString) : null;
      const userRole = user?.role ?? null;
      const authUserId = user?.id ?? user?.user_id ?? null;
      setRole(userRole);
      setCurrentUserId(authUserId);

      const myFiles = await fileApi.listMyFiles().catch(() => [] as FileRecord[]);
      const patientFiles = authUserId
        ? await fileApi.listPatientFiles(authUserId).catch(() => [] as FileRecord[])
        : [];
      const merged = Array.from(new Map([...myFiles, ...patientFiles].map((file) => [file.id, file])).values());
      setFiles(sortNewestFirst(merged));
    } catch (error) {
      console.error('Failed to load files:', error);
      setUploadError('We could not load your records right now.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFiles();
  }, []);

  useEffect(() => {
    return () => {
      if (previewUrl) URL.revokeObjectURL(previewUrl);
    };
  }, [previewUrl]);

  const filteredFiles = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    return files.filter((file) => {
      const bucket = classifyRecord(file);
      const matchesQuery =
        query === '' ||
        file.original_name.toLowerCase().includes(query) ||
        file.mime_type.toLowerCase().includes(query) ||
        file.kind.toLowerCase().includes(query);
      return (activeFilter === 'all' || activeFilter === bucket) && matchesQuery;
    });
  }, [activeFilter, files, searchQuery]);

  const prescriptions = files.filter((file) => classifyRecord(file) === 'prescriptions');

  const clearPreview = () => {
    if (previewUrl) URL.revokeObjectURL(previewUrl);
    setPreviewUrl(null);
    setPreviewFile(null);
  };

  const handleUpload = async (fileToUpload: File) => {
    try {
      setUploading(true);
      setUploadError(null);
      if (role === 'doctor') {
        await fileApi.uploadDocument(fileToUpload);
      } else if (currentUserId) {
        await fileApi.uploadPatientFile(currentUserId, fileToUpload);
      } else {
        throw new Error('Patient profile is unavailable');
      }
      setSelectedFile(null);
      await loadFiles();
    } catch (error) {
      console.error('Failed to upload file:', error);
      let message = 'Upload failed. Please try again.';
      if (axios.isAxiosError(error)) {
        message = error.response?.data?.error || message;
      }
      setUploadError(message);
    } finally {
      setUploading(false);
    }
  };

  const handleFileInputChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;
    setSelectedFile(file);
    await handleUpload(file);
    event.target.value = '';
  };

  const handleDrop = async (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragActive(false);
    const file = event.dataTransfer.files?.[0];
    if (!file) return;
    setSelectedFile(file);
    await handleUpload(file);
  };

  const openPreview = async (file: FileRecord) => {
    try {
      setPreviewLoading(true);
      setPreviewFile(file);
      const blob = await fileApi.downloadFile(file.id);
      setPreviewUrl(URL.createObjectURL(blob));
    } catch (error) {
      console.error('Failed to preview file:', error);
      setUploadError('We could not open this file right now.');
      setPreviewFile(null);
    } finally {
      setPreviewLoading(false);
    }
  };

  const downloadRecord = async (file: FileRecord) => {
    try {
      const blob = await fileApi.downloadFile(file.id);
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement('a');
      anchor.href = url;
      anchor.download = file.original_name;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to download file:', error);
      setUploadError('We could not download this file right now.');
    }
  };

  const handleDelete = async (file: FileRecord) => {
    if (!window.confirm(`Delete ${file.original_name}? This cannot be undone.`)) return;
    try {
      await fileApi.deleteFile(file.id);
      await loadFiles();
    } catch (error) {
      console.error('Failed to delete file:', error);
      setUploadError('We could not delete this file right now.');
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    window.location.href = '/auth/login';
  };

  const canPreview = previewFile ? isPreviewable(previewFile) : false;
  const isImagePreview = previewFile?.mime_type.startsWith('image/') ?? false;
  const isPatient = role !== 'doctor';

  return (
    <div className="min-h-screen bg-[#f2fbfa] flex font-sans text-slate-900">
      <aside className="w-60 bg-white/90 backdrop-blur border-r border-teal-100 hidden lg:flex flex-col sticky top-0 h-screen">
        <div className="px-6 pt-6 pb-5">
          <Link to="/dashboard" className="flex items-center gap-2.5">
            <div className="h-9 w-9 rounded-full bg-brand flex items-center justify-center shadow-sm shadow-brand/20">
              <Activity className="h-4 w-4 text-white" />
            </div>
            <span className="text-lg font-medium tracking-tight text-slate-900">MediPulse SriLanka</span>
          </Link>
        </div>

        <nav className="flex-1 px-4 space-y-1">
          <p className="text-[10px] font-semibold text-slate-400 uppercase tracking-[0.32em] px-3 mb-3">Menu</p>
          <Link
            to="/dashboard"
            className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/dashboard'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <Activity className="h-[18px] w-[18px]" /> Dashboard
          </Link>
          <Link
            to="/profile"
            className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/profile'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <User className="h-[18px] w-[18px]" /> Profile
          </Link>
          <Link
            to="/appointments"
            className={`flex items-center justify-between px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/appointments'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <div className="flex items-center gap-3">
              <Calendar className="h-[18px] w-[18px]" /> Appointments
            </div>
            <ChevronDown className="h-4 w-4 opacity-50" />
          </Link>
          {isPatient && (
            <Link
              to="/payments"
              className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
                location.pathname === '/payments'
                  ? 'bg-brand/10 text-brand font-semibold'
                  : 'text-slate-500 hover:bg-teal-50'
              }`}
            >
              <CreditCard className="h-[18px] w-[18px]" /> Payments
            </Link>
          )}
          {isPatient && (
            <Link to="/symptom-checker" className="flex items-center gap-3 px-3 py-2.5 text-slate-500 hover:bg-teal-50 rounded-full text-sm">
              <Stethoscope className="h-[18px] w-[18px]" /> Symptom checker
            </Link>
          )}
          <Link
            to="/records"
            className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/records'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <ClipboardList className="h-[18px] w-[18px]" /> Records
          </Link>
          <Link
            to="/bmi-calculator"
            className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/bmi-calculator'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <Scale className="h-[18px] w-[18px]" /> BMI Calculator
          </Link>
        </nav>

        <div className="p-4 border-t border-teal-100 mx-4 mb-4">
          <button
            onClick={handleLogout}
            className="flex items-center gap-2.5 w-full px-3 py-2.5 text-red-500 hover:bg-red-50 rounded-full transition-colors text-sm font-medium"
          >
            <LogOut className="h-[18px] w-[18px]" /> Sign Out
          </button>
        </div>
      </aside>

      <div className="flex-1 min-h-screen flex flex-col">
        <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-6 md:px-8 sticky top-0 z-10">
          <div>
            <p className="text-[10px] uppercase tracking-[0.32em] text-slate-400">Medical Records</p>
            <h1 className="text-lg font-medium text-slate-900">Your chart</h1>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden md:flex items-center gap-2 rounded-full bg-brand/10 px-3 py-1.5 text-brand text-xs">
              <CheckCircle2 className="h-4 w-4" />
              Secure upload
            </div>
            <div className="w-10 h-10 bg-brand-light rounded-full flex items-center justify-center text-brand border border-brand/20">
              <User className="h-5 w-5" />
            </div>
          </div>
        </header>

        <main className="flex-1 p-4 sm:p-6 lg:p-8 overflow-y-auto">
          <div className="max-w-5xl mx-auto space-y-5">
            <section className="rounded-[1.5rem] border border-slate-200 bg-white shadow-sm overflow-hidden">
              <div className="flex flex-col gap-5 p-5 sm:p-6 lg:flex-row lg:items-start lg:justify-between">
                <div className="space-y-3 max-w-2xl">
                  <div className="inline-flex items-center gap-2 rounded-full bg-brand/10 px-3 py-1 text-[11px] text-brand border border-brand/10">
                    <Shield className="h-3.5 w-3.5" />
                    Secure chart
                  </div>
                  <div>
                    <h2 className="text-xl font-medium tracking-tight text-slate-900">Medical records</h2>
                    <p className="mt-1 text-sm text-slate-600 leading-6">Prescriptions and reports in one compact view.</p>
                  </div>
                  <div className="flex flex-wrap gap-2 text-sm text-slate-600">
                    <span className="inline-flex items-center gap-2 rounded-full bg-slate-50 px-3 py-2 border border-slate-100">
                      <FileText className="h-4 w-4 text-brand" />
                      {files.length} records
                    </span>
                    <span className="inline-flex items-center gap-2 rounded-full bg-slate-50 px-3 py-2 border border-slate-100">
                      <User className="h-4 w-4 text-brand" />
                      {prescriptions.length} prescriptions
                    </span>
                    <span className="inline-flex items-center gap-2 rounded-full bg-slate-50 px-3 py-2 border border-slate-100">
                      <Activity className="h-4 w-4 text-brand" />
                      {files.length - prescriptions.length} reports
                    </span>
                  </div>
                </div>

                <div
                  className={`w-full lg:w-[340px] rounded-[1.5rem] border border-slate-200 bg-slate-50 p-4 ${dragActive ? 'border-brand bg-white' : ''}`}
                  onDragOver={(event) => {
                    event.preventDefault();
                    setDragActive(true);
                  }}
                  onDragLeave={() => setDragActive(false)}
                  onDrop={handleDrop}
                >
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <p className="text-sm text-slate-900">Upload a record</p>
                      <p className="text-xs text-slate-500 mt-1">Prescriptions, reports, notes.</p>
                    </div>
                    <div className="h-11 w-11 rounded-full bg-brand/10 flex items-center justify-center text-brand">
                      <CloudUpload className="h-5 w-5" />
                    </div>
                  </div>

                  <label className="mt-4 block cursor-pointer rounded-[1.25rem] border border-dashed border-slate-200 bg-white px-4 py-5 text-center">
                    <input
                      type="file"
                      className="hidden"
                      onChange={handleFileInputChange}
                      accept=".pdf,.png,.jpg,.jpeg,.webp,.heic,.doc,.docx,.txt,image/*,application/pdf"
                    />
                    <Upload className="mx-auto h-7 w-7 text-brand" />
                    <p className="mt-3 text-sm text-slate-900">{uploading ? 'Uploading...' : 'Click to upload or drop a file'}</p>
                    <p className="mt-1 text-xs text-slate-500">Keeps the original file name.</p>
                  </label>

                  {selectedFile && (
                    <div className="mt-4 rounded-full bg-white px-4 py-3 text-sm text-slate-700 flex items-center gap-3 border border-slate-200">
                      <FileText className="h-4 w-4 text-brand" />
                      <span className="truncate">{selectedFile.name}</span>
                    </div>
                  )}

                  {uploadError && (
                    <div className="mt-4 rounded-[1.25rem] border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 flex items-start gap-2">
                      <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                      <span>{uploadError}</span>
                    </div>
                  )}
                </div>
              </div>
            </section>

            <section className="rounded-[1.5rem] bg-white border border-slate-200 shadow-sm overflow-hidden">
              <div className="flex flex-col gap-4 border-b border-slate-100 p-5 sm:p-6 lg:flex-row lg:items-center lg:justify-between">
                <div>
                  <h3 className="text-base text-slate-900">All records</h3>
                  <p className="text-sm text-slate-500 mt-1">Search prescriptions and reports.</p>
                </div>
                <div className="flex flex-col sm:flex-row gap-3">
                  <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                    <input
                      value={searchQuery}
                      onChange={(event) => setSearchQuery(event.target.value)}
                      placeholder="Search records..."
                      className="w-full sm:w-64 rounded-full border border-slate-200 bg-slate-50 pl-9 pr-4 py-2.5 text-sm outline-none focus:border-brand/40 focus:ring-2 focus:ring-brand/10"
                    />
                  </div>
                  <div className="flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 p-1">
                    {(Object.keys(filterLabels) as FilterKey[]).map((filter) => (
                      <button
                        key={filter}
                        type="button"
                        onClick={() => setActiveFilter(filter)}
                        className={`rounded-full px-3 py-1.5 text-xs transition-colors ${
                          activeFilter === filter ? 'bg-white text-brand shadow-sm' : 'text-slate-500 hover:text-slate-800'
                        }`}
                      >
                        {filterLabels[filter]}
                      </button>
                    ))}
                  </div>
                </div>
              </div>

              {loading ? (
                <div className="p-12 flex items-center justify-center">
                  <div className="h-9 w-9 animate-spin rounded-full border-[3px] border-brand border-t-transparent" />
                </div>
              ) : filteredFiles.length === 0 ? (
                <div className="p-12 text-center">
                  <div className="mx-auto h-14 w-14 rounded-full bg-brand/10 flex items-center justify-center text-brand">
                    <FileText className="h-7 w-7" />
                  </div>
                  <h4 className="mt-4 text-lg text-slate-900">No records found</h4>
                  <p className="mt-2 text-sm text-slate-500 max-w-lg mx-auto">
                    {files.length === 0
                      ? 'Once a doctor uploads a prescription or report, it will show up here.'
                      : 'Try a different search term or filter to narrow the list.'}
                  </p>
                </div>
              ) : (
                <div className="divide-y divide-slate-100">
                  {filteredFiles.map((file) => {
                    const uploadedByYou = file.owner_id === file.uploader_id;

                    return (
                      <div key={file.id} className="p-4 sm:p-5">
                        <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
                          <div className="flex items-start gap-4 min-w-0">
                            <div className="h-11 w-11 rounded-full bg-brand/10 text-brand flex items-center justify-center shrink-0">
                              {getFileIcon(file)}
                            </div>
                            <div className="min-w-0">
                              <div className="flex flex-wrap items-center gap-2">
                                <h4 className="font-medium text-slate-900 truncate max-w-[18rem]">{file.original_name}</h4>
                                <span className="inline-flex items-center rounded-full bg-slate-100 px-2.5 py-1 text-[11px] text-slate-600">
                                  {getFileTypeLabel(file)}
                                </span>
                                {uploadedByYou ? (
                                  <span className="inline-flex items-center gap-1 rounded-full bg-brand/10 px-2.5 py-1 text-[11px] text-brand-dark">
                                    <User className="h-3 w-3" />
                                    Uploaded by you
                                  </span>
                                ) : (
                                  <span className="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-1 text-[11px] text-amber-700">
                                    <Shield className="h-3 w-3" />
                                    Doctor record
                                  </span>
                                )}
                              </div>

                              <div className="mt-2 flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-slate-500">
                                <span className="inline-flex items-center gap-1.5">
                                  <Calendar className="h-4 w-4" />
                                  {formatDate(file.created_at)}
                                </span>
                                <span className="inline-flex items-center gap-1.5">
                                  <FileText className="h-4 w-4" />
                                  {formatFileSize(file.size_bytes)}
                                </span>
                                <span className="inline-flex items-center gap-1.5">
                                  <ImageIcon className="h-4 w-4" />
                                  {file.mime_type}
                                </span>
                              </div>
                            </div>
                          </div>

                          <div className="flex flex-wrap items-center gap-2 xl:justify-end">
                            <button
                              type="button"
                              onClick={() => openPreview(file)}
                              className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white px-4 py-2 text-sm text-slate-700 hover:border-brand/30 hover:text-brand"
                            >
                              <Eye className="h-4 w-4" />
                              View
                            </button>
                            <button
                              type="button"
                              onClick={() => downloadRecord(file)}
                              className="inline-flex items-center gap-2 rounded-full bg-slate-900 px-4 py-2 text-sm text-white hover:bg-brand"
                            >
                              <ArrowDownToLine className="h-4 w-4" />
                              Download
                            </button>
                            <button
                              type="button"
                              onClick={() => handleDelete(file)}
                              className="inline-flex items-center gap-2 rounded-full border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-600 hover:bg-red-100"
                            >
                              <Trash2 className="h-4 w-4" />
                              Delete
                            </button>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </section>
          </div>
        </main>
      </div>

      {(previewFile || previewLoading) && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 px-4 py-6 backdrop-blur-sm">
          <div className="w-full max-w-5xl overflow-hidden rounded-[1.5rem] bg-white shadow-2xl">
            <div className="flex items-center justify-between border-b border-slate-100 px-5 py-4">
              <div>
                <p className="text-xs uppercase tracking-[0.3em] text-slate-400">Record preview</p>
                <h3 className="text-lg text-slate-900">{previewFile?.original_name}</h3>
              </div>
              <button type="button" onClick={clearPreview} className="rounded-full p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-900">
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="bg-slate-50">
              {previewLoading ? (
                <div className="flex min-h-[28rem] items-center justify-center">
                  <div className="h-10 w-10 animate-spin rounded-full border-[3px] border-brand border-t-transparent" />
                </div>
              ) : canPreview && previewUrl && previewFile ? (
                isImagePreview ? (
                  <img src={previewUrl} alt={previewFile.original_name} className="mx-auto max-h-[75vh] w-full object-contain" />
                ) : (
                  <iframe title={previewFile.original_name} src={previewUrl} className="h-[75vh] w-full border-0" />
                )
              ) : (
                <div className="flex min-h-[28rem] flex-col items-center justify-center px-6 text-center">
                  <div className="h-16 w-16 rounded-full bg-white shadow-sm flex items-center justify-center text-brand">
                    <FileText className="h-8 w-8" />
                  </div>
                  <h4 className="mt-4 text-lg text-slate-900">Record preview unavailable</h4>
                  <p className="mt-2 max-w-md text-sm text-slate-500">
                    This record type cannot be previewed inline. Use download to open it locally.
                  </p>
                  {previewFile && (
                    <button
                      type="button"
                      onClick={() => downloadRecord(previewFile)}
                      className="mt-5 inline-flex items-center gap-2 rounded-full bg-slate-900 px-5 py-2.5 text-sm text-white hover:bg-brand"
                    >
                      <ArrowDownToLine className="h-4 w-4" />
                      Download file
                    </button>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
