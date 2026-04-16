import React, { useState } from 'react';
import { Link} from 'react-router-dom';
import { Mail, Lock, LogIn, AlertCircle, ArrowRight } from 'lucide-react';
import axios from 'axios';
import apiClient from '../api/client';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const response = await apiClient.post('/auth/login', { email, password });
      
      // The backend returns { "success": true, "data": { "access_token": "..." } }
      const token = response.data?.data?.access_token;
      const user = response.data?.data?.user;
      
      if (token) {
        localStorage.setItem('access_token', token);
        if (user) {
          localStorage.setItem('user', JSON.stringify(user));
        }
        if (user?.role === 'admin') {
          window.location.href = '/admin';
        } else {
          window.location.href = '/dashboard';
        }
      } else {
        setError('Login failed: Invalid response from server');
      }
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        setError(err.response?.data?.error || 'Invalid credentials or server error.');
      } else {
        setError('An unexpected error occurred.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full">
      <div className="text-center mb-6">
        <h2 className="text-2xl font-semibold text-slate-900 mb-1 tracking-tight">Welcome back</h2>
        <p className="text-slate-500 text-sm">Please enter your details to sign in.</p>
      </div>

      {error && (
        <div className={`mb-6 p-5 rounded-2xl border transition-all animate-in fade-in slide-in-from-top-2 ${
          error.includes('deactivated') 
            ? 'bg-amber-50 border-amber-100' 
            : 'bg-red-50 border-red-100 text-red-700'
        }`}>
          {error.includes('deactivated') ? (
            <div className="flex gap-4">
              <div className="shrink-0 w-10 h-10 bg-amber-100 rounded-xl flex items-center justify-center text-amber-600">
                <AlertCircle className="h-6 w-6" />
              </div>
              <div className="flex flex-col">
                <span className="text-amber-900 font-bold text-base leading-tight">Account Suspended</span>
                <span className="text-amber-700 text-sm mt-1 font-medium leading-relaxed">
                  Your profile has been deactivated. If you believe this is an error or wish to return, please submit a request.
                </span>
                <Link 
                  to="/auth/reactivate" 
                  className="mt-3 inline-flex items-center text-amber-900 font-bold text-sm bg-white px-4 py-2 rounded-lg border border-amber-200 shadow-sm hover:bg-amber-100 transition-all w-fit"
                >
                  Request account reactivation
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Link>
              </div>
            </div>
          ) : (
            <span className="font-medium">{error}</span>
          )}
        </div>
      )}

      <form onSubmit={handleLogin} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1" htmlFor="email">Email address</label>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Mail className="h-5 w-5 text-gray-400" />
            </div>
            <input
              id="email"
              type="email"
              required
              className="block w-full pl-10 px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-shadow sm:text-sm bg-gray-50 hover:bg-white"
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1" htmlFor="password">Password</label>
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <Lock className="h-5 w-5 text-gray-400" />
            </div>
            <input
              id="password"
              type="password"
              required
              className="block w-full pl-10 px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-brand focus:border-transparent transition-shadow sm:text-sm bg-gray-50 hover:bg-white"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          <div className="flex items-center justify-end mt-2">
            <a href="#" className="font-medium text-sm text-brand hover:text-brand-dark transition-colors">
              Forgot your password?
            </a>
          </div>
        </div>

        <button
          type="submit"
          disabled={isLoading}
          className="w-full flex justify-center items-center py-3 px-4 border border-transparent rounded-xl shadow-sm text-sm font-semibold text-white bg-brand hover:bg-brand-dark focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-brand disabled:opacity-70 disabled:cursor-not-allowed transition-all"
        >
          {isLoading ? (
            <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
          ) : (
            <>
              Sign In
              <LogIn className="ml-2 h-4 w-4" />
            </>
          )}
        </button>
      </form>

      <div className="mt-5 text-center text-sm">
        <span className="text-gray-500 font-medium">Don't have an account? </span>
        <Link to="/auth/register" className="font-semibold text-brand hover:text-brand-dark transition-colors">
          Create an account
        </Link>
      </div>
    </div>
  );
}
