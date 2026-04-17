import { useEffect, useMemo, useState } from 'react';
import { Activity, Calendar, CreditCard, ShieldCheck, Users, Search, CheckCircle, XCircle, UserPlus, Clock, LogOut, ArrowLeft, RefreshCw, LayoutDashboard, Ticket, Pencil, Trash2, X } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import { Toaster, toast } from 'react-hot-toast';
import {
  adminApi,
  type AdminTransaction,
  type AdminUser,
} from '../api/admin.ts';
import { appointmentApi, type Appointment } from '../api/appointments';
import { supportApi, type AdminSupportTicket } from '../api/support.ts';

type TabKey = 'patients' | 'doctors' | 'appointments' | 'payments' | 'tickets';

function fmtDate(value: string) {
  if (!value) return 'N/A';
  return new Date(value).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

function fmtTime(value: string) {
  if (!value) return '--:--';
  return new Date(value).toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

function compactId(value: string, length = 8) {
  if (!value) return 'N/A';
  return value.slice(0, length);
}

function displayName(first?: string, last?: string) {
  const name = [first, last].filter(Boolean).join(' ').trim();
  return name || 'Unknown';
}

function appointmentTitle(appointment: Appointment) {
  const patient = displayName(appointment.patient_first_name, appointment.patient_last_name);
  return patient !== 'Unknown' ? patient : compactId(appointment.patient_id);
}

function modeLabel(mode?: string) {
  switch ((mode || '').toLowerCase()) {
    case 'jitsi':
    case 'video':
      return 'Telemedicine';
    case 'physical':
    case 'in-person':
      return 'Physical';
    default:
      return mode ? mode.charAt(0).toUpperCase() + mode.slice(1) : 'Physical';
  }
}

function statusTone(status?: string) {
  switch ((status || '').toLowerCase()) {
    case 'confirmed':
    case 'paid':
    case 'succeeded':
      return 'bg-emerald-50 text-emerald-700 border-emerald-100';
    case 'completed':
      return 'bg-blue-50 text-blue-700 border-blue-100';
    case 'cancelled':
    case 'failed':
    case 'expired':
      return 'bg-rose-50 text-rose-700 border-rose-100';
    case 'pending':
    default:
      return 'bg-amber-50 text-amber-700 border-amber-100';
  }
}

function appoinmentStatusLabel(status?: string) {
  return status ? status.replace(/_/g, ' ') : 'pending';
}

export default function AdminConsole() {
  const navigate = useNavigate();
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [transactions, setTransactions] = useState<AdminTransaction[]>([]);
  const [tickets, setTickets] = useState<AdminSupportTicket[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<TabKey>('patients');
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  // CRUD States
  const [userModalOpen, setUserModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<AdminUser | null>(null);
  
  const [apptModalOpen, setApptModalOpen] = useState(false);
  const [editingAppt, setEditingAppt] = useState<Appointment | null>(null);

  const [saving, setSaving] = useState(false);

  const user = useMemo(() => {
    const raw = localStorage.getItem('user');
    if (!raw) return null;
    try { return JSON.parse(raw); } catch { return null; }
  }, []);

  const role = user?.role ?? '';

  const loadData = async () => {
    setLoading(true);
    setError('');
    try {
      const [u, a, t, tick] = await Promise.all([
        adminApi.listUsers(),
        appointmentApi.listAppointments(),
        adminApi.listTransactions(),
        supportApi.listTickets().catch(() => []), // Fallback to empty if 404
      ]);
      setUsers(u || []);
      setAppointments(a || []);
      setTransactions(t || []);
      setTickets(tick || []);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load admin data';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (role !== 'admin') {
      if (!loading) navigate('/dashboard');
      return;
    }
    loadData();
  }, [role]);

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    navigate('/login');
  };

  const handleDeactivateUser = async (userId: string) => {
    if (!confirm('Deactivate this account?')) return;
    setActionLoadingId(userId);
    try {
      await adminApi.deactivateUser(userId);
      toast.success('Account deactivated');
      await loadData();
    } catch (e) {
      toast.error('Deactivation failed');
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleReactivateUser = async (userId: string) => {
    setActionLoadingId(userId);
    try {
      await adminApi.reactivateUser(userId);
      toast.success('Account reactivated');
      await loadData();
    } catch (e) {
      toast.error('Reactivation failed');
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleResolveTicket = async (ticket: AdminSupportTicket) => {
    setActionLoadingId(ticket.id);
    try {
      await supportApi.resolveTicket(ticket.id);
      const u = users.find(x => x.email === ticket.email);
      if (u) await adminApi.reactivateUser(u.id);
      toast.success('Ticket resolved and account restored');
      await loadData();
    } catch (e) {
      toast.error('Action failed');
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleSync = async () => {
    setLoading(true);
    try {
      await adminApi.syncData();
      toast.success('Platform sync initiated');
      await loadData();
    } catch (e) {
      toast.error('Sync failed');
    } finally {
      setLoading(false);
    }
  };

  const handleSaveUser = async (user: any) => {
    setSaving(true);
    try {
      if (editingUser) {
        await adminApi.updateUser(editingUser.id, user);
        toast.success('User updated');
      } else {
        await adminApi.createUser(user);
        toast.success('New user profile created');
      }
      setUserModalOpen(false);
      setEditingUser(null);
      await loadData();
    } catch (e) {
      toast.error('Failed to save user profile');
    } finally {
      setSaving(false);
    }
  };

  const handleSaveAppt = async (appt: any) => {
    setSaving(true);
    try {
      if (editingAppt) {
        await adminApi.updateAppointment(editingAppt.id, appt);
        toast.success('Appointment details updated');
      } else {
        await adminApi.createAppointment(appt);
        toast.success('Manual booking confirmed');
      }
      setApptModalOpen(false);
      setEditingAppt(null);
      await loadData();
    } catch (e) {
      toast.error('Failed to finalize booking');
    } finally {
      setSaving(false);
    }
  };

  const handleCancelAppt = async (apptId: string) => {
    if (!confirm('Are you certain you want to cancel this appointment?')) return;
    setActionLoadingId(apptId);
    try {
      await adminApi.cancelAppointment(apptId);
      toast.success('Appointment cancelled');
      await loadData();
    } catch (e) {
      toast.error('Cancellation failed');
    } finally {
      setActionLoadingId(null);
    }
  };

  const filteredPatients = useMemo(() => 
    users.filter(u => u.role === 'patient' && 
      (u.first_name.toLowerCase().includes(searchQuery.toLowerCase()) || 
       u.email.toLowerCase().includes(searchQuery.toLowerCase()))),
    [users, searchQuery]
  );

  const filteredDoctors = useMemo(() => 
    users.filter(u => u.role === 'doctor' && 
      (u.first_name.toLowerCase().includes(searchQuery.toLowerCase()) || 
       u.email.toLowerCase().includes(searchQuery.toLowerCase()))),
    [users, searchQuery]
  );

  if (role !== 'admin') return null;

  return (
    <div className="min-h-screen bg-[#eff9f8] font-sans text-slate-900">
      <div className="flex min-h-screen flex-col overflow-hidden">
        <Toaster position="top-right" />
        
        <header className="h-16 bg-white border-b border-teal-100 flex items-center justify-between px-6 lg:px-8 sticky top-0 z-10 transition-all duration-300">
          <div className="flex items-center gap-6">
            <Link to="/dashboard" className="h-10 w-10 rounded-full bg-slate-50 border border-slate-100 flex items-center justify-center text-slate-400 hover:text-brand hover:border-brand/20 transition-all" title="Back to Dashboard">
              <ArrowLeft className="h-5 w-5" />
            </Link>
            <div className="h-8 w-[1px] bg-teal-100/50 hidden sm:block" />
            <div>
              <p className="text-[10px] uppercase tracking-[0.32em] text-slate-400 font-semibold leading-none mb-1">
                Infrastructure Control
              </p>
              <h1 className="text-lg font-bold text-slate-900 leading-none">
                Admin Console
              </h1>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="relative hidden sm:block">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <input
                type="text"
                placeholder="Search resources..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-64 pl-9 pr-4 py-2 bg-slate-50 border border-teal-100/50 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20 transition-all"
              />
            </div>
            <button
              onClick={handleLogout}
              className="h-10 w-10 rounded-full bg-red-50 border border-red-100 flex items-center justify-center text-red-500 hover:bg-red-100 transition-colors"
              title="Sign Out"
            >
              <LogOut className="h-5 w-5" />
            </button>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
          <div className="max-w-7xl mx-auto space-y-8">
            {/* Hero Section */}
            <section className="rounded-[2rem] border border-teal-100 bg-white p-6 sm:p-8 shadow-sm relative overflow-hidden">
              <div className="absolute top-0 right-0 p-8 opacity-[0.03] pointer-events-none">
                <LayoutDashboard size={200} />
              </div>
              
              <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between relative z-10">
                <div className="max-w-3xl space-y-4">
                  <div className="inline-flex items-center gap-2 rounded-full bg-brand/10 px-3 py-1 text-xs font-medium text-brand border border-brand/10">
                    <ShieldCheck className="h-3.5 w-3.5" />
                    Operational Workspace
                  </div>
                  <div>
                    <h2 className="text-3xl sm:text-4xl font-medium tracking-tight text-slate-900">
                      Welcome Back, Admin.
                    </h2>
                    <p className="mt-3 max-w-2xl text-sm sm:text-base leading-6 text-slate-600">
                      Monitor appointments, manage user accounts, oversee payments, and resolve support tickets in one focused operational hub.
                    </p>
                  </div>
                </div>

                <div className="flex flex-wrap gap-3">
                  {(activeTab === 'patients' || activeTab === 'doctors') && (
                    <button
                      onClick={() => { setEditingUser(null); setUserModalOpen(true); }}
                      className="inline-flex items-center gap-2 rounded-full bg-brand px-5 py-3 text-sm font-medium text-white shadow-sm hover:bg-brand-dark transition-all"
                    >
                      <UserPlus className="h-4 w-4" />
                      Add Profile
                    </button>
                  )}
                  {activeTab === 'appointments' && (
                    <button
                      onClick={() => { setEditingAppt(null); setApptModalOpen(true); }}
                      className="inline-flex items-center gap-2 rounded-full bg-brand px-5 py-3 text-sm font-medium text-white shadow-sm hover:bg-brand-dark transition-all"
                    >
                      <Calendar className="h-4 w-4" />
                      New Booking
                    </button>
                  )}
                  <button
                    onClick={handleSync}
                    disabled={loading}
                    className="inline-flex items-center gap-2 rounded-full bg-slate-900 px-5 py-3 text-sm font-medium text-white shadow-sm hover:bg-slate-800 transition-all disabled:opacity-50"
                  >
                    <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                    Sync
                  </button>
                </div>
              </div>
            </section>

            {/* Stats Grid */}
            <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              <StatCard
                icon={<Users className="h-5 w-5" />}
                title="Total Patients"
                value={loading ? '...' : String(users.filter(u => u.role === 'patient').length)}
                caption="Registered on platform"
                tone="teal"
              />
              <StatCard
                icon={<UserPlus className="h-5 w-5" />}
                title="Total Doctors"
                value={loading ? '...' : String(users.filter(u => u.role === 'doctor').length)}
                caption="Medical professionals"
                tone="blue"
              />
              <StatCard
                icon={<Calendar className="h-5 w-5" />}
                title="Appointments"
                value={loading ? '...' : String(appointments.length)}
                caption="Scheduled sessions"
                tone="emerald"
              />
              <StatCard
                icon={<Ticket className="h-5 w-5" />}
                title="Pending Tickets"
                value={loading ? '...' : String(tickets.filter(t => t.status === 'pending').length)}
                caption="Requires attention"
                tone="slate"
              />
            </section>

            {/* Main Content Area */}
            <div className="space-y-6">
              <div className="flex items-center justify-between border-b border-teal-100/50 pb-2">
                <div className="flex flex-wrap gap-2">
                  {(['patients', 'doctors', 'appointments', 'payments', 'tickets'] as TabKey[]).map((tab) => (
                    <button
                      key={tab}
                      onClick={() => { setActiveTab(tab); setSearchQuery(''); }}
                      className={`px-4 py-2 rounded-full text-sm font-semibold transition-all ${
                        activeTab === tab 
                          ? 'bg-brand/10 text-brand' 
                          : 'text-slate-500 hover:bg-slate-100'
                      }`}
                    >
                      {tab.charAt(0).toUpperCase() + tab.slice(1)}
                    </button>
                  ))}
                </div>
                
                {error && (
                  <div className="flex items-center gap-2 text-xs font-medium text-red-500 bg-red-50 px-3 py-1.5 rounded-full border border-red-100">
                    <XCircle className="h-3.5 w-3.5" />
                    {error}
                  </div>
                )}
              </div>

              <section className="rounded-[2rem] border border-white bg-white shadow-[0_20px_60px_rgba(8,47,73,0.06)] overflow-hidden min-h-[400px]">
                {loading ? (
                  <div className="flex flex-col items-center justify-center py-32 space-y-4">
                    <div className="h-12 w-12 animate-spin rounded-full border-4 border-brand border-t-transparent" />
                    <p className="text-sm font-medium text-slate-400">Loading resources...</p>
                  </div>
                ) : (
                  <div className="overflow-x-auto">
                    {activeTab === 'patients' && (
                      <div className="p-2">
                        <UserTable 
                          users={filteredPatients} 
                          onDeactivate={handleDeactivateUser} 
                          onReactivate={handleReactivateUser} 
                          onEdit={(u: any) => { setEditingUser(u); setUserModalOpen(true); }}
                          actionId={actionLoadingId} 
                        />
                      </div>
                    )}
                    {activeTab === 'doctors' && (
                      <div className="p-2">
                        <UserTable 
                          users={filteredDoctors} 
                          onDeactivate={handleDeactivateUser} 
                          onReactivate={handleReactivateUser} 
                          onEdit={(u: any) => { setEditingUser(u); setUserModalOpen(true); }}
                          actionId={actionLoadingId} 
                        />
                      </div>
                    )}
                    {activeTab === 'appointments' && (
                      <AppointmentList 
                        appointments={appointments} 
                        onEdit={(a: any) => { setEditingAppt(a); setApptModalOpen(true); }}
                        onCancel={handleCancelAppt}
                      />
                    )}
                    {activeTab === 'payments' && (
                      <div className="p-2">
                        <PaymentTable transactions={transactions} />
                      </div>
                    )}
                    {activeTab === 'tickets' && (
                      <div className="p-2">
                        <TicketTable tickets={tickets} onResolve={handleResolveTicket} actionId={actionLoadingId} />
                      </div>
                    )}
                  </div>
                )}
              </section>
            </div>
            
            {/* Modals */}
            {userModalOpen && (
              <UserModal 
                isOpen={userModalOpen} 
                onClose={() => setUserModalOpen(false)} 
                onSave={handleSaveUser} 
                editingUser={editingUser}
                saving={saving}
              />
            )}
            {apptModalOpen && (
              <ApptModal 
                isOpen={apptModalOpen} 
                onClose={() => setApptModalOpen(false)} 
                onSave={handleSaveAppt} 
                editingAppt={editingAppt}
                saving={saving}
                patients={users.filter(u => u.role === 'patient')}
                doctors={users.filter(u => u.role === 'doctor' && u.is_verified)}
              />
            )}
          </div>
        </main>
      </div>
    </div>
  );
}

// ─── Sub-Components ─────────────────────────────────────────────────────────

function UserModal({ isOpen, onClose, onSave, editingUser, saving }: any) {
  const [form, setForm] = useState({
    first_name: editingUser?.first_name || '',
    last_name: editingUser?.last_name || '',
    email: editingUser?.email || '',
    role: editingUser?.role || 'patient',
    password: '',
  });

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/40 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-white w-full max-w-md rounded-[2.5rem] shadow-2xl overflow-hidden border border-teal-50 animate-in zoom-in-95 duration-200">
        <div className="p-8 border-b border-slate-50 flex items-center justify-between bg-white">
          <div>
            <h3 className="text-xl font-bold text-slate-900">{editingUser ? 'Edit Profile' : 'Create User'}</h3>
            <p className="text-xs text-slate-400 font-medium mt-1">Fill in the administrative records</p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-slate-100 rounded-full transition-colors"><X className="h-5 w-5 text-slate-400" /></button>
        </div>
        
        <form onSubmit={(e) => { e.preventDefault(); onSave(form); }} className="p-8 space-y-6">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">First Name</label>
              <input required value={form.first_name} onChange={e=>setForm({...form, first_name: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20" />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Last Name</label>
              <input required value={form.last_name} onChange={e=>setForm({...form, last_name: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20" />
            </div>
          </div>
          
          <div className="space-y-2">
            <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Email Address</label>
            <input type="email" required value={form.email} onChange={e=>setForm({...form, email: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20" />
          </div>

          {!editingUser && (
            <>
              <div className="space-y-2">
                <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Account Role</label>
                <select value={form.role} onChange={e=>setForm({...form, role: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20">
                  <option value="patient">Patient</option>
                  <option value="doctor">Doctor</option>
                  <option value="admin">Administrator</option>
                </select>
              </div>
              <div className="space-y-2">
                <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Temporary Password</label>
                <input type="password" required value={form.password} onChange={e=>setForm({...form, password: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20" />
              </div>
            </>
          )}

          <div className="pt-4 flex gap-3">
             <button type="button" onClick={onClose} className="flex-1 px-6 py-4 border border-slate-100 font-bold text-slate-600 rounded-2xl hover:bg-slate-50 transition-all">Cancel</button>
             <button disabled={saving} className="flex-1 px-6 py-4 bg-brand text-white font-bold rounded-2xl hover:bg-brand-dark transition-all shadow-lg shadow-brand/20 disabled:opacity-50">
               {saving ? 'Processing...' : editingUser ? 'Update Profile' : 'Create Account'}
             </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function ApptModal({ isOpen, onClose, onSave, editingAppt, saving, patients, doctors }: any) {
  const [form, setForm] = useState({
    patient_id: editingAppt?.patient_id || '',
    doctor_id: editingAppt?.doctor_id || '',
    scheduled_at: editingAppt?.scheduled_at ? new Date(editingAppt.scheduled_at).toISOString().slice(0, 16) : '',
    status: editingAppt?.status || 'confirmed',
    reason: editingAppt?.reason || '',
  });

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/40 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-white w-full max-w-md rounded-[2.5rem] shadow-2xl overflow-hidden border border-teal-50 animate-in zoom-in-95 duration-200">
        <div className="p-8 border-b border-slate-50 bg-white flex items-center justify-between">
          <div>
            <h3 className="text-xl font-bold text-slate-900">{editingAppt ? 'Edit Appointment' : 'Manual Booking'}</h3>
            <p className="text-xs text-slate-400 font-medium mt-1">Schedule and manage medical sessions</p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-slate-100 rounded-full transition-colors"><X className="h-5 w-5 text-slate-400" /></button>
        </div>
        
        <form onSubmit={(e) => { e.preventDefault(); onSave(form); }} className="p-8 space-y-6">
          {!editingAppt && (
            <>
               <div className="space-y-2">
                <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Patient</label>
                <select required value={form.patient_id} onChange={e=>setForm({...form, patient_id: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20">
                  <option value="">Select Patient</option>
                  {patients.map((p: any) => <option key={p.id} value={p.id}>{p.first_name} {p.last_name} ({p.email})</option>)}
                </select>
              </div>
              <div className="space-y-2">
                <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Doctor</label>
                <select required value={form.doctor_id} onChange={e=>setForm({...form, doctor_id: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20">
                  <option value="">Select Doctor</option>
                  {doctors.map((d: any) => <option key={d.id} value={d.id}>{d.first_name} {d.last_name} ({d.email})</option>)}
                </select>
              </div>
            </>
          )}

          <div className="space-y-2">
            <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Date & Time</label>
            <input type="datetime-local" required value={form.scheduled_at} onChange={e=>setForm({...form, scheduled_at: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20" />
          </div>

          <div className="space-y-2">
            <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Status</label>
            <select value={form.status} onChange={e=>setForm({...form, status: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20">
              <option value="pending">Pending</option>
              <option value="confirmed">Confirmed</option>
              <option value="completed">Completed</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>

          <div className="space-y-2">
            <label className="text-xs font-bold text-slate-400 uppercase tracking-wider">Notes/Reason</label>
            <textarea value={form.reason} onChange={e=>setForm({...form, reason: e.target.value})} className="w-full px-4 py-3 bg-slate-50 border border-slate-100 rounded-2xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20 min-h-[80px]" />
          </div>

          <div className="pt-4 flex gap-3">
             <button type="button" onClick={onClose} className="flex-1 px-6 py-4 border border-slate-100 font-bold text-slate-600 rounded-2xl hover:bg-slate-50 transition-all">Cancel</button>
             <button disabled={saving} className="flex-1 px-6 py-4 bg-brand text-white font-bold rounded-2xl hover:bg-brand-dark transition-all shadow-lg shadow-brand/20 disabled:opacity-50">
               {saving ? 'Processing...' : editingAppt ? 'Update Booking' : 'Finalize Booking'}
             </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function StatCard({
  icon,
  title,
  value,
  caption,
  tone,
}: {
  icon: React.ReactNode;
  title: string;
  value: string;
  caption: string;
  tone: 'teal' | 'blue' | 'emerald' | 'slate';
}) {
  const toneStyles: Record<'teal' | 'blue' | 'emerald' | 'slate', string> = {
    teal: 'text-brand',
    blue: 'text-sky-600',
    emerald: 'text-emerald-600',
    slate: 'text-slate-600',
  };

  return (
    <div className={`rounded-[1.75rem] border border-slate-100 bg-white ${toneStyles[tone]} p-5 shadow-sm`}>
      <div className="flex items-center justify-between gap-3">
        <div className="flex h-11 w-11 items-center justify-center rounded-full bg-white shadow-sm border border-slate-100">
          {icon}
        </div>
      </div>
      <p className="mt-5 text-sm font-medium text-slate-500">{title}</p>
      <p className="mt-2 text-2xl font-bold tracking-tight text-slate-900">{value}</p>
      <p className="mt-1 text-xs leading-5 text-slate-400 font-medium">{caption}</p>
    </div>
  );
}

function UserTable({ users, onDeactivate, onReactivate, onEdit, actionId }: any) {
  if (users.length === 0) return <Empty message="No matching records found" />;
  return (
    <table className="w-full text-left text-sm border-separate border-spacing-y-2">
      <thead className="text-slate-400 font-semibold text-[11px] uppercase tracking-wider">
        <tr>
          <th className="px-6 py-2">Identity</th>
          <th className="px-5 py-2">Status</th>
          <th className="px-5 py-2">Verification</th>
          <th className="px-5 py-2 text-right">Actions</th>
        </tr>
      </thead>
      <tbody className="">
        {users.map((u: any) => (
          <tr key={u.id} className="group hover:bg-slate-50/50 transition-colors bg-white">
            <td className="px-6 py-4 rounded-l-2xl border-y border-l border-slate-50 shadow-sm first:border-none">
              <div className="flex items-center gap-4">
                <div className="w-10 h-10 rounded-xl bg-brand/5 text-brand flex items-center justify-center font-bold text-sm">
                  {u.first_name?.[0] || 'U'}
                </div>
                <div>
                  <p className="font-bold text-slate-800">{u.first_name} {u.last_name}</p>
                  <p className="text-xs text-slate-400 font-medium">{u.email}</p>
                </div>
              </div>
            </td>
            <td className="px-5 py-4 border-y border-slate-50 shadow-sm">
              <span className={`inline-flex px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${u.is_active ? 'bg-emerald-50 text-emerald-600' : 'bg-red-50 text-red-600'}`}>
                {u.is_active ? 'Active' : 'Deactivated'}
              </span>
            </td>
            <td className="px-5 py-4 border-y border-slate-50 shadow-sm">
              <div className="flex items-center gap-2 text-xs font-semibold text-slate-600">
                <ShieldCheck className={`h-4 w-4 ${u.is_verified ? 'text-emerald-500' : 'text-slate-300'}`} />
                {u.is_verified ? 'Verified' : 'Pending'}
              </div>
            </td>
            <td className="px-5 py-4 rounded-r-2xl border-y border-r border-slate-50 shadow-sm text-right">
              <div className="flex justify-end gap-2 items-center">
                <button 
                  onClick={() => onEdit(u)}
                  className="p-2 text-slate-400 hover:text-brand hover:bg-brand/5 rounded-lg transition-all"
                  title="Edit Profile"
                >
                  <Pencil className="h-4 w-4" />
                </button>
                <div className="relative group/actions">
                  {u.is_active ? (
                    <button 
                      onClick={()=>onDeactivate(u.id)} 
                      disabled={actionId===u.id} 
                      className="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-all"
                      title="Deactivate Account"
                    >
                      <XCircle className="h-4.5 w-4.5" />
                    </button>
                  ) : (
                    <button 
                      onClick={()=>onReactivate(u.id)} 
                      disabled={actionId===u.id} 
                      className="px-4 py-1.5 bg-brand text-white text-[11px] rounded-lg font-bold hover:bg-brand-dark transition-all disabled:opacity-50"
                    >
                      Reactivate
                    </button>
                  )}
                </div>
              </div>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function AppointmentList({ appointments, onEdit, onCancel }: any) {
  if (appointments.length === 0) return <Empty message="No appointments scheduled" />;
  return (
    <div className="p-6 space-y-4">
      {appointments.map((a: any) => (
        <div key={a.id} className="flex flex-col gap-4 p-5 rounded-[1.75rem] border border-slate-100 bg-white hover:border-brand/20 transition-all sm:flex-row sm:items-center">
          <div className="flex h-16 w-16 flex-col items-center justify-center rounded-2xl bg-white border border-brand/10 text-center shadow-sm shrink-0">
            <span className="text-[11px] font-medium uppercase tracking-[0.2em] text-slate-400">
              {new Date(a.scheduled_at).toLocaleDateString("en-US", { month: "short" })}
            </span>
            <span className="text-xl font-bold text-slate-900">
              {new Date(a.scheduled_at).getDate()}
            </span>
          </div>

          <div className="min-w-0 flex-1 space-y-1">
            <div className="flex flex-wrap items-center gap-2">
              <h4 className="text-base font-bold text-slate-900 truncate">
                {appointmentTitle(a)}
              </h4>
              <span className={`inline-flex px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider border ${statusTone(a.status)}`}>
                {appoinmentStatusLabel(a.status)}
              </span>
              {a.payment_status && (
                <span className={`inline-flex px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider border ${statusTone(a.payment_status)}`}>
                  Payment {a.payment_status}
                </span>
              )}
            </div>
            
            <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500 font-medium">
              <span className="flex items-center gap-1.5"><Clock className="h-3.5 w-3.5 text-brand" /> {fmtTime(a.scheduled_at)} · {a.duration_minutes || 30}m</span>
              <span className="flex items-center gap-1.5"><Activity className="h-3.5 w-3.5 text-brand" /> {modeLabel(a.consultation_mode)}</span>
              <span className="flex items-center gap-1.5"><ShieldCheck className="h-3.5 w-3.5 text-brand" /> Ref: {compactId(a.id, 10).toUpperCase()}</span>
            </div>
          </div>

          <div className="flex items-center gap-3">
             <button 
                onClick={() => onEdit(a)}
                className="p-2 text-slate-400 hover:text-brand hover:bg-brand/5 rounded-lg transition-all"
                title="Edit Booking"
              >
                <Pencil className="h-4 w-4" />
              </button>
              {a.status !== 'cancelled' && (
                <button 
                  onClick={() => onCancel(a.id)}
                  className="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-all"
                  title="Cancel Appointment"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              )}
          </div>
        </div>
      ))}
    </div>
  );
}

function PaymentTable({ transactions }: any) {
  if (transactions.length === 0) return <Empty message="No transaction history" />;
  return (
    <div className="p-2">
      <table className="w-full text-left text-sm border-separate border-spacing-y-2">
        <thead className="text-slate-400 font-semibold text-[11px] uppercase tracking-wider">
          <tr>
            <th className="px-6 py-2">Transaction Details</th>
            <th className="px-5 py-2">Status</th>
            <th className="px-5 py-2 text-right">Date & Time</th>
          </tr>
        </thead>
        <tbody>
          {transactions.map((t: any) => (
            <tr key={t.id} className="bg-white border-y border-slate-50 shadow-sm first:rounded-t-2xl">
              <td className="px-6 py-4 rounded-l-2xl border-y border-l border-slate-50 shadow-sm first:border-none">
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center">
                    <CreditCard className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="font-bold text-slate-900">{t.currency || 'LKR'} {Number(t.amount || 0).toLocaleString()}</p>
                    <p className="text-[10px] font-bold text-slate-400 uppercase tracking-tighter">{t.provider || 'Gateway'} Transaction</p>
                  </div>
                </div>
              </td>
              <td className="px-5 py-4 border-y border-slate-50 shadow-sm">
                <span className={`inline-flex px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${t.status === 'succeeded' ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-50 text-slate-400'}`}>
                  {t.status || 'pending'}
                </span>
              </td>
              <td className="px-5 py-4 rounded-r-2xl border-y border-r border-slate-50 shadow-sm text-right">
                <p className="text-sm font-bold text-slate-800">{fmtDate(t.created_at)}</p>
                <p className="text-[10px] font-medium text-slate-400">{fmtTime(t.created_at)}</p>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function TicketTable({ tickets, onResolve, actionId }: any) {
  if (tickets.length === 0) return <Empty message="No active support tickets" />;
  return (
    <div className="p-2">
      <table className="w-full text-left text-sm border-separate border-spacing-y-2">
        <thead className="text-slate-400 font-semibold text-[11px] uppercase tracking-wider">
          <tr>
            <th className="px-6 py-2">Requester</th>
            <th className="px-5 py-2">Problem Description</th>
            <th className="px-5 py-2 text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {tickets.map((t: any) => (
            <tr key={t.id} className="bg-white group transition-all">
              <td className="px-6 py-4 rounded-l-2xl border-y border-l border-slate-50 shadow-sm">
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold">
                    {t.name?.[0] || 'T'}
                  </div>
                  <div>
                    <p className="font-bold text-slate-800">{t.name}</p>
                    <p className="text-[10px] text-slate-400 font-medium tracking-wide">{t.email}</p>
                  </div>
                </div>
              </td>
              <td className="px-5 py-4 border-y border-slate-50 shadow-sm max-w-sm">
                <div className="text-xs text-slate-600 font-medium bg-slate-50/50 p-2.5 rounded-xl italic border border-slate-100/50 line-clamp-2">
                  "{t.reason}"
                </div>
              </td>
              <td className="px-5 py-4 rounded-r-2xl border-y border-r border-slate-50 shadow-sm text-right">
                {t.status === 'pending' ? (
                  <button 
                    onClick={()=>onResolve(t)} 
                    disabled={actionId===t.id} 
                    className="px-4 py-2 bg-brand text-white text-[11px] rounded-xl font-bold hover:bg-brand-dark transition-all shadow-sm hover:shadow-brand/20 disabled:opacity-50"
                  >
                    Resolve Case
                  </button>
                ) : (
                  <span className="inline-flex items-center gap-1.5 text-emerald-500 text-[11px] font-bold bg-emerald-50 px-3 py-1.5 rounded-full">
                    <CheckCircle className="h-4 w-4" /> Resolved
                  </span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function Empty({ message }: { message: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-32 text-center">
      <div className="mb-6 flex h-20 w-20 items-center justify-center rounded-[2rem] border border-dashed border-slate-200 bg-slate-50 text-slate-300">
        <Activity className="h-8 w-8" />
      </div>
      <h3 className="text-lg font-bold text-slate-900">{message}</h3>
      <p className="mt-2 text-sm text-slate-400 max-w-xs px-4">There aren't any records matching your current filter in this workspace.</p>
    </div>
  );
}
