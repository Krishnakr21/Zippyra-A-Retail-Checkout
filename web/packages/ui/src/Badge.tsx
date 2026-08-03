import React from 'react';

export interface BadgeProps {
  status: string;
}

export const Badge: React.FC<BadgeProps> = ({ status }) => {
  const s = status.toUpperCase();

  let colorClasses = 'bg-gray-100 text-gray-900 border border-gray-300';
  if (['COMPLETED', 'RECEIVED', 'PASSED', 'ACTIVE', 'APPROVED'].includes(s)) {
    colorClasses = 'bg-emerald-100 text-emerald-950 border border-emerald-300 font-bold';
  } else if (['DRAFT', 'REQUESTED', 'QC_PENDING', 'SUBMITTED', 'PENDING'].includes(s)) {
    colorClasses = 'bg-amber-100 text-amber-950 border border-amber-300 font-bold';
  } else if (['IN_TRANSIT', 'PARTIALLY_RECEIVED', 'RETURN_REQUESTED'].includes(s)) {
    colorClasses = 'bg-blue-100 text-blue-950 border border-blue-300 font-bold';
  } else if (['REJECTED', 'CANCELLED', 'RETURNED', 'INACTIVE', 'FAILED', 'EXPIRED'].includes(s)) {
    colorClasses = 'bg-rose-100 text-rose-950 border border-rose-300 font-bold';
  }

  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold ${colorClasses}`}>
      {status}
    </span>
  );
};
