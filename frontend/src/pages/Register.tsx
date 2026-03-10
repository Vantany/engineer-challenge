import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AuthLayout } from '../components/AuthLayout';
import { Input } from '../components/Input';
import { Button } from '../components/Button';

export const Register = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (password !== passwordConfirm) {
      setError('Пароли не совпадают');
      return;
    }
    
    setLoading(true);
    setError('');
    
    try {
      const response = await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      });
      
      if (!response.ok) {
        throw new Error('Ошибка регистрации');
      }
      
      const data = await response.json();
      console.log('Register success:', data);
      navigate('/login');
    } catch (err: any) {
      console.error(err);
      setError(err.message || 'Ошибка регистрации');
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout>
      <div className="w-full max-w-[440px] mx-auto">
        <h1 className="text-3xl font-semibold text-gray-900 mb-10">Регистрация в системе</h1>
        
        <form onSubmit={handleSubmit}>
          <Input 
            label="E-mail"
            placeholder="Введите e-mail"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          
          <Input 
            label="Пароль"
            placeholder="Введите пароль"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          <Input 
            label="Повторите пароль"
            placeholder="Повторите пароль"
            type="password"
            value={passwordConfirm}
            onChange={(e) => setPasswordConfirm(e.target.value)}
            error={error}
            required
          />
          
          <div className="mt-8 mb-4">
            <Button type="submit" disabled={loading}>
              {loading ? 'Регистрация...' : 'Зарегистрироваться'}
            </Button>
          </div>
          
          <p className="text-xs text-gray-400 text-center leading-tight">
            Зарегистрировавшись пользователь принимает условия{' '}
            <a href="#" className="text-gray-400 underline decoration-gray-300 underline-offset-2">договора оферты</a>
            {' '}и{' '}
            <a href="#" className="text-gray-400 underline decoration-gray-300 underline-offset-2">политики конфиденциальности</a>
          </p>
        </form>
      </div>

      <div className="absolute bottom-8 left-0 right-0 text-center border-t border-gray-100 pt-6">
        <span className="text-sm text-gray-500">Уже есть аккаунт? </span>
        <Link to="/login" className="text-sm text-blue-500 hover:text-blue-600 font-medium">
          Войти
        </Link>
      </div>
    </AuthLayout>
  );
};
