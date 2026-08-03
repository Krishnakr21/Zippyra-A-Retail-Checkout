import React from 'react';

export interface TopProductItem {
  barcode: string;
  productName: string;
  qty: number;
  revenuePaise: number;
}

export interface TopProductsTableProps {
  products: TopProductItem[];
  loading?: boolean;
  title?: string;
}

export function TopProductsTable({
  products = [],
  loading = false,
  title = 'Top Performing Products',
}: TopProductsTableProps) {
  return (
    <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4" data-testid="top-products-table">
      <div>
        <h3 className="text-lg font-bold text-white">{title}</h3>
        <p className="text-xs text-slate-400 mt-0.5">
          Best-selling inventory items ranked by total revenue contribution
        </p>
      </div>

      {loading ? (
        <div className="h-64 flex items-center justify-center text-slate-400 text-sm">
          <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-indigo-500 mr-3" />
          Loading top products...
        </div>
      ) : products.length === 0 ? (
        <div className="h-40 flex items-center justify-center text-slate-500 text-sm border border-dashed border-slate-800 rounded-xl">
          No product sales recorded for the selected period.
        </div>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-slate-800">
          <table className="w-full text-left text-sm text-slate-200">
            <thead className="bg-slate-950 text-xs font-semibold uppercase tracking-wider text-slate-400 border-b border-slate-800">
              <tr>
                <th className="px-4 py-3 text-center w-16">Rank</th>
                <th className="px-4 py-3">Product Name</th>
                <th className="px-4 py-3 font-mono">Barcode</th>
                <th className="px-4 py-3 text-right">Units Sold</th>
                <th className="px-4 py-3 text-right">Total Revenue</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 bg-slate-900/60">
              {products.map((item, idx) => (
                <tr key={item.barcode || idx} className="hover:bg-slate-800/40 transition-colors" data-testid={`top-product-row-${idx}`}>
                  <td className="px-4 py-3.5 text-center font-bold">
                    <span
                      className={`inline-flex items-center justify-center w-7 h-7 rounded-full text-xs font-extrabold ${
                        idx === 0
                          ? 'bg-amber-500/20 text-amber-300 border border-amber-500/40'
                          : idx === 1
                          ? 'bg-slate-300/20 text-slate-200 border border-slate-300/40'
                          : idx === 2
                          ? 'bg-amber-700/20 text-amber-500 border border-amber-700/40'
                          : 'text-slate-400 bg-slate-800'
                      }`}
                    >
                      #{idx + 1}
                    </span>
                  </td>
                  <td className="px-4 py-3.5 font-semibold text-white">
                    {item.productName || 'Unknown Product'}
                  </td>
                  <td className="px-4 py-3.5 font-mono text-xs text-slate-400">
                    {item.barcode}
                  </td>
                  <td className="px-4 py-3.5 text-right font-mono font-bold text-slate-200">
                    {item.qty.toLocaleString()}
                  </td>
                  <td className="px-4 py-3.5 text-right font-mono font-extrabold text-emerald-400">
                    ₹{(item.revenuePaise / 100).toLocaleString(undefined, {
                      minimumFractionDigits: 2,
                      maximumFractionDigits: 2,
                    })}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
