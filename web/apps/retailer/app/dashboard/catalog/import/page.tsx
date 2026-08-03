'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';

export default function BulkImportPage() {
  const router = useRouter();
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [jobId, setJobId] = useState<string | null>(null);
  const [status, setStatus] = useState<string>('');
  const [progress, setProgress] = useState<number>(0);
  const [failedRows, setFailedRows] = useState<{ row: number; error: string }[]>([]);

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return;
    setUploading(true);

    const formData = new FormData();
    formData.append('file', file);
    formData.append('store_id', 'store-001');

    try {
      const res = await fetch('/api/catalog/import', {
        method: 'POST',
        body: formData,
      });
      const data = await res.json();
      if (data.job_id) {
        setJobId(data.job_id);
        startPolling(data.job_id);
      }
    } catch (err: any) {
      alert(err.message || 'CSV Import upload failed');
      setUploading(false);
    }
  };

  const startPolling = (id: string) => {
    const interval = setInterval(async () => {
      try {
        const res = await fetch(`/api/catalog/import/${id}`);
        const data = await res.json();
        setStatus(data.status);
        setProgress(data.progress || 50);

        if (data.status === 'COMPLETED' || data.status === 'FAILED') {
          clearInterval(interval);
          setUploading(false);
          if (data.failed_rows) {
            setFailedRows(data.failed_rows);
          }
        }
      } catch (_) {
        clearInterval(interval);
        setUploading(false);
      }
    }, 2000);
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <button onClick={() => router.back()} className="text-xs text-blue-600 hover:underline mb-1">
          ← Back to Catalog
        </button>
        <h1 className="text-2xl font-bold text-gray-900">Bulk CSV Product Import</h1>
        <p className="text-sm text-gray-500 mt-1">Upload a CSV file containing barcode, name, price, mrp, and hsn_code headers</p>
      </div>

      {!jobId ? (
        <form onSubmit={handleUpload} className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm space-y-4">
          <div>
            <label className="block text-xs font-semibold text-gray-700 mb-1">Select CSV File</label>
            <input
              type="file"
              accept=".csv"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
              className="w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-xs file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
              required
            />
          </div>

          <div className="pt-4 flex justify-end">
            <button
              type="submit"
              disabled={uploading || !file}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-md disabled:opacity-50"
            >
              {uploading ? 'Uploading...' : 'Start Import'}
            </button>
          </div>
        </form>
      ) : (
        <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm space-y-4">
          <div className="flex justify-between items-center text-sm font-semibold">
            <span>Import Job #{jobId.substring(0, 8)}</span>
            <span className="px-2 py-0.5 rounded bg-blue-100 text-blue-800 text-xs">{status || 'PROCESSING'}</span>
          </div>

          <div className="w-full bg-gray-200 h-2.5 rounded-full overflow-hidden">
            <div className="bg-blue-600 h-2.5 rounded-full transition-all duration-300" style={{ width: `${progress}%` }} />
          </div>

          {failedRows.length > 0 && (
            <div className="mt-6 space-y-2">
              <h3 className="font-bold text-red-600 text-sm">Failed Rows ({failedRows.length})</h3>
              <div className="overflow-x-auto border rounded-lg">
                <table className="min-w-full divide-y text-xs">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-2 text-left font-semibold">Row #</th>
                      <th className="px-4 py-2 text-left font-semibold">Error Reason</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {failedRows.map((r, i) => (
                      <tr key={i}>
                        <td className="px-4 py-2 font-mono">{r.row}</td>
                        <td className="px-4 py-2 text-red-700">{r.error}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
