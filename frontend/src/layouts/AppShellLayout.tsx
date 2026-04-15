import { NavLink, Outlet, Link } from 'react-router-dom';
import {
    Activity,
    Calendar,
    ClipboardList,
    CreditCard,
    LogOut,
    User,
} from 'lucide-react';

function readRole(): string | null {
    try {
        const raw = localStorage.getItem('user');
        if (!raw) return null;
        return JSON.parse(raw)?.role ?? null;
    } catch {
        return null;
    }
}

export default function AppShellLayout() {
    const isDoctor = readRole() === 'doctor';

    const handleLogout = () => {
        localStorage.removeItem('access_token');
        localStorage.removeItem('user');
        window.location.href = '/auth/login';
    };

    const linkClass = ({ isActive }: { isActive: boolean }) =>
        `flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm transition-colors ${
            isActive ? 'bg-brand/10 text-brand font-semibold shadow-sm' : 'text-gray-600 hover:bg-gray-50'
        }`;

    return (
        <div className="min-h-screen bg-[#f6f8fa]">
            <aside
                className="fixed inset-y-0 left-0 z-50 hidden w-60 flex-col border-r border-gray-200 bg-white shadow-sm lg:flex"
                aria-label="Main navigation"
            >
                <div className="shrink-0 border-b border-gray-100 px-6 pb-5 pt-6">
                    <Link to="/dashboard" className="flex items-center gap-2.5 outline-none ring-brand/30 focus-visible:ring-2 rounded-lg">
                        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand">
                            <Activity className="h-4 w-4 text-white" />
                        </div>
                        <span className="text-lg font-bold tracking-tight text-gray-900">AyaRX</span>
                    </Link>
                </div>

                <nav className="flex flex-1 flex-col gap-0.5 overflow-y-auto px-4 py-4">
                    <p className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-wider text-gray-400">Menu</p>
                    <NavLink to="/dashboard" end className={linkClass}>
                        <Activity className="h-[18px] w-[18px] shrink-0" />
                        Dashboard
                    </NavLink>
                    <NavLink to="/appointments" className={linkClass}>
                        <Calendar className="h-[18px] w-[18px] shrink-0" />
                        Appointments
                    </NavLink>
                    <NavLink to="/profile" className={linkClass}>
                        <User className="h-[18px] w-[18px] shrink-0" />
                        Profile
                    </NavLink>
                    {!isDoctor && (
                        <NavLink to="/payments" className={linkClass}>
                            <CreditCard className="h-[18px] w-[18px] shrink-0" />
                            Payments
                        </NavLink>
                    )}
                    <NavLink to="/records" className={linkClass}>
                        <ClipboardList className="h-[18px] w-[18px] shrink-0" />
                        Records
                    </NavLink>
                </nav>

                <div className="shrink-0 border-t border-gray-100 p-4">
                    <button
                        type="button"
                        onClick={handleLogout}
                        className="flex w-full items-center gap-2.5 rounded-xl px-3 py-2.5 text-sm font-medium text-red-600 transition-colors hover:bg-red-50"
                    >
                        <LogOut className="h-[18px] w-[18px]" />
                        Sign out
                    </button>
                </div>
            </aside>

            <div className="min-h-screen lg:pl-60">
                <Outlet />
            </div>
        </div>
    );
}
