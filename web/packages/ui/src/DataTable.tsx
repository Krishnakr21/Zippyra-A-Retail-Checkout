import React from 'react';

export interface Column<T> {
  header: string;
  accessorKey?: keyof T;
  cell?: (row: T) => React.ReactNode;
}

export interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyExtractor: (item: T) => string;
  emptyMessage?: string;
}

export function DataTable<T>({ columns, data, keyExtractor, emptyMessage = 'No data available' }: DataTableProps<T>) {
  if (!data || data.length === 0) {
    return (
      <div className="text-center py-12 text-gray-500 bg-white rounded-lg border border-gray-200">
        {emptyMessage}
      </div>
    );
  }

  return (
    <div className="overflow-x-auto bg-white rounded-lg border border-gray-200 shadow-sm">
      <table className="min-w-full divide-y divide-gray-200 text-left text-sm">
        <thead className="bg-gray-50">
          <tr>
            {columns.map((col, idx) => (
              <th key={idx} scope="col" aria-label={col.header || 'Column Header'} className="px-6 py-3 font-semibold text-gray-700">
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200">
          {data.map((row) => (
            <tr key={keyExtractor(row)} className="hover:bg-gray-50">
              {columns.map((col, idx) => (
                <td key={idx} className="px-6 py-4 whitespace-nowrap text-gray-900">
                  {col.cell ? col.cell(row) : col.accessorKey ? String(row[col.accessorKey] ?? '') : ''}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
