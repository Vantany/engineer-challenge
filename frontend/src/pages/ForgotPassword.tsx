import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AuthLayout } from '../components/AuthLayout';
import { Input } from '../components/Input';
import { Button } from '../components/Button';

export const ForgotPassword = () => {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await fetch('/api/v1/auth/password-reset/request', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email })
      });
      
      if (!response.ok) {
        throw new Error('Ошибка при запросе восстановления пароля');
      }
      
      navigate('/check-email');
    } catch (err: any) {
      console.error(err);
      setError(err.message || 'Произошла ошибка');
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout showRightPanel={false}>
      <div className="w-full max-w-[480px] mx-auto mt-20">
        <div className="flex items-center gap-4 mb-4">
          <Link to="/login" className="text-gray-900 hover:text-gray-600 transition-colors">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M15.41 16.59L10.83 12L15.41 7.41L14 6L8 12L14 18L15.41 16.59Z" fill="currentColor"/>
            </svg>
          </Link>
          <h1 className="text-3xl font-semibold text-gray-900">Восстановление пароля</h1>
        </div>
        
        <p className="text-sm text-gray-500 mb-10 pl-10">
          Укажите адрес почты на который был зарегистрирован аккаунт
        </p>
        
        <div className="pl-10">
          <form onSubmit={handleSubmit}>
            <Input 
              label="E-mail"
              placeholder="Введите e-mail"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              error={error}
              required
            />
            
            <div className="mt-8">
              <Button 
                type="submit" 
                disabled={loading || !email}
                className={!email ? 'bg-blue-100 text-blue-400 hover:bg-blue-100' : ''}
              >
                {loading ? 'Отправка...' : 'Восстановить пароль'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </AuthLayout>
  );
};
