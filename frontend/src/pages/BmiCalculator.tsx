import { useState, useEffect } from 'react';
import { User, Scale, HeartPulse, RefreshCcw, Info } from 'lucide-react';
import { patientApi, type PatientProfile } from '../api/patient';
import { doctorApi, type DoctorProfile } from '../api/doctor';

export default function BmiCalculator() {
  const [weight, setWeight] = useState<string>('');
  const [height, setHeight] = useState<string>('');
  const [bmi, setBmi] = useState<number | null>(null);
  const [role] = useState<string | null>(() => {
    const userString = localStorage.getItem('user');
    const user = userString ? JSON.parse(userString) : null;
    return user?.role || null;
  });
  const [profile, setProfile] = useState<PatientProfile | DoctorProfile | null>(null);

  useEffect(() => {
    const userString = localStorage.getItem('user');
    const user = userString ? JSON.parse(userString) : null;
    const userRole = user?.role;    const fetchProfile = async () => {
        try {
            if (userRole === 'doctor') {
               const docProfile = await doctorApi.getProfile();
               setProfile(docProfile);
            } else if (userRole === 'patient') {
               const patProfile = await patientApi.getProfile();
               setProfile(patProfile);
            }
        } catch (err) {
            console.error('Failed to fetch profile', err);
        }
    };
    if (userRole) {
        fetchProfile();
    }
  }, []);

  const calculateBmi = () => {
    const w = parseFloat(weight);
    const h = parseFloat(height) / 100; // convert cm to meters

    if (w > 0 && h > 0) {
      const result = w / (h * h);
      setBmi(parseFloat(result.toFixed(1)));
    } else {
      setBmi(null);
    }
  };

  const resetCalculator = () => {
    setWeight('');
    setHeight('');
    setBmi(null);
  };

  const getBmiCategory = (score: number) => {
    if (score < 18.5) return { category: 'Underweight', color: 'text-blue-500', bg: 'bg-blue-50', border: 'border-blue-200' };
    if (score >= 18.5 && score < 25) return { category: 'Normal weight', color: 'text-green-500', bg: 'bg-green-50', border: 'border-green-200' };
    if (score >= 25 && score < 30) return { category: 'Overweight', color: 'text-amber-500', bg: 'bg-amber-50', border: 'border-amber-200' };
    if (score >= 30 && score < 35) return { category: 'Obesity Class I', color: 'text-orange-500', bg: 'bg-orange-50', border: 'border-orange-200' };
    if (score >= 35 && score < 40) return { category: 'Obesity Class II', color: 'text-red-500', bg: 'bg-red-50', border: 'border-red-200' };
    return { category: 'Obesity Class III', color: 'text-rose-600', bg: 'bg-rose-50', border: 'border-rose-200' };
  };

  return (
    <div className="min-h-screen bg-[#f6f8fa] font-sans">
      <div className="flex min-h-screen flex-col">
        <header className="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-8 bg-white/80 backdrop-blur-md sticky top-0 z-10">
          <h2 className="text-lg font-semibold tracking-tight text-slate-900">Health tools</h2>
          <div className="flex items-center space-x-4">
            <p className="text-sm font-medium text-gray-600 hidden sm:block">
              {role === 'doctor' ? (profile as DoctorProfile)?.name : profile ? `${(profile as PatientProfile)?.first_name} ${(profile as PatientProfile)?.last_name}` : ''}
            </p>
            <div className="w-10 h-10 bg-brand-light rounded-full flex items-center justify-center text-brand border-2 border-brand/20">
              <User className="h-6 w-6" />
            </div>
          </div>
        </header>

        <main className="p-8 max-w-5xl mx-auto w-full">
          <div className="mb-8">
            <h3 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
              <HeartPulse className="h-7 w-7 text-brand" /> Body Mass Index (BMI)
            </h3>
            <p className="text-gray-500 mt-2 max-w-2xl">
              Calculate your BMI to determine if you are at a healthy weight relative to your height, according to standard international results.
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
            {/* Calculator Form */}
            <div className="lg:col-span-5 bg-white rounded-3xl border border-gray-100 shadow-sm p-8 h-fit">
              <h4 className="text-lg font-bold text-gray-800 mb-6">Your Measurements</h4>

              <div className="space-y-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Weight (kg)</label>
                  <div className="relative">
                    <input
                      type="number"
                      placeholder="e.g. 70"
                      value={weight}
                      onChange={(e) => setWeight(e.target.value)}
                      className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand/20 focus:border-brand/40 transition-all font-medium text-gray-900 [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                    />
                    <span className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 font-semibold text-sm">kg</span>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Height (cm)</label>
                  <div className="relative">
                    <input
                      type="number"
                      placeholder="e.g. 175"
                      value={height}
                      onChange={(e) => setHeight(e.target.value)}
                      className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand/20 focus:border-brand/40 transition-all font-medium text-gray-900 [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                    />
                    <span className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 font-semibold text-sm">cm</span>
                  </div>
                </div>

                <div className="pt-2 flex gap-3">
                  <button
                    onClick={calculateBmi}
                    disabled={!weight || !height}
                    className="flex-1 py-3.5 bg-brand text-white rounded-xl font-bold hover:bg-brand-dark transition-colors shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Calculate BMI
                  </button>
                  <button
                    onClick={resetCalculator}
                    className="p-3.5 bg-gray-100 text-gray-500 rounded-xl hover:bg-gray-200 transition-colors"
                    title="Reset"
                  >
                    <RefreshCcw className="h-5 w-5" />
                  </button>
                </div>
              </div>
            </div>

            {/* Results Display */}
            <div className="lg:col-span-7 space-y-6">
              {/* Score Card */}
              {bmi !== null ? (
                <div className={`rounded-3xl border-2 p-8 shadow-sm transition-all duration-500 animate-in zoom-in-95 ${getBmiCategory(bmi).bg} ${getBmiCategory(bmi).border}`}>
                  <div className="text-center mb-6">
                    <p className="text-sm font-bold uppercase tracking-widest text-gray-500 mb-2">Your BMI Score</p>
                    <div className="flex items-baseline justify-center gap-2">
                      <h2 className={`text-6xl font-black ${getBmiCategory(bmi).color}`}>{bmi}</h2>
                      <span className="text-gray-500 font-semibold">kg/m²</span>
                    </div>
                  </div>

                  <div className="bg-white/60 backdrop-blur-sm rounded-2xl p-6 text-center shadow-sm border border-white/50">
                    <p className="text-gray-600 mb-1">According to the international classification, your result is:</p>
                    <h3 className={`text-2xl font-bold ${getBmiCategory(bmi).color}`}>{getBmiCategory(bmi).category}</h3>
                  </div>
                </div>
              ) : (
                <div className="h-64 rounded-3xl border border-gray-100 bg-white shadow-sm flex flex-col items-center justify-center p-8 text-center text-gray-400 border-dashed">
                  <Scale className="h-12 w-12 mb-4 text-gray-200" />
                  <p className="font-medium text-gray-500">Enter your weight and height to view your BMI result.</p>
                </div>
              )}

              {/* Reference Table */}
              <div className="bg-white rounded-3xl border border-gray-100 shadow-sm overflow-hidden">
                <div className="px-6 py-5 border-b border-gray-100 bg-gray-50/50 flex items-center gap-2">
                  <Info className="h-5 w-5 text-gray-400" />
                  <h4 className="font-bold text-gray-800">International Reference Results</h4>
                </div>
                <div className="p-0">
                  <table className="w-full text-left text-sm">
                    <thead className="bg-gray-50/80 text-gray-500">
                      <tr>
                        <th className="px-6 py-3.5 font-bold uppercase tracking-wider">Classification</th>
                        <th className="px-6 py-3.5 font-bold uppercase tracking-wider">BMI Range (kg/m²)</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      <tr className={bmi !== null && bmi < 18.5 ? "bg-blue-50/50 font-semibold" : ""}>
                        <td className="px-6 py-4 text-blue-600 flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-blue-500"></div> Underweight</td>
                        <td className="px-6 py-4 text-gray-600">&lt; 18.5</td>
                      </tr>
                      <tr className={bmi !== null && bmi >= 18.5 && bmi < 25 ? "bg-green-50/50 font-semibold" : ""}>
                        <td className="px-6 py-4 text-green-600 flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-green-500"></div> Normal weight</td>
                        <td className="px-6 py-4 text-gray-600">18.5 - 24.9</td>
                      </tr>
                      <tr className={bmi !== null && bmi >= 25 && bmi < 30 ? "bg-amber-50/50 font-semibold" : ""}>
                        <td className="px-6 py-4 text-amber-600 flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-amber-500"></div> Overweight</td>
                        <td className="px-6 py-4 text-gray-600">25.0 - 29.9</td>
                      </tr>
                      <tr className={bmi !== null && bmi >= 30 && bmi < 35 ? "bg-orange-50/50 font-semibold" : ""}>
                        <td className="px-6 py-4 text-orange-600 flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-orange-500"></div> Obesity Class I</td>
                        <td className="px-6 py-4 text-gray-600">30.0 - 34.9</td>
                      </tr>
                      <tr className={bmi !== null && bmi >= 35 && bmi < 40 ? "bg-red-50/50 font-semibold" : ""}>
                        <td className="px-6 py-4 text-red-600 flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-red-500"></div> Obesity Class II</td>
                        <td className="px-6 py-4 text-gray-600">35.0 - 39.9</td>
                      </tr>
                      <tr className={bmi !== null && bmi >= 40 ? "bg-rose-50/50 font-semibold" : ""}>
                        <td className="px-6 py-4 text-rose-600 flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-rose-600"></div> Obesity Class III</td>
                        <td className="px-6 py-4 text-gray-600">&ge; 40.0</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
