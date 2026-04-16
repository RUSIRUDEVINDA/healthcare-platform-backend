import React, { useState } from 'react';
import {
  Mail,
  User,
  MessageSquare,
  Send,
  CheckCircle2,
  ArrowLeft,
  ShieldAlert,
  Sparkles,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { supportApi } from '../api/support';
import toast from 'react-hot-toast';

export default function Reactivate() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [reason, setReason] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      await supportApi.submitReactivationTicket({ name, email, reason });
      setIsSubmitted(true);
      toast.success('Request submitted successfully');
    } catch (err) {
      console.error(err);
      toast.error('Failed to submit request. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-white text-slate-900 font-sans selection:bg-brand/20 relative overflow-hidden flex items-center justify-center px-4 py-16">
      {/* Decorative blobs — matching Landing page */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
        <div className="absolute -top-[10%] -left-[10%] w-[40%] h-[40%] bg-brand/5 blur-[120px] rounded-full" />
        <div className="absolute top-[30%] -right-[5%] w-[30%] h-[30%] bg-brand/10 blur-[100px] rounded-full" />
        <div className="absolute bottom-0 left-[20%] w-[25%] h-[25%] bg-amber-400/5 blur-[80px] rounded-full" />
      </div>

      <div className="relative z-10 w-full max-w-lg">
        {/* Brand badge */}
        <div className="flex justify-center mb-10">
          <Link to="/" className="flex items-center gap-3 group">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-brand to-brand-dark text-white shadow-lg shadow-brand/20 transition-transform group-hover:scale-105">
              <Sparkles className="h-5 w-5" />
            </div>
            <div>
              <p className="text-sm font-bold uppercase tracking-[0.2em] text-slate-900">MediPulse</p>
              <p className="text-[10px] font-medium text-slate-500">Sri Lanka Edition</p>
            </div>
          </Link>
        </div>

        {isSubmitted ? (
          /* ── Success State ── */
          <div className="rounded-[2rem] bg-slate-900 p-10 shadow-2xl text-white text-center relative overflow-hidden">
            <div className="absolute top-0 right-0 w-1/2 h-full bg-brand/10 blur-[80px] rounded-full -translate-y-1/2 translate-x-1/4 pointer-events-none" />
            <div className="relative z-10">
              <div className="flex justify-center mb-8">
                <div className="w-20 h-20 rounded-3xl bg-brand/20 border border-brand/30 flex items-center justify-center text-brand">
                  <CheckCircle2 className="h-10 w-10" />
                </div>
              </div>
              <h2 className="text-3xl font-bold tracking-tight mb-4">Request Received</h2>
              <p className="text-slate-400 leading-relaxed max-w-sm mx-auto">
                Your reactivation request has been sent to our administration team.
                We will review your details and get back to you via email shortly.
              </p>
              <div className="mt-10 inline-flex items-center gap-2 rounded-2xl bg-white/10 border border-white/10 px-5 py-2.5 text-sm font-medium text-white/80 hover:bg-white/20 transition-all cursor-pointer">
                <Link to="/auth/login" className="flex items-center gap-2">
                  <ArrowLeft className="h-4 w-4" />
                  Back to Sign In
                </Link>
              </div>
            </div>
          </div>
        ) : (
          /* ── Form State ── */
          <div className="rounded-[2rem] border border-slate-200/60 bg-white shadow-2xl shadow-slate-200/50 overflow-hidden">
            {/* Card header */}
            <div className="relative bg-gradient-to-br from-amber-50 to-orange-50/50 border-b border-amber-100 px-10 pt-10 pb-8">
              <div className="absolute top-0 right-0 w-1/2 h-full bg-brand/5 blur-[60px] rounded-full translate-x-1/4 pointer-events-none" />
              <div className="relative z-10">
                <div className="inline-flex items-center justify-center w-14 h-14 bg-white rounded-2xl shadow-sm border border-amber-100 text-amber-500 mb-6">
                  <ShieldAlert className="h-7 w-7" />
                </div>
                <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Account Reactivation</h1>
                <p className="mt-2 text-sm text-slate-500 max-w-xs leading-relaxed">
                  Submit a request and our team will restore your access after a quick review.
                </p>
              </div>
            </div>

            {/* Form body */}
            <form onSubmit={handleSubmit} className="px-10 py-8 space-y-5">
              {/* Full Name */}
              <div>
                <label className="block text-xs font-bold uppercase tracking-widest text-slate-400 mb-2" htmlFor="name">
                  Full Name
                </label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-brand transition-colors">
                    <User className="h-4 w-4" />
                  </div>
                  <input
                    id="name"
                    type="text"
                    required
                    className="block w-full pl-11 pr-4 py-3 border border-slate-200 rounded-xl focus:ring-4 focus:ring-brand/10 focus:border-brand transition-all bg-slate-50 hover:bg-white text-sm text-slate-900 placeholder:text-slate-400"
                    placeholder="Your full name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                  />
                </div>
              </div>

              {/* Email */}
              <div>
                <label className="block text-xs font-bold uppercase tracking-widest text-slate-400 mb-2" htmlFor="email">
                  Email Address
                </label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-400 group-focus-within:text-brand transition-colors">
                    <Mail className="h-4 w-4" />
                  </div>
                  <input
                    id="email"
                    type="email"
                    required
                    className="block w-full pl-11 pr-4 py-3 border border-slate-200 rounded-xl focus:ring-4 focus:ring-brand/10 focus:border-brand transition-all bg-slate-50 hover:bg-white text-sm text-slate-900 placeholder:text-slate-400"
                    placeholder="you@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                  />
                </div>
              </div>

              {/* Reason */}
              <div>
                <label className="block text-xs font-bold uppercase tracking-widest text-slate-400 mb-2" htmlFor="reason">
                  Reason for Reactivation
                </label>
                <div className="relative group">
                  <div className="absolute top-3.5 left-4 pointer-events-none text-slate-400 group-focus-within:text-brand transition-colors">
                    <MessageSquare className="h-4 w-4" />
                  </div>
                  <textarea
                    id="reason"
                    required
                    rows={4}
                    className="block w-full pl-11 pr-4 py-3 border border-slate-200 rounded-xl focus:ring-4 focus:ring-brand/10 focus:border-brand transition-all bg-slate-50 hover:bg-white resize-none text-sm text-slate-900 placeholder:text-slate-400"
                    placeholder="Briefly explain why you'd like to reactivate your account…"
                    value={reason}
                    onChange={(e) => setReason(e.target.value)}
                  />
                </div>
              </div>

              {/* Submit */}
              <button
                type="submit"
                disabled={isLoading}
                className="w-full flex justify-center items-center gap-2 py-3.5 rounded-2xl bg-slate-900 text-white text-sm font-bold shadow-xl shadow-slate-900/10 transition-all hover:bg-slate-800 hover:-translate-y-0.5 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed disabled:translate-y-0"
              >
                {isLoading ? (
                  <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                ) : (
                  <>
                    Submit Request
                    <Send className="h-4 w-4" />
                  </>
                )}
              </button>
            </form>

            {/* Card footer */}
            <div className="px-10 pb-8 text-center">
              <Link
                to="/auth/login"
                className="inline-flex items-center gap-2 text-xs font-medium text-slate-400 hover:text-brand transition-colors"
              >
                <ArrowLeft className="h-3.5 w-3.5" />
                Back to Sign In
              </Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
