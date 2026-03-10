import { useNavigate } from 'react-router-dom';
import { AuthLayout } from '../components/AuthLayout';
import { Button } from '../components/Button';

export const PasswordRecovered = () => {
  const navigate = useNavigate();

  return (
    <AuthLayout showRightPanel={false}>
      <div className="w-full max-w-[480px] mx-auto mt-20">
        <h1 className="text-3xl font-semibold text-gray-900 mb-4">Пароль был восстановлен</h1>
        <p className="text-sm text-gray-500 mb-10">
          Перейдите на страницу авторизации, чтобы войти в систему с новым паролем
        </p>
        
        <Button onClick={() => navigate('/login')} variant="secondary">
          Назад в авторизацию
        </Button>
      </div>
    </AuthLayout>
  );
};
