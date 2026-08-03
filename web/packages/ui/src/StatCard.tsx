import React from 'react';

export interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon?: React.ReactNode;
  isComingSoon?: boolean;
}

export const StatCard: React.FC<StatCardProps> = ({ title, value, subtitle, icon, isComingSoon }) => {
  return (
    <div className="p-6 bg-white rounded-lg border border-gray-200 shadow-sm flex items-center justify-between">
      <div>
        <p className="text-sm font-medium text-gray-500">{title}</p>
        {isComingSoon ? (
          <span className="inline-block mt-2 px-2.5 py-0.5 rounded text-xs font-semibold bg-amber-100 text-amber-800">
            Coming soon
          </span>
        ) : (
          <p className="text-3xl font-bold text-gray-900 mt-1">{value}</p>
        )}
        {subtitle && <p className="text-xs text-gray-400 mt-1">{subtitle}</p>}
      </div>
      {icon && <div className="p-3 bg-blue-50 text-blue-600 rounded-full">{icon}</div>}
    </div>
  );
};
