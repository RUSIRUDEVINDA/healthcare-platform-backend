import { Navigate, Route, Routes } from 'react-router-dom';
import AuthLayout from './layouts/AuthLayout';
import Login from './pages/Login';
import Register from './pages/Register.tsx';
import Dashboard from './pages/Dashboard';
import Profile from './pages/Profile';
import Appointments from './pages/Appointments';
import Payments from './pages/Payments';
import PaymentStatus from './pages/PaymentStatus';
import MedicalRecords from './pages/MedicalRecords';
import AppShellLayout from './layouts/AppShellLayout';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/auth/login" replace />} />
      <Route path="/auth" element={<AuthLayout />}>
        <Route path="login" element={<Login />} />
        <Route path="register" element={<Register />} />
      </Route>
      <Route element={<AppShellLayout />}>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/appointments" element={<Appointments />} />
        <Route path="/payments" element={<Payments />} />
        <Route path="/records" element={<MedicalRecords />} />
      </Route>
      
      {/* Payment redirection targets */}
      <Route path="/payment/success" element={<PaymentStatus type="success" />} />
      <Route path="/payment/cancel" element={<PaymentStatus type="cancel" />} />
      
      {/* Fallback route */}
      <Route path="*" element={<Navigate to="/auth/login" replace />} />
    </Routes>
  );
}

export default App;
