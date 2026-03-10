import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'text';
  fullWidth?: boolean;
}

export const Button: React.FC<ButtonProps> = ({ 
  children, 
  variant = 'primary', 
  fullWidth = true,
  className = '',
  disabled,
  ...props 
}) => {
  const baseStyles = "py-3 px-4 rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2";
  const widthStyles = fullWidth ? "w-full" : "";
  
  const variants = {
    primary: "bg-[#3b82f6] hover:bg-blue-600 text-white focus:ring-blue-500 disabled:bg-blue-200 disabled:cursor-not-allowed",
    secondary: "bg-[#eff6ff] hover:bg-[#e0e7ff] text-[#3b82f6] focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-400 disabled:cursor-not-allowed",
    text: "bg-transparent hover:text-blue-700 text-[#3b82f6] focus:ring-blue-500 disabled:text-gray-400 disabled:cursor-not-allowed"
  };

  return (
    <button 
      className={`${baseStyles} ${widthStyles} ${variants[variant]} ${className}`}
      disabled={disabled}
      {...props}
    >
      {children}
    </button>
  );
};
