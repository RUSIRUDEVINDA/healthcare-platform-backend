import { LogOut, User, Activity, Calendar, ClipboardList } from 'lucide-react';

export default function Dashboard() {
  const handleLogout = () => {
    localStorage.removeItem('access_token');
    window.location.href = '/auth/login';
  };

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Sidebar */}
      <div className="w-64 bg-white border-r border-gray-200 hidden md:flex flex-col">
        <div className="p-6">
          <h1 className="text-2xl font-bold text-brand">Healthcare</h1>
        </div>
        <nav className="flex-1 px-4 space-y-2">
          <a href="#" className="flex items-center px-4 py-3 bg-brand/10 text-brand rounded-xl font-medium">
            <Activity className="mr-3 h-5 w-5" /> Dashboard
          </a>
          <a href="#" className="flex items-center px-4 py-3 text-gray-600 hover:bg-gray-100 rounded-xl transition-colors">
            <Calendar className="mr-3 h-5 w-5" /> Appointments
          </a>
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
      <div className="flex-1 flex flex-col">
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-8">
          <h2 className="text-xl font-semibold text-gray-800">Overview</h2>
          <div className="flex items-center space-x-4">
            <div className="w-10 h-10 bg-brand-light rounded-full flex items-center justify-center text-brand">
              <User className="h-6 w-6" />
            </div>
          </div>
        </header>
        
        <main className="p-8">
          <div className="bg-white rounded-3xl p-8 shadow-sm border border-gray-100">
            <h3 className="text-2xl font-bold text-gray-900 mb-4">Welcome to your portal</h3>
            <p className="text-gray-600 mb-6">Your authentication was successful. You are now logged in securely.</p>
            <div className="p-4 bg-brand-light/30 rounded-2xl border border-brand/10 inline-block">
              <span className="text-brand font-semibold text-sm">Status: Protected Session Active</span>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
