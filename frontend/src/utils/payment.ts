import type { CheckoutResponse } from '../api/payment';

/**
 * Dynamically creates and submits a hidden form to redirect the user to PayHere.
 */
export const submitPayHereForm = (checkout: CheckoutResponse) => {
    const form = document.createElement('form');
    form.method = 'POST';
    form.action = checkout.checkout_url;

    // Add all fields from the backend
    Object.entries(checkout.fields).forEach(([name, value]) => {
        const input = document.createElement('input');
        input.type = 'hidden';
        input.name = name;
        input.value = value;
        form.appendChild(input);
    });

    document.body.appendChild(form);
    form.submit();
    document.body.removeChild(form);
};
