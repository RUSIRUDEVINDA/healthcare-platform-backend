import { Outlet } from 'react-router-dom';

export default function AuthLayout() {
  return (
    <div className="flex min-h-screen bg-[#f2fbfa] font-sans text-gray-800">
      <div className="hidden lg:flex flex-1 items-center justify-center p-12">
        <div className="max-w-xl">
          <div className="inline-flex items-center gap-2 rounded-full border border-brand/10 bg-brand/10 px-3 py-1 text-xs text-brand">
            MediPulse SriLanka
          </div>
          <h1 className="mt-5 text-5xl leading-tight text-slate-900">
            Care, records, and appointments in one place.
          </h1>
          <p className="mt-4 max-w-lg text-base leading-7 text-slate-600">
            Log in to manage bookings, view medical records, and use the AI symptom checker in a calm, modern interface.
          </p>
          <div className="mt-8 grid gap-3 sm:grid-cols-2">
            {[
              'Patient and doctor workflow',
              'Medical records and prescriptions',
              'Appointment booking',
              'AI symptom guidance',
            ].map((item) => (
              <div key={item} className="rounded-[1.25rem] border border-slate-200 bg-white px-4 py-3 text-sm text-slate-600">
                {item}
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="flex-1 flex flex-col justify-center px-4 sm:px-6 lg:p-12 bg-white">
        <div className="mx-auto w-full max-w-md rounded-[1.5rem] border border-slate-200 bg-white p-8 shadow-sm">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
