import { useState, useEffect } from 'react';
import {
    Search,
    Calendar,
    Clock,
    Video,
    MapPin,
    LogOut,
    Activity,
    ClipboardList,
    Plus,
    User,
    ChevronRight,
    ChevronDown,
    Building2,
    Briefcase,
    CreditCard,
    Pencil,
    Trash2,
} from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import type { Slot, Appointment, BookAppointmentRequest } from '../api/appointments';
import { appointmentApi } from '../api/appointments';
import { doctorApi as doctorsListApi, type Doctor } from '../api/doctors';
import { doctorApi as doctorProfileApi, type DoctorProfile } from '../api/doctor';
import { patientApi } from '../api/patient';
import { paymentApi } from '../api/payment';
import { submitPayHereForm } from '../utils/payment';
import BookingModal from '../components/appointments/BookingModal';
import DoctorSlotFormModal from '../components/appointments/DoctorSlotFormModal';

type TabKey = 'doctors' | 'appointments' | 'slots';

const HOSPITAL_FEE = 500;

function readUserRole(): string | null {
    try {
        const raw = localStorage.getItem('user');
        if (!raw) return null;
        return JSON.parse(raw)?.role ?? null;
    } catch {
        return null;
    }
}

function initialAppointmentsTab(): TabKey {
    return readUserRole() === 'doctor' ? 'appointments' : 'doctors';
}

