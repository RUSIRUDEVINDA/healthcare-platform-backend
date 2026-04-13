import { Outlet } from 'react-router-dom';

export default function AuthLayout() {
  return (
    <div className="flex min-h-screen bg-brand-light font-sans text-gray-800">
      {/* Left side: Branding/Illustrations (Hidden on small screens) */}
      <div className="hidden lg:flex flex-col flex-1 p-12 justify-center relative items-center">
        <div className="max-w-xl w-full">
          <h1 className="text-5xl font-semibold leading-tight text-brand-dark mb-4 drop-shadow-sm">
            Enabling exceptional<br />telehealth at every<br />touchpoint.
          </h1>
          <p className="text-lg text-emerald-800 opacity-90 mb-10 max-w-md">
            Innovative telehealth solutions proven to deliver seamless virtual care.
          </p>
          <div className="bg-white/60 backdrop-blur-md p-6 rounded-2xl shadow-sm border border-white/40 max-w-sm">
            <div className="flex items-center space-x-4">
              <div className="w-12 h-12 bg-brand rounded-full flex items-center justify-center text-white shrink-0">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path>
                </svg>
              </div>
              <div>
                <p className="font-semibold text-gray-900">1000+ Happy patients</p>
                <p className="text-sm text-gray-600 font-medium">Join our growing virtual clinic</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Right side: Authentication forms */}
      <div className="flex-1 flex flex-col justify-center px-4 sm:px-6 lg:p-12 lg:bg-transparent bg-brand-light">
        <div className="mx-auto w-full max-w-md bg-white p-8 sm:p-10 rounded-3xl shadow-xl shadow-brand-dark/5 border border-gray-100">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
