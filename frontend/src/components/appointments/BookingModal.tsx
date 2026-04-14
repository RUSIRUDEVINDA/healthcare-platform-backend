import { useState } from 'react';
import { X, Clock, Video, Home } from 'lucide-react';
import type { Slot, BookAppointmentRequest } from '../../api/appointments';

interface BookingModalProps {
    isOpen: boolean;
    onClose: () => void;
    doctor: {
        id: string;
        name: string;
        specialty: string;
        hospital: string;
        experience: number;
    };
    slots: Slot[];
    onBook: (data: BookAppointmentRequest) => Promise<void>;
}

export default function BookingModal({ isOpen, onClose, doctor, slots, onBook }: BookingModalProps) {
    const [selectedSlot, setSelectedSlot] = useState<Slot | null>(null);
    const [consultationMode, setConsultationMode] = useState<'jitsi' | 'physical'>('jitsi');
    const [paymentMode, setPaymentMode] = useState<'pay_now' | 'pay_later'>('pay_now');
    const [notes, setNotes] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);

    if (!isOpen) return null;

    const handleBookSubmit = async () => {
        if (!selectedSlot) return;

        setIsSubmitting(true);
        try {
            await onBook({
                doctor_id: doctor.id,
                slot_id: selectedSlot.id,
                scheduled_at: selectedSlot.start_time,
                consultation_mode: consultationMode,
                notes,
                payment_mode: paymentMode,
            });
            onClose();
        } catch (error) {
            console.error('Booking failed:', error);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
            <div className="bg-white rounded-3xl w-full max-w-4xl shadow-2xl animate-in fade-in zoom-in duration-300">
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

                <div className="p-8 flex flex-col md:flex-row gap-8">
                    {/* Left Column - Slots */}
                    <div className="flex-[1.3] space-y-6">
                        {/* Slot Selection */}
                        <div>
                            <label className="block text-sm font-semibold text-gray-700 mb-3">Available Slots</label>
                            <div className="grid grid-cols-2 gap-3">
                                {slots.map((slot) => {
                                    const isBooked = slot.is_booked || slot.status === 'booked';
                                    return (
                                        <button
                                            key={slot.id}
                                            type="button"
                                            disabled={isBooked}
                                            onClick={() => setSelectedSlot(slot)}
                                            className={`p-3 rounded-2xl border text-left transition-all relative ${selectedSlot?.id === slot.id
                                                ? 'border-brand bg-brand/5 ring-1 ring-brand'
                                                : isBooked
                                                    ? 'border-gray-100 bg-gray-50 opacity-60 cursor-not-allowed'
                                                    : 'border-gray-200 hover:border-brand/50 bg-white'
                                                }`}
                                        >
                                            <div className="flex items-center text-sm font-medium">
                                                <Clock className={`h-4 w-4 mr-2 ${isBooked ? 'text-gray-400' : 'text-brand'}`} />
                                                <span className={isBooked ? 'line-through text-gray-400' : ''}>
                                                    {new Date(slot.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                                </span>
                                                {isBooked && (
                                                    <span className="ml-auto text-[10px] bg-gray-200 text-gray-500 px-1.5 py-0.5 rounded-md font-bold uppercase tracking-tighter">
                                                        Booked
                                                    </span>
                                                )}
                                            </div>
                                            <div className="text-xs text-gray-500 mt-1">
                                                {new Date(slot.start_time).toLocaleDateString([], { month: 'short', day: 'numeric' })}
                                            </div>
                                        </button>
                                    );
                                })}
                                {slots.length === 0 && (
                                    <p className="col-span-2 text-sm text-gray-500 py-4 text-center bg-gray-50 rounded-2xl">
                                        No slots available for this doctor.
                                    </p>
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Right Column - Details */}
                    <div className="flex-1 flex flex-col space-y-6">
                        {/* Consultation Mode */}
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

                        {/* Reason / Notes */}
                        <div>
                            <label className="block text-sm font-semibold text-gray-700 mb-2">Reason for Visit</label>
                            <textarea
                                value={notes}
                                onChange={(e) => setNotes(e.target.value)}
                                placeholder="Briefly describe your symptoms or reason for the visit..."
                                className="w-full p-4 rounded-2xl border border-gray-200 focus:outline-none focus:ring-2 focus:ring-brand/20 focus:border-brand transition-all resize-none h-24"
                            />
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
                            <button
                                type="button"
                                onClick={handleBookSubmit}
                                disabled={!selectedSlot || isSubmitting}
                                className={`w-full py-4 rounded-2xl font-bold text-white transition-all shadow-lg ${!selectedSlot || isSubmitting
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
    );
}
