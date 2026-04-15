import { useState, useEffect } from 'react';
import { LogOut, User, Activity, Calendar, ClipboardList, CreditCard, CheckCircle, Clock, AlertCircle, Scale } from 'lucide-react';
import { Link } from 'react-router-dom';
import { patientApi, type PatientProfile } from '../api/patient';
import { paymentApi, type Payment } from '../api/payment';
import { appointmentApi, type Appointment } from '../api/appointments';

export default function Payments() {
  const [profile, setProfile] = useState<PatientProfile | null>(null);
  const [payments, setPayments] = useState<Payment[]>([]);
  const [appointments, setAppointments] = useState<Record<string, Appointment>>({});
  const [loading, setLoading] = useState(true);
  const [userRole, setUserRole] = useState<string | null>(null);

  useEffect(() => {
    const userJson = localStorage.getItem('user');
    if (userJson) {
      const user = JSON.parse(userJson);
      setUserRole(user.role);
    }
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const patientData = await patientApi.getProfile();
      setProfile(patientData);

      // Fetch payments and appointments in parallel
      const [paymentData, apptData] = await Promise.all([
        paymentApi.listPayments(patientData.user_id),
        appointmentApi.listAppointments()
      ]);

      setPayments(paymentData);
      
      // Create a map of appointments for easy lookup
      const apptMap: Record<string, Appointment> = {};
      apptData.forEach(appt => {
        apptMap[appt.id] = appt;
      });
      setAppointments(apptMap);
    } catch (err) {
      console.error('Failed to fetch payment data:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    window.location.href = '/auth/login';
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusStyle = (status: string) => {
    const s = status.toLowerCase();
    switch (s) {
      case 'completed':
      case 'confirmed':
      case 'paid':
        return 'bg-green-100 text-green-700 border-green-200';
      case 'pending':
        return 'bg-amber-100 text-amber-700 border-amber-200';
      case 'cancelled':
      case 'failed':
      case 'expired':
        return 'bg-red-100 text-red-700 border-red-200';
      default:
        return 'bg-gray-100 text-gray-700 border-gray-200';
    }
  };

  const getStatusIcon = (status: string) => {
    const s = status.toLowerCase();
    switch (s) {
      case 'completed':
      case 'confirmed':
      case 'paid':
        return <CheckCircle className="h-4 w-4 mr-1.5" />;
      case 'pending':
        return <Clock className="h-4 w-4 mr-1.5" />;
      default:
        return <AlertCircle className="h-4 w-4 mr-1.5" />;
    }
  };

  return (
    <div className="min-h-screen bg-[#f6f8fa] flex font-sans">
      {/* Sidebar */}
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
          <Link
            to="/dashboard"
            className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
          >
            <Activity className="h-[18px] w-[18px]" /> Dashboard
          </Link>
          <Link
            to="/profile"
            className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
          >
            <User className="h-[18px] w-[18px]" /> Profile
          </Link>
          <Link
            to="/appointments"
            className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
          >
            <Calendar className="h-[18px] w-[18px]" /> Appointments
          </Link>
          {userRole !== 'doctor' && (
            <Link
              to="/payments"
              className="flex items-center gap-3 px-3 py-2.5 bg-brand/10 text-brand rounded-xl font-semibold transition-all text-sm shadow-sm"
            >
              <CreditCard className="h-[18px] w-[18px]" /> Payments
            </Link>
          )}
          <Link
            to="/bmi-calculator"
            className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
          >
            <Scale className="h-[18px] w-[18px]" /> BMI Calculator
          </Link>
          <a
            href="#"
            className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
          >
            <ClipboardList className="h-[18px] w-[18px]" /> Records
          </a>
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

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-h-screen">
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-8 bg-white/80 backdrop-blur-md sticky top-0 z-10">
          <h2 className="text-xl font-semibold text-gray-800">Payment History</h2>
          <div className="flex items-center space-x-4">
            <p className="text-sm font-medium text-gray-600 hidden sm:block">
              {profile ? `${profile.first_name} ${profile.last_name}` : ''}
            </p>
            <div className="w-10 h-10 bg-brand-light rounded-full flex items-center justify-center text-brand border-2 border-brand/20">
              <User className="h-6 w-6" />
            </div>
          </div>
        </header>

        <main className="p-8">
          <div className="mb-8">
            <h3 className="text-2xl font-bold text-gray-900">Billing Overview</h3>
            <p className="text-gray-500 mt-1">Manage your payments and download receipts.</p>
          </div>

          {loading ? (
            <div className="flex flex-col items-center justify-center py-20">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand"></div>
              <p className="text-gray-500 mt-4">Loading data...</p>
            </div>
          ) : (
            <div className="bg-white rounded-3xl border border-gray-200 shadow-sm overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead className="bg-gray-50 border-b border-gray-200">
                    <tr>
                      <th className="px-6 py-4 text-xs font-bold text-gray-500 uppercase tracking-wider">Date</th>
                      <th className="px-6 py-4 text-xs font-bold text-gray-500 uppercase tracking-wider">Appointment ID</th>
                      <th className="px-6 py-4 text-xs font-bold text-gray-500 uppercase tracking-wider">Amount</th>
                      <th className="px-6 py-4 text-xs font-bold text-gray-500 uppercase tracking-wider text-center">Appt Status</th>
                      <th className="px-6 py-4 text-xs font-bold text-gray-500 uppercase tracking-wider text-right">Payment</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {payments.length > 0 ? (
                      payments.map((payment) => {
                        const appt = appointments[payment.appointment_id];
                        return (
                          <tr key={payment.id} className="hover:bg-gray-50 transition-colors">
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-medium">
                              {formatDate(payment.created_at)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 font-mono">
                              {payment.appointment_id}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-bold">
                              {payment.currency} {payment.amount.toFixed(2)}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-center">
                              {appt ? (
                                <span className={`inline-flex items-center px-3 py-1 rounded-full text-[10px] font-bold border uppercase ${getStatusStyle(appt.status)}`}>
                                  {getStatusIcon(appt.status)}
                                  {appt.status}
                                </span>
                              ) : (
                                <span className={`inline-flex items-center px-3 py-1 rounded-full text-[10px] font-bold border uppercase translate-y-[-1px] ${getStatusStyle(payment.status)}`}>
                                  {getStatusIcon(payment.status)}
                                  {payment.status === 'completed' ? 'confirmed' : payment.status}
                                </span>
                              )}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-right">
                              <span className={`text-xs font-bold uppercase ${
                                (appt?.payment_status || payment.status) === 'paid' || (appt?.payment_status || payment.status) === 'completed' 
                                  ? 'text-green-600' 
                                  : 'text-amber-600'
                              }`}>
                                {appt?.payment_status || (payment.status === 'completed' ? 'paid' : payment.status)}
                              </span>
                            </td>
                          </tr>
                        );
                      })
                    ) : (
                      <tr>
                        <td colSpan={5} className="px-6 py-20 text-center">
                          <CreditCard className="h-12 w-12 text-gray-300 mx-auto mb-4" />
                          <p className="text-gray-500 font-medium">No payment records found.</p>
                          <Link to="/appointments" className="text-brand text-sm font-bold mt-2 inline-block hover:underline">
                            Book an appointment to start
                          </Link>
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
