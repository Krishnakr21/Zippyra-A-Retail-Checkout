import React from 'react';

export interface ComingSoonBannerProps {
  serviceName: string;
  description: string;
}

export const ComingSoonBanner: React.FC<ComingSoonBannerProps> = ({ serviceName, description }) => {
  return (
    <div className="p-8 bg-amber-50 border border-amber-200 rounded-xl shadow-sm text-center my-6">
      <div className="inline-flex p-3 bg-amber-100 text-amber-800 rounded-full mb-3 font-semibold text-xs tracking-wide uppercase">
        Coming Soon
      </div>
      <h2 className="text-xl font-bold text-gray-900">Backend Dependency Pending: {serviceName}</h2>
      <p className="text-sm text-gray-600 mt-2 max-w-xl mx-auto">{description}</p>
    </div>
  );
};
