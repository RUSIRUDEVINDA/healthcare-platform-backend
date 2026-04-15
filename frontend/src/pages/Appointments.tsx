import { useState, useEffect } from 'react';
import {
    Search,
    Calendar,
    Clock,
    Video,
    MapPin,
    Plus,
    User,
    ChevronRight,
    Building2,
    Briefcase,
    Pencil,
    Trash2,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import type { Slot, Appointment, BookAppointmentRequest } from '../api/appointments';
import { appointmentApi } from '../api/appointments';
import { doctorApi, type Doctor } from '../api/doctors';
import { doctorApi as doctorProfileApi, type DoctorProfile } from '../api/doctor';
import { patientApi } from '../api/patient';
import { paymentApi } from '../api/payment';
import { submitPayHereForm } from '../utils/payment';
import BookingModal from '../components/appointments/BookingModal';
import DoctorSlotFormModal from '../components/appointments/DoctorSlotFormModal';

type TabKey = 'doctors' | 'appointments' | 'slots';

function getPatientDisplayName(appt: Appointment): string {
    const f = appt.patient_first_name?.trim() || '';
    const l = appt.patient_last_name?.trim() || '';
    if (f || l) {
        return `${f} ${l}`.trim();
    }
    const id = appt.patient_id || '';
    return id.length > 10 ? `Patient ${id.slice(0, 8)}…` : 'Patient';
}

function getPatientInitials(appt: Appointment): string {
    const f = appt.patient_first_name?.trim();
    const l = appt.patient_last_name?.trim();
    if (f && l) {
        return `${f[0] ?? ''}${l[0] ?? ''}`.toUpperCase();
    }
    if (f) {
        return f.slice(0, 2).toUpperCase();
    }
    return 'PT';
}

function canJoinJitsiVisit(appt: Appointment): boolean {
    const mode = appt.consultation_mode;
    const video = mode === 'jitsi' || mode === 'video';
    if (!video || !(appt.join_url && appt.join_url.trim())) {
        return false;
    }
    if (appt.status === 'cancelled' || appt.status === 'completed') {
        return false;
    }
    return true;
}

function initialAppointmentsTab(): TabKey {
    try {
        const raw = localStorage.getItem('user');
        if (!raw) return 'doctors';
        return JSON.parse(raw)?.role === 'doctor' ? 'appointments' : 'doctors';
    } catch {
        return 'doctors';
    }
}

export default function Appointments() {
    const [role, setRole] = useState<string | null>(() => {
        try {
            const raw = localStorage.getItem('user');
            return raw ? JSON.parse(raw)?.role ?? null : null;
        } catch {
            return null;
        }
    });
    const [appointments, setAppointments] = useState<Appointment[]>([]);
    const [doctors, setDoctors] = useState<Doctor[]>([]);
    const [doctorSlots, setDoctorSlots] = useState<Record<string, Slot[]>>({});
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedDoctor, setSelectedDoctor] = useState<Doctor | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [activeTab, setActiveTab] = useState<TabKey>(initialAppointmentsTab);
    const [isRedirecting, setIsRedirecting] = useState(false);
    const [mySlots, setMySlots] = useState<Slot[]>([]);
    const [doctorProfile, setDoctorProfile] = useState<DoctorProfile | null>(null);
    const [slotModalOpen, setSlotModalOpen] = useState(false);
    const [slotModalMode, setSlotModalMode] = useState<'create' | 'edit'>('create');
    const [editingSlot, setEditingSlot] = useState<Slot | null>(null);
    const isDoctor = role === 'doctor';

    useEffect(() => {
        fetchData();
    }, []);

    useEffect(() => {
        if (role === 'doctor' && activeTab === 'doctors') {
            setActiveTab('appointments');
        }
        if (role !== 'doctor' && activeTab === 'slots') {
            setActiveTab('doctors');
        }
    }, [role, activeTab]);

    const fetchData = async () => {
        setIsLoading(true);
        try {
            const userRaw = localStorage.getItem('user');
            const userRole = userRaw ? JSON.parse(userRaw)?.role : null;
            setRole(userRole ?? null);

            // Fetch doctors from doctor-service (patients only; doctors manage visits elsewhere)
            let docs: Doctor[] = [];
            if (userRole !== 'doctor') {
                try {
                    const docResult = await doctorApi.listDoctors();
                    docs = Array.isArray(docResult) ? docResult : [];
                } catch (err) {
                    console.error('Failed to fetch doctors:', err);
                }
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

            if (userRole === 'doctor') {
                try {
                    const prof = await doctorProfileApi.getProfile();
                    setDoctorProfile(prof);
                } catch (err) {
                    console.error('Failed to fetch doctor profile:', err);
                    setDoctorProfile(null);
                }
                try {
                    const s = await appointmentApi.listMySlots();
                    setMySlots(Array.isArray(s) ? s : []);
                } catch (err) {
                    console.error('Failed to fetch slots:', err);
                    setMySlots([]);
                }
            } else {
                setDoctorProfile(null);
                setMySlots([]);
            }

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
        const query = searchQuery.toLowerCase();
        if (isDoctor) {
            const patientLabel = getPatientDisplayName(appt).toLowerCase();
            return (
                patientLabel.includes(query) ||
                appt.patient_id.toLowerCase().includes(query) ||
                appt.status.toLowerCase().includes(query)
            );
        }
        const docName = getDoctorName(appt.doctor_id).toLowerCase();
        const specialty = getDoctorSpecialty(appt.doctor_id).toLowerCase();
        return docName.includes(query) || specialty.includes(query) || appt.status.toLowerCase().includes(query);
    });

    const sortedMySlots = [...mySlots].sort(
        (a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime()
    );
    const filteredMySlots = sortedMySlots.filter((slot) => {
        const q = searchQuery.toLowerCase();
        if (!q) return true;
        const hosp = (slot.hospital || '').toLowerCase();
        const startLabel = new Date(slot.start_time).toLocaleString().toLowerCase();
        return hosp.includes(q) || startLabel.includes(q) || slot.id.toLowerCase().includes(q);
    });

    const statusColor: Record<string, string> = {
        confirmed: 'bg-green-50 text-green-600',
        pending: 'bg-amber-50 text-amber-600',
        cancelled: 'bg-red-50 text-red-500',
        completed: 'bg-slate-100 text-slate-600',
    };

    const tabBtn = (active: boolean) =>
        `rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors ${
            active ? 'bg-brand/10 text-brand ring-1 ring-brand/20' : 'text-gray-600 hover:bg-gray-100'
        }`;

    return (
        <div className="flex min-h-screen flex-1 flex-col bg-[#f6f8fa] font-sans">
            <div className="flex min-h-screen flex-1 flex-col">
                <header className="sticky top-0 z-10 border-b border-gray-100 bg-white px-4 sm:px-8">
                    <div className="flex flex-col gap-3 py-3 sm:flex-row sm:items-center sm:justify-between sm:py-3">
                        <div className="min-w-0 flex-1">
                            <h2 className="text-[15px] font-bold text-gray-900">
                                {activeTab === 'doctors'
                                    ? 'Book appointment'
                                    : activeTab === 'slots'
                                      ? 'My availability'
                                      : 'My appointments'}
                            </h2>
                            <div className="mt-2 flex flex-wrap gap-2">
                                {!isDoctor && (
                                    <button type="button" onClick={() => setActiveTab('doctors')} className={tabBtn(activeTab === 'doctors')}>
                                        Book an appointment
                                    </button>
                                )}
                                <button
                                    type="button"
                                    onClick={() => setActiveTab('appointments')}
                                    className={tabBtn(activeTab === 'appointments')}
                                >
                                    My appointments
                                </button>
                                {isDoctor && (
                                    <button type="button" onClick={() => setActiveTab('slots')} className={tabBtn(activeTab === 'slots')}>
                                        Manage availability
                                    </button>
                                )}
                            </div>
                        </div>
                    <div className="flex flex-shrink-0 flex-wrap items-center gap-4">
                        <div className="relative">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                            <input
                                type="text"
                                placeholder={`Search ${activeTab === 'doctors' ? 'doctors...' : activeTab === 'slots' ? 'slots...' : 'appointments...'}`}
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
                                    setEditingSlot(null);
                                    setSlotModalOpen(true);
                                }}
                                className="flex items-center gap-2 px-4 py-2 bg-brand text-white rounded-xl text-sm font-medium hover:bg-brand-dark transition-colors shadow-sm"
                            >
                                <Plus className="h-4 w-4" /> Add slot
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
                            {!doctorProfile?.hospital?.trim() && (
                                <div className="bg-amber-50 border border-amber-100 rounded-2xl px-4 py-3 text-sm text-amber-900">
                                    Add your hospital on{' '}
                                    <Link to="/profile" className="font-semibold underline">
                                        Profile
                                    </Link>{' '}
                                    before patients can book your slots.
                                </div>
                            )}
                            <div className="flex items-center justify-between">
                                <div>
                                    <h3 className="text-base font-semibold text-gray-800">Your slots</h3>
                                    <p className="text-sm text-gray-400 mt-1">
                                        {filteredMySlots.length} slot{filteredMySlots.length !== 1 ? 's' : ''} shown · at your
                                        profile hospital
                                    </p>
                                </div>
                            </div>
                            {filteredMySlots.length === 0 ? (
                                <div className="bg-white rounded-2xl border border-gray-100 p-14 text-center">
                                    <Calendar className="h-10 w-10 text-gray-300 mx-auto mb-3" />
                                    <p className="text-sm font-medium text-gray-500">No slots yet</p>
                                    <p className="text-sm text-gray-400 mt-1">
                                        Create availability so patients can book you.
                                    </p>
                                </div>
                            ) : (
                                <div className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
                                    <table className="w-full text-sm">
                                        <thead>
                                            <tr className="bg-gray-50 text-left text-xs font-bold text-gray-500 uppercase tracking-wider border-b border-gray-100">
                                                <th className="px-5 py-3">Date</th>
                                                <th className="px-5 py-3">Start</th>
                                                <th className="px-5 py-3">End</th>
                                                <th className="px-5 py-3">Hospital</th>
                                                <th className="px-5 py-3">Status</th>
                                                <th className="px-5 py-3 text-right">Actions</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {filteredMySlots.map((slot) => {
                                                const booked = slot.is_booked || slot.status === 'booked';
                                                return (
                                                    <tr
                                                        key={slot.id}
                                                        className="border-b border-gray-50 last:border-0 hover:bg-gray-50/80"
                                                    >
                                                        <td className="px-5 py-3 text-gray-700 font-medium">
                                                            {new Date(slot.start_time).toLocaleDateString([], {
                                                                weekday: 'short',
                                                                month: 'short',
                                                                day: 'numeric',
                                                                year: 'numeric',
                                                            })}
                                                        </td>
                                                        <td className="px-5 py-3 text-gray-600">
                                                            {new Date(slot.start_time).toLocaleTimeString([], {
                                                                hour: '2-digit',
                                                                minute: '2-digit',
                                                            })}
                                                        </td>
                                                        <td className="px-5 py-3 text-gray-600">
                                                            {new Date(slot.end_time).toLocaleTimeString([], {
                                                                hour: '2-digit',
                                                                minute: '2-digit',
                                                            })}
                                                        </td>
                                                        <td className="px-5 py-3 text-gray-600 max-w-[180px] truncate">
                                                            {slot.hospital || '—'}
                                                        </td>
                                                        <td className="px-5 py-3">
                                                            <span
                                                                className={`text-[10px] font-bold uppercase px-2 py-1 rounded-lg ${
                                                                    booked
                                                                        ? 'bg-slate-100 text-slate-600'
                                                                        : 'bg-green-50 text-green-700'
                                                                }`}
                                                            >
                                                                {booked ? 'Booked' : 'Available'}
                                                            </span>
                                                        </td>
                                                        <td className="px-5 py-3 text-right">
                                                            <div className="flex items-center justify-end gap-1">
                                                                <button
                                                                    type="button"
                                                                    disabled={booked}
                                                                    onClick={() => {
                                                                        setEditingSlot(slot);
                                                                        setSlotModalMode('edit');
                                                                        setSlotModalOpen(true);
                                                                    }}
                                                                    className="p-2 rounded-lg text-gray-500 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed"
                                                                    title={booked ? 'Cannot edit a booked slot' : 'Edit'}
                                                                >
                                                                    <Pencil className="h-4 w-4" />
                                                                </button>
                                                                <button
                                                                    type="button"
                                                                    disabled={booked}
                                                                    onClick={async () => {
                                                                        if (
                                                                            !window.confirm(
                                                                                'Remove this slot? This cannot be undone.'
                                                                            )
                                                                        ) {
                                                                            return;
                                                                        }
                                                                        try {
                                                                            await appointmentApi.deleteSlot(slot.id);
                                                                            await fetchData();
                                                                        } catch (e: unknown) {
                                                                            const msg =
                                                                                e &&
                                                                                typeof e === 'object' &&
                                                                                'response' in e &&
                                                                                (e as { response?: { data?: { error?: string } } })
                                                                                    .response?.data?.error;
                                                                            alert(msg || 'Could not delete slot');
                                                                        }
                                                                    }}
                                                                    className="p-2 rounded-lg text-red-500 hover:bg-red-50 disabled:opacity-40 disabled:cursor-not-allowed"
                                                                    title={booked ? 'Cannot remove a booked slot' : 'Delete'}
                                                                >
                                                                    <Trash2 className="h-4 w-4" />
                                                                </button>
                                                            </div>
                                                        </td>
                                                    </tr>
                                                );
                                            })}
                                        </tbody>
                                    </table>
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
                                    <p className="text-sm text-gray-400 mt-1">
                                        {isDoctor
                                            ? 'No scheduled appointments yet.'
                                            : 'Book your first appointment with a doctor'}
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
                                            {/* Patient (doctor) or doctor (patient) */}
                                            <div className="w-14 h-14 bg-gray-50 border border-gray-100 rounded-xl flex items-center justify-center text-brand font-bold shrink-0 shadow-sm group-hover:border-brand/20 transition-colors">
                                                {isDoctor
                                                    ? getPatientInitials(appt)
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
                                                    {isDoctor ? 'Assigned patient' : getDoctorSpecialty(appt.doctor_id)}
                                                </p>
                                                {isDoctor && appt.room_name ? (
                                                    <p className="text-[10px] text-gray-400 mt-0.5 font-mono truncate">
                                                        Room: {appt.room_name}
                                                    </p>
                                                ) : null}
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

                                                {canJoinJitsiVisit(appt) && (
                                                    <a
                                                        href={appt.join_url}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="flex items-center gap-2 px-5 py-2.5 bg-gray-900 text-white rounded-xl text-xs font-bold hover:bg-brand transition-all active:scale-95 shadow-sm shadow-black/5"
                                                        title={
                                                            isDoctor
                                                                ? 'Open the same Jitsi room as your patient'
                                                                : 'Join video consultation'
                                                        }
                                                    >
                                                        <Video className="h-3.5 w-3.5" />
                                                        {isDoctor ? 'Join Jitsi' : 'Join'}
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

            {isDoctor && (
                <DoctorSlotFormModal
                    isOpen={slotModalOpen}
                    onClose={() => {
                        setSlotModalOpen(false);
                        setEditingSlot(null);
                    }}
                    mode={slotModalMode}
                    slot={editingSlot}
                    profileHospital={doctorProfile?.hospital ?? ''}
                    onSubmit={async (payload) => {
                        if (slotModalMode === 'create') {
                            await appointmentApi.createSlot({
                                doctor_id: doctorProfile ? String(doctorProfile.id) : undefined,
                                start_time: payload.start_time,
                                end_time: payload.end_time,
                                hospital: payload.hospital,
                            });
                        } else if (editingSlot) {
                            await appointmentApi.updateSlot(editingSlot.id, {
                                start_time: payload.start_time,
                                end_time: payload.end_time,
                                hospital: payload.hospital,
                            });
                        }
                        await fetchData();
                    }}
                />
            )}

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
