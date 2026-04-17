import { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { CheckCircle2, XCircle, ArrowLeft, Calendar } from 'lucide-react';
import { appointmentApi, type BookAppointmentRequest } from '../api/appointments';
import { paymentApi } from '../api/payment';

interface PaymentStatusProps {
    type: 'success' | 'cancel';
}

type PendingBookingDraft = BookAppointmentRequest & {
    appointment_id: string;
};

export default function PaymentStatus({ type }: PaymentStatusProps) {
     const navigate = useNavigate();
     const [searchParams] = useSearchParams();
     const orderId = searchParams.get('order_id');
     const [isFinalizing, setIsFinalizing] = useState(type === 'success');
     const [finalizeError, setFinalizeError] = useState<string | null>(null);
     const [finalized, setFinalized] = useState(false);

     useEffect(() => {
         if (type !== 'success') {
             if (orderId) {
                 localStorage.removeItem(`pending-booking:${orderId}`);
             }
             return;
         }

         if (!orderId) {
             setIsFinalizing(false);
             setFinalizeError('Missing payment reference.');
             return;
         }

         const storageKey = `pending-booking:${orderId}`;
         const rawDraft = localStorage.getItem(storageKey);

         let isMounted = true;
         const finalize = async () => {
             try {
                 if (rawDraft) {
                     let draft: PendingBookingDraft | null = null;
                     try {
                         draft = JSON.parse(rawDraft) as PendingBookingDraft;
                     } catch {
                         throw new Error('The pending appointment details were corrupted.');
                     }

                     let appointmentExists = false;
                     try {
                         await appointmentApi.getAppointmentStatus(draft!.appointment_id);
                         appointmentExists = true;
                     } catch {
                         appointmentExists = false;
                     }

                     if (!appointmentExists) {
                         try {
                             await appointmentApi.bookAppointment({
                                 appointment_id: draft!.appointment_id,
                                 doctor_id: draft!.doctor_id,
                                 slot_id: draft!.slot_id,
                                 scheduled_at: draft!.scheduled_at,
                                 notes: draft!.notes,
                                 payment_mode: draft!.payment_mode || 'pay_now',
                                 consultation_mode: draft!.consultation_mode || 'jitsi',
                                 payment_completed: true,
                             });
                         } catch {
                             // If the appointment was created by a concurrent/previous step,
                             // we treat that as success and continue with payment completion.
                             await appointmentApi.getAppointmentStatus(draft!.appointment_id);
                         }
                     }
                 }

                 await paymentApi.completePayment(orderId);
                 if (rawDraft) {
                     localStorage.removeItem(storageKey);
                 }
                 if (!isMounted) return;
                 setFinalized(true);
                 setIsFinalizing(false);
             } catch (error) {
                 console.error('Failed to finalize appointment after payment:', error);
                 if (!isMounted) return;
                 setIsFinalizing(false);
                 setFinalizeError(
                     rawDraft
                         ? 'Payment went through, but we could not create the appointment yet. Please go to Appointments and try again.'
                         : 'Payment went through, but we could not update the appointment yet. Please go to Appointments and refresh.'
                 );
             }
         };

         finalize();

         return () => {
             isMounted = false;
         };
     }, [orderId, type]);

    return (
        <div className="min-h-screen bg-[#f6f8fa] flex items-center justify-center p-6 font-sans">
            <div className="bg-white rounded-3xl w-full max-w-md p-8 shadow-2xl animate-in fade-in zoom-in duration-300 text-center">
                <div className="flex justify-center mb-6">
                    {type === 'success' ? (
                        <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center">
                            <CheckCircle2 className="h-10 w-10 text-green-600" />
                        </div>
                    ) : (
                        <div className="w-20 h-20 bg-red-100 rounded-full flex items-center justify-center">
                            <XCircle className="h-10 w-10 text-red-600" />
                        </div>
                    )}
                </div>

                <h1 className="text-2xl font-bold text-gray-900 mb-2">
                    {type === 'success'
                        ? isFinalizing
                            ? 'Processing Payment...'
                            : finalized
                                ? 'Appointment Confirmed!'
                                : 'Payment Successful!'
                        : 'Payment Canceled'}
                </h1>
                <p className="text-gray-500 mb-8 leading-relaxed">
                    {type === 'success'
                        ? isFinalizing
                            ? 'We are creating your appointment and reserving the slot now.'
                            : finalized
                                ? 'Your appointment has been added to My Appointments. If it is a Jitsi visit, the join link will appear there once payment is completed.'
                                : 'Your payment was received.'
                        : 'Your payment was not completed. If this was an error, please try booking again.'}
                </p>

                {finalizeError && (
                    <div className="mb-6 rounded-2xl border border-amber-100 bg-amber-50 px-4 py-3 text-left text-sm text-amber-800">
                        {finalizeError}
                    </div>
                )}

                {orderId && (
                    <div className="bg-gray-50 rounded-2xl p-4 mb-8 text-left border border-gray-100">
                        <span className="text-[10px] text-gray-400 uppercase font-bold tracking-tight block">Reference ID</span>
                        <code className="text-sm font-semibold text-gray-700">{orderId}</code>
                    </div>
                )}

                <div className="space-y-3">
                    {type === 'success' ? (
                        <Link
                            to="/appointments?tab=appointments"
                            onClick={(event) => {
                                if (!finalized && isFinalizing) {
                                    event.preventDefault();
                                } else if (finalized) {
                                    navigate('/appointments?tab=appointments', { replace: true });
                                }
                            }}
                            className="flex items-center justify-center gap-2 w-full py-4 bg-brand text-white rounded-2xl font-bold hover:bg-brand-dark transition-all shadow-lg shadow-brand/20 active:scale-[0.98]"
                        >
                            <Calendar className="h-5 w-5" />
                            {isFinalizing ? 'Finalizing...' : 'View Appointments'}
                        </Link>
                    ) : (
                        <Link
                            to="/appointments"
                            className="flex items-center justify-center gap-2 w-full py-4 bg-gray-900 text-white rounded-2xl font-bold hover:bg-gray-800 transition-all shadow-lg shadow-black/10 active:scale-[0.98]"
                        >
                             Try Again
                        </Link>
                    )}
                    
                    <Link
                        to="/dashboard"
                        className="flex items-center justify-center gap-2 w-full py-4 bg-transparent text-gray-500 rounded-2xl font-semibold hover:bg-gray-50 transition-all"
                    >
                        <ArrowLeft className="h-5 w-5" /> Go to Dashboard
                    </Link>
                </div>
            </div>
        </div>
    );
}
