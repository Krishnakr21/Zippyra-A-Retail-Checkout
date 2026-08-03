import React from 'react';
import './globals.css';
import { AxeCoreDevInit } from '@zippyra/ui';

export const metadata = {
  title: 'Zippyra Admin',
  description: 'Internal administration dashboard for Zippyra platform operations',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="dark">
      <head>
        <link rel="icon" href="/zippyra_icon.svg" type="image/svg+xml" />
        <link
          href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="bg-slate-950 text-slate-100 font-sans antialiased">
        <AxeCoreDevInit />
        {children}
        <footer className="border-t border-slate-800 px-6 py-3 text-xs text-slate-400 flex justify-between items-center">
          <span>© 2026 Zippyra Retail Technologies Private Limited</span>
          <a href="https://status.zippyra.com" target="_blank" rel="noopener noreferrer" className="text-emerald-400 hover:underline font-semibold">
            🟢 System Status (status.zippyra.com)
          </a>
        </footer>
      </body>
    </html>
  );
}