export default function Appointments() {
    const navigate = useNavigate();
    const isDoctor = readUserRole() === 'doctor';
    const [appointments, setAppointments] = useState<Appointment[]>([]);
    const [doctors, setDoctors] = useState<Doctor[]>([]);
    const [doctorSlots, setDoctorSlots] = useState<Record<string, Slot[]>>({});
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedDoctor, setSelectedDoctor] = useState<Doctor | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [activeTab, setActiveTab] = useState<TabKey>(initialAppointmentsTab);
    const [isApptMenuOpen, setIsApptMenuOpen] = useState(true);
    const [isRedirecting, setIsRedirecting] = useState(false);
    const [mySlots, setMySlots] = useState<Slot[]>([]);
    const [doctorProfile, setDoctorProfile] = useState<DoctorProfile | null>(null);
    const [slotModalOpen, setSlotModalOpen] = useState(false);
    const [slotModalMode, setSlotModalMode] = useState<'create' | 'edit'>('create');
    const [slotEditing, setSlotEditing] = useState<Slot | null>(null);

    useEffect(() => {
        fetchData();
    }, []);

    const fetchData = async () => {
        setIsLoading(true);
        try {
            let appts: Appointment[] = [];
            try {
                const apptData = await appointmentApi.listAppointments();
                appts = Array.isArray(apptData) ? apptData : [];
            } catch (err) {
                console.error('Failed to fetch appointments:', err);
            }
            setAppointments(appts);

            if (isDoctor) {
                let profile: DoctorProfile | null = null;
                let slotsMine: Slot[] = [];
                try {
                    profile = await doctorProfileApi.getProfile();
                } catch (err) {
                    console.error('Failed to fetch doctor profile:', err);
                }
                setDoctorProfile(profile);
                try {
                    const s = await appointmentApi.listMySlots();
                    slotsMine = Array.isArray(s) ? s : [];
                } catch (err) {
                    console.error('Failed to fetch my slots:', err);
                }
                slotsMine.sort(
                    (a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
                );
                setMySlots(slotsMine);
                setDoctors([]);
                setDoctorSlots({});
                return;
            }

            let docs: Doctor[] = [];
            try {
                const docResult = await doctorsListApi.listDoctors();
                docs = Array.isArray(docResult) ? docResult : [];
            } catch (err) {
                console.error('Failed to fetch doctors:', err);
            }
            setDoctors(docs);

            if (docs.length > 0) {
                const slotResults: Record<string, Slot[]> = {};
                await Promise.allSettled(
                    docs.map(async (doc) => {
                        try {
                            const slots = await appointmentApi.getDoctorSlots(String(doc.id));
                            slotResults[String(doc.id)] = Array.isArray(slots) ? slots : [];
                        } catch {
                            slotResults[String(doc.id)] = [];
                        }
                    })
                );
                setDoctorSlots(slotResults);
            }
        } catch (error) {
            console.error('Failed to fetch data:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleLogout = () => {
        localStorage.removeItem('access_token');
        window.location.href = '/auth/login';
    };

    const handleBook = async (data: BookAppointmentRequest) => {
        try {
            const appt = await appointmentApi.bookAppointment(data);
            
            if (data.payment_mode === 'pay_now' && appt.id) {
                // Fetch patient profile to get customer details for PayHere
                const profile = await patientApi.getProfile();
                
                // Get checkout parameters from our payment service
                const checkout = await paymentApi.checkout({
                    appointment_id: appt.id,
                    items: `Consultation with ${getDoctorName(data.doctor_id)}`,
                    customer: {
                        first_name: profile.first_name,
                        last_name: profile.last_name,
                        email: profile.email,
                        phone: profile.phone_number || '0000000000',
                        address: profile.address || 'Colombo',
                        city: 'Colombo',
                        country: 'Sri Lanka'
                    }
                });

                // Redirect to PayHere Checkout
                setIsRedirecting(true);
                submitPayHereForm(checkout);
            } else {
                await fetchData();
                setActiveTab('appointments');
            }
        } catch (error: unknown) {
            const apiMessage =
                error && typeof error === 'object' && 'response' in error
                    ? (error as { response?: { data?: { error?: string } } }).response?.data?.error
                    : undefined;
            const fallbackMessage = error instanceof Error ? error.message : 'Booking failed.';
            const message = apiMessage || fallbackMessage;
            console.error('Booking failed:', message, error);
            throw new Error(message);
        }
    };

    const openBooking = (doc: Doctor) => {
        setSelectedDoctor(doc);
        setIsModalOpen(true);
    };

    const getDoctorName = (doctorId: string) => {
        const doc = doctors.find((d) => String(d.id) === doctorId);
        return doc?.name ?? 'Doctor';
    };

    const getDoctorSpecialty = (doctorId: string) => {
        const doc = doctors.find((d) => String(d.id) === doctorId);
        return doc?.specialization ?? '';
    };

    const getPatientDisplayName = (appt: Appointment) => {
        const parts = [appt.patient_first_name, appt.patient_last_name].filter(Boolean);
        return parts.length ? parts.join(' ') : 'Patient';
    };

    const profileHospitalStr = (doctorProfile?.hospital ?? '').trim();

    const handleSaveSlot = async (payload: { start_time: string; end_time: string; hospital: string }) => {
        if (!doctorProfile) {
            throw new Error('Doctor profile not loaded');
        }
        const doctorId = String(doctorProfile.id);
        if (slotModalMode === 'create') {
            await appointmentApi.createSlot({
                doctor_id: doctorId,
                hospital: payload.hospital,
                start_time: payload.start_time,
                end_time: payload.end_time,
            });
        } else if (slotEditing) {
            await appointmentApi.updateSlot(slotEditing.id, {
                start_time: payload.start_time,
                end_time: payload.end_time,
                hospital: payload.hospital,
            });
        }
        await fetchData();
    };

    const handleDeleteSlot = async (slot: Slot) => {
        if (slot.is_booked) return;
        if (!window.confirm('Remove this availability slot? Patients can no longer book it.')) return;
        try {
            await appointmentApi.deleteSlot(slot.id);
            await fetchData();
        } catch (e) {
            console.error(e);
            window.alert('Could not delete slot. It may already be booked.');
        }
    };

    const handleCancelBookedSlot = async (slot: Slot) => {
        const appt = appointments.find(
            (a) => a.slot_id === slot.id && a.status !== 'cancelled' && a.status !== 'completed'
        );
        if (!appt) {
            window.alert('No appointment found for this slot.');
            return;
        }
        if (!window.confirm('Cancel this appointment and release the time slot?')) return;
        try {
            await appointmentApi.cancelAppointment(appt.id);
            await fetchData();
        } catch (e) {
            console.error(e);
            window.alert('Could not cancel appointment.');
        }
    };

    const hasAppointmentEnded = (appt: Appointment) => {
        const scheduledAtRaw = appt.scheduled_at || appt.scheduled_time;
        if (!scheduledAtRaw) return false;

        const startTime = new Date(scheduledAtRaw);
        if (Number.isNaN(startTime.getTime())) return false;

        const durationMinutes = appt.duration_minutes ?? 30;
        const endTime = new Date(startTime.getTime() + durationMinutes * 60 * 1000);
        return endTime.getTime() <= Date.now();
    };

    const filteredDoctors = doctors.filter(
        (d) =>
            d.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            d.specialization.toLowerCase().includes(searchQuery.toLowerCase()) ||
            d.hospital.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const filteredAppointments = appointments.filter((appt) => {
        const query = searchQuery.toLowerCase();
        if (isDoctor) {
            const patientName = getPatientDisplayName(appt).toLowerCase();
            return (
                patientName.includes(query) ||
                appt.status.toLowerCase().includes(query) ||
                (appt.payment_status && appt.payment_status.toLowerCase().includes(query))
            );
        }
        const docName = getDoctorName(appt.doctor_id).toLowerCase();
        const specialty = getDoctorSpecialty(appt.doctor_id).toLowerCase();
        return docName.includes(query) || specialty.includes(query) || appt.status.toLowerCase().includes(query);
    });

    const filteredMySlots = mySlots.filter((slot) => {
        const q = searchQuery.toLowerCase();
        const hosp = (slot.hospital ?? '').toLowerCase();
        const startLabel = new Date(slot.start_time).toLocaleString().toLowerCase();
        return hosp.includes(q) || startLabel.includes(q) || (slot.is_booked ? 'booked' : 'available').includes(q);
    });

    const statusColor: Record<string, string> = {
        confirmed: 'bg-green-50 text-green-600',
        pending: 'bg-amber-50 text-amber-600',
        cancelled: 'bg-red-50 text-red-500',
        completed: 'bg-slate-100 text-slate-600',
    };

    const initialsFromName = (name: string) =>
        name
            .split(/\s+/)
            .filter(Boolean)
            .map((w) => w[0])
            .join('')
            .slice(0, 2)
            .toUpperCase() || '?';

    return (
        <div className="min-h-screen bg-[#f6f8fa] flex font-sans">
            {/* ─── Sidebar ─── */}
            <aside className="w-60 bg-white border-r border-gray-100 hidden lg:flex flex-col sticky top-0 h-screen">
                <div className="px-6 pt-6 pb-5">
                    <Link to="/dashboard" className="flex items-center gap-2.5">
                        <div className="w-8 h-8 bg-brand rounded-lg flex items-center justify-center">
                            <Activity className="h-4 w-4 text-white" />
                        </div>
                        <span className="text-lg font-medium text-gray-900 tracking-tight">MediPulse SriLanka</span>
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
                    {/* Appointments Dropdown */}
                    <div className="space-y-1">
                        <button
                            onClick={() => setIsApptMenuOpen(!isApptMenuOpen)}
                            className={`flex items-center justify-between w-full px-3 py-2.5 rounded-xl text-sm font-semibold transition-all ${
                                location.pathname === '/appointments' 
                                ? 'bg-brand/10 text-brand border border-brand/10' 
                                : 'text-gray-500 hover:bg-gray-50'
                            }`}
                        >
                            <div className="flex items-center gap-3">
                                <Calendar className="h-[18px] w-[18px]" /> Appointments
                            </div>
                            <ChevronDown className={`h-4 w-4 transition-transform duration-200 ${isApptMenuOpen ? 'rotate-180' : ''}`} />
                        </button>
                        
                        {isApptMenuOpen && (
                            <div className="ml-9 flex flex-col gap-1 mt-1 animate-in fade-in slide-in-from-top-1 duration-200">
                                {!isDoctor && (
                                <button
                                    onClick={() => setActiveTab('doctors')}
                                    className={`text-left px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                                        activeTab === 'doctors' 
                                        ? 'text-brand font-bold bg-brand/5' 
                                        : 'text-gray-400 hover:text-gray-600 hover:bg-gray-50'
                                    }`}
                                >
                                    Book an Appointment
                                </button>
                                )}
                                <button
                                    onClick={() => setActiveTab('appointments')}
                                    className={`text-left px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                                        activeTab === 'appointments' 
                                        ? 'text-brand font-bold bg-brand/5' 
                                        : 'text-gray-400 hover:text-gray-600 hover:bg-gray-50'
                                    }`}
                                >
                                    {isDoctor ? 'Consultations' : 'See My Appointments'}
                                </button>
                                {isDoctor && (
                                    <button
                                        onClick={() => setActiveTab('slots')}
                                        className={`text-left px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                                            activeTab === 'slots'
                                                ? 'text-brand font-bold bg-brand/5'
                                                : 'text-gray-400 hover:text-gray-600 hover:bg-gray-50'
                                        }`}
                                    >
                                        Availability
                                    </button>
                                )}
                            </div>
                        )}
                    </div>
                    <Link
                        to="/profile"
                        className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
                    >
                        <User className="h-[18px] w-[18px]" /> Profile
                    </Link>
                    {!isDoctor && (
                    <Link
                        to="/payments"
                        className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
                    >
                        <CreditCard className="h-[18px] w-[18px]" /> Payments
                    </Link>
                    )}
                    <Link
                        to="/records"
                        className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
                    >
                        <ClipboardList className="h-[18px] w-[18px]" /> Records
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

            {/* ─── Main Content ─── */}
            <div className="flex-1 flex flex-col min-h-screen">
                {/* Top bar */}
                <header className="h-14 bg-white border-b border-gray-100 flex items-center justify-between px-8 sticky top-0 z-10">
                    <div className="flex items-center gap-8">
                        <h2 className="text-[15px] font-bold text-gray-900">
                            {activeTab === 'doctors'
                                ? 'Book Appointment'
                                : activeTab === 'slots'
                                  ? 'Availability'
                                  : isDoctor
                                    ? 'Consultations'
                                    : 'My Appointments'}
                        </h2>
                    </div>
                    <div className="flex items-center gap-4">
                        <div className="relative">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                            <input
                                type="text"
                                placeholder={`Search ${
                                    activeTab === 'doctors'
                                        ? 'doctors...'
                                        : activeTab === 'slots'
                                          ? 'slots...'
                                          : 'appointments...'
                                }`}
                                className="w-56 pl-9 pr-4 py-2 bg-gray-50 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20 focus:border-brand/40 transition-all"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                            />
                        </div>
                        {isDoctor && activeTab === 'slots' && (
                            <button
                                type="button"
                                onClick={() => {
                                    setSlotModalMode('create');
                                    setSlotEditing(null);
                                    setSlotModalOpen(true);
                                }}
                                className="flex items-center gap-2 px-4 py-2 bg-brand text-white rounded-xl text-sm font-medium hover:bg-brand-dark transition-colors shadow-sm"
                            >
                                <Plus className="h-4 w-4" /> New slot
                            </button>
                        )}
                        {!isDoctor && (
                        <button
                            onClick={() => {
                                if (activeTab !== 'doctors') setActiveTab('doctors');
                                if (doctors.length > 0) {
                                    setSelectedDoctor(doctors[0]);
                                    setIsModalOpen(true);
                                }
                            }}
                            className="flex items-center gap-2 px-4 py-2 bg-brand text-white rounded-xl text-sm font-medium hover:bg-brand-dark transition-colors shadow-sm"
                        >
                            <Plus className="h-4 w-4" /> Book New
                        </button>
                        )}
                    </div>
                </header>

                <main className="flex-1 p-8 overflow-y-auto">
                    {isLoading ? (
                        <div className="flex items-center justify-center py-24">
                            <div className="animate-spin rounded-full h-8 w-8 border-[3px] border-brand border-t-transparent" />
                        </div>
                    ) : activeTab === 'doctors' ? (
                        /* ─── Doctors Grid ─── */
                        <div className="space-y-6">
                            <div className="flex items-center justify-between">
                                <div>
                                    <h3 className="text-base font-semibold text-gray-800">Available Doctors</h3>
                                    <p className="text-sm text-gray-400 mt-1">{filteredDoctors.length} physician{filteredDoctors.length !== 1 ? 's' : ''} found</p>
                                </div>
                            </div>

                            {filteredDoctors.length === 0 ? (
                                <div className="bg-white rounded-2xl border border-gray-100 p-14 text-center">
                                    <User className="h-10 w-10 text-gray-300 mx-auto mb-3" />
                                    <p className="text-sm font-medium text-gray-500">No doctors found</p>
                                    <p className="text-sm text-gray-400 mt-1">Try adjusting your search query</p>
                                </div>
                            ) : (
                                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
                                    {filteredDoctors.map((doc) => {
                                        const docSlots = doctorSlots[String(doc.id)] || [];
                                        const count = docSlots.filter((s: Slot) => !s.is_booked && s.status !== 'booked').length;
                                        const initials = doc.name
                                            .split(' ')
                                            .map((w: string) => w[0])
                                            .join('')
                                            .slice(0, 2)
                                            .toUpperCase();

                                        return (
                                            <div
                                                key={doc.id}
                                                className="bg-white rounded-2xl border border-gray-100 p-5 hover:shadow-lg hover:border-brand/20 transition-all duration-300 group"
                                            >
                                                <div className="flex items-start gap-4">
                                                    <div className="w-12 h-12 bg-brand/10 rounded-xl flex items-center justify-center text-brand text-sm font-bold shrink-0">
                                                        {initials}
                                                    </div>
                                                    <div className="flex-1 min-w-0">
                                                        <h4 className="text-[15px] font-semibold text-gray-900 truncate group-hover:text-brand transition-colors">
                                                            {doc.name}
                                                        </h4>
                                                        <p className="text-sm text-gray-500 mt-0.5">{doc.specialization}</p>
                                                    </div>
                                                </div>

                                                <div className="mt-4 flex flex-wrap gap-2">
                                                    <span className="inline-flex items-center gap-1.5 text-xs text-gray-500 bg-gray-50 px-2.5 py-1 rounded-lg">
                                                        <Building2 className="h-3.5 w-3.5" /> {doc.hospital}
                                                    </span>
                                                    <span className="inline-flex items-center gap-1.5 text-xs text-gray-500 bg-gray-50 px-2.5 py-1 rounded-lg">
                                                        <Briefcase className="h-3.5 w-3.5" /> {doc.experience} yrs experience
                                                    </span>
                                                </div>

                                                <div className="mt-4 flex items-center gap-2">
                                                    <span className="inline-flex items-center gap-1.5 text-xs text-brand bg-brand/10 px-2.5 py-1 rounded-lg font-medium">
                                                        <Clock className="h-3.5 w-3.5" /> {count} slot{count !== 1 ? 's' : ''} available
                                                    </span>
                                                </div>

                                                <button
                                                    onClick={() => openBooking(doc)}
                                                    className="mt-6 w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-gray-900 text-white rounded-xl text-sm font-medium hover:bg-brand transition-colors"
                                                >
                                                    Book Appointment <ChevronRight className="h-4 w-4" />
                                                </button>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                        </div>
                    ) : activeTab === 'slots' ? (
                        <div className="space-y-6">
                            <div>
                                <h3 className="text-base font-semibold text-gray-800">Your availability</h3>
                                <p className="text-sm text-gray-400 mt-1">
                                    Create slots at your profile hospital. Patients book into open times; booked slots
                                    can be released by cancelling the consultation.
                                </p>
                            </div>

                            {!profileHospitalStr && (
                                <div className="rounded-2xl border border-amber-200 bg-amber-50 px-5 py-4 text-sm text-amber-900">
                                    Add your <strong>hospital</strong> on{' '}
                                    <Link to="/profile" className="font-semibold underline underline-offset-2">
                                        Profile
                                    </Link>{' '}
                                    before creating slots.
                                </div>
                            )}

                            {filteredMySlots.length === 0 ? (
                                <div className="bg-white rounded-2xl border border-gray-100 p-14 text-center">
                                    <Clock className="h-10 w-10 text-gray-300 mx-auto mb-3" />
                                    <p className="text-sm font-medium text-gray-500">No slots yet</p>
                                    <p className="text-sm text-gray-400 mt-1">
                                        Add your first availability window so patients can book.
                                    </p>
                                    <button
                                        type="button"
                                        onClick={() => {
                                            setSlotModalMode('create');
                                            setSlotEditing(null);
                                            setSlotModalOpen(true);
                                        }}
                                        disabled={!profileHospitalStr}
                                        className="mt-5 px-5 py-2.5 bg-brand text-white rounded-xl text-sm font-medium hover:bg-brand-dark transition-colors shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        Create slot
                                    </button>
                                </div>
                            ) : (
                                <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
                                    <div className="overflow-x-auto">
                                        <table className="w-full text-sm">
                                            <thead>
                                                <tr className="border-b border-gray-100 bg-gray-50/80 text-left text-[11px] font-bold uppercase tracking-wider text-gray-400">
                                                    <th className="px-5 py-3">Hospital</th>
                                                    <th className="px-5 py-3">Start</th>
                                                    <th className="px-5 py-3">End</th>
                                                    <th className="px-5 py-3">Status</th>
                                                    <th className="px-5 py-3 text-right">Actions</th>
                                                </tr>
                                            </thead>
                                            <tbody className="divide-y divide-gray-100">
                                                {filteredMySlots.map((slot) => {
                                                    const booked = slot.is_booked;
                                                    return (
                                                        <tr key={slot.id} className="hover:bg-gray-50/60">
                                                            <td className="px-5 py-3.5 text-gray-800">
                                                                {slot.hospital || '—'}
                                                            </td>
                                                            <td className="px-5 py-3.5 text-gray-600 whitespace-nowrap">
                                                                {new Date(slot.start_time).toLocaleString([], {
                                                                    dateStyle: 'medium',
                                                                    timeStyle: 'short',
                                                                })}
                                                            </td>
                                                            <td className="px-5 py-3.5 text-gray-600 whitespace-nowrap">
                                                                {new Date(slot.end_time).toLocaleString([], {
                                                                    dateStyle: 'medium',
                                                                    timeStyle: 'short',
                                                                })}
                                                            </td>
                                                            <td className="px-5 py-3.5">
                                                                <span
                                                                    className={`inline-flex px-2.5 py-1 rounded-lg text-[10px] font-bold uppercase tracking-wider ${
                                                                        booked
                                                                            ? 'bg-amber-50 text-amber-700'
                                                                            : 'bg-emerald-50 text-emerald-700'
                                                                    }`}
                                                                >
                                                                    {booked ? 'Booked' : 'Available'}
                                                                </span>
                                                            </td>
                                                            <td className="px-5 py-3.5 text-right">
                                                                <div className="inline-flex items-center gap-2 justify-end">
                                                                    {!booked ? (
                                                                        <>
                                                                            <button
                                                                                type="button"
                                                                                onClick={() => {
                                                                                    setSlotModalMode('edit');
                                                                                    setSlotEditing(slot);
                                                                                    setSlotModalOpen(true);
                                                                                }}
                                                                                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-gray-200 text-xs font-semibold text-gray-700 hover:bg-gray-50"
                                                                            >
                                                                                <Pencil className="h-3.5 w-3.5" />
                                                                                Edit
                                                                            </button>
                                                                            <button
                                                                                type="button"
                                                                                onClick={() => handleDeleteSlot(slot)}
                                                                                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-red-100 text-xs font-semibold text-red-600 hover:bg-red-50"
                                                                            >
                                                                                <Trash2 className="h-3.5 w-3.5" />
                                                                                Remove
                                                                            </button>
                                                                        </>
                                                                    ) : (
                                                                        <button
                                                                            type="button"
                                                                            onClick={() =>
                                                                                handleCancelBookedSlot(slot)
                                                                            }
                                                                            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-amber-200 text-xs font-semibold text-amber-800 hover:bg-amber-50"
                                                                        >
                                                                            Cancel booking
                                                                        </button>
                                                                    )}
                                                                </div>
                                                            </td>
                                                        </tr>
                                                    );
                                                })}
                                            </tbody>
                                        </table>
                                    </div>
                                </div>
                            )}
                        </div>
                    ) : (
                        /* ─── My Appointments Table ─── */
                        <div className="space-y-6">
                            <div>
                                <h3 className="text-base font-semibold text-gray-800">
                                    {isDoctor ? 'Consultations' : 'My Appointments'}
                                </h3>
                                <p className="text-sm text-gray-400 mt-1">{appointments.length} total</p>
                            </div>

                            {appointments.length === 0 ? (
                                <div className="bg-white rounded-2xl border border-gray-100 p-14 text-center">
                                    <Calendar className="h-10 w-10 text-gray-300 mx-auto mb-3" />
                                    <p className="text-sm font-medium text-gray-500">No appointments yet</p>
                                    <p className="text-sm text-gray-400 mt-1">
                                        {isDoctor ? 'Scheduled consultations will appear here.' : 'Book your first appointment with a doctor'}
                                    </p>
                                    {!isDoctor && (
                                    <button
                                        onClick={() => setActiveTab('doctors')}
                                        className="mt-5 px-5 py-2.5 bg-brand text-white rounded-xl text-sm font-medium hover:bg-brand-dark transition-colors shadow-sm"
                                    >
                                        Browse Doctors
                                    </button>
                                    )}
                                </div>
                            ) : filteredAppointments.length === 0 ? (
                                <div className="bg-white rounded-2xl border border-gray-100 p-14 text-center">
                                    <Search className="h-10 w-10 text-gray-300 mx-auto mb-3" />
                                    <p className="text-sm font-medium text-gray-500">No matching appointments</p>
                                    <p className="text-sm text-gray-400 mt-1">Try adjusting your search query</p>
                                </div>
                            ) : (
                                <div className="flex flex-col gap-3">
                                    {filteredAppointments.map((appt) => (
                                        <div key={appt.id} className="bg-white rounded-2xl border border-gray-100 p-4 hover:shadow-md transition-all group flex items-center gap-6">
                                            <div className="w-14 h-14 bg-gray-50 border border-gray-100 rounded-xl flex items-center justify-center text-brand font-bold shrink-0 shadow-sm group-hover:border-brand/20 transition-colors">
                                                {isDoctor
                                                    ? initialsFromName(getPatientDisplayName(appt))
                                                    : getDoctorName(appt.doctor_id)
                                                          .split(' ')
                                                          .map((n: string) => n[0])
                                                          .join('')
                                                          .slice(0, 2)
                                                          .toUpperCase()}
                                            </div>

                                            <div className="w-56 shrink-0">
                                                <h4 className="text-[14px] font-bold text-gray-900 truncate">
                                                    {isDoctor ? getPatientDisplayName(appt) : getDoctorName(appt.doctor_id)}
                                                </h4>
                                                <p className="text-xs text-brand font-medium mt-0.5 truncate">
                                                    {isDoctor
                                                        ? `${appt.consultation_mode === 'jitsi' || appt.consultation_mode === 'video' ? 'Video' : 'Physical'} · ${appt.status}`
                                                        : getDoctorSpecialty(appt.doctor_id)}
                                                </p>
                                            </div>

                                            {/* Date & Time */}
                                            <div className="flex-1 flex items-center gap-8">
                                                <div className="flex items-center gap-2.5 min-w-[120px]">
                                                    <Calendar className="h-4 w-4 text-gray-400" />
                                                    <span className="text-sm text-gray-600 font-medium">
                                                        {new Date(appt.scheduled_at || appt.scheduled_time || '').toLocaleDateString([], { month: 'short', day: 'numeric', year: 'numeric' })}
                                                    </span>
                                                </div>
                                                <div className="flex items-center gap-2.5">
                                                    <Clock className="h-4 w-4 text-gray-400" />
                                                    <span className="text-sm text-gray-600 font-medium">
                                                        {new Date(appt.scheduled_at || appt.scheduled_time || '').toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                                    </span>
                                                </div>
                                            </div>

                                            {/* Mode & Status */}
                                            <div className="w-56 flex items-center gap-4 justify-center">
                                                <span className="inline-flex items-center gap-1.5 text-[11px] font-medium text-gray-500 bg-gray-50 px-2.5 py-1 rounded-lg border border-gray-100">
                                                    {appt.consultation_mode === 'jitsi' || appt.consultation_mode === 'video' ? (
                                                        <><Video className="h-3.5 w-3.5 text-brand" /> Video</>
                                                    ) : (
                                                        <><MapPin className="h-3.5 w-3.5 text-blue-500" /> Physical</>
                                                    )}
                                                </span>
                                                <span 
                                                    className={`px-2.5 py-1 rounded-lg text-[10px] font-bold uppercase tracking-wider ${
                                                        statusColor[appt.status] || 'bg-gray-100 text-gray-500'
                                                    }`}
                                                >
                                                    {appt.status}
                                                </span>
                                            </div>

                                            {/* Payment & CTA */}
                                            <div className="w-64 flex items-center justify-end gap-6">
                                                 <div className="text-right">
                                                    <span className="text-[10px] text-gray-400 uppercase font-bold tracking-tight block">Payment</span>
                                                    <span 
                                                        className={`text-xs font-bold mt-0.5 ${
                                                            appt.payment_status === 'paid' ? 'text-green-600' : 'text-amber-600'
                                                        }`}
                                                    >
                                                        {appt.payment_status || 'Pending'}
                                                    </span>
                                                </div>

                                                {(appt.consultation_mode === 'jitsi' || appt.consultation_mode === 'video') && appt.join_url && appt.status !== 'cancelled' && (
                                                    <button 
                                                        type="button"
                                                        onClick={() =>
                                                            navigate(
                                                                `/telemedicine?join_url=${encodeURIComponent(appt.join_url || '')}&peer=${encodeURIComponent(isDoctor ? getPatientDisplayName(appt) : getDoctorName(appt.doctor_id))}&title=${encodeURIComponent('Telemedicine Session')}`
                                                            )
                                                        }
                                                        disabled={hasAppointmentEnded(appt)}
                                                        title={hasAppointmentEnded(appt) ? 'This meeting has ended' : 'Join meeting'}
                                                        className={`flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs font-bold shadow-sm shadow-black/5 transition-all ${
                                                            hasAppointmentEnded(appt)
                                                                ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
                                                                : 'bg-gray-900 text-white hover:bg-brand active:scale-95'
                                                        }`}
                                                    >
                                                        <Video className="h-3.5 w-3.5" /> 
                                                        {hasAppointmentEnded(appt) ? 'Ended' : 'Join'}
                                                    </button>
                                                )}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    )}
                </main>
            </div>

            {/* Booking Modal */}
            {selectedDoctor && (
                <BookingModal
                    isOpen={isModalOpen}
                    onClose={() => {
                        setIsModalOpen(false);
                        setSelectedDoctor(null);
                    }}
                    doctor={{
                        id: String(selectedDoctor.id),
                        name: selectedDoctor.name,
                        specialty: selectedDoctor.specialization,
                        hospital: selectedDoctor.hospital,
                        experience: selectedDoctor.experience,
                        channeling_fee: selectedDoctor.channeling_fee,
                    }}
                    slots={doctorSlots[String(selectedDoctor.id)] || []}
                    consultationFee={Number(selectedDoctor.channeling_fee || 0)}
                    hospitalFee={HOSPITAL_FEE}
                    onBook={handleBook}
                />
            )}
            {/* Fallback modal when no doctor selected */}
            {!selectedDoctor && isModalOpen && doctors.length > 0 && (
                <BookingModal
                    isOpen={isModalOpen}
                    onClose={() => setIsModalOpen(false)}
                    doctor={{
                        id: String(doctors[0].id),
                        name: doctors[0].name,
                        specialty: doctors[0].specialization,
                        hospital: doctors[0].hospital,
                        experience: doctors[0].experience,
                        channeling_fee: doctors[0].channeling_fee,
                    }}
                    slots={doctorSlots[String(doctors[0].id)] || []}
                    consultationFee={Number(doctors[0].channeling_fee || 0)}
                    hospitalFee={HOSPITAL_FEE}
                    onBook={handleBook}
                />
            )}

            {isDoctor && (
                <DoctorSlotFormModal
                    isOpen={slotModalOpen}
                    onClose={() => {
                        setSlotModalOpen(false);
                        setSlotEditing(null);
                    }}
                    mode={slotModalMode}
                    slot={slotEditing}
                    profileHospital={profileHospitalStr}
                    onSubmit={handleSaveSlot}
                />
            )}

            {/* Redirecting Overlay */}
            {isRedirecting && (
                <div className="fixed inset-0 z-[100] flex items-center justify-center bg-white/80 backdrop-blur-md animate-in fade-in duration-300">
                    <div className="text-center">
                        <div className="w-16 h-16 border-4 border-brand border-t-transparent rounded-full animate-spin mx-auto mb-4" />
                        <h3 className="text-xl font-bold text-gray-900">Redirecting to Payment</h3>
                        <p className="text-gray-500 mt-2">Please do not close your browser...</p>
                    </div>
                </div>
            )}
        </div>
    );
}
