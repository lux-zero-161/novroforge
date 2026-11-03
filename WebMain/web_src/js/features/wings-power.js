function updateKillConfirmation({currentTarget: action}) {
  const form = action.closest('[data-wings-power-form]');
  const confirmation = form.querySelector('[data-wings-kill-confirmation]');
  const checkbox = confirmation.querySelector('input[name="confirm_kill"]');
  const needsConfirmation = action.value === 'kill';
  confirmation.hidden = !needsConfirmation;
  checkbox.required = needsConfirmation;
  if (!needsConfirmation) checkbox.checked = false;
}

export function initWingsPower() {
  for (const form of document.querySelectorAll('[data-wings-power-form]')) {
    const action = form.querySelector('[data-wings-power-action]');
    const confirmation = form.querySelector('[data-wings-kill-confirmation]');
    const checkbox = confirmation?.querySelector('input[name="confirm_kill"]');
    if (!action || !confirmation || !checkbox) continue;

    action.addEventListener('change', updateKillConfirmation);
    updateKillConfirmation({currentTarget: action});
  }
}
