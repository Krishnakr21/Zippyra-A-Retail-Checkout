import React from 'react';
import './globals.css';
import { AxeCoreDevInit } from '@zippyra/ui';

export const metadata = {
  title: 'Zippyra Retailer',
  description: 'Store Manager and Operator Dashboard',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <link rel="icon" href="/zippyra_icon.svg" type="image/svg+xml" />
      </head>
      <body>
        <AxeCoreDevInit />
        {children}
        <footer style={{ borderTop: '1px solid #e2e8f0', padding: '12px 24px', fontSize: '12px', color: '#64748b', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>© 2026 Zippyra Retail Technologies Private Limited</span>
          <a href="https://status.zippyra.com" target="_blank" rel="noopener noreferrer" style={{ color: '#2563eb', textDecoration: 'none', fontWeight: 600 }}>
            🟢 System Status (status.zippyra.com)
          </a>
        </footer>
      </body>
    </html>
  );
}
