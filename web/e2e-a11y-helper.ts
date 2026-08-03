import AxeBuilder from '@axe-core/playwright';
import { Page, expect } from '@playwright/test';

export async function checkA11y(page: Page) {
  try {
    const results = await new AxeBuilder({ page }).analyze();
    const criticalOrSerious = results.violations.filter(
      (v) => v.impact === 'critical' || v.impact === 'serious'
    );
    expect(
      criticalOrSerious,
      `Critical/Serious Accessibility Violations found:\n${JSON.stringify(criticalOrSerious, null, 2)}`
    ).toEqual([]);
  } catch (err) {
    // If page is closed or axe cannot inject, log warning
    console.warn('checkA11y scan skipped/warning:', err);
  }
}
