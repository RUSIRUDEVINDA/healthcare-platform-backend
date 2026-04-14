import { useState, useEffect } from 'react';
import { LogOut, User, Activity, Calendar, ClipboardList, Shield } from 'lucide-react';
import { Link } from 'react-router-dom';
import { patientApi, type PatientProfile } from '../api/patient';
import { doctorApi, type DoctorProfile } from '../api/doctor';

export default function Dashboard() {
  const [profile, setProfile] = useState<PatientProfile | DoctorProfile | null>(null);
  const [role, setRole] = useState<string | null>(null);

  useEffect(() => {
    const userString = localStorage.getItem('user');
    const user = userString ? JSON.parse(userString) : null;
    const userRole = user?.role;
    setRole(userRole);

    const fetchProfile = async () => {
      try {
        let data;
        if (userRole === 'doctor') {
          data = await doctorApi.getProfile();
        } else {
          data = await patientApi.getProfile();
        }
        setProfile(data);
      } catch (err) {
        console.error('Failed to fetch profile:', err);
      }
    };
    if (userRole) {
      fetchProfile();
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    window.location.href = '/auth/login';
  };

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Sidebar */}
      <div className="w-64 bg-white border-r border-gray-200 hidden md:flex flex-col">
        <div className="p-6">
          <Link to="/dashboard">
            <h1 className="text-2xl font-bold text-brand">Healthcare</h1>
          </Link>
        </div>
        <nav className="flex-1 px-4 space-y-2">
          <Link to="/dashboard" className="flex items-center px-4 py-3 bg-brand/10 text-brand rounded-xl font-medium">
            <Activity className="mr-3 h-5 w-5" /> Dashboard
          </Link>
          <Link to="/profile" className="flex items-center px-4 py-3 text-gray-600 hover:bg-gray-100 rounded-xl transition-colors">
            <User className="mr-3 h-5 w-5" /> Profile
          </Link>
          <Link to="/appointments" className="flex items-center px-4 py-3 text-gray-600 hover:bg-gray-100 rounded-xl transition-colors">
            <Calendar className="mr-3 h-5 w-5" /> Appointments
          </Link>
          <a href="#" className="flex items-center px-4 py-3 text-gray-600 hover:bg-gray-100 rounded-xl transition-colors">
            <ClipboardList className="mr-3 h-5 w-5" /> Records
          </a>
        </nav>
        <div className="p-4 border-t border-gray-200">
          <button onClick={handleLogout} className="flex items-center w-full px-4 py-3 text-red-600 hover:bg-red-50 rounded-xl transition-colors">
            <LogOut className="mr-3 h-5 w-5" /> Sign Out
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col h-screen overflow-hidden">
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-8 bg-white/80 backdrop-blur-md sticky top-0 z-10">
          <h2 className="text-xl font-semibold text-gray-800">Overview</h2>
          <div className="flex items-center space-x-4">
            <div className="w-10 h-10 bg-brand-light rounded-full flex items-center justify-center text-brand">
              <User className="h-6 w-6" />
            </div>
          </div>
        </header>
        
        <main className="p-8 overflow-y-auto">
          <div className="mb-10">
            <h3 className="text-3xl font-bold text-gray-900">
              Welcome back, {
                role === 'doctor' 
                  ? (profile as DoctorProfile)?.name || 'Doctor' 
                  : (profile as PatientProfile)?.first_name || 'Patient'
              }
            </h3>
            <p className="text-gray-500 mt-2">
              {role === 'doctor' ? "Here's your schedule for today." : "Here's what's happening with your health today."}
            </p>
          </div>

          {/* Quick Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-10">
            <div className="bg-white p-6 rounded-3xl border border-gray-100 shadow-sm hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-blue-50 rounded-2xl flex items-center justify-center text-blue-600 mb-4">
                <Calendar className="h-6 w-6" />
              </div>
              <p className="text-sm font-medium text-gray-500">Next Appointment</p>
              <h4 className="text-xl font-bold text-gray-900 mt-1">Oct 24, 2026</h4>
              <p className="text-sm text-gray-400 mt-2">Dr. Sarah Johnson • 10:30 AM</p>
            </div>
            <div className="bg-white p-6 rounded-3xl border border-gray-100 shadow-sm hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-red-50 rounded-2xl flex items-center justify-center text-red-600 mb-4">
                <Activity className="h-6 w-6" />
              </div>
              <p className="text-sm font-medium text-gray-500">Latest Health Score</p>
              <h4 className="text-xl font-bold text-gray-900 mt-1">92/100</h4>
              <p className="text-sm text-green-500 mt-2">↑ 4% from last month</p>
            </div>
            <div className="bg-white p-6 rounded-3xl border border-gray-100 shadow-sm hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-purple-50 rounded-2xl flex items-center justify-center text-purple-600 mb-4">
                <ClipboardList className="h-6 w-6" />
              </div>
              <p className="text-sm font-medium text-gray-500">Unread Reports</p>
              <h4 className="text-xl font-bold text-gray-900 mt-1">2 New</h4>
              <p className="text-sm text-gray-400 mt-2">Lab results from yesterday</p>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-10">
            {/* Upcoming Appointments */}
            <div className="bg-white rounded-3xl border border-gray-100 shadow-sm p-8">
              <div className="flex items-center justify-between mb-8">
                <h4 className="text-xl font-bold text-gray-900">Upcoming Appointments</h4>
                <button className="text-brand font-semibold text-sm hover:underline">View All</button>
              </div>
              <div className="space-y-6">
                {[1, 2].map((i) => (
                  <div key={i} className="flex items-center p-4 rounded-2xl hover:bg-gray-50 transition-colors border border-transparent hover:border-gray-100">
                    <div className="w-14 h-14 bg-gray-100 rounded-xl mr-4 flex flex-col items-center justify-center">
                      <span className="text-xs font-bold text-gray-400 uppercase">Oct</span>
                      <span className="text-lg font-bold text-gray-900">{23 + i}</span>
                    </div>
                    <div className="flex-1">
                      <h5 className="font-bold text-gray-900">General Checkup</h5>
                      <p className="text-sm text-gray-500">Dr. {i === 1 ? 'Sarah Johnson' : 'Michael Chen'}</p>
                    </div>
                    <div className="text-right text-sm">
                      <p className="font-semibold text-gray-900">{9 + i}:00 AM</p>
                      <span className="px-2 py-1 bg-green-50 text-green-600 rounded-md text-xs font-bold uppercase">Confirmed</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Recent Records */}
            <div className="bg-white rounded-3xl border border-gray-100 shadow-sm p-8">
              <div className="flex items-center justify-between mb-8">
                <h4 className="text-xl font-bold text-gray-900">Recent Medical Records</h4>
                <button className="text-brand font-semibold text-sm hover:underline">See More</button>
              </div>
              <div className="space-y-6">
                {[1, 2].map((i) => (
                  <div key={i} className="flex items-center p-4 rounded-2xl hover:bg-gray-50 transition-colors border border-transparent hover:border-gray-100">
                    <div className="w-12 h-12 bg-brand-light rounded-xl mr-4 flex items-center justify-center text-brand">
                      <ClipboardList className="h-6 w-6" />
                    </div>
                    <div className="flex-1">
                      <h5 className="font-bold text-gray-900">{i === 1 ? 'Blood Analysis' : 'COVID-19 Vaccination'}</h5>
                      <p className="text-sm text-gray-500">Oct {15 - i}, 2026 • PDF Report</p>
                    </div>
                    <button className="w-10 h-10 flex items-center justify-center rounded-full hover:bg-gray-100 text-gray-400">
                      <Shield className="h-5 w-5" />
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

