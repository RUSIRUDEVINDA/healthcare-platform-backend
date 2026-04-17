import { Navigate, Route, Routes } from 'react-router-dom';
import Landing from './pages/Landing';
import AuthLayout from './layouts/AuthLayout';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import Profile from './pages/Profile';
import Appointments from './pages/Appointments';
import Payments from './pages/Payments';
import PaymentStatus from './pages/PaymentStatus';
import Telemedicine from './pages/Telemedicine';
import Records from './pages/Records';
import Reactivate from './pages/Reactivate';
import BmiCalculator from './pages/BmiCalculator';
import SymptomChecker from './pages/SymptomChecker';

import AppShellLayout from './layouts/AppShellLayout';
import AdminConsole from './pages/AdminConsole';
import NotificationMonitor from './pages/NotificationMonitor';

import { Toaster } from 'react-hot-toast';
import MedicalRecords from './pages/MedicalRecords';

function App() {
  return (
    <>
      <Toaster position="top-right" reverseOrder={false} />
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/auth" element={<AuthLayout />}>
          <Route index element={<Navigate to="login" replace />} />
          <Route path="login" element={<Login />} />
          <Route path="register" element={<Register />} />
          <Route path="reactivate" element={<Reactivate />} />
        </Route>

        <Route element={<AppShellLayout />}>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/appointments" element={<Appointments />} />
          <Route path="/payments" element={<Payments />} />
          <Route path="/records" element={<Records />} />
          <Route path="/bmi-calculator" element={<BmiCalculator />} />
          <Route path="/symptom-checker" element={<SymptomChecker />} />
          <Route path="/telemedicine" element={<Telemedicine />} />
          <Route path="/admin" element={<AdminConsole />} />
          <Route path="/notifications" element={<NotificationMonitor />} />
        </Route>

        {/* Payment redirection targets */}
        <Route path="/payment/success" element={<PaymentStatus type="success" />} />
        <Route path="/payment/cancel" element={<PaymentStatus type="cancel" />} />

        {/* Fallback route */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </>
  );
}

export default App;
