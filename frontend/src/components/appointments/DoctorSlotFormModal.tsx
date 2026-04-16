import { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import type { Slot } from '../../api/appointments';

function toDatetimeLocalValue(iso: string): string {
    const d = new Date(iso);
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

/** Current local time floored to the minute — minimum for datetime-local inputs. */
function currentMinuteDatetimeLocal(): string {
    const d = new Date();
    d.setSeconds(0, 0);
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export interface DoctorSlotFormModalProps {
    isOpen: boolean;
    onClose: () => void;
    mode: 'create' | 'edit';
    slot?: Slot | null;
    profileHospital: string;
    onSubmit: (payload: { start_time: string; end_time: string; hospital: string }) => Promise<void>;
}

export default function DoctorSlotFormModal({
    isOpen,
    onClose,
    mode,
    slot,
    profileHospital,
    onSubmit,
}: DoctorSlotFormModalProps) {
    const [startLocal, setStartLocal] = useState('');
    const [endLocal, setEndLocal] = useState('');
    const [hospital, setHospital] = useState('');
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!isOpen) return;
        setError(null);
        if (mode === 'edit' && slot) {
            setStartLocal(toDatetimeLocalValue(slot.start_time));
            setEndLocal(toDatetimeLocalValue(slot.end_time));
            setHospital(slot.hospital || profileHospital);
        } else {
            setHospital(profileHospital);
            const start = new Date();
            start.setSeconds(0, 0);
            const step = 15;
            const rem = start.getMinutes() % step;
            if (rem !== 0) start.setMinutes(start.getMinutes() + (step - rem), 0, 0);
            if (start.getTime() <= Date.now()) {
                start.setMinutes(start.getMinutes() + step, 0, 0);
            }
            const end = new Date(start.getTime() + 30 * 60 * 1000);
            setStartLocal(toDatetimeLocalValue(start.toISOString()));
            setEndLocal(toDatetimeLocalValue(end.toISOString()));
        }
    }, [isOpen, mode, slot, profileHospital]);

    if (!isOpen) return null;

    const minStart = currentMinuteDatetimeLocal();
    const minEnd = startLocal && startLocal >= minStart ? startLocal : minStart;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        const start = new Date(startLocal);
        const end = new Date(endLocal);
        if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
            setError('Please enter valid start and end times.');
            return;
        }
        const nowMs = Date.now();
        if (start.getTime() < nowMs) {
            setError('Start time cannot be in the past.');
            return;
        }
        if (end.getTime() < nowMs) {
            setError('End time cannot be in the past.');
            return;
        }
        if (end <= start) {
            setError('End time must be after start time.');
            return;
        }
        const hosp = hospital.trim();
        if (!hosp) {
            setError('Please specify the hospital for this slot.');
            return;
        }
        setSubmitting(true);
        try {
            await onSubmit({
                start_time: start.toISOString(),
                end_time: end.toISOString(),
                hospital: hosp,
            });
            onClose();
        } catch (err: unknown) {
            const msg =
                err && typeof err === 'object' && 'response' in err
                    ? (err as { response?: { data?: { error?: string } } }).response?.data?.error
                    : null;
            setError(msg || 'Could not save slot. Please try again.');
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div className="fixed inset-0 z-[60] flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
            <div className="bg-white rounded-2xl w-full max-w-md shadow-2xl border border-gray-100">
                <div className="flex items-center justify-between px-5 py-4 border-b border-gray-100">
                    <h3 className="text-lg font-bold text-gray-900">
                        {mode === 'create' ? 'New availability slot' : 'Update slot'}
                    </h3>
                    <button
                        type="button"
                        onClick={onClose}
                        className="p-2 rounded-full hover:bg-gray-100 text-gray-500"
                        aria-label="Close"
                    >
                        <X className="h-5 w-5" />
                    </button>
                </div>
                <form onSubmit={handleSubmit} className="p-5 space-y-4">
                    <div>
                        <label htmlFor="slot-hospital" className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">
                            Hospital
                        </label>
                        <input
                            id="slot-hospital"
                            type="text"
                            required
                            placeholder="e.g. Asiri Central, Colombo"
                            value={hospital}
                            onChange={(e) => setHospital(e.target.value)}
                            className="w-full px-3 py-2.5 rounded-xl border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-brand/20 focus:border-brand/40"
                        />
                        <p className="text-xs text-gray-400 mt-1">
                            Specify the hospital location for this specific availability slot.
                        </p>
                    </div>
                    <div>
                        <label htmlFor="slot-start" className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">
                            Start
                        </label>
                        <input
                            id="slot-start"
                            type="datetime-local"
                            required
                            min={minStart}
                            value={startLocal}
                            onChange={(e) => {
                                const v = e.target.value;
                                setStartLocal(v);
                                setEndLocal((prev) => {
                                    if (!v || !prev) return prev;
                                    if (new Date(prev).getTime() <= new Date(v).getTime()) {
                                        const s = new Date(v);
                                        return toDatetimeLocalValue(
                                            new Date(s.getTime() + 30 * 60 * 1000).toISOString()
                                        );
                                    }
                                    return prev;
                                });
                            }}
                            className="w-full px-3 py-2.5 rounded-xl border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-brand/20 focus:border-brand/40"
                        />
                    </div>
                    <div>
                        <label htmlFor="slot-end" className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">
                            End
                        </label>
                        <input
                            id="slot-end"
                            type="datetime-local"
                            required
                            min={minEnd}
                            value={endLocal}
                            onChange={(e) => setEndLocal(e.target.value)}
                            className="w-full px-3 py-2.5 rounded-xl border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-brand/20 focus:border-brand/40"
                        />
                    </div>
                    {error && <p className="text-sm text-red-600 bg-red-50 border border-red-100 rounded-xl px-3 py-2">{error}</p>}
                    <div className="flex gap-2 pt-2">
                        <button
                            type="button"
                            onClick={onClose}
                            className="flex-1 px-4 py-2.5 rounded-xl border border-gray-200 text-sm font-medium text-gray-600 hover:bg-gray-50"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={submitting || !hospital.trim()}
                            className="flex-1 px-4 py-2.5 rounded-xl bg-brand text-white text-sm font-semibold hover:bg-brand-dark disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {submitting ? 'Saving…' : mode === 'create' ? 'Create slot' : 'Save changes'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}
