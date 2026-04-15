import { Link } from 'react-router-dom';
import {
  ArrowRight,
  CalendarCheck,
  MessageCircleHeart,
  ShieldCheck,
  Sparkles,
  Stethoscope,
  FileText,
  HeartPulse,
  CheckCircle2,
  Zap,
  Activity,
  UserCheck,
} from 'lucide-react';

const services = [
  {
    title: 'Smart Appointments',
    text: 'Seamlessly book visits and manage doctor slots with real-time availability updates.',
    icon: CalendarCheck,
    color: 'text-emerald-500',
    bgColor: 'bg-emerald-50',
  },
  {
    title: 'Integrated Records',
    text: 'A unified digital health record system for prescriptions, reports, and history.',
    icon: FileText,
    color: 'text-blue-500',
    bgColor: 'bg-blue-50',
  },
  {
    title: 'AI Symptom Checker',
    text: 'Intelligent guidance that helps patients understand their symptoms and next steps.',
    icon: MessageCircleHeart,
    color: 'text-rose-500',
    bgColor: 'bg-rose-50',
  },
  {
    title: 'Telemedicine Hub',
    text: 'Direct video consultations and remote care management from any device.',
    icon: Stethoscope,
    color: 'text-amber-500',
    bgColor: 'bg-amber-50',
  },
];

const whyChooseUs = [
  {
    title: 'Efficient Workflows',
    text: 'Optimized for the specific needs of Sri Lankan healthcare providers and patients.',
    icon: Zap,
  },
  {
    title: 'Modern Experience',
    text: 'Built with a focus on simplicity, speed, and intuitive user interactions.',
    icon: Activity,
  },
  {
    title: 'Secure & Private',
    text: 'Enterprise-grade security ensuring all patient data remains confidential and protected.',
    icon: ShieldCheck,
  },
  {
    title: 'Patient-Centric',
    text: 'Empowering patients with easy access to their care journey and medical data.',
    icon: UserCheck,
  },
];

const steps = [
  {
    title: 'Join the Platform',
    desc: 'Create your secure account in minutes and set up your healthcare profile.',
  },
  {
    title: 'Access Care Services',
    desc: 'Book appointments, use the AI checker, or consult with doctors online.',
  },
  {
    title: 'Manage Your Journey',
    desc: 'Keep all your medical history, reports, and prescriptions in one secure place.',
  },
];

const stats = [
  { label: 'Care Delivery', value: 'Patient-First' },
  { label: 'Infrastructure', value: 'Secure & Scalable' },
  { label: 'Innovation', value: 'AI Integrated' },
];

