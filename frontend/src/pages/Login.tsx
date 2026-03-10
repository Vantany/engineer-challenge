import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AuthLayout } from '../components/AuthLayout';
import { Input } from '../components/Input';
import { Button } from '../components/Button';

export const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    
    try {
      const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      });
      
      if (!response.ok) {
        throw new Error('Введены неверные данные');
      }
      
      const data = await response.json();
      console.log('Login success:', data);
      navigate('/dashboard');
    } catch (err: any) {
      console.error(err);
      setError(err.message || 'Ошибка входа');
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout>
      <div className="w-full max-w-[440px] mx-auto">
        <h1 className="text-3xl font-semibold text-gray-900 mb-10">Войти в систему</h1>
        
        <form onSubmit={handleSubmit}>
          <Input 
            label="E-mail"
            placeholder="Введите e-mail"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={error && !password ? error : undefined}
            required
          />
          
          <Input 
            label="Пароль"
            placeholder="Введите пароль"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={error}
            required
          />
          
          <div className="mt-8 mb-6">
            <Button type="submit" disabled={loading}>
              {loading ? 'Вход...' : 'Войти'}
            </Button>
          </div>
          
          <div className="text-center">
            <Link to="/forgot-password" className="text-sm text-blue-500 hover:text-blue-600 font-medium">
              Забыли пароль?
            </Link>
          </div>
        </form>
      </div>

      <div className="absolute bottom-8 left-0 right-0 text-center border-t border-gray-100 pt-6">
        <span className="text-sm text-gray-500">Еще не зарегистрированы? </span>
        <Link to="/register" className="text-sm text-blue-500 hover:text-blue-600 font-medium">
          Регистрация
        </Link>
      </div>
    </AuthLayout>
  );
};
