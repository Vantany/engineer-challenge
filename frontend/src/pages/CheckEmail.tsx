import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthLayout } from '../components/AuthLayout';
import { Button } from '../components/Button';
import { Input } from '../components/Input';

export const CheckEmail = () => {
  const navigate = useNavigate();
  const [token, setToken] = useState('');

  return (
    <AuthLayout showRightPanel={false}>
      <div className="w-full max-w-[480px] mx-auto mt-20">
        <h1 className="text-3xl font-semibold text-gray-900 mb-4">Проверьте свою почту</h1>
        <p className="text-sm text-gray-500 mb-10">
          Мы отправили на почту письмо с ссылкой для восстановления пароля
        </p>
        
        <div className="flex flex-col gap-4">
          <Button onClick={() => navigate('/login')} variant="secondary">
            Назад в авторизацию
          </Button>
        </div>

        {/* Блок для ручного ввода токена (т.к. письма реально не отправляются) */}
        <div className="mt-16 p-6 border border-dashed border-gray-200 rounded-lg bg-gray-50">
          <p className="text-xs text-gray-500 mb-4">
            * Для тестирования: скопируйте токен из логов бэкенда и вставьте сюда.
          </p>
          <Input 
            label="Токен восстановления"
            placeholder="Введите токен"
            value={token}
            onChange={(e) => setToken(e.target.value)}
          />
          <Button 
            variant="primary" 
            onClick={() => navigate(`/set-password?token=${token}`)}
            disabled={!token}
          >
            Перейти к смене пароля
          </Button>
        </div>
      </div>
    </AuthLayout>
  );
};
