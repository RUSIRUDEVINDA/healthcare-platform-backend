import { useCallback, useEffect, useState, type FormEvent } from 'react';
import axios from 'axios';
import { Navigate } from 'react-router-dom';
import {
  History,
  Loader2,
  Plus,
  Sparkles,
  User,
} from 'lucide-react';
import { patientApi, type PatientProfile } from '../api/patient';
import { symptomApi, type SymptomCheckResponse } from '../api/symptom';

const HISTORY_STORAGE_PREFIX = 'ayarx_symptom_history_v1';
const MAX_HISTORY_ENTRIES = 50;

type SymptomHistoryEntry = {
  id: string;
  createdAt: string;
  symptoms: string;
  optional_context?: string;
  result: SymptomCheckResponse;
};

function readAuthUserId(): string | null {
  try {
    const raw = localStorage.getItem('user');
    if (!raw) return null;
    const u = JSON.parse(raw) as { id?: string; user_id?: string };
    const id = u?.id ?? u?.user_id;
    return typeof id === 'string' && id.length > 0 ? id : null;
  } catch {
    return null;
  }
}

function storageKeyForUser(userId: string): string {
  return `${HISTORY_STORAGE_PREFIX}:${userId}`;
}

function loadHistory(userId: string): SymptomHistoryEntry[] {
  try {
    const raw = localStorage.getItem(storageKeyForUser(userId));
    if (!raw) return [];
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(
      (item): item is SymptomHistoryEntry =>
        item !== null &&
        typeof item === 'object' &&
        typeof (item as SymptomHistoryEntry).id === 'string' &&
        typeof (item as SymptomHistoryEntry).createdAt === 'string' &&
        typeof (item as SymptomHistoryEntry).symptoms === 'string' &&
        (item as SymptomHistoryEntry).result !== null &&
        typeof (item as SymptomHistoryEntry).result === 'object'
    );
  } catch {
    return [];
  }
}

function saveHistory(userId: string, entries: SymptomHistoryEntry[]) {
  try {
    localStorage.setItem(storageKeyForUser(userId), JSON.stringify(entries));
  } catch {
    /* quota or private mode */
  }
}

