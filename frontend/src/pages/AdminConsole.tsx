import { useEffect, useMemo, useState } from 'react';
import { Activity, Calendar, CreditCard, LogOut, ShieldCheck, Users } from 'lucide-react';
import { Link } from 'react-router-dom';
import {
  adminApi,
  type AdminAppointment,
  type AdminTransaction,
  type AdminUser,
} from '../api/admin.ts';

type TabKey = 'users' | 'appointments' | 'transactions';

function fmtDate(value: string) {
  return new Date(value).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export default function AdminConsole() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [appointments, setAppointments] = useState<AdminAppointment[]>([]);
  const [transactions, setTransactions] = useState<AdminTransaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<TabKey>('users');
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);
  const [error, setError] = useState('');

  const role = useMemo(() => {
    const raw = localStorage.getItem('user');
    if (!raw) return '';
    try {
      return (JSON.parse(raw)?.role as string) ?? '';
    } catch {
      return '';
    }
  }, []);

  const loadData = async () => {
    setLoading(true);
    setError('');
    try {
      const [u, a, t] = await Promise.all([
        adminApi.listUsers(),
        adminApi.listAppointments(),
        adminApi.listTransactions(),
      ]);
      setUsers(u);
      setAppointments(a);
      setTransactions(t);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load admin data';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (role !== 'admin') {
      setLoading(false);
      return;
    }
    loadData();
  }, [role]);

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    window.location.href = '/auth/login';
  };

  const handleVerifyDoctor = async (doctorId: string) => {
    setActionLoadingId(doctorId);
    try {
      await adminApi.verifyDoctor(doctorId, 'verified from admin console');
      await loadData();
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleDeactivateUser = async (userId: string) => {
    setActionLoadingId(userId);
    try {
      await adminApi.deactivateUser(userId);
      await loadData();
    } finally {
      setActionLoadingId(null);
    }
  };

  if (role !== 'admin') {
    return (
      <div className="min-h-screen bg-[#f6f8fa] flex items-center justify-center p-6">
        <div className="bg-white rounded-3xl border border-gray-200 shadow-sm p-8 max-w-lg w-full text-center">
          <ShieldCheck className="h-12 w-12 mx-auto text-amber-500 mb-4" />
          <h2 className="text-2xl font-bold text-gray-900">Admin Access Required</h2>
          <p className="text-gray-500 mt-2">This page is only available for users with admin role.</p>
          <Link to="/dashboard" className="inline-block mt-6 px-5 py-2.5 bg-brand text-white rounded-xl text-sm font-semibold hover:bg-brand-dark transition-colors">
            Go to Dashboard
          </Link>
        </div>
      </div>
    );
  }

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
          <Link to="/admin" className="flex items-center gap-3 px-3 py-2.5 bg-brand/10 text-brand rounded-xl font-semibold transition-all text-sm shadow-sm">
            <ShieldCheck className="h-[18px] w-[18px]" /> Admin Console
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
          <h2 className="text-xl font-semibold text-gray-800">Admin Console</h2>
          <button
            onClick={loadData}
            className="px-4 py-2 bg-brand text-white rounded-xl text-sm font-semibold hover:bg-brand-dark transition-colors"
          >
            Refresh
          </button>
        </header>

        <main className="p-8 overflow-y-auto">
          <div className="mb-8">
            <h3 className="text-2xl font-bold text-gray-900">Back Office Controls</h3>
            <p className="text-gray-500 mt-1">Manage mirrored platform data from auth, appointment, and payment flows.</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-5 mb-8">
            <div className="bg-white border border-gray-200 rounded-2xl p-5">
              <div className="w-10 h-10 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center mb-3">
                <Users className="h-5 w-5" />
              </div>
              <p className="text-sm text-gray-500">Users</p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{users.length}</p>
            </div>
            <div className="bg-white border border-gray-200 rounded-2xl p-5">
              <div className="w-10 h-10 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center mb-3">
                <Calendar className="h-5 w-5" />
              </div>
              <p className="text-sm text-gray-500">Appointments</p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{appointments.length}</p>
            </div>
            <div className="bg-white border border-gray-200 rounded-2xl p-5">
              <div className="w-10 h-10 rounded-xl bg-violet-50 text-violet-600 flex items-center justify-center mb-3">
                <CreditCard className="h-5 w-5" />
              </div>
              <p className="text-sm text-gray-500">Transactions</p>
              <p className="text-2xl font-bold text-gray-900 mt-1">{transactions.length}</p>
            </div>
          </div>

          <div className="flex gap-3 mb-5">
            <button
              onClick={() => setActiveTab('users')}
              className={`px-4 py-2 rounded-xl text-sm font-semibold transition-colors ${
                activeTab === 'users' ? 'bg-brand text-white' : 'bg-white border border-gray-200 text-gray-600'
              }`}
            >
              Users
            </button>
            <button
              onClick={() => setActiveTab('appointments')}
              className={`px-4 py-2 rounded-xl text-sm font-semibold transition-colors ${
                activeTab === 'appointments' ? 'bg-brand text-white' : 'bg-white border border-gray-200 text-gray-600'
              }`}
            >
              Appointments
            </button>
            <button
              onClick={() => setActiveTab('transactions')}
              className={`px-4 py-2 rounded-xl text-sm font-semibold transition-colors ${
                activeTab === 'transactions' ? 'bg-brand text-white' : 'bg-white border border-gray-200 text-gray-600'
              }`}
            >
              Transactions
            </button>
          </div>

          {error && (
            <div className="mb-5 p-4 rounded-xl border border-red-200 bg-red-50 text-red-700 text-sm">{error}</div>
          )}

          {loading ? (
            <div className="flex items-center justify-center py-16">
              <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-brand"></div>
            </div>
          ) : (
            <div className="bg-white rounded-3xl border border-gray-200 shadow-sm overflow-x-auto">
              {activeTab === 'users' && (
                <table className="w-full text-left border-collapse min-w-[920px]">
                  <thead className="bg-gray-50 border-b border-gray-200">
                    <tr>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">User</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Role</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Verified</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Active</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Created</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {users.map((u) => (
                      <tr key={u.id} className="hover:bg-gray-50 transition-colors">
                        <td className="px-5 py-4">
                          <p className="text-sm font-semibold text-gray-900">{u.first_name} {u.last_name}</p>
                          <p className="text-xs text-gray-500">{u.email}</p>
                        </td>
                        <td className="px-5 py-4 text-sm text-gray-700">{u.role}</td>
                        <td className="px-5 py-4 text-sm text-gray-700">{u.is_verified ? 'Yes' : 'No'}</td>
                        <td className="px-5 py-4 text-sm text-gray-700">{u.is_active ? 'Yes' : 'No'}</td>
                        <td className="px-5 py-4 text-sm text-gray-700">{fmtDate(u.created_at)}</td>
                        <td className="px-5 py-4">
                          <div className="flex justify-end gap-2">
                            {u.role === 'doctor' && !u.is_verified && (
                              <button
                                onClick={() => handleVerifyDoctor(u.id)}
                                disabled={actionLoadingId === u.id}
                                className="px-3 py-1.5 bg-emerald-500 text-white text-xs rounded-lg font-semibold hover:bg-emerald-600 disabled:opacity-60"
                              >
                                Verify
                              </button>
                            )}
                            {u.is_active && (
                              <button
                                onClick={() => handleDeactivateUser(u.id)}
                                disabled={actionLoadingId === u.id}
                                className="px-3 py-1.5 bg-red-500 text-white text-xs rounded-lg font-semibold hover:bg-red-600 disabled:opacity-60"
                              >
                                Deactivate
                              </button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}

              {activeTab === 'appointments' && (
                <table className="w-full text-left border-collapse min-w-[860px]">
                  <thead className="bg-gray-50 border-b border-gray-200">
                    <tr>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Appointment ID</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Patient ID</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Doctor ID</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Status</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Scheduled</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {appointments.map((a) => (
                      <tr key={a.id} className="hover:bg-gray-50 transition-colors">
                        <td className="px-5 py-4 text-sm text-gray-900 font-mono">{a.id}</td>
                        <td className="px-5 py-4 text-sm text-gray-700 font-mono">{a.patient_id}</td>
                        <td className="px-5 py-4 text-sm text-gray-700 font-mono">{a.doctor_id}</td>
                        <td className="px-5 py-4 text-sm text-gray-700 uppercase">{a.status}</td>
                        <td className="px-5 py-4 text-sm text-gray-700">{fmtDate(a.scheduled_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}

              {activeTab === 'transactions' && (
                <table className="w-full text-left border-collapse min-w-[860px]">
                  <thead className="bg-gray-50 border-b border-gray-200">
                    <tr>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Transaction ID</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">User ID</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Amount</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Provider</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Status</th>
                      <th className="px-5 py-4 text-xs font-bold text-gray-500 uppercase">Created</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {transactions.map((t) => (
                      <tr key={t.id} className="hover:bg-gray-50 transition-colors">
                        <td className="px-5 py-4 text-sm text-gray-900 font-mono">{t.id}</td>
                        <td className="px-5 py-4 text-sm text-gray-700 font-mono">{t.user_id}</td>
                        <td className="px-5 py-4 text-sm text-gray-700">
                          {t.currency || 'LKR'} {Number(t.amount || 0).toFixed(2)}
                        </td>
                        <td className="px-5 py-4 text-sm text-gray-700">{t.provider || '-'}</td>
                        <td className="px-5 py-4 text-sm text-gray-700 uppercase">{t.status || '-'}</td>
                        <td className="px-5 py-4 text-sm text-gray-700">{fmtDate(t.created_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
