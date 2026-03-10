import React, { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { AuthLayout } from '../components/AuthLayout';
import { Input } from '../components/Input';
import { Button } from '../components/Button';

export const SetPassword = () => {
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  
  const token = searchParams.get('token') || 'dummy-token';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (password !== passwordConfirm) {
      setError('Пароли не совпадают');
      return;
    }
    
    setLoading(true);
    setError('');
    
    try {
      const response = await fetch('/api/v1/auth/password-reset/reset', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, newPassword: password })
      });
      
      if (!response.ok) {
        throw new Error('Не удалось изменить пароль');
      }
      
      navigate('/password-recovered');
    } catch (err: any) {
      console.error(err);
      navigate('/password-not-recovered');
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout showRightPanel={false}>
      <div className="w-full max-w-[480px] mx-auto mt-20">
        <h1 className="text-3xl font-semibold text-gray-900 mb-4">Задайте пароль</h1>
        <p className="text-sm text-gray-500 mb-10">
          Напишите новый пароль, который будете использовать для входа
        </p>
        
        <form onSubmit={handleSubmit}>
          <Input 
            label="Введите пароль"
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
          
          <div className="mt-8">
            <Button type="submit" disabled={loading || !token}>
              {loading ? 'Сохранение...' : 'Изменить пароль'}
            </Button>
            {!token && <p className="text-xs text-red-500 mt-2">Отсутствует токен восстановления</p>}
          </div>
        </form>
      </div>
    </AuthLayout>
  );
};
