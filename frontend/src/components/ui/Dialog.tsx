import { useEffect } from 'react';
import { X, AlertCircle } from 'lucide-react';

interface DialogProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    description: string;
    confirmText?: string;
    cancelText?: string;
    onConfirm: () => void;
    variant?: 'danger' | 'warning' | 'info';
    isLoading?: boolean;
}

export default function Dialog({
    isOpen,
    onClose,
    title,
    description,
    confirmText = 'Confirm',
    cancelText = 'Cancel',
    onConfirm,
    variant = 'info',
    isLoading = false
}: DialogProps) {
    useEffect(() => {
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape') onClose();
        };
        if (isOpen) {
            document.addEventListener('keydown', handleEscape);
            document.body.style.overflow = 'hidden';
        }
        return () => {
            document.removeEventListener('keydown', handleEscape);
            document.body.style.overflow = 'unset';
        };
    }, [isOpen, onClose]);

    if (!isOpen) return null;

    const variantStyles = {
        danger: 'bg-red-600 hover:bg-red-700 shadow-red-200',
        warning: 'bg-amber-500 hover:bg-amber-600 shadow-amber-200',
        info: 'bg-brand hover:bg-brand-dark shadow-brand/20'
    };

    const iconStyles = {
        danger: 'text-red-600 bg-red-50',
        warning: 'text-amber-600 bg-amber-50',
        info: 'text-brand bg-brand/10'
    };

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
            {/* Backdrop */}
            <div 
                className="absolute inset-0 bg-slate-900/40 backdrop-blur-sm animate-in fade-in duration-300"
                onClick={onClose}
            />
            
            {/* Dialog Content */}
            <div className="relative bg-white w-full max-w-2xl rounded-[2rem] shadow-2xl shadow-slate-200/50 overflow-hidden animate-in zoom-in-95 duration-200">
                <div className="absolute top-6 right-6">
                    <button 
                        onClick={onClose}
                        className="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-50 rounded-full transition-all"
                    >
                        <X className="h-5 w-5" />
                    </button>
                </div>

                <div className="p-10">
                    <div className="flex flex-col items-center text-center">
                        <div className={`w-20 h-20 rounded-3xl flex items-center justify-center mb-8 ${iconStyles[variant]}`}>
                            <AlertCircle className="h-10 w-10" />
                        </div>
                        
                        <h3 className="text-3xl font-normal text-slate-900 mb-4 tracking-tight">
                            {title}
                        </h3>
                        
                        <p className="text-slate-500 text-lg leading-relaxed max-w-[480px]">
                            {description}
                        </p>
                    </div>

                    <div className="mt-12 flex flex-col gap-4 max-w-sm mx-auto">
                        <button
                            onClick={onConfirm}
                            disabled={isLoading}
                            className={`w-full py-4 text-white font-normal text-lg rounded-2xl shadow-lg transition-all active:scale-95 flex items-center justify-center gap-2 ${variantStyles[variant]}`}
                        >
                            {isLoading ? (
                                <div className="h-5 w-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                            ) : (
                                confirmText
                            )}
                        </button>
                        
                        <button
                            onClick={onClose}
                            disabled={isLoading}
                            className="w-full py-4 text-slate-500 font-normal text-lg hover:bg-slate-50 rounded-2xl transition-all active:scale-95"
                        >
                            {cancelText}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}
