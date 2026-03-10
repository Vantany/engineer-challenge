import React from 'react';

interface AuthLayoutProps {
  children: React.ReactNode;
  showRightPanel?: boolean;
}

export const AuthLayout: React.FC<AuthLayoutProps> = ({ children, showRightPanel = true }) => {
  return (
    <div className="min-h-screen flex w-full bg-white">
      {/* Left Panel */}
      <div className={`flex flex-col w-full ${showRightPanel ? 'lg:w-[600px] xl:w-[720px] flex-shrink-0' : 'max-w-[600px] mx-auto'} relative z-10 bg-white`}>
        {/* Logo (Removed) */}
        <div className="p-8 h-[88px]">
          {/* Empty space to keep the layout consistent */}
        </div>

        {/* Content */}
        <div className="flex-1 flex flex-col justify-center px-8 sm:px-12 lg:px-16 pb-12">
          {children}
        </div>
      </div>

      {/* Right Panel */}
      {showRightPanel && (
        <div className="hidden lg:flex flex-1 bg-[#f0f4f8] relative overflow-hidden items-center justify-center border-l border-gray-100">
          {/* Decorative Elements */}
          <div className="relative w-[500px] h-[500px]">
            {/* Main large translucent sphere */}
            <div className="absolute inset-0 rounded-full bg-white/40 backdrop-blur-md shadow-[0_0_80px_rgba(255,255,255,0.8)] border border-white/50 z-10"></div>
            
            {/* Small blue sphere top right */}
            <div className="absolute top-4 right-8 w-20 h-20 rounded-full bg-blue-500 blur-[2px] opacity-80 z-0"></div>
            
            {/* Small blue sphere bottom right */}
            <div className="absolute bottom-8 right-12 w-24 h-24 rounded-full bg-blue-500 blur-[2px] opacity-80 z-20"></div>
            
            {/* Small blue sphere middle left */}
            <div className="absolute top-1/2 -left-8 w-16 h-16 rounded-full bg-blue-500 blur-[2px] opacity-80 z-0"></div>
          </div>
        </div>
      )}
    </div>
  );
};
