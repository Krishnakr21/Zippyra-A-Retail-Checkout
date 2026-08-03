'use client';

import React, { useState, useEffect } from 'react';

interface MonthlyNpsModalProps {
  sourceApp: 'RETAILER_DASHBOARD' | 'CHAIN_HQ';
}

export default function MonthlyNpsModal({ sourceApp }: MonthlyNpsModalProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [score, setScore] = useState<number | null>(null);
  const [comment, setComment] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  useEffect(() => {
    const currentMonth = new Date().toISOString().slice(0, 7); // e.g. "2026-08"
    const lastSurveyMonth = localStorage.getItem(`nps_last_survey_month_${sourceApp}`);

    if (lastSurveyMonth !== currentMonth) {
      setIsOpen(true);
    }
  }, [sourceApp]);

  const handleDismiss = () => {
    const currentMonth = new Date().toISOString().slice(0, 7);
    localStorage.setItem(`nps_last_survey_month_${sourceApp}`, currentMonth);
    setIsOpen(false);
  };

  const handleSubmit = async () => {
    setIsSubmitting(true);
    try {
      const token = localStorage.getItem('token') || localStorage.getItem('jwt');
      await fetch('/v1/support/feedback', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({
          nps_score: score,
          comment: comment.trim(),
          source_app: sourceApp,
          context: 'monthly_survey',
        }),
      });
    } catch (err) {
      console.warn('Failed to submit NPS feedback', err);
    } finally {
      setIsSubmitting(false);
      setSubmitted(true);
      setTimeout(() => {
        handleDismiss();
      }, 1500);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
      <div className="w-full max-w-lg rounded-2xl bg-slate-900 border border-slate-800 p-6 shadow-2xl text-slate-100">
        {submitted ? (
          <div className="text-center py-8">
            <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-400 mb-4">
              ✓
            </div>
            <h3 className="text-xl font-bold text-white">Thank You for Your Feedback!</h3>
            <p className="text-sm text-slate-400 mt-2">Your response helps us improve Zippyra everyday.</p>
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold uppercase tracking-wider text-indigo-400">Monthly Feedback</span>
              <button
                onClick={handleDismiss}
                className="text-slate-400 hover:text-white transition-colors"
                aria-label="Close"
              >
                ✕
              </button>
            </div>

            <h3 className="mt-3 text-xl font-bold text-white">How likely are you to recommend Zippyra HQ to another enterprise chain?</h3>
            <p className="mt-1 text-sm text-slate-400">Select a rating from 0 (Not likely) to 10 (Extremely likely)</p>

            <div className="mt-6 flex flex-wrap items-center justify-between gap-1.5">
              {[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map((num) => (
                <button
                  key={num}
                  type="button"
                  onClick={() => setScore(num)}
                  className={`flex h-10 w-10 items-center justify-center rounded-lg text-sm font-bold transition-all ${
                    score === num
                      ? 'bg-indigo-600 text-white ring-2 ring-indigo-400 ring-offset-2 ring-offset-slate-900'
                      : 'bg-slate-800 text-slate-300 hover:bg-slate-700 hover:text-white'
                  }`}
                >
                  {num}
                </button>
              ))}
            </div>

            <div className="mt-5">
              <label htmlFor="comment" className="block text-xs font-semibold text-slate-300 mb-1.5">
                What could we do better? (Optional)
              </label>
              <textarea
                id="comment"
                rows={3}
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                placeholder="Share your thoughts, feature requests, or issues..."
                className="w-full rounded-xl bg-slate-800 border border-slate-700 p-3 text-sm text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
            </div>

            <div className="mt-6 flex items-center justify-end gap-3">
              <button
                type="button"
                onClick={handleDismiss}
                className="px-4 py-2 text-sm font-medium text-slate-400 hover:text-white transition-colors"
              >
                Skip for now
              </button>
              <button
                type="button"
                onClick={handleSubmit}
                disabled={isSubmitting}
                className="rounded-xl bg-indigo-600 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-indigo-600/30 hover:bg-indigo-500 transition-all disabled:opacity-50"
              >
                {isSubmitting ? 'Submitting...' : 'Submit Feedback'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
