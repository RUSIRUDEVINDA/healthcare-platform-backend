import { useEffect, useMemo, useState, type FormEvent, type ReactNode } from 'react';
import axios from 'axios';
import {
  Activity,
  ArrowRight,
  Calendar,
  ClipboardList,
  CreditCard,
  FileText,
  HeartPulse,
  Loader2,
  LogOut,
  Sparkles,
  Stethoscope,
  User,
  Scale,
  Video,
  Clock,
  Building2,
  ChevronDown,
} from 'lucide-react';

import { Link, useLocation, useNavigate } from 'react-router-dom';
import { patientApi, type PatientProfile } from '../api/patient';
import { doctorApi, type DoctorProfile } from '../api/doctor';
import { appointmentApi, type Appointment, type Slot } from '../api/appointments';
import { doctorApi as doctorsListApi, type Doctor } from '../api/doctors';
import { fileApi, type FileRecord } from '../api/files';
import { symptomApi, type SymptomCheckResponse } from '../api/symptom';

type RecordBucket = 'prescriptions' | 'reports';

function formatDate(dateStr: string) {
  const date = new Date(dateStr);
  if (Number.isNaN(date.getTime())) return 'Unknown date';
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

function formatTime(dateStr: string) {
  const date = new Date(dateStr);
  if (Number.isNaN(date.getTime())) return '--:--';
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  });
}

function uniqueStrings(values: Array<string | null | undefined>) {
  return Array.from(
    new Set(values.map((value) => value?.trim()).filter((value): value is string => Boolean(value)))
  );
}

function classifyRecord(file: FileRecord): RecordBucket {
  const haystack = `${file.original_name} ${file.stored_name}`.toLowerCase();

  if (haystack.includes('prescription') || haystack.includes('rx') || haystack.includes('medication')) {
    return 'prescriptions';
  }

  return 'reports';
}

function getFileTypeLabel(file: FileRecord) {
  return classifyRecord(file) === 'prescriptions' ? 'Prescription' : 'Report';
}

function getFileIcon() {
  return <FileText className="h-5 w-5" />;
}

function getPatientDisplayName(appt: Appointment) {
  const parts = [appt.patient_first_name, appt.patient_last_name].filter(Boolean);
  return parts.length ? parts.join(' ') : 'Patient';
}