function formatHistoryWhen(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function previewText(text: string, max = 72): string {
  const t = text.replace(/\s+/g, ' ').trim();
  if (t.length <= max) return t;
  return `${t.slice(0, max)}…`;
}

export default function SymptomChecker() {
  const [role, setRole] = useState<string | null>(null);
  const [profile, setProfile] = useState<PatientProfile | null>(null);
  const [history, setHistory] = useState<SymptomHistoryEntry[]>([]);
  const [selectedHistoryId, setSelectedHistoryId] = useState<string | null>(null);
  const [viewingHistorical, setViewingHistorical] = useState(false);
  const [symptomInput, setSymptomInput] = useState('');
  const [symptomContext, setSymptomContext] = useState('');
  const [symptomLoading, setSymptomLoading] = useState(false);
  const [symptomResult, setSymptomResult] = useState<SymptomCheckResponse | null>(null);
  const [symptomError, setSymptomError] = useState<string | null>(null);
  const [historyOpenMobile, setHistoryOpenMobile] = useState(false);

  const refreshHistoryFromStorage = useCallback(() => {
    const uid = readAuthUserId();
    if (!uid) {
      setHistory([]);
      return;
    }
    const list = loadHistory(uid);
    list.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
    setHistory(list);
  }, []);

  useEffect(() => {
    const raw = localStorage.getItem('user');
    let userRole: string | null = null;
    try {
      userRole = raw ? JSON.parse(raw)?.role ?? null : null;
    } catch {
      userRole = null;
    }
    setRole(userRole);
    if (userRole === 'patient') {
      void patientApi.getProfile().then(setProfile).catch(() => {});
    }
    refreshHistoryFromStorage();
  }, [refreshHistoryFromStorage]);

  const persistNewEntry = useCallback(
    (symptoms: string, optionalContext: string | undefined, result: SymptomCheckResponse) => {
      const uid = readAuthUserId();
      if (!uid) return;

      const entry: SymptomHistoryEntry = {
        id: crypto.randomUUID(),
        createdAt: new Date().toISOString(),
        symptoms,
        optional_context: optionalContext,
        result,
      };

      setHistory((prev) => {
        const next = [entry, ...prev.filter((e) => e.id !== entry.id)].slice(0, MAX_HISTORY_ENTRIES);
        saveHistory(uid, next);
        return next;
      });
      setSelectedHistoryId(entry.id);
      setViewingHistorical(false);
    },
    []
  );

  const openHistoryEntry = useCallback((entry: SymptomHistoryEntry) => {
    setSelectedHistoryId(entry.id);
    setViewingHistorical(true);
    setSymptomInput(entry.symptoms);
    setSymptomContext(entry.optional_context ?? '');
    setSymptomResult(entry.result);
    setSymptomError(null);
    setHistoryOpenMobile(false);
  }, []);

  const startNewCheck = useCallback(() => {
    setSelectedHistoryId(null);
    setViewingHistorical(false);
    setSymptomInput('');
    setSymptomContext('');
    setSymptomResult(null);
    setSymptomError(null);
    setHistoryOpenMobile(false);
  }, []);

  const handleSymptomCheck = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSymptomError(null);
    setSymptomResult(null);
    setSelectedHistoryId(null);
    setViewingHistorical(false);

    const symptoms = symptomInput.trim();
    if (!symptoms) {
      setSymptomError('Please describe the symptoms first.');
      return;
    }

    const ctx = symptomContext.trim() || undefined;

    try {
      setSymptomLoading(true);
      const result = await symptomApi.checkSymptoms({
        symptoms,
        optional_context: ctx,
      });
      setSymptomResult(result);
      persistNewEntry(symptoms, ctx, result);
    } catch (err) {
      console.error('Failed to check symptoms:', err);
      let message = 'We could not analyze your symptoms right now.';
      if (axios.isAxiosError(err)) {
        message = err.response?.data?.error || err.response?.data?.message || message;
      }
      setSymptomError(message);
    } finally {
      setSymptomLoading(false);
    }
  };

  if (role === 'doctor') {
    return <Navigate to="/dashboard" replace />;
  }

  const displayName = profile
    ? `${profile.first_name || 'Patient'} ${profile.last_name || ''}`.trim()
    : 'Patient';

  const isViewingPast = viewingHistorical;

  return (
    <div className="min-h-screen bg-[#f6f8fa] font-sans">
      <div className="flex min-h-screen min-w-0 flex-col">
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-6 sm:px-8 bg-white/80 backdrop-blur-md sticky top-0 z-10">
          <div>
            <p className="text-[10px] font-semibold uppercase tracking-wider text-gray-400">Health tools</p>
            <h2 className="text-lg font-semibold text-gray-900">Symptom checker</h2>
          </div>
          <div className="flex items-center gap-4">
            <p className="hidden sm:block text-sm font-medium text-gray-600">{displayName}</p>
            <div className="h-10 w-10 rounded-full bg-brand-light border border-brand/20 flex items-center justify-center text-brand">
              <User className="h-5 w-5" />
            </div>
          </div>
        </header>

        <main className="flex-1 p-6 sm:p-8 flex flex-col lg:flex-row gap-6 max-w-6xl w-full mx-auto min-w-0">
          {/* History — desktop column */}
          <section
            className={`lg:w-80 shrink-0 flex flex-col rounded-[2rem] border border-gray-100 bg-white shadow-sm overflow-hidden ${
              historyOpenMobile ? 'flex' : 'hidden lg:flex'
            }`}
            aria-labelledby="symptom-history-heading"
          >
            <div className="px-4 py-4 border-b border-gray-100 flex items-center justify-between gap-2">
              <div className="flex items-center gap-2 min-w-0">
                <History className="h-4 w-4 text-brand shrink-0" aria-hidden />
                <h3 id="symptom-history-heading" className="text-sm font-semibold text-gray-900 truncate">
                  Previous checks
                </h3>
              </div>
              <span className="text-xs font-medium text-gray-400 tabular-nums shrink-0">{history.length}</span>
            </div>
            <div className="flex-1 overflow-y-auto max-h-[50vh] lg:max-h-none lg:h-[min(32rem,calc(100vh-12rem))]">
              {history.length === 0 ? (
                <p className="px-4 py-6 text-sm text-gray-500 leading-relaxed">
                  After you run a check, it is saved here on this device so you can review it anytime.
                </p>
              ) : (
                <ul className="p-2 space-y-1">
                  {history.map((entry) => {
                    const active = entry.id === selectedHistoryId;
                    return (
                      <li key={entry.id}>
                        <button
                          type="button"
                          onClick={() => openHistoryEntry(entry)}
                          className={`w-full text-left rounded-xl px-3 py-2.5 text-sm transition-colors border ${
                            active
                              ? 'bg-brand/10 border-brand/20 text-gray-900'
                              : 'border-transparent text-gray-700 hover:bg-gray-50'
                          }`}
                        >
                          <span className="block text-xs text-gray-500 mb-0.5">{formatHistoryWhen(entry.createdAt)}</span>
                          <span className="block font-medium text-gray-900 line-clamp-2">
                            {entry.result?.suggested_specialty || 'Guidance'}
                          </span>
                          <span className="block text-xs text-gray-500 mt-1 line-clamp-2">
                            {previewText(entry.symptoms)}
                          </span>
                        </button>
                      </li>
                    );
                  })}
                </ul>
              )}
            </div>
          </section>

          <div className="flex-1 min-w-0 flex flex-col gap-4">
            <button
              type="button"
              onClick={() => setHistoryOpenMobile((o) => !o)}
              className="lg:hidden flex items-center justify-center gap-2 rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm font-medium text-gray-800 shadow-sm"
            >
              <History className="h-4 w-4" />
              {historyOpenMobile ? 'Hide history' : `Previous checks (${history.length})`}
            </button>

            <div className="rounded-[2rem] border border-gray-100 bg-white shadow-sm overflow-hidden flex-1 min-w-0">
              <div className="flex flex-wrap items-center justify-between gap-3 px-5 py-4 sm:px-6 border-b border-gray-100">
                <div>
                  <h3 className="text-lg font-medium text-gray-900">AI symptom checker</h3>
                  <p className="mt-1 text-sm text-gray-500">
                    Describe what you feel and get a gentle specialty suggestion. This is not a diagnosis.
                  </p>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  {isViewingPast && (
                    <span className="rounded-full bg-amber-50 px-3 py-1 text-xs font-semibold text-amber-900 border border-amber-100">
                      Viewing saved check
                    </span>
                  )}
                  <button
                    type="button"
                    onClick={startNewCheck}
                    className="inline-flex items-center gap-1.5 rounded-full border border-gray-200 bg-gray-50 px-3 py-1.5 text-xs font-semibold text-gray-800 hover:bg-gray-100"
                  >
                    <Plus className="h-3.5 w-3.5" />
                    New check
                  </button>
                  <div className="rounded-full bg-brand/10 px-3 py-1 text-xs font-semibold text-brand shrink-0">Guidance</div>
                </div>
              </div>

              <div className="p-5 sm:p-6 space-y-5">
                <form onSubmit={handleSymptomCheck} className="space-y-4">
                  <div>
                    <label htmlFor="symptom-desc" className="mb-2 block text-sm font-semibold text-gray-700">
                      Symptoms
                    </label>
                    <textarea
                      id="symptom-desc"
                      value={symptomInput}
                      onChange={(event) => {
                        setSymptomInput(event.target.value);
                        if (selectedHistoryId) {
                          setSelectedHistoryId(null);
                          setViewingHistorical(false);
                          setSymptomResult(null);
                        }
                      }}
                      rows={6}
                      placeholder="Example: fever, cough, chest tightness, fatigue, or pain..."
                      className="w-full rounded-2xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm outline-none transition focus:border-brand/30 focus:ring-4 focus:ring-brand/10"
                    />
                  </div>

                  <div>
                    <label htmlFor="symptom-context" className="mb-2 block text-sm font-semibold text-gray-700">
                      Optional context
                    </label>
                    <input
                      id="symptom-context"
                      value={symptomContext}
                      onChange={(event) => {
                        setSymptomContext(event.target.value);
                        if (selectedHistoryId) {
                          setSelectedHistoryId(null);
                          setViewingHistorical(false);
                          setSymptomResult(null);
                        }
                      }}
                      placeholder="Age, duration, any medical history..."
                      className="w-full rounded-full border border-gray-200 bg-gray-50 px-4 py-3 text-sm outline-none transition focus:border-brand/30 focus:ring-4 focus:ring-brand/10"
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={symptomLoading}
                    className="inline-flex w-full items-center justify-center gap-2 rounded-full bg-brand px-5 py-3 text-sm font-medium text-white shadow-sm disabled:cursor-not-allowed disabled:opacity-70"
                  >
                    {symptomLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    {symptomLoading ? 'Checking symptoms...' : 'Check symptoms'}
                  </button>
                </form>

                {symptomError && (
                  <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{symptomError}</div>
                )}

                {symptomResult ? (
                  <div className="space-y-3 rounded-[1.75rem] border border-brand/10 bg-[#f8fffe] p-5">
                    <div className="flex items-center justify-between gap-3">
                      <div>
                        <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-gray-400">Suggested specialty</p>
                        <h4 className="mt-1 text-2xl font-medium text-gray-900">
                          {symptomResult.suggested_specialty || 'General practice'}
                        </h4>
                      </div>
                      <div className="rounded-full bg-white px-3 py-1.5 text-xs font-semibold text-brand border border-brand/10">
                        AI guidance
                      </div>
                    </div>

                    <div className="rounded-2xl border border-gray-100 bg-white p-4">
                      <p className="text-xs font-semibold uppercase tracking-[0.24em] text-gray-400">Preliminary notes</p>
                      <p className="mt-2 text-sm leading-6 text-gray-700">{symptomResult.preliminary_notes}</p>
                    </div>

                    <div className="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-800">
                      {symptomResult.disclaimer}
                    </div>
                  </div>
                ) : (
                  <div className="rounded-2xl border border-dashed border-gray-200 bg-gray-50 px-5 py-6 text-sm text-gray-500">
                    Your symptom guidance will appear here after you run a check, or tap a previous check on the left.
                  </div>
                )}
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
