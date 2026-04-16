import { useEffect, useState } from 'react';
import { NavLink, Outlet, Link } from 'react-router-dom';
import {
    Activity,
    Calendar,
    ClipboardList,
    CreditCard,
    LogOut,
    Menu,
    Stethoscope,
    User,
    Scale,
    X,
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
    const [mobileNavOpen, setMobileNavOpen] = useState(false);

    useEffect(() => {
        if (mobileNavOpen) {
            document.body.style.overflow = 'hidden';
        } else {
            document.body.style.overflow = '';
        }
        return () => {
            document.body.style.overflow = '';
        };
    }, [mobileNavOpen]);

    const handleLogout = () => {
        localStorage.removeItem('access_token');
        localStorage.removeItem('user');
        window.location.href = '/auth/login';
    };

    const navClass = ({ isActive }: { isActive: boolean }) =>
        `flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
            isActive ? 'bg-brand/10 text-brand font-semibold' : 'text-slate-500 hover:bg-teal-50'
        }`;

    const closeMobile = () => setMobileNavOpen(false);

    const navLinks = (onNavigate?: () => void) => (
        <>
            <NavLink to="/dashboard" end className={navClass} onClick={onNavigate}>
                <Activity className="h-[18px] w-[18px] shrink-0" />
                Dashboard
            </NavLink>
            <NavLink to="/appointments" className={navClass} onClick={onNavigate}>
                <Calendar className="h-[18px] w-[18px] shrink-0" />
                Appointments
            </NavLink>
            <NavLink to="/profile" className={navClass} onClick={onNavigate}>
                <User className="h-[18px] w-[18px] shrink-0" />
                Profile
            </NavLink>
            {!isDoctor && (
                <NavLink to="/payments" className={navClass} onClick={onNavigate}>
                    <CreditCard className="h-[18px] w-[18px] shrink-0" />
                    Payments
                </NavLink>
            )}
            {!isDoctor && (
                <NavLink to="/symptom-checker" className={navClass} onClick={onNavigate}>
                    <Stethoscope className="h-[18px] w-[18px] shrink-0" />
                    Symptom checker
                </NavLink>
            )}
            <NavLink to="/records" className={navClass} onClick={onNavigate}>
                <ClipboardList className="h-[18px] w-[18px] shrink-0" />
                Records
            </NavLink>
            <NavLink to="/bmi-calculator" className={navClass} onClick={onNavigate}>
                <Scale className="h-[18px] w-[18px] shrink-0" />
                BMI Calculator
            </NavLink>
        </>
    );

    const brandBlock = (
        <Link
            to="/dashboard"
            onClick={closeMobile}
            className="flex items-center gap-2.5 outline-none ring-brand/30 rounded-lg focus-visible:ring-2"
        >
            <div className="flex h-9 w-9 items-center justify-center rounded-full bg-brand shadow-sm shadow-brand/20">
                <Activity className="h-4 w-4 text-white" />
            </div>
            <span className="text-lg font-medium tracking-tight text-slate-900">MediPulse SriLanka</span>
        </Link>
    );

    return (
        <div className="min-h-screen bg-[#f6f8fa]">
            <header className="sticky top-0 z-50 flex h-14 items-center justify-between border-b border-teal-100 bg-white/95 px-4 backdrop-blur lg:hidden">
                {brandBlock}
                <button
                    type="button"
                    className="flex h-10 w-10 items-center justify-center rounded-full text-slate-700 hover:bg-teal-50"
                    aria-expanded={mobileNavOpen}
                    aria-controls="mobile-app-nav"
                    aria-label={mobileNavOpen ? 'Close menu' : 'Open menu'}
                    onClick={() => setMobileNavOpen((o) => !o)}
                >
                    {mobileNavOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
                </button>
            </header>

            {mobileNavOpen && (
                <div className="fixed inset-0 z-[60] lg:hidden" id="mobile-app-nav">
                    <button
                        type="button"
                        className="absolute inset-0 bg-slate-900/40"
                        aria-label="Close menu"
                        onClick={closeMobile}
                    />
                    <div className="absolute inset-y-0 right-0 flex w-[min(100%,18rem)] flex-col border-l border-teal-100 bg-white shadow-xl">
                        <div className="flex items-center justify-between border-b border-teal-100 px-4 py-4">
                            <span className="text-sm font-semibold text-slate-800">Menu</span>
                            <button
                                type="button"
                                className="rounded-full p-2 text-slate-500 hover:bg-teal-50"
                                aria-label="Close menu"
                                onClick={closeMobile}
                            >
                                <X className="h-5 w-5" />
                            </button>
                        </div>
                        <nav className="flex flex-1 flex-col gap-0.5 overflow-y-auto px-4 py-3">{navLinks(closeMobile)}</nav>
                        <div className="border-t border-teal-100 p-4">
                            <button
                                type="button"
                                onClick={() => {
                                    closeMobile();
                                    handleLogout();
                                }}
                                className="flex w-full items-center gap-2.5 rounded-full px-3 py-2.5 text-sm font-medium text-red-500 transition-colors hover:bg-red-50"
                            >
                                <LogOut className="h-[18px] w-[18px] shrink-0" />
                                Sign Out
                            </button>
                        </div>
                    </div>
                </div>
            )}

            <aside
                className="fixed inset-y-0 left-0 z-40 hidden w-60 flex-col border-r border-teal-100 bg-white/95 shadow-sm backdrop-blur lg:flex"
                aria-label="Main navigation"
            >
                <div className="shrink-0 px-6 pb-5 pt-6">{brandBlock}</div>

                <nav className="flex flex-1 flex-col gap-0.5 overflow-y-auto px-4 py-2">
                    <p className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-[0.32em] text-slate-400">Menu</p>
                    {navLinks()}
                </nav>

                <div className="shrink-0 border-t border-teal-100 p-4">
                    <button
                        type="button"
                        onClick={handleLogout}
                        className="flex w-full items-center gap-2.5 rounded-full px-3 py-2.5 text-sm font-medium text-red-500 transition-colors hover:bg-red-50"
                    >
                        <LogOut className="h-[18px] w-[18px] shrink-0" />
                        Sign Out
                    </button>
                </div>
            </aside>

            <div className="min-h-screen min-w-0 w-full lg:pl-60">
                <Outlet />
            </div>
        </div>
    );
}
