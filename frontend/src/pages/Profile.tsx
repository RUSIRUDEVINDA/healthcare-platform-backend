import { useState, useEffect } from 'react';
import axios from 'axios';
import {
  User,
  Phone,
  MapPin,
  Calendar,
  Droplets,
  Shield,
  Save,
  Edit2,
  Medal,
  Building2,
  CreditCard,
  Stethoscope,
  Trash2,
  AlertTriangle,
  Globe,
  Activity,
  ClipboardList,
  LogOut,
  Scale,
  ChevronDown,
  Mail,
  BadgeCheck,
  CircleDollarSign,
  IdCard,
} from 'lucide-react';
import { patientApi, type PatientProfile } from '../api/patient';
import { doctorApi, type DoctorProfile } from '../api/doctor';
import { Link, useNavigate, useLocation } from 'react-router-dom';

export default function Profile() {
  const location = useLocation();
  const navigate = useNavigate();
  const [profile, setProfile] = useState<PatientProfile | DoctorProfile | null>(null);
  const [role, setRole] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [isEditing, setIsEditing] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const formatDateOnly = (value?: string | null) => {
    if (!value) return 'Not set';
    if (value.includes('T')) return value.split('T')[0];
    if (value.length >= 10) return value.slice(0, 10);
    return value;
  };

  const initialsFromName = (name: string | undefined | null) => {
    if (!name?.trim()) return 'DR';
    const parts = name.trim().split(/\s+/).filter(Boolean);
    if (parts.length === 0) return 'DR';
    if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
    return `${parts[0]![0] ?? ''}${parts[parts.length - 1]![0] ?? ''}`.toUpperCase() || 'DR';
  };

  // Unified form data for both roles
  const [formData, setFormData] = useState({
    // Patient fields
    date_of_birth: '',
    gender: '',
    phone_number: '',
    address: '',
    emergency_contact: '',
    blood_group: '',
    nationality: '',
    nic: '',
    // Doctor fields
    specialization: '',
    experience: '',
    hospital: '',
    slmc_no: '',
    channeling_fee: ''
  });

  const fetchProfile = async () => {
    try {
      setLoading(true);
      const userString = localStorage.getItem('user');
      const user = userString ? JSON.parse(userString) : null;
      const userRole = user?.role;
      setRole(userRole);

      if (userRole === 'doctor') {
        const doctorData = await doctorApi.getProfile();
        setProfile(doctorData);
        setFormData(prev => ({
          ...prev,
          specialization: doctorData.specialization || '',
          experience: doctorData.experience?.toString() || '',
          hospital: doctorData.hospital || '',
          nic: doctorData.nic || '',
          slmc_no: doctorData.slmc_no || '',
          channeling_fee: doctorData.channeling_fee?.toString() || ''
        }));
      } else {
        const patientData = await patientApi.getProfile();
        setProfile(patientData);
        setFormData(prev => ({
          ...prev,
          date_of_birth: patientData.date_of_birth ? patientData.date_of_birth.split('T')[0] : '',
          gender: patientData.gender || '',
          phone_number: patientData.phone_number || '',
          address: patientData.address || '',
          emergency_contact: patientData.emergency_contact || '',
          blood_group: patientData.blood_group || '',
          nationality: patientData.nationality || '',
          nic: patientData.nic || ''
        }));
      }
    } catch (err: unknown) {
      // Error state removed to fix build
      console.error('Error fetching profile:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      if (role === 'doctor') {
        const updateData: Record<string, string | number> = {};

        const specialization = formData.specialization.trim();
        const hospital = formData.hospital.trim();
        const nic = formData.nic.trim();
        const slmcNo = formData.slmc_no.trim();
        const experienceRaw = formData.experience.trim();

        if (specialization !== '') updateData.specialization = specialization;
        if (hospital !== '') updateData.hospital = hospital;
        if (nic !== '') updateData.nic = nic;
        if (slmcNo !== '') updateData.slmc_no = slmcNo;
        if (experienceRaw !== '') {
          const parsedExperience = Number.parseInt(experienceRaw, 10);
          if (!Number.isNaN(parsedExperience)) {
            updateData.experience = parsedExperience;
          }
        }
        const channelingFeeRaw = formData.channeling_fee.trim();
        if (channelingFeeRaw !== '') {
          const parsedFee = Number.parseFloat(channelingFeeRaw);
          if (!Number.isNaN(parsedFee)) {
            updateData.channeling_fee = parsedFee;
          }
        }

        if (Object.keys(updateData).length === 0) {
          alert('Please enter at least one doctor profile field to update.');
          setLoading(false);
          return;
        }

        await doctorApi.updateProfile((profile as DoctorProfile).id, updateData);
      } else {
        await patientApi.updateProfile(formData);
      }
      await fetchProfile();
      setIsEditing(false);
    } catch (err: unknown) {
      console.error('Error updating profile:', err);
      let message = 'Failed to update profile';
      if (axios.isAxiosError(err)) {
        message = err.response?.data?.error || message;
      }
      alert(message);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteProfile = async () => {
    try {
      setLoading(true);
      if (role === 'patient') {
        await patientApi.deleteProfile();
        // Clear auth data and redirect
        localStorage.removeItem('access_token');
        localStorage.removeItem('user');
        navigate('/auth/login');
      }
    } catch (err: unknown) {
      console.error('Error deleting profile:', err);
      let message = 'Failed to delete profile';
      if (axios.isAxiosError(err)) {
        message = err.response?.data?.error || message;
      }
      alert(message);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('user');
    window.location.href = '/auth/login';
  };

  if (loading && !profile) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#f6f8fa] flex font-sans">
      {/* Sidebar */}
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
              className="flex items-center gap-3 px-3 py-2.5 text-gray-500 hover:bg-gray-50 rounded-xl transition-colors text-sm"
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

      {/* Main Content */}
      <div className="flex-1 flex flex-col min-h-screen overflow-y-auto">
        <header className="min-h-14 bg-white border-b border-gray-100 flex items-center justify-between px-6 sm:px-8 py-3 sticky top-0 z-10">
          <div>
            <h2 className="text-[15px] font-bold text-gray-900">
              {role === 'doctor' ? 'Doctor profile' : 'Your profile'}
            </h2>
            
          </div>
          <div className="flex items-center gap-4 shrink-0">
            {!isEditing && profile && (
              <button
                onClick={() => setIsEditing(true)}
                className="flex items-center gap-2 px-4 py-2 bg-brand text-white rounded-xl text-sm font-medium hover:bg-brand-dark transition-colors shadow-sm"
              >
                <Edit2 className="h-4 w-4" /> Edit Profile
              </button>
            )}
          </div>
        </header>

        <main className="p-8">
          {!profile && !loading ? (
            <div className="bg-white p-8 rounded-3xl shadow-sm border border-gray-100 max-w-md mx-auto text-center">
              <div className="w-20 h-20 bg-brand-light rounded-full flex items-center justify-center text-brand mx-auto mb-6">
                <User size={40} />
              </div>
              <h2 className="text-2xl font-bold text-gray-900 mb-2">Profile Not Found</h2>
              <p className="text-gray-600 mb-8">
                {role === 'doctor'
                  ? "It seems you don't have a doctor profile yet. Please contact support if this persists."
                  : "It seems you don't have a patient profile yet."}
              </p>
              {role !== 'doctor' && (
                <button
                  onClick={() => setIsEditing(true)}
                  className="w-full py-3 bg-brand text-white rounded-xl font-semibold hover:bg-brand-dark transition-colors"
                >
                  Create My Profile
                </button>
              )}
            </div>
          ) : (
            <div className="max-w-4xl mx-auto">
              <div className="bg-white rounded-3xl shadow-sm border border-slate-200/80 overflow-hidden">
                {role !== 'doctor' && (
                  <div className="relative h-40 overflow-hidden bg-gradient-to-r from-brand to-brand-dark" />
                )}

                <div className={`relative bg-white px-6 pb-8 sm:px-8 ${role === 'doctor' ? 'pt-6 sm:pt-8' : ''}`}>
                  {role === 'doctor' && profile ? (
                    <div className="mb-6 flex flex-col gap-4 sm:mb-8 sm:flex-row sm:items-center sm:gap-6">
                      <div className="relative z-10 w-[7.5rem] shrink-0 sm:w-32">
                        <div className="rounded-2xl bg-white p-1 shadow-lg ring-1 ring-slate-200/80">
                          <div className="flex aspect-square w-full items-center justify-center rounded-xl bg-gradient-to-br from-slate-700 to-slate-900 text-white">
                            <span className="font-[Georgia,Cambria,serif] text-3xl font-semibold tracking-tight sm:text-4xl">
                              {initialsFromName((profile as DoctorProfile).name)}
                            </span>
                          </div>
                        </div>
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="text-[10px] font-semibold uppercase tracking-[0.28em] text-slate-400">
                          Licensed Doctor
                        </p>
                        <h2 className="mt-1.5 text-2xl font-semibold tracking-tight text-slate-900 sm:text-[1.65rem]">
                          {(profile as DoctorProfile).name}
                        </h2>
                        {!isEditing && (
                          <p className="mt-1 text-base font-medium text-brand">
                            {(profile as DoctorProfile).specialization?.trim() || 'Specialization not set'}
                          </p>
                        )}
                      </div>
                    </div>
                  ) : (
                    <div className="relative z-10 -mt-16 mb-6 w-[7.5rem] sm:w-32 sm:mb-6">
                      <div className="rounded-2xl bg-white p-1 shadow-xl ring-1 ring-black/5">
                        <div className="flex aspect-square w-full items-center justify-center rounded-xl bg-brand-light text-brand">
                          <User size={64} className="opacity-90" />
                        </div>
                      </div>
                    </div>
                  )}

                  <div className="relative z-0">
                  {isEditing ? (
                    <form onSubmit={handleSubmit} className="space-y-6">
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        {role !== 'doctor' && (
                          <div className="md:col-span-2">
                            <label className="block text-sm font-semibold text-gray-700 mb-2">Full Name (read-only)</label>
                            <input
                              type="text"
                              value={`${(profile as PatientProfile)?.first_name} ${(profile as PatientProfile)?.last_name}`}
                              disabled
                              className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-gray-500"
                            />
                          </div>
                        )}

                        {role === 'doctor' ? (
                          <>
                            <div className="md:col-span-2 pt-2">
                              <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">
                                Details
                              </p>
                            </div>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Specialization</label>
                              <div className="relative">
                                <Stethoscope className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="text"
                                  name="specialization"
                                  value={formData.specialization}
                                  onChange={handleInputChange}
                                  placeholder="e.g. Cardiologist"
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Experience (Years)</label>
                              <div className="relative">
                                <Medal className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="number"
                                  name="experience"
                                  value={formData.experience}
                                  onChange={handleInputChange}
                                  placeholder="e.g. 10"
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                            <div className="md:col-span-2">
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Hospital</label>
                              <div className="relative">
                                <Building2 className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="text"
                                  name="hospital"
                                  value={formData.hospital}
                                  onChange={handleInputChange}
                                  placeholder="e.g. General Hospital"
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                            <div className="md:col-span-2 pt-4 border-t border-slate-100">
                              <p className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">
                                Credentials & pricing
                              </p>
                            </div>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">NIC Number</label>
                              <div className="relative">
                                <IdCard className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="text"
                                  name="nic"
                                  value={formData.nic}
                                  onChange={handleInputChange}
                                  placeholder="Enter 12-digit NIC"
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">SLMC Registration No</label>
                              <div className="relative">
                                <BadgeCheck className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="text"
                                  name="slmc_no"
                                  value={formData.slmc_no}
                                  onChange={handleInputChange}
                                  placeholder="e.g. 12345"
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                            <div className="md:col-span-2">
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Channeling fee (LKR)</label>
                              <div className="relative">
                                <CircleDollarSign className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="number"
                                  name="channeling_fee"
                                  value={formData.channeling_fee}
                                  onChange={handleInputChange}
                                  placeholder="e.g. 2500"
                                  min={0}
                                  step={100}
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                          </>
                        ) : (
                          <>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Date of Birth</label>
                              <div className="relative">
                                <Calendar className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="date"
                                  name="date_of_birth"
                                  value={formData.date_of_birth}
                                  onChange={handleInputChange}
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Gender</label>
                              <div className="relative">
                                <User className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <select
                                  name="gender"
                                  value={formData.gender}
                                  onChange={handleInputChange}
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                >
                                  <option value="">Select Gender</option>
                                  <option value="male">Male</option>
                                  <option value="female">Female</option>
                                  <option value="other">Other</option>
                                </select>
                              </div>
                            </div>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Phone Number</label>
                              <div className="relative">
                                <Phone className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="tel"
                                  name="phone_number"
                                  value={formData.phone_number}
                                  onChange={handleInputChange}
                                  placeholder="+94 77 123 4567"
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">NIC Number</label>
                              <div className="relative">
                                <CreditCard className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="text"
                                  name="nic"
                                  value={formData.nic}
                                  onChange={handleInputChange}
                                  placeholder="e.g. 199512345678"
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Blood Group</label>
                              <div className="relative">
                                <Droplets className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <select
                                  name="blood_group"
                                  value={formData.blood_group}
                                  onChange={handleInputChange}
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                >
                                  <option value="">Select Blood Group</option>
                                  <option value="A+">A+</option>
                                  <option value="A-">A-</option>
                                  <option value="B+">B+</option>
                                  <option value="B-">B-</option>
                                  <option value="AB+">AB+</option>
                                  <option value="AB-">AB-</option>
                                  <option value="O+">O+</option>
                                  <option value="O-">O-</option>
                                </select>
                              </div>
                            </div>
                            <div className="md:col-span-2">
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Address</label>
                              <div className="relative">
                                <MapPin className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <textarea
                                  name="address"
                                  value={formData.address}
                                  onChange={handleInputChange}
                                  rows={2}
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                ></textarea>
                              </div>
                            </div>
                            <div className="md:col-span-2">
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Emergency Contact</label>
                              <div className="relative">
                                <Shield className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="text"
                                  name="emergency_contact"
                                  value={formData.emergency_contact}
                                  onChange={handleInputChange}
                                  placeholder="Name - Relationship - Phone"
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                            <div>
                              <label className="block text-sm font-semibold text-gray-700 mb-2">Nationality</label>
                              <div className="relative">
                                <Globe className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
                                <input
                                  type="text"
                                  name="nationality"
                                  value={formData.nationality}
                                  onChange={handleInputChange}
                                  placeholder="e.g. Sri Lankan"
                                  className="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                                />
                              </div>
                            </div>
                          </>
                        )}
                      </div>

                      <div className="flex space-x-4 pt-4">
                        <button
                          type="submit"
                          className="flex-1 flex items-center justify-center py-3 bg-brand text-white rounded-xl font-semibold hover:bg-brand-dark transition-colors shadow-lg shadow-brand/20"
                          disabled={loading}
                        >
                          <Save className="mr-2 h-5 w-5" /> {loading ? 'Saving...' : 'Save Changes'}
                        </button>
                        <button
                          type="button"
                          onClick={() => setIsEditing(false)}
                          className="flex-1 py-3 bg-gray-100 text-gray-700 rounded-xl font-semibold hover:bg-gray-200 transition-colors"
                        >
                          Cancel
                        </button>
                      </div>
                    </form>
                  ) : (
                    <div className="space-y-8">
                      {role === 'doctor' && profile ? (
                        (() => {
                          const doc = profile as DoctorProfile;
                          const fee =
                            doc.channeling_fee != null && !Number.isNaN(Number(doc.channeling_fee))
                              ? Number(doc.channeling_fee).toLocaleString('en-LK', {
                                  minimumFractionDigits: 0,
                                  maximumFractionDigits: 0,
                                })
                              : '—';
                          return (
                            <>
                              <div className="flex flex-col gap-3 border-b border-slate-100 pb-6 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
                                {doc.email?.trim() && (
                                  <a
                                    href={`mailto:${doc.email.trim()}`}
                                    className="inline-flex items-center gap-2 text-sm text-slate-600 hover:text-brand transition-colors"
                                  >
                                    <Mail className="h-4 w-4 shrink-0 text-slate-400" aria-hidden />
                                    {doc.email.trim()}
                                  </a>
                                )}
                                {doc.slmc_no?.trim() && (
                                  <div className="inline-flex w-fit items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs font-medium text-slate-700">
                                    <BadgeCheck className="h-3.5 w-3.5 shrink-0 text-emerald-600" aria-hidden />
                                    SLMC {doc.slmc_no.trim()}
                                  </div>
                                )}
                              </div>

                              <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                                <div className="lg:col-span-2 rounded-2xl border border-slate-200/90 bg-slate-50/60 p-5 sm:p-6">
                                  <h5 className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">
                                    Details
                                  </h5>
                                  <dl className="mt-4 space-y-4">
                                    <div className="flex gap-4">
                                      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white shadow-sm ring-1 ring-slate-200/80 text-slate-600">
                                        <Building2 className="h-5 w-5" aria-hidden />
                                      </div>
                                      <div className="min-w-0">
                                        <dt className="text-xs font-medium text-slate-500">Primary hospital</dt>
                                        <dd className="mt-0.5 text-sm font-medium text-slate-900">
                                          {doc.hospital?.trim() || '—'}
                                        </dd>
                                      </div>
                                    </div>
                                    <div className="flex gap-4">
                                      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white shadow-sm ring-1 ring-slate-200/80 text-slate-600">
                                        <Stethoscope className="h-5 w-5" aria-hidden />
                                      </div>
                                      <div className="min-w-0">
                                        <dt className="text-xs font-medium text-slate-500">Clinical experience</dt>
                                        <dd className="mt-0.5 text-sm font-medium text-slate-900">
                                          {doc.experience != null && !Number.isNaN(Number(doc.experience))
                                            ? `${doc.experience} years`
                                            : '—'}
                                        </dd>
                                      </div>
                                    </div>
                                  </dl>
                                </div>

                                <div className="rounded-2xl border border-slate-200/90 bg-white p-5 sm:p-6 shadow-sm">
                                  <h5 className="text-[11px] font-semibold uppercase tracking-[0.2em] text-slate-500">
                                    Registration
                                  </h5>
                                  <dl className="mt-4 space-y-3 text-sm">
                                    <div>
                                      <dt className="text-xs text-slate-500">SLMC registration</dt>
                                      <dd className="font-medium text-slate-900 tabular-nums">
                                        {doc.slmc_no?.trim() || '—'}
                                      </dd>
                                    </div>
                                    <div>
                                      <dt className="text-xs text-slate-500">NIC</dt>
                                      <dd className="font-medium text-slate-900 tabular-nums">
                                        {doc.nic?.trim() || '—'}
                                      </dd>
                                    </div>
                                  </dl>
                                </div>
                              </div>

                              <div className="flex flex-col gap-3 rounded-2xl border border-brand/20 bg-gradient-to-br from-brand/[0.06] to-teal-50/50 p-5 sm:flex-row sm:items-center sm:justify-between sm:p-6">
                                <div>
                                  <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-500">
                                    Consultation fee
                                  </p>
                                  <p className="mt-1 text-2xl font-semibold tracking-tight text-slate-900 tabular-nums">
                                    {fee === '—' ? fee : `LKR ${fee}`}
                                  </p>
                                  
                                </div>
                                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-white/80 text-brand shadow-sm ring-1 ring-brand/15">
                                  <CircleDollarSign className="h-6 w-6" aria-hidden />
                                </div>
                              </div>
                            </>
                          );
                        })()
                      ) : (
                        <>
                      <div>
                        <h4 className="text-2xl font-bold text-gray-900">
                          {(profile as PatientProfile)?.first_name} {(profile as PatientProfile)?.last_name}
                        </h4>
                        <p className="text-brand font-semibold mt-1">Registered User</p>
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                          <>
                            <div className="flex items-start">
                              <div className="w-10 h-10 bg-brand/5 rounded-lg flex items-center justify-center text-brand mr-4">
                                <Activity className="h-5 w-5" />
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Registered Email</p>
                                <p className="text-gray-900 font-semibold">{(profile as PatientProfile).email}</p>
                              </div>
                            </div>
                            <div className="flex items-start">
                              <div className="w-10 h-10 bg-blue-50 rounded-lg flex items-center justify-center text-blue-600 mr-4">
                                <Calendar className="h-5 w-5" />
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Date of Birth</p>
                                <p className="text-gray-900 font-medium">{formatDateOnly((profile as PatientProfile).date_of_birth)}</p>
                              </div>
                            </div>
                            <div className="flex items-start">
                              <div className="w-10 h-10 bg-purple-50 rounded-lg flex items-center justify-center text-purple-600 mr-4">
                                <User className="h-5 w-5" />
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Gender</p>
                                <p className="text-gray-900 font-medium capitalize">{(profile as PatientProfile).gender || 'Not set'}</p>
                              </div>
                            </div>
                            <div className="flex items-start">
                              <div className="w-10 h-10 bg-indigo-50 rounded-lg flex items-center justify-center text-indigo-600 mr-4">
                                <Phone className="h-5 w-5" />
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Phone</p>
                                <p className="text-gray-900 font-medium">{(profile as PatientProfile).phone_number || 'Not set'}</p>
                              </div>
                            </div>
                            <div className="flex items-start">
                              <div className="w-10 h-10 bg-rose-50 rounded-lg flex items-center justify-center text-rose-600 mr-4">
                                <CreditCard className="h-5 w-5" />
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">NIC Number</p>
                                <p className="text-gray-900 font-medium">{(profile as PatientProfile).nic || 'Not set'}</p>
                              </div>
                            </div>
                            <div className="flex items-start">
                              <div className="w-10 h-10 bg-red-50 rounded-lg flex items-center justify-center text-red-600 mr-4">
                                <Droplets className="h-5 w-5" />
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Blood Group</p>
                                <p className="text-gray-900 font-bold text-lg">{(profile as PatientProfile).blood_group || 'Not set'}</p>
                              </div>
                            </div>
                            <div className="flex items-start">
                              <div className="w-10 h-10 bg-amber-50 rounded-lg flex items-center justify-center text-amber-600 mr-4">
                                <MapPin className="h-5 w-5" />
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Address</p>
                                <p className="text-gray-900 font-medium">{(profile as PatientProfile).address || 'Not set'}</p>
                              </div>
                            </div>
                            <div className="flex items-start">
                              <div className="w-10 h-10 bg-green-50 rounded-lg flex items-center justify-center text-green-600 mr-4">
                                <Shield className="h-5 w-5" />
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Emergency Contact</p>
                                <p className="text-gray-900 font-bold">{(profile as PatientProfile).emergency_contact || 'Not set'}</p>
                              </div>
                            </div>
                            <div className="flex items-start">
                              <div className="w-10 h-10 bg-teal-50 rounded-lg flex items-center justify-center text-teal-600 mr-4">
                                <Globe className="h-5 w-5" />
                              </div>
                              <div>
                                <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Nationality</p>
                                <p className="text-gray-900 font-bold">{(profile as PatientProfile).nationality || 'Not set'}</p>
                              </div>
                            </div>
                          </>
                      </div>
                    </>
                      )}
                    </div>
                  )}
                  </div>
                </div>
              </div>

              {/* Danger Zone for Patients */}
              {role === 'patient' && !isEditing && (
                <div className="mt-8 bg-white rounded-3xl shadow-sm border border-red-100 overflow-hidden">
                  <div className="px-8 py-6 flex flex-col md:flex-row items-center justify-between gap-4">
                    <div>
                      <h3 className="text-lg font-bold text-red-600 flex items-center">
                        <AlertTriangle className="mr-2 h-5 w-5" /> Danger Zone
                      </h3>
                      <p className="text-gray-500 text-sm mt-1">
                        Once you delete your profile, your patient record will be removed. This action cannot be undone.
                      </p>
                    </div>
                    {!showDeleteConfirm ? (
                      <button
                        onClick={() => setShowDeleteConfirm(true)}
                        className="w-full md:w-auto px-6 py-2 border-2 border-red-600 text-red-600 rounded-xl font-semibold hover:bg-red-50 transition-colors"
                      >
                        Delete Account
                      </button>
                    ) : (
                      <div className="flex space-x-3 w-full md:w-auto">
                        <button
                          onClick={handleDeleteProfile}
                          className="flex-1 md:flex-none px-6 py-2 bg-red-600 text-white rounded-xl font-semibold hover:bg-red-700 transition-all flex items-center justify-center"
                          disabled={loading}
                        >
                          <Trash2 className="mr-2 h-4 w-4" /> {loading ? 'Deleting...' : 'Confirm'}
                        </button>
                        <button
                          onClick={() => setShowDeleteConfirm(false)}
                          className="flex-1 md:flex-none px-6 py-2 bg-gray-100 text-gray-700 rounded-xl font-semibold hover:bg-gray-200 transition-colors"
                        >
                          Cancel
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
