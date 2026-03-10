import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';

import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { ForgotPassword } from './pages/ForgotPassword';
import { CheckEmail } from './pages/CheckEmail';
import { SetPassword } from './pages/SetPassword';
import { PasswordRecovered } from './pages/PasswordRecovered';
import { PasswordNotRecovered } from './pages/PasswordNotRecovered';

const Dashboard = () => (
  <div className="min-h-screen flex items-center justify-center bg-gray-50">
    <div className="bg-white p-8 rounded-lg shadow-md">
      <h1 className="text-2xl font-bold mb-4">Добро пожаловать в систему!</h1>
      <button 
        onClick={() => window.location.href = '/login'}
        className="text-blue-500 hover:underline"
      >
        Выйти
      </button>
    </div>
  </div>
);

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Navigate to="/login" replace />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/forgot-password" element={<ForgotPassword />} />
        <Route path="/check-email" element={<CheckEmail />} />
        <Route path="/set-password" element={<SetPassword />} />
        <Route path="/password-recovered" element={<PasswordRecovered />} />
        <Route path="/password-not-recovered" element={<PasswordNotRecovered />} />
        <Route path="/dashboard" element={<Dashboard />} />
      </Routes>
    </Router>
  );
}

export default App;
