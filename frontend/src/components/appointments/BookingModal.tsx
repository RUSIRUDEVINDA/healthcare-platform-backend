import { useEffect, useState } from 'react';
import { X, Clock, Video, Home, User } from 'lucide-react';
import type { Slot, BookAppointmentRequest } from '../../api/appointments';
import { patientApi, type PatientProfile } from '../../api/patient';

interface BookingModalProps {
    isOpen: boolean;
    onClose: () => void;
    doctor: {
        id: string;
        name: string;
        specialty: string;
        hospital: string;
        experience: number;
        channeling_fee?: number;
    };
    slots: Slot[];
    consultationFee: number;
    hospitalFee: number;
    onBook: (data: BookAppointmentRequest) => Promise<void>;
}

export default function BookingModal({ isOpen, onClose, doctor, slots, consultationFee, hospitalFee, onBook }: BookingModalProps) {
    const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null);
    const [consultationMode, setConsultationMode] = useState<'jitsi' | 'physical'>('jitsi');
    const [paymentMode, setPaymentMode] = useState<'pay_now' | 'pay_later'>('pay_now');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [patientProfile, setPatientProfile] = useState<PatientProfile | null>(null);
    const [patientLoading, setPatientLoading] = useState(false);
    const [patientError, setPatientError] = useState<string | null>(null);
    const [bookingError, setBookingError] = useState<string | null>(null);

    useEffect(() => {
        if (!isOpen) {
            return;
        }

        setSelectedSlot(null);
        setConsultationMode('jitsi');
        setPaymentMode('pay_now');
        setBookingError(null);
    }, [isOpen]);

    useEffect(() => {
        if (!isOpen) {
            return;
        }

        let isMounted = true;
        const fetchPatient = async () => {
            setPatientLoading(true);
            setPatientError(null);
            try {
                const profile = await patientApi.getProfile();
                if (!isMounted) return;
                setPatientProfile(profile);
            } catch (error) {
                console.error('Failed to fetch patient profile:', error);
                if (isMounted) {
                    setPatientError('Unable to load patient details from the database.');
                }
            } finally {
                if (isMounted) {
                    setPatientLoading(false);
                }
            }
        };

        fetchPatient();

        return () => {
            isMounted = false;
        };
    }, [isOpen]);

    if (!isOpen) return null;

    const totalAmount = Number(consultationFee || 0) + Number(hospitalFee || 0);

    const handleBookSubmit = async () => {
        if (!selectedSlot) return;

        setIsSubmitting(true);
        setBookingError(null);
        try {
            await onBook({
                doctor_id: doctor.id,
                slot_id: selectedSlot.id,
                scheduled_at: selectedSlot.start_time,
                consultation_mode: consultationMode,
                payment_mode: paymentMode,
            });
            onClose();
        } catch (error: unknown) {
            const message = error instanceof Error ? error.message : 'Booking failed. Please try again.';
            setBookingError(message);
            console.error('Booking failed:', message, error);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 overflow-y-auto bg-black/50 backdrop-blur-sm">
            <div className="min-h-full flex items-start justify-center p-4 md:p-6">
                <div className="bg-white rounded-3xl w-full max-w-6xl max-h-[90vh] shadow-2xl animate-in fade-in zoom-in duration-300 flex flex-col overflow-hidden">
                    <div className="p-6 border-b border-gray-100 flex justify-between items-center">
                        <div>
                            <div className="flex items-center gap-4">
                                <div className="w-14 h-14 bg-brand/10 rounded-2xl flex items-center justify-center text-brand text-lg font-bold">
                                    {doctor.name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()}
                                </div>
                                <div>
                                    <h2 className="text-xl font-bold text-gray-900">{doctor.name}</h2>
                                    <div className="flex items-center gap-3 mt-1">
                                        <span className="text-sm text-brand font-medium">{doctor.specialty}</span>
                                        <span className="w-1 h-1 bg-gray-300 rounded-full"></span>
                                        <span className="text-sm text-gray-500">{doctor.hospital}</span>
                                        <span className="w-1 h-1 bg-gray-300 rounded-full"></span>
                                        <span className="text-sm text-gray-500">{doctor.experience} yrs exp.</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <button onClick={onClose} className="p-2 hover:bg-gray-100 rounded-full transition-colors">
                            <X className="h-5 w-5 text-gray-500" />
                        </button>
                    </div>

                    <div className="p-6 md:p-8 flex flex-col md:flex-row gap-8 overflow-y-auto flex-1 min-h-0">
                        {/* Left Column - Slots */}
                        <div className="flex-[1.3] space-y-6">
                            {/* Slot Selection */}
                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-3">Available Slots</label>
                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                    {slots.map((slot) => {
                                        const isPast = new Date(slot.start_time) < new Date();
                                        const isBooked = slot.is_booked || slot.status === 'booked' || isPast;
                                        return (
                                            <button
                                                key={slot.id}
                                                type="button"
                                                disabled={isBooked}
                                                onClick={() => setSelectedSlot(slot)}
                                                className={`relative flex min-h-[112px] w-full flex-col items-center justify-center rounded-3xl border px-4 py-4 text-center transition-all ${selectedSlot?.id === slot.id
                                                    ? 'border-brand bg-brand/5 ring-1 ring-brand'
                                                    : isBooked
                                                        ? 'border-gray-100 bg-gray-50 opacity-60 cursor-not-allowed'
                                                        : 'border-gray-200 hover:border-brand/50 bg-white'
                                                    }`}
                                            >
                                                {isBooked && (
                                                    <span className="absolute right-3 top-3 rounded-full bg-gray-200 px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-gray-500">
                                                        {isPast ? 'Expired' : 'Booked'}
                                                    </span>
                                                )}

                                                <div className="flex items-center justify-center text-sm font-medium">
                                                    <Clock className={`h-4 w-4 mr-2 ${isBooked ? 'text-gray-400' : 'text-brand'}`} />
                                                    <span className={isBooked ? 'line-through text-gray-400' : 'text-slate-900'}>
                                                        {new Date(slot.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                                    </span>
                                                </div>
                                                <div className="mt-1 text-xs text-gray-500">
                                                    {new Date(slot.start_time).toLocaleDateString([], { month: 'short', day: 'numeric' })}
                                                </div>
                                                <div className="mt-1 text-[11px] font-medium text-gray-400 truncate max-w-full">
                                                    {slot.hospital || doctor.hospital}
                                                </div>
                                            </button>
                                        );
                                    })}
                                    {slots.length === 0 && (
                                        <p className="sm:col-span-2 text-sm text-gray-500 py-4 text-center bg-gray-50 rounded-full">
                                            No slots available for this doctor.
                                        </p>
                                    )}
                                </div>
                            </div>
                        </div>

                        {/* Right Column - Details */}
                        <div className="flex-1 flex flex-col space-y-5">
                            <div className="rounded-2xl border border-gray-100 bg-gray-50 p-4">
                                <div className="flex items-center justify-between mb-3">
                                    <div>
                                        <p className="text-xs font-bold uppercase tracking-wider text-gray-400">Patient Demographics</p>

                                    </div>
                                    <User className="h-5 w-5 text-brand" />
                                </div>
                                {patientLoading ? (
                                    <div className="py-6 text-sm text-gray-500">Loading patient details...</div>
                                ) : patientError ? (
                                    <div className="rounded-xl bg-red-50 text-red-600 text-sm p-3 border border-red-100">
                                        {patientError}
                                    </div>
                                ) : patientProfile ? (
                                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                                        <div className="bg-white rounded-xl border border-gray-100 p-3">
                                            <div className="text-[10px] uppercase tracking-wider text-gray-400 font-bold">Name</div>
                                            <div className="mt-1 text-sm font-semibold text-gray-900">
                                                {patientProfile.first_name} {patientProfile.last_name}
                                            </div>
                                        </div>
                                        <div className="bg-white rounded-xl border border-gray-100 p-3">
                                            <div className="text-[10px] uppercase tracking-wider text-gray-400 font-bold">Email</div>
                                            <div className="mt-1 text-sm font-semibold text-gray-900 truncate">{patientProfile.email}</div>
                                        </div>
                                        <div className="bg-white rounded-xl border border-gray-100 p-3">
                                            <div className="text-[10px] uppercase tracking-wider text-gray-400 font-bold">Phone</div>
                                            <div className="mt-1 text-sm font-semibold text-gray-900">{patientProfile.phone_number || 'Not set'}</div>
                                        </div>
                                        <div className="bg-white rounded-xl border border-gray-100 p-3">
                                            <div className="text-[10px] uppercase tracking-wider text-gray-400 font-bold">Gender</div>
                                            <div className="mt-1 text-sm font-semibold text-gray-900">{patientProfile.gender || 'Not set'}</div>
                                        </div>
                                        <div className="bg-white rounded-xl border border-gray-100 p-3">
                                            <div className="text-[10px] uppercase tracking-wider text-gray-400 font-bold">Date of Birth</div>
                                            <div className="mt-1 text-sm font-semibold text-gray-900">
                                                {patientProfile.date_of_birth ? new Date(patientProfile.date_of_birth).toLocaleDateString() : 'Not set'}
                                            </div>
                                        </div>
                                        <div className="bg-white rounded-xl border border-gray-100 p-3">
                                            <div className="text-[10px] uppercase tracking-wider text-gray-400 font-bold">Blood Group</div>
                                            <div className="mt-1 text-sm font-semibold text-gray-900">{patientProfile.blood_group || 'Not set'}</div>
                                        </div>
                                        <div className="sm:col-span-2 bg-white rounded-xl border border-gray-100 p-3">
                                            <div className="text-[10px] uppercase tracking-wider text-gray-400 font-bold">Address</div>
                                            <div className="mt-1 text-sm font-semibold text-gray-900">{patientProfile.address || 'Not set'}</div>
                                        </div>
                                        <div className="sm:col-span-2 bg-white rounded-xl border border-gray-100 p-3">
                                            <div className="text-[10px] uppercase tracking-wider text-gray-400 font-bold">Emergency Contact</div>
                                            <div className="mt-1 text-sm font-semibold text-gray-900">{patientProfile.emergency_contact || 'Not set'}</div>
                                        </div>
                                    </div>
                                ) : (
                                    <div className="py-6 text-sm text-gray-500">No patient profile loaded.</div>
                                )}
                            </div>

                            <div>
                                <label className="block text-sm font-semibold text-gray-700 mb-3">Consultation Mode</label>
                                <div className="flex space-x-4">
                                    <button
                                        type="button"
                                        onClick={() => setConsultationMode('jitsi')}
                                        className={`flex-1 flex items-center justify-center p-4 rounded-2xl border transition-all ${consultationMode === 'jitsi'
                                            ? 'border-brand bg-brand/5 ring-1 ring-brand font-semibold text-brand'
                                            : 'border-gray-200 hover:border-brand/50'
                                            }`}
                                    >
                                        <Video className="h-5 w-5 mr-2" /> Video
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setConsultationMode('physical')}
                                        className={`flex-1 flex items-center justify-center p-4 rounded-2xl border transition-all ${consultationMode === 'physical'
                                            ? 'border-brand bg-brand/5 ring-1 ring-brand font-semibold text-brand'
                                            : 'border-gray-200 hover:border-brand/50'
                                            }`}
                                    >
                                        <Home className="h-5 w-5 mr-2" /> In-person
                                    </button>
                                </div>
                            </div>

                            <div className="rounded-2xl border border-gray-100 bg-white p-4">
                                <label className="block text-sm font-semibold text-gray-700 mb-3">Payment Summary</label>
                                <div className="space-y-2 text-sm">
                                    <div className="flex items-center justify-between text-gray-500">
                                        <span>Doctor fee</span>
                                        <span className="font-semibold text-gray-900">Rs. {Number(consultationFee || 0).toFixed(2)}</span>
                                    </div>
                                    <div className="flex items-center justify-between text-gray-500">
                                        <span>Hospital fee</span>
                                        <span className="font-semibold text-gray-900">Rs. {Number(hospitalFee || 0).toFixed(2)}</span>
                                    </div>
                                    <div className="flex items-center justify-between pt-2 border-t border-gray-100">
                                        <span className="font-semibold text-gray-900">Total</span>
                                        <span className="font-bold text-brand">Rs. {totalAmount.toFixed(2)}</span>
                                    </div>
                                </div>
                                <div className="mt-3 rounded-xl bg-gray-50 border border-gray-100 px-3 py-2 text-xs text-gray-500">
                                    This amount is read-only and will be sent to payment.
                                </div>
                            </div>

                            {/* Payment Mode */}
                            <div>
                                <label className="block text-sm font-semibold text-gray-700 mb-3">Payment Method</label>
                                <div className="flex space-x-4">
                                    <button
                                        type="button"
                                        onClick={() => setPaymentMode('pay_later')}
                                        className={`flex-1 flex flex-col items-center justify-center p-4 rounded-2xl border transition-all ${paymentMode === 'pay_later'
                                            ? 'border-brand bg-brand/5 ring-1 ring-brand font-semibold text-brand'
                                            : 'border-gray-200 hover:border-brand/50 text-gray-600'
                                            }`}
                                    >
                                        <span className="text-sm">Pay Later</span>
                                        <span className="text-[10px] opacity-60 font-normal">Expires in 1hr</span>
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setPaymentMode('pay_now')}
                                        className={`flex-1 flex flex-col items-center justify-center p-4 rounded-2xl border transition-all ${paymentMode === 'pay_now'
                                            ? 'border-brand bg-brand/5 ring-1 ring-brand font-semibold text-brand'
                                            : 'border-gray-200 hover:border-brand/50 text-gray-600'
                                            }`}
                                    >
                                        <span className="text-sm">Pay Now</span>
                                        <span className="text-[10px] opacity-60 font-normal">Instant confirm</span>
                                    </button>
                                </div>
                            </div>

                            <div className="mt-auto pt-4 border-t border-gray-100">
                                {bookingError && (
                                    <div className="mb-3 rounded-xl border border-red-100 bg-red-50 px-3 py-2 text-xs text-red-700">
                                        {bookingError}
                                    </div>
                                )}
                                <button
                                    type="button"
                                    onClick={handleBookSubmit}
                                    disabled={!selectedSlot || isSubmitting || patientLoading || !!patientError}
                                    className={`w-full py-4 rounded-2xl font-bold text-white transition-all shadow-lg ${!selectedSlot || isSubmitting || patientLoading || !!patientError
                                        ? 'bg-gray-300 cursor-not-allowed shadow-none'
                                        : 'bg-brand hover:bg-brand-dark shadow-brand/20 active:scale-[0.98]'
                                        }`}
                                >
                                    {isSubmitting ? 'Processing...' : 'Confirm Appointment'}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
