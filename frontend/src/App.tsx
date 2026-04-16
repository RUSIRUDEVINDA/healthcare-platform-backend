import { Navigate, Route, Routes } from 'react-router-dom';
import Landing from './pages/Landing';
import AuthLayout from './layouts/AuthLayout';
import Login from './pages/Login';
import Register from './pages/Register.tsx';
import Dashboard from './pages/Dashboard';
import Profile from './pages/Profile';
import Appointments from './pages/Appointments';
import Payments from './pages/Payments';
import PaymentStatus from './pages/PaymentStatus';
import Telemedicine from './pages/Telemedicine';
import Records from './pages/Records';
import BmiCalculator from './pages/BmiCalculator';
import SymptomChecker from './pages/SymptomChecker';
import AppShellLayout from './layouts/AppShellLayout';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/auth" element={<AuthLayout />}>
        <Route index element={<Navigate to="login" replace />} />
        <Route path="login" element={<Login />} />
        <Route path="register" element={<Register />} />
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
      </Route>

      <Route path="/payment/success" element={<PaymentStatus type="success" />} />
      <Route path="/payment/cancel" element={<PaymentStatus type="cancel" />} />

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default App;
