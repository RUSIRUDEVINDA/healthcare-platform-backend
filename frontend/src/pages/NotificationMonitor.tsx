import { useEffect, useState } from 'react';
import { Activity, BellRing, CheckCircle2, Clock3, ShieldCheck, TriangleAlert, RefreshCw } from 'lucide-react';
import { notificationApi, type ServiceHealth } from '../api/notification';

type HealthState = {
  health?: ServiceHealth;
  ready?: ServiceHealth;
  error?: string;
  checkedAt?: string;
};

export default function NotificationMonitor() {
  const [state, setState] = useState<HealthState>({});
  const [loading, setLoading] = useState(true);

  const loadHealth = async () => {
    setLoading(true);
    try {
      const [health, ready] = await Promise.all([
        notificationApi.getHealth(),
        notificationApi.getReady(),
      ]);
      setState({
        health,
        ready,
        checkedAt: new Date().toLocaleTimeString(),
      });
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Notification service is not reachable';
      setState({
        error: message,
        checkedAt: new Date().toLocaleTimeString(),
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadHealth();
    const timer = setInterval(loadHealth, 20000);
    return () => clearInterval(timer);
  }, []);

  const healthy = state.health?.status === 'healthy';
  const ready = state.ready?.status === 'ready';

  return (
    <div className="min-h-screen bg-[#f8fafc] flex flex-col font-sans">
      <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-8 bg-white/80 backdrop-blur-md sticky top-0 z-10">
        <div className="flex items-center gap-3">
          <BellRing className="h-5 w-5 text-brand" />
          <h2 className="text-xl font-bold text-slate-800 tracking-tight">Notification Monitoring</h2>
        </div>
        <button
          onClick={loadHealth}
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-brand text-white rounded-xl text-sm font-semibold hover:bg-brand-dark transition-all disabled:opacity-50"
        >
          {loading ? <RefreshCw className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
          Refresh Status
        </button>
      </header>

      <main className="p-8 lg:p-10 max-w-6xl mx-auto w-full">
        <div className="mb-10">
          <h3 className="text-3xl font-black text-slate-900 tracking-tight">System Reliability</h3>
          <p className="text-slate-500 mt-2 text-lg">Real-time connectivity and delivery health for the MediPulse notification engine.</p>
        </div>

        {loading && !state.checkedAt ? (
          <div className="flex flex-col items-center justify-center py-20">
            <div className="animate-spin rounded-full h-12 w-12 border-4 border-slate-100 border-t-brand mb-4"></div>
            <p className="text-slate-400 font-semibold">Probing services...</p>
          </div>
        ) : (
          <>
            {state.error && (
              <div className="mb-8 p-5 rounded-2xl border-l-4 border-red-500 bg-red-50 text-red-700 font-medium flex items-center gap-3 shadow-sm">
                <TriangleAlert className="h-5 w-5" />
                Connection Failure: {state.error}
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-10">
              <div className="bg-white rounded-3xl border border-slate-200 p-8 shadow-sm">
                <div className="flex items-center justify-between mb-6">
                  <p className="text-sm font-black text-slate-400 uppercase tracking-widest">Service Liveness</p>
                  {healthy ? <CheckCircle2 className="h-6 w-6 text-emerald-500" /> : <TriangleAlert className="h-6 w-6 text-amber-500" />}
                </div>
                <p className={`text-4xl font-black capitalize ${healthy ? 'text-emerald-600' : 'text-amber-600'}`}>
                  {state.health?.status ?? 'Unknown'}
                </p>
                <div className="mt-6 pt-6 border-t border-slate-50">
                  <code className="text-[10px] bg-slate-50 px-2 py-1 rounded text-slate-400">GET /health/notification</code>
                </div>
              </div>

              <div className="bg-white rounded-3xl border border-slate-200 p-8 shadow-sm">
                <div className="flex items-center justify-between mb-6">
                  <p className="text-sm font-black text-slate-400 uppercase tracking-widest">Message Readiness</p>
                  {ready ? <CheckCircle2 className="h-6 w-6 text-emerald-500" /> : <Clock3 className="h-6 w-6 text-amber-500" />}
                </div>
                <p className={`text-4xl font-black capitalize ${ready ? 'text-emerald-600' : 'text-amber-600'}`}>
                  {state.ready?.status ?? 'Unknown'}
                </p>
                <div className="mt-6 pt-6 border-t border-slate-50">
                  <code className="text-[10px] bg-slate-50 px-2 py-1 rounded text-slate-400">GET /health/notification/ready</code>
                </div>
              </div>
            </div>

            <div className="bg-white rounded-[2rem] border border-slate-200 shadow-xl shadow-slate-200/50 p-8 mb-8">
              <div className="flex items-center gap-3 mb-8">
                <div className="w-10 h-10 bg-slate-100 rounded-xl flex items-center justify-center">
                  <Activity className="h-5 w-5 text-slate-400" />
                </div>
                <h4 className="text-xl font-black text-slate-900">Event Subscriptions</h4>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {[
                  { name: 'appointment.booked', queue: 'notification_appointment_booked_queue' },
                  { name: 'appointment.cancelled', queue: 'notification_appointment_cancelled_queue' },
                  { name: 'consultation.completed', queue: 'notification_consultation_completed_queue' },
                ].map((event, i) => (
                  <div key={i} className="rounded-2xl border border-slate-100 p-6 bg-slate-50/50 hover:bg-white hover:border-brand/30 hover:shadow-md transition-all cursor-default">
                    <p className="font-bold text-slate-900 text-lg mb-1">{event.name}</p>
                    <p className="text-[10px] text-slate-400 font-bold uppercase tracking-widest leading-relaxed">Active RabbitMQ Binding</p>
                    <p className="text-xs text-brand font-medium mt-3 bg-brand/5 inline-block px-2 py-1 rounded-lg truncate w-full">{event.queue}</p>
                  </div>
                ))}
              </div>
            </div>

            <div className="bg-slate-900 rounded-[2rem] p-10 text-white shadow-2xl overflow-hidden relative">
              <div className="absolute top-0 right-0 w-64 h-64 bg-brand/10 blur-[100px] rounded-full -mr-32 -mt-32"></div>
              <h4 className="text-2xl font-black mb-6 flex items-center gap-3">
                <ShieldCheck className="h-6 w-6 text-brand" />
                Infrastructure Metrics
              </h4>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-10">
                <ul className="text-slate-400 space-y-4">
                  <li className="flex items-start gap-3">
                    <div className="w-1.5 h-1.5 bg-brand rounded-full mt-1.5 shrink-0"></div>
                    <p className="text-sm">Fail-safe logging: Service defaults to local I/O logs if SMTP/Twilio configuration is missing.</p>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="w-1.5 h-1.5 bg-brand rounded-full mt-1.5 shrink-0"></div>
                    <p className="text-sm">Resilience: RabbitMQ NACK/Requeue policy ensures no message loss during downtime.</p>
                  </li>
                </ul>
                <div className="bg-slate-800/50 rounded-2xl p-6 border border-slate-700/50 backdrop-blur-sm">
                  <p className="text-xs font-bold text-slate-500 uppercase tracking-[0.2em] mb-4">Management Access</p>
                  <a 
                    href="http://localhost:15672" 
                    target="_blank" 
                    rel="noreferrer"
                    className="flex items-center justify-between group"
                  >
                    <span className="text-brand font-bold hover:underline">RabbitMQ Dashboard</span>
                    <RefreshCw className="h-4 w-4 text-slate-600 group-hover:text-brand transition-colors" />
                  </a>
                  <p className="text-[10px] text-slate-500 mt-2 font-mono">guest:guest (Default credentials)</p>
                </div>
              </div>
              <div className="mt-10 pt-10 border-t border-slate-800 flex justify-between items-center text-[11px] font-bold text-slate-500 tracking-widest">
                <span>MEDIPULSE SRILANKA CORE</span>
                <span>LAST CHECKED: {state.checkedAt?.toUpperCase() ?? 'PENDING'}</span>
              </div>
            </div>
          </>
        )}
      </main>
    </div>
  );
}
