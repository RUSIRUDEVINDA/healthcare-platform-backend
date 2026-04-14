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
} from 'lucide-react';
import { Link } from 'react-router-dom';
import type { Slot, Appointment, BookAppointmentRequest } from '../api/appointments';
import { appointmentApi } from '../api/appointments';
import { doctorApi, type Doctor } from '../api/doctors';
import { patientApi } from '../api/patient';
import { paymentApi } from '../api/payment';
import { submitPayHereForm } from '../utils/payment';
import BookingModal from '../components/appointments/BookingModal';

type TabKey = 'doctors' | 'appointments';

export default function Appointments() {
    const [appointments, setAppointments] = useState<Appointment[]>([]);
    const [doctors, setDoctors] = useState<Doctor[]>([]);
    const [doctorSlots, setDoctorSlots] = useState<Record<string, Slot[]>>({});
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedDoctor, setSelectedDoctor] = useState<Doctor | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [activeTab, setActiveTab] = useState<TabKey>('doctors');
    const [isApptMenuOpen, setIsApptMenuOpen] = useState(true);
    const [isRedirecting, setIsRedirecting] = useState(false);

    useEffect(() => {
        fetchData();
    }, []);

    const fetchData = async () => {
        setIsLoading(true);
        try {
            // Fetch doctors from doctor-service (public endpoint)
            let docs: Doctor[] = [];
            try {
                const docResult = await doctorApi.listDoctors();
                docs = Array.isArray(docResult) ? docResult : [];
            } catch (err) {
                console.error('Failed to fetch doctors:', err);
            }
            setDoctors(docs);

            // Fetch appointments from appointment-service
            let appts: Appointment[] = [];
            try {
                const apptData = await appointmentApi.listAppointments();
                appts = Array.isArray(apptData) ? apptData : (apptData as any)?.appointments ?? [];
            } catch (err) {
                console.error('Failed to fetch appointments:', err);
            }
            setAppointments(appts);

            // Fetch available slots for each doctor
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
        } catch (error) {
            console.error('Booking failed:', error);
            throw error;
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

    const filteredDoctors = doctors.filter(
        (d) =>
            d.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            d.specialization.toLowerCase().includes(searchQuery.toLowerCase()) ||
            d.hospital.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const filteredAppointments = appointments.filter((appt) => {
        const docName = getDoctorName(appt.doctor_id).toLowerCase();
        const specialty = getDoctorSpecialty(appt.doctor_id).toLowerCase();
        const query = searchQuery.toLowerCase();
        return docName.includes(query) || specialty.includes(query) || appt.status.toLowerCase().includes(query);
    });

    const statusColor: Record<string, string> = {
        confirmed: 'bg-green-50 text-green-600',
        pending: 'bg-amber-50 text-amber-600',
        cancelled: 'bg-red-50 text-red-500',
        completed: 'bg-slate-100 text-slate-600',
    };

    return (
        <div className="min-h-screen bg-[#f6f8fa] flex font-sans">
            {/* ─── Sidebar ─── */}
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
                                <button
                                    onClick={() => setActiveTab('appointments')}
                                    className={`text-left px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                                        activeTab === 'appointments' 
                                        ? 'text-brand font-bold bg-brand/5' 
                                        : 'text-gray-400 hover:text-gray-600 hover:bg-gray-50'
                                    }`}
                                >
                                    See My Appointments
                                </button>
                            </div>
                        )}
                    </div>
                    <Link
                        to="/profile"
                        className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
                    >
                        <User className="h-[18px] w-[18px]" /> Profile
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

            {/* ─── Main Content ─── */}
            <div className="flex-1 flex flex-col min-h-screen">
                {/* Top bar */}
                <header className="h-14 bg-white border-b border-gray-100 flex items-center justify-between px-8 sticky top-0 z-10">
                    <div className="flex items-center gap-8">
                        <h2 className="text-[15px] font-bold text-gray-900">
                            {activeTab === 'doctors' ? 'Book Appointment' : 'My Appointments'}
                        </h2>
                    </div>
                    <div className="flex items-center gap-4">
                        <div className="relative">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                            <input
                                type="text"
                                placeholder={`Search ${activeTab === 'doctors' ? 'doctors...' : 'appointments...'}`}
                                className="w-56 pl-9 pr-4 py-2 bg-gray-50 border border-gray-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-brand/20 focus:border-brand/40 transition-all"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                            />
                        </div>
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
                    ) : (
                        /* ─── My Appointments Table ─── */
                        <div className="space-y-6">
                            <div>
                                <h3 className="text-base font-semibold text-gray-800">My Appointments</h3>
                                <p className="text-sm text-gray-400 mt-1">{appointments.length} total</p>
                            </div>

                            {appointments.length === 0 ? (
                                <div className="bg-white rounded-2xl border border-gray-100 p-14 text-center">
                                    <Calendar className="h-10 w-10 text-gray-300 mx-auto mb-3" />
                                    <p className="text-sm font-medium text-gray-500">No appointments yet</p>
                                    <p className="text-sm text-gray-400 mt-1">Book your first appointment with a doctor</p>
                                    <button
                                        onClick={() => setActiveTab('doctors')}
                                        className="mt-5 px-5 py-2.5 bg-brand text-white rounded-xl text-sm font-medium hover:bg-brand-dark transition-colors shadow-sm"
                                    >
                                        Browse Doctors
                                    </button>
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
                                            {/* Dr Avatar */}
                                            <div className="w-14 h-14 bg-gray-50 border border-gray-100 rounded-xl flex items-center justify-center text-brand font-bold shrink-0 shadow-sm group-hover:border-brand/20 transition-colors">
                                                {getDoctorName(appt.doctor_id).split(' ').map((n: string) => n[0]).join('').slice(0, 2).toUpperCase()}
                                            </div>

                                            {/* Dr Info */}
                                            <div className="w-56 shrink-0">
                                                <h4 className="text-[14px] font-bold text-gray-900 truncate">{getDoctorName(appt.doctor_id)}</h4>
                                                <p className="text-xs text-brand font-medium mt-0.5 truncate">{getDoctorSpecialty(appt.doctor_id)}</p>
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
                                                    <a 
                                                        href={appt.join_url} 
                                                        target="_blank" 
                                                        rel="noopener noreferrer"
                                                        className="flex items-center gap-2 px-5 py-2.5 bg-gray-900 text-white rounded-xl text-xs font-bold hover:bg-brand transition-all active:scale-95 shadow-sm shadow-black/5"
                                                    >
                                                        <Video className="h-3.5 w-3.5" /> Join
                                                    </a>
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
                    }}
                    slots={doctorSlots[String(selectedDoctor.id)] || []}
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
                    }}
                    slots={doctorSlots[String(doctors[0].id)] || []}
                    onBook={handleBook}
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
