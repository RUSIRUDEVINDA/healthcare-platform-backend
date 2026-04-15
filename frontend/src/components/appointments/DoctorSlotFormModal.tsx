import { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import type { Slot } from '../../api/appointments';

function toDatetimeLocalValue(iso: string): string {
    const d = new Date(iso);
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
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!isOpen) return;
        setError(null);
        if (mode === 'edit' && slot) {
            setStartLocal(toDatetimeLocalValue(slot.start_time));
            setEndLocal(toDatetimeLocalValue(slot.end_time));
        } else {
            const now = new Date();
            now.setMinutes(now.getMinutes() - (now.getMinutes() % 15), 0, 0);
            const end = new Date(now.getTime() + 30 * 60 * 1000);
            setStartLocal(toDatetimeLocalValue(now.toISOString()));
            setEndLocal(toDatetimeLocalValue(end.toISOString()));
        }
    }, [isOpen, mode, slot]);

    if (!isOpen) return null;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError(null);
        const start = new Date(startLocal);
        const end = new Date(endLocal);
        if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
            setError('Please enter valid start and end times.');
            return;
        }
        if (end <= start) {
            setError('End time must be after start time.');
            return;
        }
        const hosp = profileHospital.trim();
        if (!hosp) {
            setError('Add your hospital on your profile before managing slots.');
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
                        <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1.5">
                            Hospital
                        </label>
                        <p className="text-sm text-gray-800 bg-gray-50 border border-gray-100 rounded-xl px-3 py-2.5">
                            {profileHospital.trim() || '— Set in Profile'}
                        </p>
                        <p className="text-xs text-gray-400 mt-1">
                            Slots are offered at the hospital on your doctor profile.
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
                            value={startLocal}
                            onChange={(e) => setStartLocal(e.target.value)}
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
                            disabled={submitting || !profileHospital.trim()}
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
