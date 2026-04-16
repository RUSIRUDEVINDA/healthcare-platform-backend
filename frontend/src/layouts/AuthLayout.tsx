import { Outlet, Link } from 'react-router-dom';
import { Sparkles, Users } from 'lucide-react';

export default function AuthLayout() {
  return (
    <div className="flex h-screen w-full overflow-hidden font-sans bg-white">

      {/* ── Left Panel: Mininalist & Modern ── */}
      <div className="hidden lg:flex flex-[1.4] bg-[#e6f7f5] items-center justify-center p-12 lg:p-24 relative overflow-hidden">
        
        {/* Subtle top decoration */}
        <div className="absolute top-0 left-0 w-full h-1 bg-brand/10" />

        <div className="relative z-10 w-full max-w-lg flex flex-col space-y-12">
          
          {/* Headline Section */}
          <div className="space-y-6">
            <h1 className="text-5xl xl:text-6xl font-extrabold tracking-tight text-[#0f766e] leading-[1.15]">
              Enabling exceptional <br />
              telehealth at every <br />
              touchpoint.
            </h1>
            
            <p className="text-lg text-[#0f766e]/70 font-medium max-w-sm">
              Innovative telehealth solutions proven to deliver seamless virtual care.
            </p>
          </div>

          {/* Feature Card: Happy Patients */}
          <div className="w-fit bg-white/60 backdrop-blur-md rounded-2xl p-6 flex items-center gap-5 border border-white/40 shadow-sm transition-all hover:shadow-md hover:translate-y-[-2px]">
             <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-brand text-white shadow-lg shadow-brand/20">
                <Users className="h-6 w-6" />
             </div>
             <div className="flex flex-col">
                <p className="text-base font-bold text-slate-800">1000+ Happy patients</p>
                <p className="text-sm text-slate-500 font-medium">Join our growing virtual clinic</p>
             </div>
          </div>

          {/* Logo at bottom (optional but good for brand) */}
          <div className="absolute bottom-12 left-12 lg:left-24 opacity-20">
             <Link to="/" className="flex items-center gap-2 grayscale brightness-50">
               <Sparkles className="h-5 w-5" />
               <span className="text-sm font-black uppercase tracking-widest">MediPulse</span>
             </Link>
          </div>
        </div>
      </div>

      {/* ── Right Panel: Login Form ── */}
      <div className="flex-1 flex flex-col justify-center items-center px-6 lg:px-12 bg-white relative">
        <div className="w-full max-w-sm">
          <div className="bg-white rounded-3xl p-4 sm:p-6">
             <Outlet />
          </div>
          
          <div className="mt-8 text-center">
             <p className="text-xs text-slate-400 font-medium tracking-wide">
               &copy; {new Date().getFullYear()} MediPulse Sri Lanka Edition
             </p>
          </div>
        </div>
      </div>

    </div>
  );
}
