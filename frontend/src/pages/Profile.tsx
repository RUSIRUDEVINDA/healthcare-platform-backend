import { useState, useEffect } from 'react';
import { User, Mail, Phone, MapPin, Calendar, Droplets, Shield, ArrowLeft, Save, Edit2 } from 'lucide-react';
import { patientApi, type PatientProfile } from '../api/patient';
import { Link } from 'react-router-dom';

export default function Profile() {
  const [profile, setProfile] = useState<PatientProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({
    date_of_birth: '',
    gender: '',
    phone_number: '',
    address: '',
    emergency_contact: '',
    blood_group: ''
  });

  const fetchProfile = async () => {
    try {
      setLoading(true);
      const data = await patientApi.getProfile();
      setProfile(data);
      setFormData({
        date_of_birth: data.date_of_birth ? data.date_of_birth.split('T')[0] : '',
        gender: data.gender || '',
        phone_number: data.phone_number || '',
        address: data.address || '',
        emergency_contact: data.emergency_contact || '',
        blood_group: data.blood_group || ''
      });
      setError(null);
    } catch (err: any) {
      if (err.response?.status === 404) {
        setError('PROFILE_NOT_FOUND');
      } else {
        setError('Failed to load profile. Please try again later.');
      }
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
      await patientApi.updateProfile(formData);
      await fetchProfile();
      setIsEditing(false);
    } catch (err) {
      console.error('Error updating profile:', err);
      alert('Failed to update profile');
    } finally {
      setLoading(false);
    }
  };

  if (loading && !profile) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand"></div>
      </div>
    );
  }

  if (error === 'PROFILE_NOT_FOUND') {
    return (
      <div className="min-h-screen bg-gray-50 flex flex-col items-center justify-center p-4">
        <div className="bg-white p-8 rounded-3xl shadow-sm border border-gray-100 max-w-md w-full text-center">
          <div className="w-20 h-20 bg-brand-light rounded-full flex items-center justify-center text-brand mx-auto mb-6">
            <User size={40} />
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">Profile Not Found</h2>
          <p className="text-gray-600 mb-8">
            It seems you don't have a patient profile yet. This usually happens when registration isn't fully synchronized.
          </p>
          <button 
            onClick={() => setIsEditing(true)}
            className="w-full py-3 bg-brand text-white rounded-xl font-semibold hover:bg-brand-dark transition-colors"
          >
            Create My Profile
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="mb-8 flex items-center justify-between">
          <Link to="/dashboard" className="flex items-center text-gray-600 hover:text-brand transition-colors">
            <ArrowLeft className="mr-2 h-5 w-5" /> Back to Dashboard
          </Link>
          {!isEditing && profile && (
            <button 
              onClick={() => setIsEditing(true)}
              className="flex items-center px-4 py-2 bg-brand text-white rounded-xl font-medium hover:bg-brand-dark transition-colors"
            >
              <Edit2 className="mr-2 h-4 w-4" /> Edit Profile
            </button>
          )}
        </div>

        <div className="bg-white rounded-3xl shadow-sm border border-gray-100 overflow-hidden">
          {/* Header */}
          <div className="h-32 bg-gradient-to-r from-brand to-brand-dark"></div>
          
          <div className="px-8 pb-8">
            <div className="relative -mt-16 mb-6">
              <div className="w-32 h-32 bg-white rounded-2xl shadow-md p-1">
                <div className="w-full h-full bg-brand-light rounded-xl flex items-center justify-center text-brand">
                  <User size={64} />
                </div>
              </div>
            </div>

            {isEditing ? (
              <form onSubmit={handleSubmit} className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-2">First Name (read-only)</label>
                    <input 
                      type="text" 
                      value={profile?.first_name || ''} 
                      disabled 
                      className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-gray-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-2">Last Name (read-only)</label>
                    <input 
                      type="text" 
                      value={profile?.last_name || ''} 
                      disabled 
                      className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-gray-500"
                    />
                  </div>
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
                    <select 
                      name="gender"
                      value={formData.gender}
                      onChange={handleInputChange}
                      className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-all"
                    >
                      <option value="">Select Gender</option>
                      <option value="Male">Male</option>
                      <option value="Female">Female</option>
                      <option value="Other">Other</option>
                    </select>
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
                        placeholder="+1 (555) 000-0000"
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
                <div>
                  <h1 className="text-3xl font-bold text-gray-900">{profile?.first_name} {profile?.last_name}</h1>
                  <p className="text-gray-500 mt-1 flex items-center">
                    <Mail className="h-4 w-4 mr-2" /> {profile?.email}
                  </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                  <div className="space-y-6">
                    <div className="flex items-start">
                      <div className="w-10 h-10 bg-blue-50 rounded-lg flex items-center justify-center text-blue-600 mr-4">
                        <Calendar className="h-5 w-5" />
                      </div>
                      <div>
                        <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Date of Birth</p>
                        <p className="text-gray-900 font-medium">{profile?.date_of_birth ? new Date(profile.date_of_birth).toLocaleDateString() : 'Not set'}</p>
                      </div>
                    </div>
                    
                    <div className="flex items-start">
                      <div className="w-10 h-10 bg-purple-50 rounded-lg flex items-center justify-center text-purple-600 mr-4">
                        <User className="h-5 w-5" />
                      </div>
                      <div>
                        <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Gender</p>
                        <p className="text-gray-900 font-medium">{profile?.gender || 'Not set'}</p>
                      </div>
                    </div>

                    <div className="flex items-start">
                      <div className="w-10 h-10 bg-green-50 rounded-lg flex items-center justify-center text-green-600 mr-4">
                        <Phone className="h-5 w-5" />
                      </div>
                      <div>
                        <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Phone Number</p>
                        <p className="text-gray-900 font-medium">{profile?.phone_number || 'Not set'}</p>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-6">
                    <div className="flex items-start">
                      <div className="w-10 h-10 bg-red-50 rounded-lg flex items-center justify-center text-red-600 mr-4">
                        <Droplets className="h-5 w-5" />
                      </div>
                      <div>
                        <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Blood Group</p>
                        <p className="text-gray-900 font-medium font-bold text-lg">{profile?.blood_group || 'Not set'}</p>
                      </div>
                    </div>

                    <div className="flex items-start">
                      <div className="w-10 h-10 bg-amber-50 rounded-lg flex items-center justify-center text-amber-600 mr-4">
                        <MapPin className="h-5 w-5" />
                      </div>
                      <div>
                        <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Address</p>
                        <p className="text-gray-900 font-medium">{profile?.address || 'Not set'}</p>
                      </div>
                    </div>

                    <div className="flex items-start">
                      <div className="w-10 h-10 bg-indigo-50 rounded-lg flex items-center justify-center text-indigo-600 mr-4">
                        <Shield className="h-5 w-5" />
                      </div>
                      <div>
                        <p className="text-xs font-bold text-gray-400 uppercase tracking-wider">Emergency Contact</p>
                        <p className="text-gray-900 font-medium">{profile?.emergency_contact || 'Not set'}</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
