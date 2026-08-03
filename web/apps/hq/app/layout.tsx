import type { Metadata } from 'next';
import './globals.css';
import { AxeCoreDevInit } from '@zippyra/ui';

export const metadata: Metadata = {
  title: 'Zippyra HQ',
  description: 'Enterprise operations, multi-store order monitoring, inventory alerts, and catalog management for retail chains.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="h-full bg-slate-950 text-slate-100">
      <head>
        <link rel="icon" href="/zippyra_icon.svg" type="image/svg+xml" />
      </head>
      <body className="h-full antialiased selection:bg-indigo-500 selection:text-white font-sans">
        <AxeCoreDevInit />
        {children}
      </body>
    </html>
  );
}
