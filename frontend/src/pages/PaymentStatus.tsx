import { Link, useSearchParams } from 'react-router-dom';
import { CheckCircle2, XCircle, ArrowLeft, Calendar } from 'lucide-react';

interface PaymentStatusProps {
    type: 'success' | 'cancel';
}

export default function PaymentStatus({ type }: PaymentStatusProps) {
     const [searchParams] = useSearchParams();
     const orderId = searchParams.get('order_id');

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
                    {type === 'success' ? 'Payment Successful!' : 'Payment Canceled'}
                </h1>
                <p className="text-gray-500 mb-8 leading-relaxed">
                    {type === 'success' 
                        ? 'Your appointment has been confirmed. You can now view it in your appointments list.' 
                        : 'Your payment was not completed. If this was an error, please try booking again.'}
                </p>

                {orderId && (
                    <div className="bg-gray-50 rounded-2xl p-4 mb-8 text-left border border-gray-100">
                        <span className="text-[10px] text-gray-400 uppercase font-bold tracking-tight block">Reference ID</span>
                        <code className="text-sm font-semibold text-gray-700">{orderId}</code>
                    </div>
                )}

                <div className="space-y-3">
                    {type === 'success' ? (
                        <Link
                            to="/appointments"
                            className="flex items-center justify-center gap-2 w-full py-4 bg-brand text-white rounded-2xl font-bold hover:bg-brand-dark transition-all shadow-lg shadow-brand/20 active:scale-[0.98]"
                        >
                            <Calendar className="h-5 w-5" /> View Appointments
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
