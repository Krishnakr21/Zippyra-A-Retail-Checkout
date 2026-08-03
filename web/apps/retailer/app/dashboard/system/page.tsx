'use client';

import React from 'react';

export default function RetailerSystemPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">System Status</h1>
        <p className="text-sm text-gray-500 mt-1">Platform operational status and system health</p>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-xl p-6 text-blue-900 space-y-3 shadow-sm">
        <div className="flex items-center space-x-3">
          <span className="text-2xl">🛡️</span>
          <h2 className="text-lg font-semibold text-blue-950">Centralized Platform Operations</h2>
        </div>
        <p className="text-sm text-blue-800 leading-relaxed">
          System operational tasks (including Kafka dead-letter message replaying, platform feature flags, and payment gateway circuit breaker monitoring) are managed centrally by the Zippyra Admin Platform team.
        </p>
        <div className="pt-2 text-xs font-mono text-blue-700">
          Role-Gated Route: Managed via Admin Platform (/dashboard/system)
        </div>
      </div>
    </div>
  );
}