function isSameLocalDay(a: Date, b: Date) {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

export default function Dashboard() {
  const location = useLocation();
  const navigate = useNavigate();
  const [profile, setProfile] = useState<PatientProfile | DoctorProfile | null>(null);
  const [role, setRole] = useState<string | null>(null);
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [mySlots, setMySlots] = useState<Slot[]>([]);
  const [doctors, setDoctors] = useState<Record<string, Doctor>>({});
  const [files, setFiles] = useState<FileRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [filesLoading, setFilesLoading] = useState(true);
  const [symptomInput, setSymptomInput] = useState('');
  const [symptomContext, setSymptomContext] = useState('');
  const [symptomLoading, setSymptomLoading] = useState(false);
  const [symptomResult, setSymptomResult] = useState<SymptomCheckResponse | null>(null);
  const [symptomError, setSymptomError] = useState<string | null>(null);

  useEffect(() => {
    const userString = localStorage.getItem('user');
    const user = userString ? JSON.parse(userString) : null;
    const userRole = user?.role ?? null;
    const authUserId = user?.id ?? user?.user_id ?? null;
    setRole(userRole);

    const fetchData = async () => {
      try {
        setLoading(true);
        setFilesLoading(true);

        const profilePromise = userRole === 'doctor' ? doctorApi.getProfile() : patientApi.getProfile();

        const apptsPromise = appointmentApi.listAppointments().catch(() => [] as Appointment[]);

        let resolvedProfile: PatientProfile | DoctorProfile | null = null;

        if (userRole === 'doctor') {
          const [profileData, apptsResult, slotsResult] = await Promise.all([
            profilePromise,
            apptsPromise,
            appointmentApi.listMySlots().catch(() => [] as Slot[]),
          ]);

          resolvedProfile = profileData;
          setProfile(profileData);

          const sortedAppts = (Array.isArray(apptsResult) ? apptsResult : [])
            .filter((appointment) => appointment.status !== 'cancelled')
            .sort((a, b) => new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime());
          setAppointments(sortedAppts);

          const slots = Array.isArray(slotsResult) ? slotsResult : [];
          slots.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime());
          setMySlots(slots);
          setDoctors({});
        } else {
          const [profileData, apptsResult, docsResult] = await Promise.all([
            profilePromise,
            apptsPromise,
            doctorsListApi.listDoctors().catch(() => [] as Doctor[]),
          ]);

          resolvedProfile = profileData;
          setProfile(profileData);

          const sortedAppts = (Array.isArray(apptsResult) ? apptsResult : [])
            .filter((appointment) => appointment.status !== 'cancelled')
            .sort((a, b) => new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime());
          setAppointments(sortedAppts);

          const docsMap: Record<string, Doctor> = {};
          (Array.isArray(docsResult) ? docsResult : []).forEach((doctor) => {
            docsMap[String(doctor.id)] = doctor;
            docsMap[doctor.user_id] = doctor;
          });
          setDoctors(docsMap);
          setMySlots([]);
        }

        let records: FileRecord[] = [];
        if (userRole === 'doctor') {
          records = await fileApi.listMyFiles();
        } else {
          const patientProfile = resolvedProfile as PatientProfile;
          // File service authorises listPatientFiles when caller JWT user_id matches the requested patient id.
          const candidateOwnerIds = uniqueStrings([authUserId, patientProfile?.user_id, patientProfile?.id]);
          const results = await Promise.allSettled(
            candidateOwnerIds.map((ownerId) => fileApi.listPatientFiles(ownerId))
          );
          const combined = results.flatMap((result) => (result.status === 'fulfilled' ? result.value : []));
          const fallbackFiles = combined.length === 0 ? await fileApi.listMyFiles() : [];
          records = Array.from(new Map([...combined, ...fallbackFiles].map((file) => [file.id, file])).values());
        }

        records.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
        setFiles(records);
      } catch (err) {
        console.error('Failed to fetch dashboard data:', err);
      } finally {
        setLoading(false);
        setFilesLoading(false);
      }
    };

    if (userRole) {
      fetchData();
    } else {
      setLoading(false);
      setFilesLoading(false);
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    window.location.href = '/auth/login';
  };

  const getDoctorName = (appointment: Appointment) => {
    const doctor = doctors[appointment.doctor_id] || doctors[appointment.doctor_owner_user_id || ''];
    return doctor?.name || 'General Doctor';
  };

  const nextAppointment = appointments[0] ?? null;

  const mostVisitedDoctor = useMemo(() => {
    type DoctorVisit = {
      doctorId: string;
      count: number;
    };

    const counts = new Map<string, { count: number; doctorId: string }>();

    appointments.forEach((appointment) => {
      const doctorKey = appointment.doctor_owner_user_id || appointment.doctor_id;
      const current = counts.get(doctorKey);
      counts.set(doctorKey, {
        doctorId: doctorKey,
        count: current ? current.count + 1 : 1,
      });
    });

    let topEntry: DoctorVisit | null = null;
    counts.forEach((value: DoctorVisit) => {
      if (!topEntry || value.count > topEntry.count) {
        topEntry = value;
      }
    });

    if (!topEntry) {
      return null;
    }

    const best = topEntry as { doctorId: string; count: number };
    const doctor = doctors[best.doctorId] || doctors[String(best.doctorId)] || null;
    return {
      name: doctor?.name || 'General Doctor',
      specialty: doctor?.specialization || 'Specialty not set',
      count: best.count,
    };
  }, [appointments, doctors]);

  const openFutureSlotsCount = useMemo(() => {
    if (role !== 'doctor') return 0;
    const now = Date.now();
    return mySlots.filter((s) => !s.is_booked && new Date(s.start_time).getTime() > now).length;
  }, [role, mySlots]);

  const todaysConsultationCount = useMemo(() => {
    if (role !== 'doctor') return 0;
    const now = new Date();
    return appointments.filter((a) => isSameLocalDay(new Date(a.scheduled_at), now)).length;
  }, [role, appointments]);

  const paymentAttentionCount = useMemo(() => {
    if (role !== 'doctor') return 0;
    return appointments.filter((a) => {
      const ps = (a.payment_status || '').toLowerCase();
      return ps === 'pending' || ps === 'overdue';
    }).length;
  }, [role, appointments]);

  const doctorMeetingEnded = (appointment: Appointment) => {
    const raw = appointment.scheduled_at || appointment.scheduled_time || '';
    const start = new Date(raw);
    if (Number.isNaN(start.getTime())) return false;
    const endMs = start.getTime() + (appointment.duration_minutes ?? 30) * 60 * 1000;
    return endMs <= Date.now();
  };

  const records = useMemo(() => files, [files]);
  const prescriptions = useMemo(() => records.filter((file) => classifyRecord(file) === 'prescriptions'), [records]);

  const latestPrescription = prescriptions[0] ?? null;

  const displayName =
    role === 'doctor'
      ? (profile as DoctorProfile | null)?.name || 'Doctor'
      : `${(profile as PatientProfile | null)?.first_name || 'Patient'} ${(profile as PatientProfile | null)?.last_name || ''}`.trim();

  const handleSymptomCheck = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSymptomError(null);
    setSymptomResult(null);

    const symptoms = symptomInput.trim();
    if (!symptoms) {
      setSymptomError('Please describe the symptoms first.');
      return;
    }

    try {
      setSymptomLoading(true);
      const result = await symptomApi.checkSymptoms({
        symptoms,
        optional_context: symptomContext.trim() || undefined,
      });
      setSymptomResult(result);
    } catch (err) {
      console.error('Failed to check symptoms:', err);
      let message = 'We could not analyze your symptoms right now.';
      if (axios.isAxiosError(err)) {
        message = err.response?.data?.error || err.response?.data?.message || message;
      }
      setSymptomError(message);
    } finally {
      setSymptomLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#eff9f8] flex items-center justify-center">
        <div className="h-12 w-12 animate-spin rounded-full border-[3px] border-brand border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#eff9f8] flex font-sans text-slate-900">
      <aside className="w-60 bg-white/90 backdrop-blur border-r border-teal-100 hidden lg:flex flex-col sticky top-0 h-screen">
        <div className="px-6 pt-6 pb-5">
          <Link to="/dashboard" className="flex items-center gap-2.5">
            <div className="h-9 w-9 rounded-full bg-brand flex items-center justify-center shadow-sm shadow-brand/20">
              <Activity className="h-4 w-4 text-white" />
            </div>
            <span className="text-lg font-medium tracking-tight text-slate-900">MediPulse SriLanka</span>
          </Link>
        </div>

        <nav className="flex-1 px-4 space-y-1">
          <p className="text-[10px] font-semibold text-slate-400 uppercase tracking-[0.32em] px-3 mb-3">Menu</p>
          <Link
            to="/dashboard"
            className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/dashboard'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <Activity className="h-[18px] w-[18px]" /> Dashboard
          </Link>
          <Link
            to="/profile"
            className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/profile'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <User className="h-[18px] w-[18px]" /> Profile
          </Link>
          <Link
            to="/appointments"
            className={`flex items-center justify-between px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/appointments'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <div className="flex items-center gap-3">
              <Calendar className="h-[18px] w-[18px]" /> Appointments
            </div>
            <ChevronDown className="h-4 w-4 opacity-50" />
          </Link>
          {role !== 'doctor' && (
            <Link
              to="/payments"
              className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
                location.pathname === '/payments'
                  ? 'bg-brand/10 text-brand font-semibold'
                  : 'text-slate-500 hover:bg-teal-50'
              }`}
            >
              <CreditCard className="h-[18px] w-[18px]" /> Payments
            </Link>
          )}
          {role !== 'doctor' && (
            <Link
              to="/symptom-checker"
              className="flex items-center gap-3 px-3 py-2.5 text-slate-500 hover:bg-teal-50 rounded-full transition-colors text-sm"
            >
              <Stethoscope className="h-[18px] w-[18px]" /> Symptom checker
            </Link>
          )}
          <Link
            to="/records"
            className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/records'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <ClipboardList className="h-[18px] w-[18px]" /> Records
          </Link>
          <Link
            to="/bmi-calculator"
            className={`flex items-center gap-3 px-3 py-2.5 rounded-full text-sm transition-colors ${
              location.pathname === '/bmi-calculator'
                ? 'bg-brand/10 text-brand font-semibold'
                : 'text-slate-500 hover:bg-teal-50'
            }`}
          >
            <Scale className="h-[18px] w-[18px]" /> BMI Calculator
          </Link>
        </nav>

        <div className="p-4 border-t border-teal-100 mx-4 mb-4">
          <button
            onClick={handleLogout}
            className="flex items-center gap-2.5 w-full px-3 py-2.5 text-red-500 hover:bg-red-50 rounded-full transition-colors text-sm font-medium"
          >
            <LogOut className="h-[18px] w-[18px]" /> Sign Out
          </button>
        </div>
      </aside>

      <div className="flex-1 flex flex-col min-h-screen overflow-hidden">
        <header className="h-16 bg-white border-b border-teal-100 flex items-center justify-between px-6 lg:px-8 sticky top-0 z-10">
          <div>
            <p className="text-[10px] uppercase tracking-[0.32em] text-slate-400 font-semibold">
              {role === 'doctor' ? 'Clinical overview' : 'Overview'}
            </p>
            <h1 className="text-lg font-bold text-slate-900">
              {role === 'doctor' ? 'Doctor dashboard' : 'Dashboard'}
            </h1>
          </div>
          <div className="flex items-center gap-4">
            <p className="hidden sm:block text-sm font-medium text-slate-600">{displayName}</p>
            <div className="h-10 w-10 rounded-full bg-brand-light border border-brand/20 flex items-center justify-center text-brand">
              <User className="h-5 w-5" />
            </div>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
          <div className="max-w-7xl mx-auto space-y-6">
            {role === 'doctor' ? (
              <section className="rounded-[2rem] border border-teal-100 bg-white p-6 sm:p-8 shadow-sm">
                <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
                  <div className="max-w-3xl space-y-4">
                    <div className="inline-flex items-center gap-2 rounded-full bg-brand/10 px-3 py-1 text-xs font-medium text-brand border border-brand/10">
                      <Stethoscope className="h-3.5 w-3.5" />
                      Clinical workspace
                    </div>
                    <div>
                      <h2 className="text-3xl sm:text-4xl font-medium tracking-tight text-slate-900">
                        Welcome back, {(profile as DoctorProfile | null)?.name || 'Doctor'}.
                      </h2>
                      <p className="mt-3 max-w-2xl text-sm sm:text-base leading-6 text-slate-600">
                        A focused view of your schedule, availability, and shared records—everything patients see about
                        bookings stays under Appointments.
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-wrap gap-3">
                    <Link
                      to="/appointments"
                      className="inline-flex items-center gap-2 rounded-full bg-slate-900 px-5 py-3 text-sm font-medium text-white shadow-sm"
                    >
                      Consultations · slots
                      <ArrowRight className="h-4 w-4" />
                    </Link>
                    <Link
                      to="/profile"
                      className="inline-flex items-center gap-2 rounded-full bg-white px-5 py-3 text-sm font-medium text-brand shadow-sm border border-brand/10"
                    >
                      Profile
                      <User className="h-4 w-4" />
                    </Link>
                    <Link
                      to="/records"
                      className="inline-flex items-center gap-2 rounded-full bg-white px-5 py-3 text-sm font-medium text-brand shadow-sm border border-brand/10"
                    >
                      Medical records
                      <ClipboardList className="h-4 w-4" />
                    </Link>
                  </div>
                </div>
              </section>
            ) : (
              <section className="rounded-[2rem] border border-teal-100 bg-white p-6 sm:p-8 shadow-sm">
                <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
                  <div className="max-w-3xl space-y-4">
                    <div className="inline-flex items-center gap-2 rounded-full bg-brand/10 px-3 py-1 text-xs font-medium text-brand border border-brand/10">
                      <Sparkles className="h-3.5 w-3.5" />
                      Care workspace
                    </div>
                    <div>
                      <h2 className="text-3xl sm:text-4xl font-medium tracking-tight text-slate-900">
                        Welcome back, {(profile as PatientProfile | null)?.first_name || 'Patient'}.
                      </h2>
                      <p className="mt-3 max-w-2xl text-sm sm:text-base leading-6 text-slate-600">
                        A calmer view of your care history, bookings, and medication flow. Prescriptions, reports, and
                        symptom guidance stay close without clutter.
                      </p>
                    </div>
                  </div>

                  <div className="flex flex-wrap gap-3">
                    <Link
                      to="/appointments"
                      className="inline-flex items-center gap-2 rounded-full bg-slate-900 px-5 py-3 text-sm font-medium text-white shadow-sm"
                    >
                      Book appointment
                      <ArrowRight className="h-4 w-4" />
                    </Link>
                    <Link
                      to="/records"
                      className="inline-flex items-center gap-2 rounded-full bg-white px-5 py-3 text-sm font-medium text-brand shadow-sm border border-brand/10"
                    >
                      View records
                      <ClipboardList className="h-4 w-4" />
                    </Link>
                  </div>
                </div>
              </section>
            )}

            {role === 'doctor' ? (
              <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <StatCard
                  icon={<Calendar className="h-5 w-5" />}
                  title="Next consultation"
                  value={nextAppointment ? formatDate(nextAppointment.scheduled_at) : 'None scheduled'}
                  caption={
                    nextAppointment
                      ? `${getPatientDisplayName(nextAppointment)} · ${formatTime(nextAppointment.scheduled_at)}`
                      : 'Add availability so patients can book'
                  }
                  tone="teal"
                />
                <StatCard
                  icon={<Clock className="h-5 w-5" />}
                  title="Today"
                  value={`${todaysConsultationCount}`}
                  caption={
                    todaysConsultationCount === 1
                      ? 'Consultation on your calendar today'
                      : 'Consultations on your calendar today'
                  }
                  tone="blue"
                />
                <StatCard
                  icon={<Calendar className="h-5 w-5" />}
                  title="Open slots"
                  value={`${openFutureSlotsCount}`}
                  caption="Future times patients can still book"
                  tone="emerald"
                />
                <StatCard
                  icon={<CreditCard className="h-5 w-5" />}
                  title="Payments to track"
                  value={`${paymentAttentionCount}`}
                  caption={
                    paymentAttentionCount > 0
                      ? 'Consultations with pending or overdue payment'
                      : 'No pending payment flags on your list'
                  }
                  tone="slate"
                />
              </section>
            ) : (
              <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                <StatCard
                  icon={<Calendar className="h-5 w-5" />}
                  title="Next appointment"
                  value={nextAppointment ? formatDate(nextAppointment.scheduled_at) : 'No appointment'}
                  caption={
                    nextAppointment
                      ? `${getDoctorName(nextAppointment)} • ${formatTime(nextAppointment.scheduled_at)}`
                      : 'Book your next visit when ready'
                  }
                  tone="teal"
                />
                <StatCard
                  icon={<Stethoscope className="h-5 w-5" />}
                  title="Most visited doctor"
                  value={mostVisitedDoctor ? mostVisitedDoctor.name : 'No visits yet'}
                  caption={
                    mostVisitedDoctor
                      ? `${mostVisitedDoctor.count} visits • ${mostVisitedDoctor.specialty}`
                      : 'It will appear after a few appointments'
                  }
                  tone="blue"
                />
                <StatCard
                  icon={<HeartPulse className="h-5 w-5" />}
                  title="Active prescriptions"
                  value={`${prescriptions.length}`}
                  caption={
                    latestPrescription
                      ? `Latest: ${latestPrescription.original_name}`
                      : 'Your medications will appear here'
                  }
                  tone="emerald"
                />
                <StatCard
                  icon={<ClipboardList className="h-5 w-5" />}
                  title="Total appointments"
                  value={`${appointments.length}`}
                  caption="Across all periods"
                  tone="slate"
                />
              </section>
            )}

            {role === 'doctor' ? (
              <section className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
                <div className="space-y-6">
                  <section className="rounded-[2rem] border border-white bg-white shadow-[0_20px_60px_rgba(8,47,73,0.06)] overflow-hidden">
                    <div className="flex items-center justify-between gap-3 px-5 py-4 sm:px-6 border-b border-slate-100">
                      <div>
                        <h3 className="text-lg font-medium text-slate-900">Upcoming consultations</h3>
                        <p className="mt-1 text-sm text-slate-500">Next patients on your schedule.</p>
                      </div>
                      <Link
                        to="/appointments"
                        className="inline-flex items-center gap-2 rounded-full bg-brand/10 px-4 py-2 text-sm font-semibold text-brand"
                      >
                        View all
                        <ArrowRight className="h-4 w-4" />
                      </Link>
                    </div>

                    <div className="p-5 sm:p-6">
                      {appointments.length > 0 ? (
                        <div className="space-y-3">
                          {appointments.slice(0, 5).map((appointment) => (
                            <div
                              key={appointment.id}
                              className="flex flex-col gap-3 rounded-[1.5rem] border border-slate-100 bg-white px-4 py-4 sm:flex-row sm:items-center sm:justify-between"
                            >
                              <div className="flex items-center gap-4 min-w-0">
                                <div className="flex h-14 w-14 flex-col items-center justify-center rounded-[1.25rem] bg-white border border-brand/10 text-center shadow-sm shrink-0">
                                  <span className="text-[11px] font-medium uppercase tracking-[0.2em] text-slate-400">
                                    {new Date(appointment.scheduled_at).toLocaleDateString('en-US', { month: 'short' })}
                                  </span>
                                  <span className="text-lg font-medium text-slate-900">
                                    {new Date(appointment.scheduled_at).getDate()}
                                  </span>
                                </div>
                                <div className="min-w-0">
                                  <h4 className="text-base font-medium text-slate-900 truncate">
                                    {getPatientDisplayName(appointment)}
                                  </h4>
                                  <p className="text-sm text-slate-500">
                                    {appointment.consultation_mode === 'jitsi' || appointment.consultation_mode === 'video'
                                      ? 'Video'
                                      : 'Physical'}{' '}
                                    · {appointment.status}
                                    {appointment.payment_status ? ` · Payment: ${appointment.payment_status}` : ''}
                                  </p>
                                </div>
                              </div>
                              <div className="flex items-center gap-3 shrink-0">
                                <p className="text-sm font-semibold text-slate-900 tabular-nums">
                                  {formatTime(appointment.scheduled_at)}
                                </p>
                                {(appointment.consultation_mode === 'jitsi' || appointment.consultation_mode === 'video') &&
                                  appointment.join_url &&
                                  appointment.status !== 'cancelled' && (
                                    <button
                                      type="button"
                                      onClick={() =>
                                        navigate(
                                          `/telemedicine?join_url=${encodeURIComponent(appointment.join_url || '')}&peer=${encodeURIComponent(getPatientDisplayName(appointment))}&title=${encodeURIComponent('Telemedicine Session')}`
                                        )
                                      }
                                      disabled={doctorMeetingEnded(appointment)}
                                      className={`inline-flex items-center gap-1.5 rounded-full px-4 py-2 text-xs font-semibold ${
                                        doctorMeetingEnded(appointment)
                                          ? 'bg-slate-200 text-slate-400 cursor-not-allowed'
                                          : 'bg-slate-900 text-white hover:bg-brand'
                                      }`}
                                    >
                                      <Video className="h-3.5 w-3.5" />
                                      {doctorMeetingEnded(appointment) ? 'Ended' : 'Join'}
                                    </button>
                                  )}
                              </div>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <div className="rounded-[1.5rem] border border-dashed border-slate-200 bg-slate-50 px-6 py-12 text-center">
                          <Calendar className="mx-auto h-10 w-10 text-slate-300" />
                          <h4 className="mt-4 text-lg font-medium text-slate-900">No upcoming consultations</h4>
                          <p className="mt-2 text-sm text-slate-500">When patients book you, they will appear here.</p>
                        </div>
                      )}
                    </div>
                  </section>

                  <section className="rounded-[2rem] border border-white bg-white shadow-[0_20px_60px_rgba(8,47,73,0.06)] overflow-hidden">
                    <div className="flex items-center justify-between gap-3 px-5 py-4 sm:px-6 border-b border-slate-100">
                      <div>
                        <h3 className="text-lg font-medium text-slate-900">Your recent uploads</h3>
                        <p className="mt-1 text-sm text-slate-500">Files you have added for patients.</p>
                      </div>
                      <Link
                        to="/records"
                        className="inline-flex items-center gap-2 rounded-full bg-brand/10 px-4 py-2 text-sm font-semibold text-brand"
                      >
                        See more
                        <ArrowRight className="h-4 w-4" />
                      </Link>
                    </div>

                    <div className="p-5 sm:p-6">
                      {filesLoading ? (
                        <div className="flex items-center justify-center py-12">
                          <Loader2 className="h-8 w-8 animate-spin text-brand" />
                        </div>
                      ) : records.length > 0 ? (
                        <div className="space-y-3">
                          {records.slice(0, 3).map((file) => (
                            <div
                              key={file.id}
                              className="flex items-center gap-4 rounded-[1.5rem] border border-slate-100 bg-white px-4 py-4"
                            >
                              <div className="flex h-12 w-12 items-center justify-center rounded-[1.1rem] bg-white text-brand shadow-sm border border-brand/10">
                                {getFileIcon()}
                              </div>
                              <div className="min-w-0 flex-1">
                                <div className="flex flex-wrap items-center gap-2">
                                  <p className="truncate text-sm font-medium text-slate-900">{file.original_name}</p>
                                  <span className="rounded-full bg-white px-2.5 py-1 text-[11px] font-semibold text-slate-500 border border-slate-100">
                                    {getFileTypeLabel(file)}
                                  </span>
                                </div>
                                <p className="mt-1 text-xs text-slate-500">
                                  {formatDate(file.created_at)} • {file.mime_type}
                                </p>
                              </div>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <div className="rounded-[1.5rem] border border-dashed border-slate-200 bg-slate-50 px-6 py-12 text-center">
                          <FileText className="mx-auto h-10 w-10 text-slate-300" />
                          <h4 className="mt-4 text-lg font-medium text-slate-900">No uploads yet</h4>
                          <p className="mt-2 text-sm text-slate-500">
                            Shared documents will show up after you add them from Records.
                          </p>
                        </div>
                      )}
                    </div>
                  </section>
                </div>

                <section className="rounded-[2rem] border border-white bg-white shadow-[0_20px_60px_rgba(8,47,73,0.06)] overflow-hidden">
                  <div className="px-5 py-4 sm:px-6 border-b border-slate-100">
                    <h3 className="text-lg font-medium text-slate-900">Practice snapshot</h3>
                    <p className="mt-1 text-sm text-slate-500">Pulled from your profile.</p>
                  </div>
                  <div className="p-5 sm:p-6 space-y-4">
                    <div className="flex items-start gap-3 rounded-[1.25rem] border border-slate-100 bg-slate-50/80 px-4 py-3">
                      <Building2 className="h-5 w-5 text-brand shrink-0 mt-0.5" />
                      <div>
                        <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-400">Hospital</p>
                        <p className="text-sm font-medium text-slate-900 mt-0.5">
                          {(profile as DoctorProfile | null)?.hospital?.trim() || '—'}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-start gap-3 rounded-[1.25rem] border border-slate-100 bg-slate-50/80 px-4 py-3">
                      <Stethoscope className="h-5 w-5 text-brand shrink-0 mt-0.5" />
                      <div>
                        <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-400">Specialty</p>
                        <p className="text-sm font-medium text-slate-900 mt-0.5">
                          {(profile as DoctorProfile | null)?.specialization?.trim() || '—'}
                        </p>
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div className="rounded-[1.25rem] border border-slate-100 bg-white px-4 py-3">
                        <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-400">SLMC</p>
                        <p className="text-sm font-medium text-slate-900 mt-1">
                          {(profile as DoctorProfile | null)?.slmc_no?.trim() || '—'}
                        </p>
                      </div>
                      <div className="rounded-[1.25rem] border border-slate-100 bg-white px-4 py-3">
                        <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-400">Channeling</p>
                        <p className="text-sm font-medium text-slate-900 mt-1">
                          {(profile as DoctorProfile | null)?.channeling_fee != null
                            ? `LKR ${Number((profile as DoctorProfile).channeling_fee).toLocaleString()}`
                            : '—'}
                        </p>
                      </div>
                    </div>
                    <Link
                      to="/appointments"
                      className="inline-flex w-full items-center justify-center gap-2 rounded-full border border-brand/20 bg-brand/5 px-4 py-3 text-sm font-semibold text-brand hover:bg-brand/10 transition-colors"
                    >
                      Manage availability and consultations
                      <ArrowRight className="h-4 w-4" />
                    </Link>
                  </div>
                </section>
              </section>
            ) : (
              <section className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
                <div className="space-y-6">
                  <section className="rounded-[2rem] border border-white bg-white shadow-[0_20px_60px_rgba(8,47,73,0.06)] overflow-hidden">
                    <div className="flex items-center justify-between gap-3 px-5 py-4 sm:px-6 border-b border-slate-100">
                      <div>
                        <h3 className="text-lg font-medium text-slate-900">Upcoming appointments</h3>
                        <p className="mt-1 text-sm text-slate-500">Your nearest visit shows up first.</p>
                      </div>
                      <Link to="/appointments" className="inline-flex items-center gap-2 rounded-full bg-brand/10 px-4 py-2 text-sm font-semibold text-brand">
                        View all
                        <ArrowRight className="h-4 w-4" />
                      </Link>
                    </div>

                    <div className="p-5 sm:p-6">
                      {appointments.length > 0 ? (
                        <div className="space-y-3">
                          {appointments.slice(0, 3).map((appointment) => (
                            <div
                              key={appointment.id}
                              className="flex items-center gap-4 rounded-[1.5rem] border border-slate-100 bg-white px-4 py-4"
                            >
                              <div className="flex h-14 w-14 flex-col items-center justify-center rounded-[1.25rem] bg-white border border-brand/10 text-center shadow-sm">
                                <span className="text-[11px] font-medium uppercase tracking-[0.2em] text-slate-400">
                                  {new Date(appointment.scheduled_at).toLocaleDateString('en-US', { month: 'short' })}
                                </span>
                                <span className="text-lg font-medium text-slate-900">{new Date(appointment.scheduled_at).getDate()}</span>
                              </div>

                              <div className="min-w-0 flex-1">
                                <div className="flex flex-wrap items-center gap-2">
                                  <h4 className="truncate text-base font-medium text-slate-900 capitalize">
                                    {appointment.consultation_mode} consultation
                                  </h4>
                                  <span className="rounded-full bg-brand/10 px-2.5 py-1 text-[11px] font-semibold text-brand">
                                    {appointment.status}
                                  </span>
                                </div>
                                <p className="mt-1 text-sm text-slate-500">{getDoctorName(appointment)}</p>
                              </div>

                              <div className="text-right">
                                <p className="text-sm font-semibold text-slate-900">{formatTime(appointment.scheduled_at)}</p>
                                <p className="mt-1 text-xs text-slate-400 uppercase tracking-[0.2em]">{appointment.consultation_mode === 'jitsi' || appointment.consultation_mode === 'video' ? 'Video' : 'Physical'}</p>
                              </div>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <div className="rounded-[1.5rem] border border-dashed border-slate-200 bg-slate-50 px-6 py-12 text-center">
                          <Calendar className="mx-auto h-10 w-10 text-slate-300" />
                          <h4 className="mt-4 text-lg font-medium text-slate-900">No upcoming appointments</h4>
                          <p className="mt-2 text-sm text-slate-500">Once you book a session, it will show up here.</p>
                        </div>
                      )}
                    </div>
                  </section>

                  <section className="rounded-[2rem] border border-white bg-white shadow-[0_20px_60px_rgba(8,47,73,0.06)] overflow-hidden">
                    <div className="flex items-center justify-between gap-3 px-5 py-4 sm:px-6 border-b border-slate-100">
                      <div>
                        <h3 className="text-lg font-medium text-slate-900">Recent medical records</h3>
                        <p className="mt-1 text-sm text-slate-500">Prescriptions and reports from your chart.</p>
                      </div>
                      <Link to="/records" className="inline-flex items-center gap-2 rounded-full bg-brand/10 px-4 py-2 text-sm font-semibold text-brand">
                        See more
                        <ArrowRight className="h-4 w-4" />
                      </Link>
                    </div>

                    <div className="p-5 sm:p-6">
                      {filesLoading ? (
                        <div className="flex items-center justify-center py-12">
                          <Loader2 className="h-8 w-8 animate-spin text-brand" />
                        </div>
                      ) : records.length > 0 ? (
                        <div className="space-y-3">
                          {records.slice(0, 3).map((file) => (
                            <div
                              key={file.id}
                              className="flex items-center gap-4 rounded-[1.5rem] border border-slate-100 bg-white px-4 py-4"
                            >
                                <div className="flex h-12 w-12 items-center justify-center rounded-[1.1rem] bg-white text-brand shadow-sm border border-brand/10">
                                {getFileIcon()}
                                </div>
                              <div className="min-w-0 flex-1">
                                <div className="flex flex-wrap items-center gap-2">
                                  <p className="truncate text-sm font-medium text-slate-900">{file.original_name}</p>
                                  <span className="rounded-full bg-white px-2.5 py-1 text-[11px] font-semibold text-slate-500 border border-slate-100">
                                    {getFileTypeLabel(file)}
                                  </span>
                                </div>
                                <p className="mt-1 text-xs text-slate-500">
                                  {formatDate(file.created_at)} • {file.mime_type}
                                </p>
                              </div>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <div className="rounded-[1.5rem] border border-dashed border-slate-200 bg-slate-50 px-6 py-12 text-center">
                          <FileText className="mx-auto h-10 w-10 text-slate-300" />
                          <h4 className="mt-4 text-lg font-medium text-slate-900">No records yet</h4>
                          <p className="mt-2 text-sm text-slate-500">Prescriptions and reports will appear here after upload.</p>
                        </div>
                      )}
                    </div>
                  </section>
                </div>

                <section className="rounded-[2rem] border border-white bg-white shadow-[0_20px_60px_rgba(8,47,73,0.06)] overflow-hidden">
                  <div className="flex items-center justify-between gap-3 px-5 py-4 sm:px-6 border-b border-slate-100">
                    <div>
                      <h3 className="text-lg font-medium text-slate-900">AI symptom checker</h3>
                      <p className="mt-1 text-sm text-slate-500">Describe what you feel and get a gentle specialty suggestion.</p>
                    </div>
                    <div className="rounded-full bg-brand/10 px-3 py-1 text-xs font-semibold text-brand">
                      Connected
                    </div>
                  </div>

                  <div className="p-5 sm:p-6 space-y-5">
                    <form onSubmit={handleSymptomCheck} className="space-y-4">
                      <div>
                        <label className="mb-2 block text-sm font-semibold text-slate-700">Symptoms</label>
                        <textarea
                          value={symptomInput}
                          onChange={(event) => setSymptomInput(event.target.value)}
                          rows={6}
                          placeholder="Example: fever, cough, chest tightness, fatigue, or pain..."
                          className="w-full rounded-[1.5rem] border border-slate-200 bg-slate-50 px-4 py-3 text-sm outline-none transition focus:border-brand/30 focus:ring-4 focus:ring-brand/10"
                        />
                      </div>

                      <div>
                        <label className="mb-2 block text-sm font-semibold text-slate-700">Optional context</label>
                        <input
                          value={symptomContext}
                          onChange={(event) => setSymptomContext(event.target.value)}
                          placeholder="Age, duration, any medical history..."
                          className="w-full rounded-full border border-slate-200 bg-slate-50 px-4 py-3 text-sm outline-none transition focus:border-brand/30 focus:ring-4 focus:ring-brand/10"
                        />
                      </div>

                      <button
                        type="submit"
                        disabled={symptomLoading}
                        className="inline-flex w-full items-center justify-center gap-2 rounded-full bg-brand px-5 py-3 text-sm font-medium text-white shadow-sm disabled:cursor-not-allowed disabled:opacity-70"
                      >
                        {symptomLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                        {symptomLoading ? 'Checking symptoms...' : 'Check symptoms'}
                      </button>
                    </form>

                    {symptomError && (
                      <div className="rounded-[1.5rem] border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
                        {symptomError}
                      </div>
                    )}

                    {symptomResult ? (
                      <div className="space-y-3 rounded-[1.75rem] border border-brand/10 bg-[#f8fffe] p-5">
                        <div className="flex items-center justify-between gap-3">
                          <div>
                            <p className="text-[11px] font-semibold uppercase tracking-[0.28em] text-slate-400">Suggested specialty</p>
                            <h4 className="mt-1 text-2xl font-medium text-slate-900">{symptomResult.suggested_specialty || 'General practice'}</h4>
                          </div>
                          <div className="rounded-full bg-white px-3 py-1.5 text-xs font-semibold text-brand border border-brand/10">
                            AI guidance
                          </div>
                        </div>

                        <div className="rounded-[1.5rem] border border-slate-100 bg-white p-4">
                          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-slate-400">Preliminary notes</p>
                          <p className="mt-2 text-sm leading-6 text-slate-700">{symptomResult.preliminary_notes}</p>
                        </div>

                        <div className="rounded-[1.5rem] border border-amber-100 bg-amber-50 px-4 py-3 text-sm leading-6 text-amber-800">
                          {symptomResult.disclaimer}
                        </div>
                      </div>
                    ) : (
                      <div className="rounded-[1.5rem] border border-dashed border-slate-200 bg-slate-50 px-5 py-6 text-sm text-slate-500">
                        Your symptom guidance will appear here after you run a check.
                      </div>
                    )}
                  </div>
                </section>
              </section>
            )}
          </div>
        </main>
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
  icon: ReactNode;
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
      <p className="mt-2 break-words text-[1.1rem] font-medium tracking-tight text-slate-900">{value}</p>
      <p className="mt-2 text-sm leading-5 text-slate-500">{caption}</p>
    </div>
  );
}