export default function Landing() {
  return (
    <div className="min-h-screen bg-white text-slate-900 font-sans selection:bg-brand/20">
      {/* Decorative Background Elements */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
        <div className="absolute -top-[10%] -left-[10%] w-[40%] h-[40%] bg-brand/5 blur-[120px] rounded-full" />
        <div className="absolute top-[20%] -right-[5%] w-[30%] h-[30%] bg-brand/10 blur-[100px] rounded-full" />
        <img 
          src="/assets/bg-pattern.png" 
          alt="" 
          className="absolute bottom-0 right-0 w-1/2 opacity-[0.03] grayscale pointer-events-none"
        />
      </div>

      <header className="fixed top-0 left-0 right-0 z-50 border-b border-slate-200/50 bg-white/70 backdrop-blur-xl">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8">
          <Link to="/" className="flex items-center gap-3 group">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-brand to-brand-dark text-white shadow-lg shadow-brand/20 transition-transform group-hover:scale-105">
              <Sparkles className="h-5 w-5" />
            </div>
            <div>
              <p className="text-sm font-bold uppercase tracking-[0.2em] text-slate-900">MediPulse</p>
              <p className="text-[10px] font-medium text-slate-500">Sri Lanka Edition</p>
            </div>
          </Link>

          <nav className="hidden md:flex items-center gap-8 text-sm font-medium text-slate-600">
            <a href="#services" className="hover:text-brand transition-colors">Services</a>
            <a href="#features" className="hover:text-brand transition-colors">Features</a>
            <a href="#process" className="hover:text-brand transition-colors">Workflow</a>
          </nav>

          <div className="flex items-center gap-3">
            <Link
              to="/auth/login"
              className="px-4 py-2 text-sm font-medium text-slate-700 transition-colors hover:text-brand"
            >
              Log in
            </Link>
            <Link
              to="/auth/register"
              className="inline-flex items-center gap-2 rounded-full bg-brand px-5 py-2.5 text-sm font-semibold text-white shadow-lg shadow-brand/25 transition-all hover:bg-brand-dark hover:shadow-brand/40 active:scale-95"
            >
              Get started
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </header>

      <main className="relative z-10 pt-24">
        {/* Hero Section */}
        <section className="mx-auto grid max-w-7xl gap-16 px-4 py-16 sm:px-6 lg:grid-cols-2 lg:px-8 lg:py-24 items-center">
          <div className="space-y-8">
            <div className="inline-flex items-center gap-2 rounded-full border border-brand/20 bg-brand/5 px-4 py-1.5 text-xs font-semibold text-brand animate-in fade-in slide-in-from-bottom-4 duration-1000">
              <CheckCircle2 className="h-3.5 w-3.5" />
              <span>Future of Sri Lankan Healthcare</span>
            </div>

            <div className="space-y-6">
              <h1 className="text-5xl font-extrabold tracking-tight text-slate-900 sm:text-6xl lg:text-7xl leading-[1.1]">
                Modern care, <br />
                <span className="bg-gradient-to-r from-brand to-brand-dark bg-clip-text text-transparent">unified.</span>
              </h1>
              <p className="max-w-xl text-lg leading-relaxed text-slate-600 sm:text-xl">
                MediPulse SriLanka bridges the gap between patients and quality care. 
                Manage appointments, digital health records, and AI-guided symptom checks in one intuitive platform.
              </p>
            </div>

            <div className="flex flex-wrap gap-4">
              <Link
                to="/auth/register"
                className="inline-flex items-center gap-2 rounded-2xl bg-slate-900 px-8 py-4 text-base font-bold text-white shadow-2xl shadow-slate-900/20 transition-all hover:bg-slate-800 hover:-translate-y-0.5"
              >
                Join the platform
                <ArrowRight className="h-5 w-5" />
              </Link>
              <Link
                to="/auth/login"
                className="inline-flex items-center gap-2 rounded-2xl border border-slate-200 bg-white px-8 py-4 text-base font-bold text-slate-700 transition-all hover:border-brand/30 hover:text-brand hover:bg-slate-50"
              >
                Open Dashboard
              </Link>
            </div>

            <div className="grid grid-cols-3 gap-6 pt-4 border-t border-slate-100">
              {stats.map((item) => (
                <div key={item.label}>
                  <p className="text-[10px] font-bold uppercase tracking-widest text-slate-400">{item.label}</p>
                  <p className="mt-1 text-sm font-semibold text-slate-700">{item.value}</p>
                </div>
              ))}
            </div>
          </div>

          <div className="relative lg:ml-auto">
            <div className="absolute -inset-4 bg-gradient-to-tr from-brand/20 to-transparent blur-3xl opacity-50 rounded-[3rem]" />
            <div className="relative rounded-[2.5rem] border border-slate-200/60 bg-white p-3 shadow-2xl">
              <div className="overflow-hidden rounded-[2rem]">
                <img 
                  src="/assets/hero.png" 
                  alt="MediPulse Dashboard Mockup" 
                  className="w-full h-auto transition-transform duration-700 hover:scale-105"
                />
              </div>
              
              {/* Floating Element 1 */}
              <div className="absolute -bottom-6 -left-6 hidden sm:flex items-center gap-4 rounded-2xl border border-slate-200 bg-white/80 backdrop-blur-md p-4 shadow-xl animate-bounce-slow">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-brand/10 text-brand">
                  <CalendarCheck className="h-5 w-5" />
                </div>
                <div>
                  <p className="text-xs font-bold text-slate-900">Next Appointment</p>
                  <p className="text-[10px] text-slate-500">Today at 2:00 PM</p>
                </div>
              </div>

              {/* Floating Element 2 */}
              <div className="absolute top-10 -right-10 hidden sm:flex items-center gap-4 rounded-2xl border border-white/40 bg-brand/90 backdrop-blur-md p-4 shadow-xl text-white">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-white/20">
                  <HeartPulse className="h-5 w-5" />
                </div>
                <div>
                  <p className="text-xs font-bold">Vitals Tracked</p>
                  <p className="text-[10px] opacity-80">Syncing live...</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Services Section */}
        <section id="services" className="mx-auto max-w-7xl px-4 py-24 sm:px-6 lg:px-8">
          <div className="text-center max-w-3xl mx-auto space-y-4 mb-16">
            <h2 className="text-sm font-bold uppercase tracking-widest text-brand">The Platform</h2>
            <h3 className="text-3xl font-extrabold text-slate-900 sm:text-4xl">Everything you need for smart healthcare</h3>
            <p className="text-slate-600">Built to handle the complexities of medical workflows with a simple, modern interface.</p>
          </div>

          <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-4">
            {services.map((item) => {
              const Icon = item.icon;
              return (
                <div key={item.title} className="group relative rounded-3xl border border-slate-100 bg-white p-8 transition-all hover:border-brand/20 hover:shadow-2xl hover:shadow-brand/5 hover:-translate-y-1">
                  <div className={`flex h-12 w-12 items-center justify-center rounded-2xl ${item.bgColor} ${item.color} mb-6 transition-transform group-hover:scale-110`}>
                    <Icon className="h-6 w-6" />
                  </div>
                  <h4 className="text-xl font-bold text-slate-900 mb-3">{item.title}</h4>
                  <p className="text-sm leading-relaxed text-slate-500">{item.text}</p>
                </div>
              );
            })}
          </div>
        </section>

        {/* Features Section */}
        <section id="features" className="bg-slate-50/50 py-24">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid gap-12 lg:grid-cols-2 items-center">
              <div className="space-y-8">
                <div>
                  <h2 className="text-sm font-bold uppercase tracking-widest text-brand">Why MediPulse</h2>
                  <h3 className="mt-4 text-3xl font-extrabold text-slate-900 sm:text-4xl">Designed for the modern patient flow</h3>
                </div>
                
                <div className="grid gap-6">
                  {whyChooseUs.map((item) => {
                    const Icon = item.icon;
                    return (
                      <div key={item.title} className="flex gap-5">
                        <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-white shadow-sm border border-slate-100 text-brand">
                          <Icon className="h-5 w-5" />
                        </div>
                        <div>
                          <h4 className="text-base font-bold text-slate-900">{item.title}</h4>
                          <p className="mt-1 text-sm text-slate-500">{item.text}</p>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              <div className="relative">
                <div className="absolute -inset-10 bg-brand/5 blur-[80px] rounded-full" />
                <div className="relative rounded-[2.5rem] bg-slate-900 p-8 shadow-2xl text-white">
                  <div className="flex items-center justify-between mb-8">
                    <div className="flex items-center gap-3">
                      <div className="h-3 w-3 rounded-full bg-red-500" />
                      <div className="h-3 w-3 rounded-full bg-yellow-500" />
                      <div className="h-3 w-3 rounded-full bg-green-500" />
                    </div>
                    <div className="px-3 py-1 rounded-full bg-white/10 text-[10px] font-mono tracking-tighter">SECURE CHANNEL</div>
                  </div>
                  
                  <div className="space-y-6">
                    <div className="p-4 rounded-2xl bg-white/5 border border-white/10">
                      <div className="h-2 w-24 bg-brand/40 rounded-full mb-3" />
                      <div className="h-2 w-full bg-white/10 rounded-full" />
                    </div>
                    <div className="p-4 rounded-2xl bg-brand/20 border border-brand/30 translate-x-4">
                      <div className="flex items-center gap-3 mb-3">
                        <div className="h-6 w-6 rounded-full bg-brand" />
                        <div className="h-2 w-16 bg-white/40 rounded-full" />
                      </div>
                      <div className="h-2 w-3/4 bg-white/30 rounded-full" />
                    </div>
                    <div className="p-4 rounded-2xl bg-white/5 border border-white/10">
                      <div className="h-2 w-32 bg-white/20 rounded-full mb-3" />
                      <div className="h-2 w-1/2 bg-white/10 rounded-full" />
                    </div>
                  </div>
                  
                  <div className="mt-8 flex justify-center">
                    <div className="px-6 py-2 rounded-xl bg-brand font-bold text-xs">READY FOR APPOINTMENT</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Process Section */}
        <section id="process" className="mx-auto max-w-7xl px-4 py-24 sm:px-6 lg:px-8">
          <div className="rounded-[3rem] bg-gradient-to-br from-slate-900 to-slate-800 p-8 sm:p-16 text-white overflow-hidden relative">
            <div className="absolute top-0 right-0 w-1/2 h-full bg-brand/10 blur-[100px] rounded-full -translate-y-1/2 translate-x-1/2" />
            
            <div className="relative z-10 grid gap-16 lg:grid-cols-2 items-center">
              <div>
                <h2 className="text-sm font-bold uppercase tracking-widest text-brand">The Journey</h2>
                <h3 className="mt-4 text-3xl font-extrabold sm:text-5xl leading-tight">Start your health journey in three simple steps</h3>
                <p className="mt-6 text-slate-400 text-lg leading-relaxed">
                  We've simplified healthcare management so you can focus on what matters most—your well-being. 
                  Experience the MediPulse way today.
                </p>
                
                <div className="mt-10 flex flex-wrap gap-4">
                  <Link
                    to="/auth/register"
                    className="inline-flex items-center gap-2 rounded-2xl bg-brand px-8 py-4 text-base font-bold text-white transition-all hover:bg-brand-dark hover:shadow-lg hover:shadow-brand/20"
                  >
                    Get started now
                    <ArrowRight className="h-5 w-5" />
                  </Link>
                </div>
              </div>

              <div className="space-y-6">
                {steps.map((step, index) => (
                  <div key={step.title} className="flex gap-6 items-start group">
                    <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-white/10 border border-white/10 text-brand text-xl font-bold transition-all group-hover:bg-brand group-hover:text-white">
                      {index + 1}
                    </div>
                    <div>
                      <h4 className="text-xl font-bold text-white">{step.title}</h4>
                      <p className="mt-2 text-slate-400 leading-relaxed">{step.desc}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* Footer */}
        <footer className="border-t border-slate-100 py-12">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="flex flex-col md:flex-row justify-between items-center gap-8">
              <div className="flex items-center gap-3">
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-brand text-white">
                  <Sparkles className="h-4 w-4" />
                </div>
                <p className="text-sm font-bold uppercase tracking-[0.2em] text-slate-900 underline decoration-brand/30 decoration-2 underline-offset-4">MediPulse</p>
              </div>
              
              <p className="text-sm text-slate-500">
                &copy; {new Date().getFullYear()} MediPulse Sri Lanka. Modern healthcare for a modern nation.
              </p>

              <div className="flex items-center gap-6 text-sm font-medium text-slate-500">
                <a href="#" className="hover:text-brand transition-colors">Privacy</a>
                <a href="#" className="hover:text-brand transition-colors">Terms</a>
                <a href="#" className="hover:text-brand transition-colors">Contact</a>
              </div>
            </div>
          </div>
        </footer>
      </main>

      <style>{`
        @keyframes bounce-slow {
          0%, 100% { transform: translateY(-5%); animation-timing-function: cubic-bezier(0.8, 0, 1, 1); }
          50% { transform: translateY(0); animation-timing-function: cubic-bezier(0, 0, 0.2, 1); }
        }
        .animate-bounce-slow {
          animation: bounce-slow 4s infinite;
        }
      `}</style>
    </div>
  );
}
