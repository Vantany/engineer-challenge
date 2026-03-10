import { useNavigate } from 'react-router-dom';
import { AuthLayout } from '../components/AuthLayout';
import { Button } from '../components/Button';

export const PasswordNotRecovered = () => {
  const navigate = useNavigate();

  return (
    <AuthLayout showRightPanel={false}>
      <div className="w-full max-w-[480px] mx-auto mt-20">
        <h1 className="text-3xl font-semibold text-gray-900 mb-4">Пароль не был восстановлен</h1>
        <p className="text-sm text-gray-500 mb-10">
          По каким-то причинам мы не смогли изменить ваш пароль.<br />
          Попробуйте еще раз через некоторое время.
        </p>
        
        <div className="flex flex-col gap-4">
          <Button onClick={() => navigate('/login')} variant="secondary">
            Назад в авторизацию
          </Button>
          <Button onClick={() => navigate('/forgot-password')} variant="text">
            Попробовать заново
          </Button>
        </div>
      </div>
    </AuthLayout>
  );
};
