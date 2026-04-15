import { useState, useEffect } from 'react';
import { User, Activity, Calendar, ClipboardList, Video } from 'lucide-react';
import { Link } from 'react-router-dom';
import { patientApi, type PatientProfile } from '../api/patient';
import { doctorApi, type DoctorProfile } from '../api/doctor';
import { appointmentApi, type Appointment } from '../api/appointments';
import { doctorApi as doctorsListApi, type Doctor } from '../api/doctors';

function getPatientDisplayName(appt: Appointment): string {
  const f = appt.patient_first_name?.trim() || '';
  const l = appt.patient_last_name?.trim() || '';
  if (f || l) {
    return `${f} ${l}`.trim();
  }
  const id = appt.patient_id || '';
  return id.length > 10 ? `Patient ${id.slice(0, 8)}…` : 'Patient';
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

export default function Dashboard() {
  const [profile, setProfile] = useState<PatientProfile | DoctorProfile | null>(null);
  const [role, setRole] = useState<string | null>(null);
  const [appointments, setAppointments] = useState<Appointment[]>([]);
  const [doctors, setDoctors] = useState<Record<string, Doctor>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const userString = localStorage.getItem('user');
    const user = userString ? JSON.parse(userString) : null;
    const userRole = user?.role;
    setRole(userRole);

    const fetchData = async () => {
      try {
        setLoading(true);
        // Fetch profile
        let profileData;
        if (userRole === 'doctor') {
          profileData = await doctorApi.getProfile();
        } else {
          profileData = await patientApi.getProfile();
        }
        setProfile(profileData);

        // Fetch appointments
        const appts = await appointmentApi.listAppointments();
        // Filter upcoming ones and sort by date
        const sortedAppts = appts
          .filter(a => a.status !== 'cancelled')
          .sort((a, b) => new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime());
        setAppointments(sortedAppts);

        if (userRole !== 'doctor') {
          const docs = await doctorsListApi.listDoctors();
          const docsMap: Record<string, Doctor> = {};
          docs.forEach(d => {
            docsMap[String(d.id)] = d;
            docsMap[d.user_id] = d;
          });
          setDoctors(docsMap);
        }

      } catch (err) {
        console.error('Failed to fetch dashboard data:', err);
      } finally {
        setLoading(false);
      }
    };

    if (userRole) {
      fetchData();
    }
  }, []);

  const getDoctorName = (appt: Appointment) => {
    // Try mapping by doctor_id (integer ID) first, then doctor_owner_user_id
    const doc = doctors[appt.doctor_id] || doctors[appt.doctor_owner_user_id || ''];
    return doc?.name || 'General Doctor';
  };

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const formatTime = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
  };

  const getDay = (dateStr: string) => new Date(dateStr).getDate();
  const getMonth = (dateStr: string) => new Date(dateStr).toLocaleDateString('en-US', { month: 'short' });

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand"></div>
      </div>
    );
  }

  const nextAppointment = appointments.length > 0 ? appointments[0] : null;

  return (
    <div className="flex min-h-screen flex-1 flex-col overflow-hidden bg-gray-50">
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-8 bg-white/80 backdrop-blur-md sticky top-0 z-10">
          <h2 className="text-xl font-semibold text-gray-800">Overview</h2>
          <div className="flex items-center space-x-4">
            <p className="text-sm font-medium text-gray-600 hidden sm:block">
              {role === 'doctor' ? (profile as DoctorProfile)?.name : `${(profile as PatientProfile)?.first_name} ${(profile as PatientProfile)?.last_name}`}
            </p>
            <div className="w-10 h-10 bg-brand-light rounded-full flex items-center justify-center text-brand border-2 border-brand/20">
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
              }!
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
              {nextAppointment ? (
                <>
                  <h4 className="text-xl font-bold text-gray-900 mt-1">{formatDate(nextAppointment.scheduled_at)}</h4>
                  <p className="text-sm text-gray-400 mt-2">
                    {role === 'doctor'
                      ? `${getPatientDisplayName(nextAppointment)} • ${formatTime(nextAppointment.scheduled_at)}`
                      : `${getDoctorName(nextAppointment)} • ${formatTime(nextAppointment.scheduled_at)}`}
                  </p>
                </>
              ) : (
                <>
                  <h4 className="text-xl font-bold text-gray-400 mt-1">No upcoming</h4>
                  <p className="text-sm text-gray-400 mt-2">
                    {role === 'doctor' ? 'No visits scheduled yet' : 'Book your next session'}
                  </p>
                </>
              )}
            </div>
            <div className="bg-white p-6 rounded-3xl border border-gray-100 shadow-sm hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-red-50 rounded-2xl flex items-center justify-center text-red-600 mb-4">
                <Activity className="h-6 w-6" />
              </div>
              <p className="text-sm font-medium text-gray-500">Waitlist Status</p>
              <h4 className="text-xl font-bold text-gray-900 mt-1">N/A</h4>
              <p className="text-sm text-gray-400 mt-2">No active waitlists</p>
            </div>
            <div className="bg-white p-6 rounded-3xl border border-gray-100 shadow-sm hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-purple-50 rounded-2xl flex items-center justify-center text-purple-600 mb-4">
                <ClipboardList className="h-6 w-6" />
              </div>
              <p className="text-sm font-medium text-gray-500">Total Appointments</p>
              <h4 className="text-xl font-bold text-gray-900 mt-1">{appointments.length} Records</h4>
              <p className="text-sm text-gray-400 mt-2">Across all periods</p>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-10">
            {/* Upcoming Appointments */}
            <div className="bg-white rounded-3xl border border-gray-100 shadow-sm p-8">
              <div className="flex items-center justify-between mb-8">
                <h4 className="text-xl font-bold text-gray-900">Upcoming Appointments</h4>
                <Link to="/appointments" className="text-brand font-semibold text-sm hover:underline">View All</Link>
              </div>
              <div className="space-y-6">
                {appointments.length > 0 ? (
                  appointments.slice(0, 3).map((appt) => (
                    <div key={appt.id} className="flex items-center p-4 rounded-2xl hover:bg-gray-50 transition-colors border border-transparent hover:border-gray-100">
                      <div className="w-14 h-14 bg-gray-100 rounded-xl mr-4 flex flex-col items-center justify-center">
                        <span className="text-xs font-bold text-gray-400 uppercase">{getMonth(appt.scheduled_at)}</span>
                        <span className="text-lg font-bold text-gray-900">{getDay(appt.scheduled_at)}</span>
                      </div>
                      <div className="flex-1">
                        <h5 className="font-bold text-gray-900 capitalize">{appt.consultation_mode} Consultation</h5>
                        <p className="text-sm text-gray-500">
                          {role === 'doctor' ? getPatientDisplayName(appt) : `Dr. ${getDoctorName(appt)}`}
                        </p>
                      </div>
                      <div className="text-right text-sm flex flex-col items-end gap-2">
                        <div>
                          <p className="font-semibold text-gray-900">{formatTime(appt.scheduled_at)}</p>
                          <span className={`px-2 py-1 rounded-md text-xs font-bold uppercase ${
                            appt.status === 'confirmed' ? 'bg-green-50 text-green-600' : 
                            appt.status === 'pending' ? 'bg-amber-50 text-amber-600' : 'bg-gray-50 text-gray-600'
                          }`}>
                            {appt.status}
                          </span>
                        </div>
                        {canJoinJitsiVisit(appt) && (
                          <a
                            href={appt.join_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="inline-flex items-center gap-1 text-xs font-bold text-brand hover:underline"
                          >
                            <Video className="h-3.5 w-3.5" />
                            {role === 'doctor' ? 'Join Jitsi' : 'Join'}
                          </a>
                        )}
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-10">
                    <Calendar className="h-10 w-10 text-gray-300 mx-auto mb-4" />
                    <p className="text-gray-500">No upcoming appointments found.</p>
                    {role !== 'doctor' && (
                      <Link to="/appointments" className="text-brand text-sm font-semibold mt-2 inline-block">Book Now</Link>
                    )}
                  </div>
                )}
              </div>
            </div>

            {/* Recent Records Placeholder (since we don't have a records service fully integrated in frontend yet) */}
            <div className="bg-white rounded-3xl border border-gray-100 shadow-sm p-8">
              <div className="flex items-center justify-between mb-8">
                <h4 className="text-xl font-bold text-gray-900">Recent Medical Records</h4>
                <button className="text-brand font-semibold text-sm hover:underline">See More</button>
              </div>
              <div className="space-y-6">
                <div className="text-center py-10">
                  <ClipboardList className="h-10 w-10 text-gray-300 mx-auto mb-4" />
                  <p className="text-gray-500">No medical records available yet.</p>
                  <p className="text-xs text-gray-400 mt-1">Your reports will appear here after consultation.</p>
                </div>
              </div>
            </div>
          </div>
        </main>
    </div>
  );
}

