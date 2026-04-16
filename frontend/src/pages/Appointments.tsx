import { useState, useEffect, useMemo } from 'react';
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
import { Link, useNavigate } from 'react-router-dom';
import type { Slot, Appointment, BookAppointmentRequest } from '../api/appointments';
import { appointmentApi } from '../api/appointments';
import { doctorApi as doctorsListApi, type Doctor } from '../api/doctors';
import { doctorApi as doctorProfileApi, type DoctorProfile } from '../api/doctor';
import { patientApi } from '../api/patient';
import { paymentApi } from '../api/payment';
import { submitPayHereForm } from '../utils/payment';
import toast from 'react-hot-toast';
import BookingModal from '../components/appointments/BookingModal';
import DoctorSlotFormModal from '../components/appointments/DoctorSlotFormModal';
import Dialog from '../components/ui/Dialog';

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
    const [isRedirecting, setIsRedirecting] = useState(false);
    const [mySlots, setMySlots] = useState<Slot[]>([]);
    const [doctorProfile, setDoctorProfile] = useState<DoctorProfile | null>(null);
    const [slotModalOpen, setSlotModalOpen] = useState(false);
    const [slotModalMode, setSlotModalMode] = useState<'create' | 'edit'>('create');
    const [slotEditing, setSlotEditing] = useState<Slot | null>(null);
    const [cancelDialogOpen, setCancelDialogOpen] = useState(false);
    const [appointmentToCancel, setAppointmentToCancel] = useState<Appointment | null>(null);
    const [isCancelling, setIsCancelling] = useState(false);
    
    // Slot deletion dialog state
    const [deleteSlotDialogOpen, setDeleteSlotDialogOpen] = useState(false);
    const [slotToDelete, setSlotToDelete] = useState<Slot | null>(null);
    const [consultationsPage, setConsultationsPage] = useState(1);
    const [consultationsPageSize, setConsultationsPageSize] = useState(10);
    const [availabilityPage, setAvailabilityPage] = useState(1);
    const [availabilityPageSize, setAvailabilityPageSize] = useState(10);

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

    const handleDeleteSlot = (slot: Slot) => {
        if (slot.is_booked) return;
        setSlotToDelete(slot);
        setDeleteSlotDialogOpen(true);
    };

    const confirmDeleteSlot = async () => {
        if (!slotToDelete) return;
        setIsLoading(true);
        try {
            await appointmentApi.deleteSlot(slotToDelete.id);
            toast.success('Availability slot removed');
            await fetchData();
        } catch (e) {
            console.error(e);
            toast.error('Could not delete slot. It may already be booked.');
        } finally {
            setIsLoading(false);
            setDeleteSlotDialogOpen(false);
            setSlotToDelete(null);
        }
    };

    const handleCancelBookedSlot = (slot: Slot) => {
        const appt = appointments.find(
            (a) => a.slot_id === slot.id && a.status !== 'cancelled' && a.status !== 'completed'
        );
        if (!appt) {
            toast.error('No appointment found for this slot.');
            return;
        }
        setAppointmentToCancel(appt);
        setCancelDialogOpen(true);
    };

    const handleConfirmCancel = async () => {
        if (!appointmentToCancel) return;
        setIsCancelling(true);
        try {
            await appointmentApi.cancelAppointment(appointmentToCancel.id);
            toast.success('Appointment cancelled successfully');
            await fetchData();
        } catch (e) {
            console.error(e);
            toast.error('Failed to cancel appointment.');
        } finally {
            setIsCancelling(false);
            setCancelDialogOpen(false);
            setAppointmentToCancel(null);
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

    const consultationsTotalPages = Math.max(1, Math.ceil(filteredAppointments.length / consultationsPageSize));
    const availabilityTotalPages = Math.max(1, Math.ceil(filteredMySlots.length / availabilityPageSize));

    const paginatedConsultations = useMemo(() => {
        const start = (consultationsPage - 1) * consultationsPageSize;
        return filteredAppointments.slice(start, start + consultationsPageSize);
    }, [consultationsPage, consultationsPageSize, filteredAppointments]);

    const paginatedAvailability = useMemo(() => {
        const start = (availabilityPage - 1) * availabilityPageSize;
        return filteredMySlots.slice(start, start + availabilityPageSize);
    }, [availabilityPage, availabilityPageSize, filteredMySlots]);

    useEffect(() => {
        setConsultationsPage(1);
    }, [searchQuery, consultationsPageSize, activeTab]);

    useEffect(() => {
        setAvailabilityPage(1);
    }, [searchQuery, availabilityPageSize, activeTab]);

    useEffect(() => {
        if (consultationsPage > consultationsTotalPages) {
            setConsultationsPage(consultationsTotalPages);
        }
    }, [consultationsPage, consultationsTotalPages]);

    useEffect(() => {
        if (availabilityPage > availabilityTotalPages) {
            setAvailabilityPage(availabilityTotalPages);
        }
    }, [availabilityPage, availabilityTotalPages]);

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

    const tabBtn = (on: boolean) =>
        `rounded-full px-3 py-1.5 text-xs font-semibold transition-colors ${
            on ? 'bg-brand/10 text-brand' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
        }`;

    return (
        <div className="min-h-screen bg-[#f6f8fa] font-sans">
            <div className="flex min-h-screen flex-col">
                <header className="min-h-14 flex flex-col gap-3 border-b border-gray-100 bg-white px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:gap-6 sm:px-8 sticky top-0 z-10">
                    <div className="flex min-w-0 flex-col gap-2">
                        <h2 className="text-[15px] font-bold text-gray-900">
                            {activeTab === 'doctors'
                                ? 'Book Appointment'
                                : activeTab === 'slots'
                                  ? 'Availability'
                                  : isDoctor
                                    ? 'Consultations'
                                    : 'My Appointments'}
                        </h2>
                        <div
                            className="flex flex-wrap gap-1.5"
                            role="tablist"
                            aria-label="Switch appointments view"
                        >
                            {!isDoctor && (
                                <button
                                    type="button"
                                    role="tab"
                                    aria-selected={activeTab === 'doctors'}
                                    className={tabBtn(activeTab === 'doctors')}
                                    onClick={() => setActiveTab('doctors')}
                                >
                                    Book appointment
                                </button>
                            )}
                            <button
                                type="button"
                                role="tab"
                                aria-selected={activeTab === 'appointments'}
                                className={tabBtn(activeTab === 'appointments')}
                                onClick={() => setActiveTab('appointments')}
                            >
                                {isDoctor ? 'Consultations' : 'My appointments'}
                            </button>
                            {isDoctor && (
                                <button
                                    type="button"
                                    role="tab"
                                    aria-selected={activeTab === 'slots'}
                                    className={tabBtn(activeTab === 'slots')}
                                    onClick={() => setActiveTab('slots')}
                                >
                                    Availability
                                </button>
                            )}
                        </div>
                    </div>
                    <div className="flex shrink-0 flex-wrap items-center gap-3 sm:gap-4">
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
                                        const count = docSlots.filter((s: Slot) => !s.is_booked && s.status !== 'booked' && new Date(s.start_time) > new Date()).length;
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
                            <div className="rounded-2xl border border-gray-100 bg-white p-5 shadow-sm">
                                <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                                    <div>
                                        <h3 className="text-lg font-semibold text-gray-900">Availability management</h3>
                                        <p className="mt-1 text-sm text-gray-500 max-w-2xl">
                                            Define your consulting windows so patients can book reliably. Booked slots can be
                                            released by cancelling the associated consultation.
                                        </p>
                                    </div>
                                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 lg:min-w-[27rem]">
                                        <div className="rounded-xl border border-gray-100 bg-gray-50 px-4 py-3">
                                            <p className="text-[11px] uppercase tracking-wider text-gray-400 font-semibold">Total slots</p>
                                            <p className="mt-1 text-xl font-semibold text-gray-900">{filteredMySlots.length}</p>
                                        </div>
                                        <div className="rounded-xl border border-emerald-100 bg-emerald-50 px-4 py-3">
                                            <p className="text-[11px] uppercase tracking-wider text-emerald-600 font-semibold">Available</p>
                                            <p className="mt-1 text-xl font-semibold text-emerald-700">
                                                {filteredMySlots.filter((s) => !s.is_booked).length}
                                            </p>
                                        </div>
                                        <div className="rounded-xl border border-amber-100 bg-amber-50 px-4 py-3">
                                            <p className="text-[11px] uppercase tracking-wider text-amber-700 font-semibold">Booked</p>
                                            <p className="mt-1 text-xl font-semibold text-amber-700">
                                                {filteredMySlots.filter((s) => s.is_booked).length}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </div>


                            {!profileHospitalStr && (
                                <div className="rounded-2xl border border-amber-200 bg-amber-50 px-5 py-4 text-sm text-amber-900 shadow-sm">
                                    Add your <strong>hospital</strong> on{' '}
                                    <Link to="/profile" className="font-semibold underline underline-offset-2">
                                        Profile
                                    </Link>{' '}
                                    before creating slots.
                                </div>
                            )}

                            {filteredMySlots.length === 0 ? (
                                <div className="bg-white rounded-2xl border border-gray-100 p-14 text-center shadow-sm">
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
                                    <div className="flex flex-col gap-3 border-b border-gray-100 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
                                        <p className="text-sm text-gray-500">
                                            Showing{' '}
                                            <span className="font-medium text-gray-700">
                                                {Math.min((availabilityPage - 1) * availabilityPageSize + 1, filteredMySlots.length)}
                                            </span>
                                            {' '}-{' '}
                                            <span className="font-medium text-gray-700">
                                                {Math.min(availabilityPage * availabilityPageSize, filteredMySlots.length)}
                                            </span>
                                            {' '}of <span className="font-medium text-gray-700">{filteredMySlots.length}</span>
                                        </p>
                                        <label className="inline-flex items-center gap-2 rounded-full border border-gray-200 bg-gray-50 px-3 py-1.5 text-xs text-gray-600">
                                            <span>Show</span>
                                            <select
                                                value={availabilityPageSize}
                                                onChange={(event) => setAvailabilityPageSize(Number(event.target.value))}
                                                className="bg-transparent text-xs font-medium text-gray-700 outline-none"
                                            >
                                                <option value={10}>10</option>
                                                <option value={20}>20</option>
                                                <option value={30}>30</option>
                                            </select>
                                        </label>
                                    </div>
                                    <div className="overflow-x-auto">
                                        <table className="w-full text-sm">
                                            <thead>
                                                <tr className="border-b border-gray-100 bg-gray-50 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500">
                                                    <th className="px-5 py-3.5">Hospital</th>
                                                    <th className="px-5 py-3.5">Start</th>
                                                    <th className="px-5 py-3.5">End</th>
                                                    <th className="px-5 py-3.5">Status</th>
                                                    <th className="px-5 py-3.5 text-right">Actions</th>
                                                </tr>
                                            </thead>
                                            <tbody className="divide-y divide-gray-100">
                                                {paginatedAvailability.map((slot) => {
                                                    const booked = slot.is_booked;
                                                    return (
                                                        <tr key={slot.id} className="hover:bg-gray-50/70 transition-colors">
                                                            <td className="px-5 py-3.5 text-gray-800">
                                                                <div className="font-medium text-gray-900">{slot.hospital || '—'}</div>
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
                                                                    className={`inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${
                                                                        booked
                                                                            ? 'bg-amber-50 text-amber-700 border border-amber-100'
                                                                            : 'bg-emerald-50 text-emerald-700 border border-emerald-100'
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
                                                                                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-gray-200 bg-white text-xs font-semibold text-gray-700 hover:bg-gray-50"
                                                                            >
                                                                                <Pencil className="h-3.5 w-3.5" />
                                                                                Edit
                                                                            </button>
                                                                            <button
                                                                                type="button"
                                                                                onClick={() => handleDeleteSlot(slot)}
                                                                                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-red-200 bg-red-50 text-xs font-semibold text-red-600 hover:bg-red-100"
                                                                            >
                                                                                <Trash2 className="h-3.5 w-3.5" />
                                                                                Remove
                                                                            </button>
                                                                        </>
                                                                    ) : (
                                                                        <button
                                                                            type="button"
                                                                            onClick={() => handleCancelBookedSlot(slot)}
                                                                            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-amber-200 bg-amber-50 text-xs font-semibold text-amber-800 hover:bg-amber-100"
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
                                    <div className="flex flex-col gap-3 border-t border-gray-100 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
                                        <span className="text-sm text-gray-500">
                                            Page <span className="font-medium text-gray-700">{availabilityPage}</span> of{' '}
                                            <span className="font-medium text-gray-700">{availabilityTotalPages}</span>
                                        </span>
                                        <div className="flex items-center gap-2">
                                            <button
                                                type="button"
                                                onClick={() => setAvailabilityPage((page) => Math.max(1, page - 1))}
                                                disabled={availabilityPage === 1}
                                                className="rounded-full border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-600 hover:border-brand/30 hover:text-brand disabled:cursor-not-allowed disabled:opacity-50"
                                            >
                                                Previous
                                            </button>
                                            <button
                                                type="button"
                                                onClick={() => setAvailabilityPage((page) => Math.min(availabilityTotalPages, page + 1))}
                                                disabled={availabilityPage === availabilityTotalPages}
                                                className="rounded-full border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-600 hover:border-brand/30 hover:text-brand disabled:cursor-not-allowed disabled:opacity-50"
                                            >
                                                Next
                                            </button>
                                        </div>
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
                                    <div className="flex flex-col gap-3 rounded-2xl border border-gray-100 bg-white px-4 py-4 sm:flex-row sm:items-center sm:justify-between">
                                        <p className="text-sm text-gray-500">
                                            Showing{' '}
                                            <span className="font-medium text-gray-700">
                                                {Math.min((consultationsPage - 1) * consultationsPageSize + 1, filteredAppointments.length)}
                                            </span>
                                            {' '}-{' '}
                                            <span className="font-medium text-gray-700">
                                                {Math.min(consultationsPage * consultationsPageSize, filteredAppointments.length)}
                                            </span>
                                            {' '}of <span className="font-medium text-gray-700">{filteredAppointments.length}</span>
                                        </p>
                                        <label className="inline-flex items-center gap-2 rounded-full border border-gray-200 bg-gray-50 px-3 py-1.5 text-xs text-gray-600">
                                            <span>Show</span>
                                            <select
                                                value={consultationsPageSize}
                                                onChange={(event) => setConsultationsPageSize(Number(event.target.value))}
                                                className="bg-transparent text-xs font-medium text-gray-700 outline-none"
                                            >
                                                <option value={10}>10</option>
                                                <option value={20}>20</option>
                                                <option value={30}>30</option>
                                            </select>
                                        </label>
                                    </div>

                                    {paginatedConsultations.map((appt) => (
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

                                                {appt.status !== 'cancelled' && appt.status !== 'completed' && !hasAppointmentEnded(appt) && (
                                                    <button
                                                        type="button"
                                                        onClick={() => {
                                                            setAppointmentToCancel(appt);
                                                            setCancelDialogOpen(true);
                                                        }}
                                                        className="flex items-center justify-center p-2.5 rounded-xl border border-red-100 text-red-500 hover:bg-red-50 hover:border-red-200 transition-all active:scale-95"
                                                        title="Cancel Appointment"
                                                    >
                                                        <Trash2 className="h-4 w-4" />
                                                    </button>
                                                )}
                                            </div>
                                        </div>
                                    ))}

                                    <div className="flex flex-col gap-3 rounded-2xl border border-gray-100 bg-white px-4 py-4 sm:flex-row sm:items-center sm:justify-between">
                                        <span className="text-sm text-gray-500">
                                            Page <span className="font-medium text-gray-700">{consultationsPage}</span> of{' '}
                                            <span className="font-medium text-gray-700">{consultationsTotalPages}</span>
                                        </span>
                                        <div className="flex items-center gap-2">
                                            <button
                                                type="button"
                                                onClick={() => setConsultationsPage((page) => Math.max(1, page - 1))}
                                                disabled={consultationsPage === 1}
                                                className="rounded-full border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-600 hover:border-brand/30 hover:text-brand disabled:cursor-not-allowed disabled:opacity-50"
                                            >
                                                Previous
                                            </button>
                                            <button
                                                type="button"
                                                onClick={() => setConsultationsPage((page) => Math.min(consultationsTotalPages, page + 1))}
                                                disabled={consultationsPage === consultationsTotalPages}
                                                className="rounded-full border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-600 hover:border-brand/30 hover:text-brand disabled:cursor-not-allowed disabled:opacity-50"
                                            >
                                                Next
                                            </button>
                                        </div>
                                    </div>
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

            {/* Cancellation Confirmation Dialog */}
            <Dialog 
                isOpen={cancelDialogOpen}
                onClose={() => {
                    if (!isCancelling) {
                        setCancelDialogOpen(false);
                        setAppointmentToCancel(null);
                    }
                }}
                title="Cancel Appointment?"
                description="We understand plans change. Please note that refunds are not issued for cancellations. Do you wish to continue?"
                confirmText="Yes, Cancel"
                cancelText="No, Keep it"
                variant="danger"
                onConfirm={handleConfirmCancel}
                isLoading={isCancelling}
            />

            {/* Slot Removal Dialog */}
            <Dialog
                isOpen={deleteSlotDialogOpen}
                onClose={() => setDeleteSlotDialogOpen(false)}
                title="Remove Slot?"
                description="This availability slot will be removed. Patients will no longer be able to book this time."
                confirmText="Remove Slot"
                variant="danger"
                onConfirm={confirmDeleteSlot}
            />

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
