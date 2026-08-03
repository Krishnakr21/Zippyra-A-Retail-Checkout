'use client';

import React, { useEffect } from 'react';

export function AxeCoreDevInit() {
  useEffect(() => {
    if (process.env.NODE_ENV !== 'production' && typeof window !== 'undefined') {
      // @ts-ignore
      import('@axe-core/react')
        .then((axe) => {
          axe.default(React, 1000);
        })
        .catch(() => {
          // Dev logging fallback
        });
    }
  }, []);

  return null;
}
