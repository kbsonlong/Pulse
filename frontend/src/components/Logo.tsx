import React from 'react';

interface LogoProps {
  variant?: 'full' | 'icon' | 'white';
  size?: 'sm' | 'md' | 'lg' | 'xl';
  animated?: boolean;
  className?: string;
}

const Logo: React.FC<LogoProps> = ({ 
  variant = 'full', 
  size = 'md', 
  animated = true, 
  className = '' 
}) => {
  const sizeClasses = {
    sm: 'w-6 h-6',
    md: 'w-8 h-8',
    lg: 'w-12 h-12',
    xl: 'w-16 h-16'
  };

  const fullSizeClasses = {
    sm: 'w-20 h-6',
    md: 'w-24 h-8',
    lg: 'w-32 h-12',
    xl: 'w-40 h-16'
  };

  const getLogoSrc = () => {
    switch (variant) {
      case 'icon':
        return '/logo-icon.svg';
      case 'white':
        return '/logo-white.svg';
      default:
        return '/logo.svg';
    }
  };

  const getSizeClass = () => {
    return variant === 'full' ? fullSizeClasses[size] : sizeClasses[size];
  };

  const animationClass = animated ? 'pulse-logo-animation' : '';

  return (
    <div className={`inline-flex items-center ${className}`}>
      <img 
        src={getLogoSrc()} 
        alt="Pulse Logo" 
        className={`${getSizeClass()} ${animationClass} transition-all duration-300 hover:scale-105`}
      />
      {variant === 'icon' && (
        <span className="ml-2 font-bold text-pulse-blue text-lg">
          Pulse
        </span>
      )}
    </div>
  );
};

// 纯图标组件，用于特殊场景
export const LogoIcon: React.FC<Omit<LogoProps, 'variant'>> = (props) => (
  <Logo {...props} variant="icon" />
);

// 白色版本组件，用于深色背景
export const LogoWhite: React.FC<Omit<LogoProps, 'variant'>> = (props) => (
  <Logo {...props} variant="white" />
);

// 脉搏动画组件
export const PulseBeat: React.FC<{ className?: string }> = ({ className = '' }) => (
  <div className={`inline-flex items-center space-x-1 ${className}`}>
    <div className="w-2 h-2 bg-pulse-red rounded-full animate-pulse"></div>
    <div className="w-1 h-1 bg-pulse-red rounded-full animate-pulse" style={{ animationDelay: '0.2s' }}></div>
    <div className="w-1 h-1 bg-pulse-red rounded-full animate-pulse" style={{ animationDelay: '0.4s' }}></div>
  </div>
);

export default Logo;