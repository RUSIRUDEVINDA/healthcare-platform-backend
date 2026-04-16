import { useEffect, useState } from 'react';
import { Activity, BellRing, CheckCircle2, Clock3, LogOut, ShieldCheck, TriangleAlert } from 'lucide-react';
import { Link } from 'react-router-dom';
import { notificationApi, type ServiceHealth } from '../api/notification';
import { redirectToLogin } from '../utils/navigation';

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

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    redirectToLogin();
  };

  const healthy = state.health?.status === 'healthy';
  const ready = state.ready?.status === 'ready';

  return (
    <div className="min-h-screen bg-[#f6f8fa] flex font-sans">
      <aside className="w-60 bg-white border-r border-gray-100 hidden lg:flex flex-col sticky top-0 h-screen">
        <div className="px-6 pt-6 pb-5">
          <Link to="/dashboard" className="flex items-center gap-2.5">
            <div className="w-8 h-8 bg-brand rounded-lg flex items-center justify-center">
              <Activity className="h-4 w-4 text-white" />
            </div>
            <span className="text-lg font-bold text-gray-900 tracking-tight">AyaRX</span>
          </Link>
        </div>

        <nav className="flex-1 px-4 space-y-1">
          <p className="text-[10px] font-semibold text-gray-400 uppercase tracking-wider px-3 mb-3">Menu</p>
          <Link to="/dashboard" className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm">
            <Activity className="h-[18px] w-[18px]" /> Dashboard
          </Link>
          <Link to="/admin" className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm">
            <ShieldCheck className="h-[18px] w-[18px]" /> Admin Console
          </Link>
          <Link to="/notifications" className="flex items-center gap-3 px-3 py-2.5 bg-brand/10 text-brand rounded-xl font-semibold transition-all text-sm shadow-sm">
            <BellRing className="h-[18px] w-[18px]" /> Notification Monitor
          </Link>
        </nav>

        <div className="p-4 border-t border-gray-100 mx-4 mb-4">
          <button
            onClick={handleLogout}
            className="flex items-center gap-2.5 w-full px-3 py-2.5 text-red-500 hover:bg-red-50 rounded-xl transition-colors text-sm font-medium"
          >
            <LogOut className="h-[18px] w-[18px]" /> Sign Out
          </button>
        </div>
      </aside>

      <div className="flex-1 flex flex-col min-h-screen">
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-8 bg-white/80 backdrop-blur-md sticky top-0 z-10">
          <h2 className="text-xl font-semibold text-gray-800">Notification Service Monitor</h2>
          <button
            onClick={loadHealth}
            className="px-4 py-2 bg-brand text-white rounded-xl text-sm font-semibold hover:bg-brand-dark transition-colors"
          >
            Refresh
          </button>
        </header>

        <main className="p-8 overflow-y-auto">
          <div className="mb-8">
            <h3 className="text-2xl font-bold text-gray-900">Event Delivery Readiness</h3>
            <p className="text-gray-500 mt-1">Live health view for notification-service and expected event subscriptions.</p>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-16">
              <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-brand"></div>
            </div>
          ) : (
            <>
              {state.error && (
                <div className="mb-6 p-4 rounded-xl border border-red-200 bg-red-50 text-red-700 text-sm">
                  Unable to query notification service: {state.error}
                </div>
              )}

              <div className="grid grid-cols-1 md:grid-cols-2 gap-5 mb-7">
                <div className="bg-white rounded-2xl border border-gray-200 p-5">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium text-gray-500">Liveness</p>
                    {healthy ? <CheckCircle2 className="h-5 w-5 text-emerald-500" /> : <TriangleAlert className="h-5 w-5 text-amber-500" />}
                  </div>
                  <p className="text-2xl font-bold text-gray-900 mt-2">{state.health?.status ?? 'unknown'}</p>
                  <p className="text-xs text-gray-500 mt-2">GET /health/notification</p>
                </div>

                <div className="bg-white rounded-2xl border border-gray-200 p-5">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium text-gray-500">Readiness</p>
                    {ready ? <CheckCircle2 className="h-5 w-5 text-emerald-500" /> : <Clock3 className="h-5 w-5 text-amber-500" />}
                  </div>
                  <p className="text-2xl font-bold text-gray-900 mt-2">{state.ready?.status ?? 'unknown'}</p>
                  <p className="text-xs text-gray-500 mt-2">GET /health/notification/ready</p>
                </div>
              </div>

              <div className="bg-white rounded-3xl border border-gray-200 shadow-sm p-6 mb-6">
                <h4 className="text-lg font-bold text-gray-900 mb-4">Subscribed Event Routes</h4>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                  <div className="rounded-xl border border-gray-200 p-4">
                    <p className="font-semibold text-gray-800">appointment.booked</p>
                    <p className="text-gray-500 mt-1">Queue: notification_appointment_booked_queue</p>
                  </div>
                  <div className="rounded-xl border border-gray-200 p-4">
                    <p className="font-semibold text-gray-800">appointment.cancelled</p>
                    <p className="text-gray-500 mt-1">Queue: notification_appointment_cancelled_queue</p>
                  </div>
                  <div className="rounded-xl border border-gray-200 p-4">
                    <p className="font-semibold text-gray-800">consultation.completed</p>
                    <p className="text-gray-500 mt-1">Queue: notification_consultation_completed_queue</p>
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-3xl border border-gray-200 shadow-sm p-6">
                <h4 className="text-lg font-bold text-gray-900 mb-3">Operational Notes</h4>
                <ul className="text-sm text-gray-600 space-y-2">
                  <li>If SMTP/Twilio are not configured, the service logs delivery attempts instead of failing.</li>
                  <li>Failed message handlers are NACKed and requeued by RabbitMQ.</li>
                  <li>Open RabbitMQ management at <a className="text-brand font-semibold hover:underline" href="http://localhost:15672" target="_blank" rel="noreferrer">http://localhost:15672</a> to publish test events.</li>
                </ul>
                <p className="text-xs text-gray-400 mt-4">Last checked: {state.checkedAt ?? 'never'}</p>
              </div>
            </>
          )}
        </main>
      </div>
    </div>
  );
}
